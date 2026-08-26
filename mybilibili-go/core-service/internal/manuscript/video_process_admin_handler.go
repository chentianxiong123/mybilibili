package manuscript

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"mybilibili/pkg/httputil"
)

// VideoProcessAdminHandler 转码流水线看板（还原旧版 admin-web /video-process 页面）。
// current/queue/statistics 直查 videos 表；stream 用 SSE 实时推送进度事件。
// ponytail: stream 端点未做鉴权 —— EventSource 无法带 Authorization 头，
// 旧版网关鉴权在新架构无对应物；只读监控可接受，后续可改 cookie 校验。
type VideoProcessAdminHandler struct {
	db  *sql.DB
	hub *sseHub
}

func NewVideoProcessAdminHandler(db *sql.DB) *VideoProcessAdminHandler {
	return &VideoProcessAdminHandler{db: db, hub: newSSEHub()}
}

// Hub 供 main.go 的 MQ 订阅协程投递进度事件。
func (h *VideoProcessAdminHandler) Hub() *sseHub { return h.hub }

type sseHub struct {
	mu      sync.Mutex
	clients map[chan ProgressEvt]struct{}
}

type ProgressEvt struct {
	VideoID      int64  `json:"video_id"`
	ManuscriptID int64  `json:"manuscript_id"`
	Title        string `json:"title"`
	Stage        string `json:"stage"`
	StageText    string `json:"stage_text"`
	Progress     int32  `json:"progress"`
	Status       int32  `json:"status"`
	Done         bool   `json:"done"`
	Error        string `json:"error"`
}

func newSSEHub() *sseHub {
	return &sseHub{clients: make(map[chan ProgressEvt]struct{})}
}

func (s *sseHub) add(ch chan ProgressEvt) {
	s.mu.Lock()
	s.clients[ch] = struct{}{}
	s.mu.Unlock()
}

func (s *sseHub) remove(ch chan ProgressEvt) {
	s.mu.Lock()
	delete(s.clients, ch)
	s.mu.Unlock()
}

func (s *sseHub) Broadcast(evt ProgressEvt) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.clients {
		select {
		case ch <- evt:
		default: // 慢客户端丢帧，不阻塞流水线
		}
	}
}

func (h *VideoProcessAdminHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/video/process/admin/current", h.handleCurrent)
	mux.HandleFunc("/api/v1/video/process/admin/queue", h.handleQueue)
	mux.HandleFunc("/api/v1/video/process/admin/statistics", h.handleStatistics)
	mux.HandleFunc("/api/v1/video/process/admin/stream", h.handleStream)
}

// handleCurrent GET 当前处理中的任务（process_status 1-4，取最近更新的一个）。
func (h *VideoProcessAdminHandler) handleCurrent(w http.ResponseWriter, r *http.Request) {
	var videoID, manuscriptID int64
	var title, stage, stageText string
	var status, progress int32
	err := h.db.QueryRowContext(r.Context(),
		`SELECT v.id, v.manuscript_id, COALESCE(NULLIF(v.title,''), m.title),
		        v.process_stage,
		        CASE v.process_status
		            WHEN 1 THEN '视频转码中' WHEN 2 THEN '音频提取中'
		            WHEN 3 THEN '字幕生成中' WHEN 4 THEN 'AI总结中'
		            ELSE '' END,
		        v.process_status, v.process_progress
		 FROM videos v JOIN manuscripts m ON m.id = v.manuscript_id
		 WHERE v.process_status IN (1,2,3,4)
		 ORDER BY v.updated_at DESC LIMIT 1`).Scan(
		&videoID, &manuscriptID, &title, &stage, &stageText, &status, &progress)
	if err == sql.ErrNoRows {
		httputil.WriteOK(w, map[string]interface{}{"processing": false})
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	httputil.WriteOK(w, map[string]interface{}{
		"processing": true, "videoId": videoID, "manuscriptId": manuscriptID,
		"videoTitle": title, "stage": stage, "stageText": stageText,
		"status": status, "progress": progress,
	})
}

// handleQueue GET 各阶段排队数量。
func (h *VideoProcessAdminHandler) handleQueue(w http.ResponseWriter, r *http.Request) {
	counts := map[string]int{
		"waitingTranscode": 0, "waitingAudio": 0, "waitingSubtitle": 0, "waitingAi": 0,
	}
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT process_status, COUNT(*) FROM videos WHERE process_status IN (0,11,21,31) GROUP BY process_status`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	total := 0
	for rows.Next() {
		var st, cnt int
		_ = rows.Scan(&st, &cnt)
		switch st {
		case 0:
			counts["waitingTranscode"] = cnt
		case 11:
			counts["waitingAudio"] = cnt
		case 21:
			counts["waitingSubtitle"] = cnt
		case 31:
			counts["waitingAi"] = cnt
		}
		total += cnt
	}
	counts["queueSize"] = total
	httputil.WriteOK(w, counts)
}

// handleStatistics GET 各阶段完成数（对齐旧版看板统计卡片）。
func (h *VideoProcessAdminHandler) handleStatistics(w http.ResponseWriter, r *http.Request) {
	stats := map[string]int{
		"pending": 0, "transcoding": 0, "audioExtracting": 0,
		"subtitleGenerating": 0, "aiSummarizing": 0, "completed": 0,
	}
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT process_status, COUNT(*) FROM videos WHERE process_status IN (0,5,11,21,31,41) GROUP BY process_status`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var st, cnt int
		_ = rows.Scan(&st, &cnt)
		switch st {
		case 0:
			stats["pending"] = cnt
		case 11:
			stats["transcoding"] = cnt
		case 21:
			stats["audioExtracting"] = cnt
		case 31:
			stats["subtitleGenerating"] = cnt
		case 41:
			stats["aiSummarizing"] = cnt
		case 5:
			stats["completed"] = cnt
		}
	}
	httputil.WriteOK(w, stats)
}

// handleStream GET SSE 推流。事件名对齐旧版前端：snapshot/progress/complete/error。
func (h *VideoProcessAdminHandler) handleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", 500)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan ProgressEvt, 16)
	h.hub.add(ch)
	defer h.hub.remove(ch)

	// 连接建立即推一次快照（旧版前端 onopen 后靠 snapshot 初始化）
	writeEvent := func(name string, payload []byte) bool {
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", name, payload); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}
	snapshot := h.buildSnapshot(r.Context())
	if !writeEvent("snapshot", snapshot) {
		return
	}

	ctx := r.Context()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-ch:
			payload, _ := json.Marshal(map[string]interface{}{
				"type": eventName(evt), "videoId": evt.VideoID, "manuscriptId": evt.ManuscriptID,
				"title": evt.Title, "stage": evt.Stage, "stageText": evt.StageText,
				"progress": evt.Progress, "status": evt.Status, "error": evt.Error,
				"statusText": evt.StageText,
			})
			if !writeEvent(eventName(evt), payload) {
				return
			}
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (h *VideoProcessAdminHandler) buildSnapshot(ctx context.Context) []byte {
	current := map[string]interface{}{"processing": false}
	var videoID, manuscriptID int64
	var title, stageText string
	var status, progress int32
	err := h.db.QueryRowContext(ctx,
		`SELECT v.id, v.manuscript_id, COALESCE(NULLIF(v.title,''), m.title),
		        CASE v.process_status
		            WHEN 1 THEN '视频转码中' WHEN 2 THEN '音频提取中'
		            WHEN 3 THEN '字幕生成中' WHEN 4 THEN 'AI总结中'
		            ELSE '' END,
		        v.process_status, v.process_progress
		 FROM videos v JOIN manuscripts m ON m.id = v.manuscript_id
		 WHERE v.process_status IN (1,2,3,4)
		 ORDER BY v.updated_at DESC LIMIT 1`).
		Scan(&videoID, &manuscriptID, &title, &stageText, &status, &progress)
	if err == nil {
		current = map[string]interface{}{
			"processing": true, "videoId": videoID, "manuscriptId": manuscriptID,
			"videoTitle": title, "stageText": stageText, "status": status, "progress": progress,
		}
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"current":    current,
		"statistics": h.collectStats(ctx),
	})
	return payload
}

func (h *VideoProcessAdminHandler) collectStats(ctx context.Context) map[string]int {
	stats := map[string]int{
		"pending": 0, "transcoding": 0, "audioExtracting": 0,
		"subtitleGenerating": 0, "aiSummarizing": 0, "completed": 0,
	}
	rows, err := h.db.QueryContext(ctx,
		`SELECT process_status, COUNT(*) FROM videos WHERE process_status IN (0,5,11,21,31,41) GROUP BY process_status`)
	if err != nil {
		return stats
	}
	defer rows.Close()
	for rows.Next() {
		var st, cnt int
		_ = rows.Scan(&st, &cnt)
		switch st {
		case 0:
			stats["pending"] = cnt
		case 11:
			stats["transcoding"] = cnt
		case 21:
			stats["audioExtracting"] = cnt
		case 31:
			stats["subtitleGenerating"] = cnt
		case 41:
			stats["aiSummarizing"] = cnt
		case 5:
			stats["completed"] = cnt
		}
	}
	return stats
}

func eventName(evt ProgressEvt) string {
	switch {
	case evt.Error != "":
		return "error"
	case evt.Done:
		return "complete"
	default:
		return "progress"
	}
}

var _ = json.Marshal
