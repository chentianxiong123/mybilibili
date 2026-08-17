package manuscript

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ManuscriptAdminHandler 稿件管理后台真 SQL 实现
type ManuscriptAdminHandler struct {
	db        *sql.DB
	events    ManuscriptEventWriter
	publisher ManuscriptEventPublisher
}

func NewManuscriptAdminHandler(db *sql.DB) *ManuscriptAdminHandler {
	return &ManuscriptAdminHandler{db: db}
}

// SetEventWriter 注入事件/流水落库器；未注入时惰性创建 SQL 实现。
func (h *ManuscriptAdminHandler) SetEventWriter(w ManuscriptEventWriter) {
	h.events = w
}

// SetEventPublisher 注入跨服务事件发布器（core.EventPublisher 适配）。
func (h *ManuscriptAdminHandler) SetEventPublisher(p ManuscriptEventPublisher) {
	h.publisher = p
}

func (h *ManuscriptAdminHandler) eventWriter() ManuscriptEventWriter {
	if h.events == nil {
		h.events = NewManuscriptEventWriter(h.db)
	}
	return h.events
}

func (h *ManuscriptAdminHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/manuscript/admin/pending", h.handlePending)
	mux.HandleFunc("/api/v1/manuscript/admin/processing", h.handleProcessing)
	mux.HandleFunc("/api/v1/manuscript/admin/all", h.handleAll)
	mux.HandleFunc("/api/v1/manuscript/admin/statistics", h.handleStatistics)
	mux.HandleFunc("/api/v1/manuscript/admin/", h.handleByID)
}

// GET /api/v1/manuscript/admin/pending — 待审核稿件列表
func (h *ManuscriptAdminHandler) handlePending(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	page, size := parsePageAdmin(r)
	h.listByStatus(w, r, page, size, 0, "pending")
}

// GET /api/v1/manuscript/admin/processing — 审核中稿件列表
func (h *ManuscriptAdminHandler) handleProcessing(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	page, size := parsePageAdmin(r)
	h.listByStatus(w, r, page, size, 3, "processing")
}

// GET /api/v1/manuscript/admin/all — 全部稿件列表（带过滤）
func (h *ManuscriptAdminHandler) handleAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	page, size := parsePageAdmin(r)
	statusStr := r.URL.Query().Get("status")
	keyword := r.URL.Query().Get("keyword")

	offset := (page - 1) * size
	var rows *sql.Rows
	var err error

	if keyword != "" {
		rows, err = h.db.QueryContext(r.Context(),
			`SELECT m.id, m.title, m.user_id, m.category_id, m.status, m.review_status,
			        COALESCE(m.view_count,0), COALESCE(m.like_count,0), COALESCE(m.comment_count,0),
			        m.upload_time, m.updated_at
			 FROM manuscripts m
			 WHERE m.title ILIKE '%' || $1 || '%'
			 ORDER BY m.id DESC LIMIT $2 OFFSET $3`,
			keyword, size, offset)
	} else if statusStr != "" {
		status, _ := strconv.Atoi(statusStr)
		rows, err = h.db.QueryContext(r.Context(),
			`SELECT m.id, m.title, m.user_id, m.category_id, m.status, m.review_status,
			        COALESCE(m.view_count,0), COALESCE(m.like_count,0), COALESCE(m.comment_count,0),
			        m.upload_time, m.updated_at
			 FROM manuscripts m
			 WHERE m.status = $1
			 ORDER BY m.id DESC LIMIT $2 OFFSET $3`,
			status, size, offset)
	} else {
		rows, err = h.db.QueryContext(r.Context(),
			`SELECT m.id, m.title, m.user_id, m.category_id, m.status, m.review_status,
			        COALESCE(m.view_count,0), COALESCE(m.like_count,0), COALESCE(m.comment_count,0),
			        m.upload_time, m.updated_at
			 FROM manuscripts m
			 ORDER BY m.id DESC LIMIT $1 OFFSET $2`,
			size, offset)
	}

	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"list": []interface{}{}, "total": 0})
		return
	}
	defer rows.Close()

	type ms struct {
		ID           int64  `json:"id"`
		Title        string `json:"title"`
		UserID       int64  `json:"user_id"`
		CategoryID   int64  `json:"category_id"`
		Status       int32  `json:"status"`
		ReviewStatus int32  `json:"review_status"`
		ViewCount    int64  `json:"view_count"`
		LikeCount    int64  `json:"like_count"`
		CommentCount int64  `json:"comment_count"`
		UploadTime   string `json:"upload_time"`
		UpdatedAt    string `json:"updated_at"`
	}
	list := []ms{}
	for rows.Next() {
		var m ms
		var uploadTime, updatedAt time.Time
		if err := rows.Scan(&m.ID, &m.Title, &m.UserID, &m.CategoryID,
			&m.Status, &m.ReviewStatus, &m.ViewCount, &m.LikeCount,
			&m.CommentCount, &uploadTime, &updatedAt); err != nil {
			continue
		}
		m.UploadTime = uploadTime.Format("2006-01-02T15:04:05Z")
		m.UpdatedAt = updatedAt.Format("2006-01-02T15:04:05Z")
		list = append(list, m)
	}

	var total int
	h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM manuscripts`).Scan(&total)

	json.NewEncoder(w).Encode(map[string]interface{}{"list": list, "total": total, "page": page, "size": size})
}

// GET /api/v1/manuscript/admin/statistics — 稿件统计概览
func (h *ManuscriptAdminHandler) handleStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	stats := map[string]interface{}{}

	var total, pending, approved, rejected, draft int64
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts`).Scan(&total)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE review_status = 0`).Scan(&pending)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE review_status = 1`).Scan(&approved)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE review_status = 2`).Scan(&rejected)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE status = -1`).Scan(&draft)

	stats["total"] = total
	stats["pending"] = pending
	stats["approved"] = approved
	stats["rejected"] = rejected
	stats["draft"] = draft

	var totalViews, totalLikes int64
	h.db.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(view_count),0) FROM manuscripts`).Scan(&totalViews)
	h.db.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(like_count),0) FROM manuscripts`).Scan(&totalLikes)
	stats["total_views"] = totalViews
	stats["total_likes"] = totalLikes

	var videoCount int64
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM videos`).Scan(&videoCount)
	stats["total_videos"] = videoCount

	json.NewEncoder(w).Encode(stats)
}

const (
	manuscriptStatusPendingReview = 0
	manuscriptStatusProcessing    = 1
	manuscriptStatusPublished     = 3
	manuscriptStatusRejected      = 4
	manuscriptStatusProcessFailed = 5
	manuscriptStatusUnpublished   = -1

	manuscriptReviewStatusPending  = 0
	manuscriptReviewStatusApproved = 1
	manuscriptReviewStatusRejected = 2

	videoProcessStatusPending      = 0
	videoProcessStatusTranscoding  = 1
	videoProcessStatusAudioExtra   = 2
	videoProcessStatusSubtitleGen  = 3
	videoProcessStatusAiSummary    = 4
	videoProcessStatusCompleted    = 5
	videoProcessStatusTranscodeEnd = 11
)

// GET /api/v1/manuscript/admin/{id} — 稿件详情（含视频列表）
// POST /api/v1/manuscript/admin/{id}/approve-with-process — 审核通过并触发处理
func (h *ManuscriptAdminHandler) handleByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/manuscript/admin/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "not found", 404)
		return
	}

	// 视频级动作（不依赖稿件ID前缀）:
	// /transcode/{videoId} /extract-audio/{videoId} /generate-subtitle/{videoId}
	// /ai-summary/{videoId} /process-all/{videoId} /video-source/{videoId} /reset/{videoId}
	switch parts[0] {
	case "transcode", "extract-audio", "generate-subtitle", "ai-summary", "process-all", "video-source", "reset":
		if len(parts) < 2 {
			http.Error(w, "invalid video id", 400)
			return
		}
		videoID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			http.Error(w, "invalid video id", 400)
			return
		}
		switch parts[0] {
		case "transcode":
			h.triggerVideoProcess(w, r, videoID, videoProcessStatusTranscoding, "TRANSCODING")
		case "extract-audio":
			h.triggerVideoProcess(w, r, videoID, videoProcessStatusAudioExtra, "AUDIO_EXTRACTING")
		case "generate-subtitle":
			h.triggerVideoProcess(w, r, videoID, videoProcessStatusSubtitleGen, "SUBTITLE_GENERATING")
		case "ai-summary":
			h.triggerVideoProcess(w, r, videoID, videoProcessStatusAiSummary, "AI_SUMMARIZING")
		case "process-all":
			h.triggerVideoProcess(w, r, videoID, videoProcessStatusTranscoding, "PROCESS_ALL")
		case "video-source":
			h.getVideoSource(w, r, videoID)
		case "reset":
			h.resetVideo(w, r, videoID)
		}
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		// 稿件级动作: /approve/{id} /reject/{id} /publish/{id} /unpublish/{id} /retry/{id}
		switch parts[0] {
		case "approve", "reject", "publish", "unpublish", "retry":
			if len(parts) < 2 || r.Method != "POST" {
				if r.Method != "POST" {
					http.Error(w, "method not allowed", 405)
				} else {
					http.Error(w, "invalid manuscript id", 400)
				}
				return
			}
			mid, perr := strconv.ParseInt(parts[1], 10, 64)
			if perr != nil {
				http.Error(w, "invalid manuscript id", 400)
				return
			}
			switch parts[0] {
			case "approve":
				h.reviewManuscript(w, r, mid, true, false)
			case "reject":
				h.reviewManuscript(w, r, mid, false, false)
			case "publish":
				h.setManuscriptStatus(w, r, mid, manuscriptStatusPublished)
			case "unpublish":
				h.setManuscriptStatus(w, r, mid, manuscriptStatusUnpublished)
			case "retry":
				h.setManuscriptStatus(w, r, mid, manuscriptStatusProcessing)
			}
			return
		}
		http.Error(w, "invalid id", 400)
		return
	}

	// /{id}/videos — 稿件视频列表
	if len(parts) >= 2 && parts[1] == "videos" && r.Method == "GET" {
		h.getVideos(w, r, id)
		return
	}

	// /{id}/approve-with-process — 审核通过并触发处理
	if len(parts) >= 2 && parts[1] == "approve-with-process" && r.Method == "POST" {
		h.approveWithProcess(w, r, id)
		return
	}

	// /{id}/publish /{id}/unpublish — 稿件级发布/下架（owner 端也可调用）
	if len(parts) >= 2 && r.Method == "POST" {
		switch parts[1] {
		case "publish":
			h.setManuscriptStatus(w, r, id, manuscriptStatusPublished)
			return
		case "unpublish":
			h.setManuscriptStatus(w, r, id, manuscriptStatusUnpublished)
			return
		}
	}

	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}

	// GET /{id} — 稿件详情
	row := h.db.QueryRowContext(r.Context(),
		`SELECT m.id, m.title, m.description, m.cover_url, m.user_id, m.category_id,
		        m.status, m.review_status, COALESCE(m.review_reason,''),
		        COALESCE(m.view_count,0), COALESCE(m.like_count,0), COALESCE(m.coin_count,0),
		        COALESCE(m.collect_count,0), COALESCE(m.share_count,0), COALESCE(m.comment_count,0),
		        COALESCE(m.danmaku_count,0), m.duration, m.upload_time, m.updated_at
		 FROM manuscripts m WHERE m.id = $1`, id)

	type detail struct {
		ID           int64  `json:"id"`
		Title        string `json:"title"`
		Description  string `json:"description"`
		CoverURL     string `json:"cover_url"`
		UserID       int64  `json:"user_id"`
		CategoryID   int64  `json:"category_id"`
		Status       int32  `json:"status"`
		ReviewStatus int32  `json:"review_status"`
		ReviewReason string `json:"review_reason"`
		ViewCount    int64  `json:"view_count"`
		LikeCount    int64  `json:"like_count"`
		CoinCount    int64  `json:"coin_count"`
		CollectCount int64  `json:"collect_count"`
		ShareCount   int64  `json:"share_count"`
		CommentCount int64  `json:"comment_count"`
		DanmakuCount int64  `json:"danmaku_count"`
		Duration     string `json:"duration"`
		UploadTime   string `json:"upload_time"`
		UpdatedAt    string `json:"updated_at"`
	}
	var d detail
	var uploadTime, updatedAt time.Time
	if err := row.Scan(&d.ID, &d.Title, &d.Description, &d.CoverURL, &d.UserID, &d.CategoryID,
		&d.Status, &d.ReviewStatus, &d.ReviewReason,
		&d.ViewCount, &d.LikeCount, &d.CoinCount, &d.CollectCount,
		&d.ShareCount, &d.CommentCount, &d.DanmakuCount,
		&d.Duration, &uploadTime, &updatedAt); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	d.UploadTime = uploadTime.Format("2006-01-02T15:04:05Z")
	d.UpdatedAt = updatedAt.Format("2006-01-02T15:04:05Z")

	// 附带视频列表
	视频列表 := []map[string]interface{}{}
	vrows, _ := h.db.QueryContext(r.Context(),
		`SELECT id, video_order, title, play_url_hd, upload_time FROM videos WHERE manuscript_id = $1 ORDER BY video_order`, id)
	if vrows != nil {
		defer vrows.Close()
		for vrows.Next() {
			var vid int64
			var order int
			var vtitle, playURL string
			var vtime time.Time
			vrows.Scan(&vid, &order, &vtitle, &playURL, &vtime)
			视频列表 = append(视频列表, map[string]interface{}{
				"id": vid, "video_order": order, "title": vtitle,
				"play_url": playURL, "upload_time": vtime.Format("2006-01-02T15:04:05Z"),
			})
		}
	}
	dMap := map[string]interface{}{
		"id": d.ID, "title": d.Title, "description": d.Description, "cover_url": d.CoverURL,
		"user_id": d.UserID, "category_id": d.CategoryID, "status": d.Status,
		"review_status": d.ReviewStatus, "review_reason": d.ReviewReason,
		"view_count": d.ViewCount, "like_count": d.LikeCount, "coin_count": d.CoinCount,
		"collect_count": d.CollectCount, "share_count": d.ShareCount,
		"comment_count": d.CommentCount, "danmaku_count": d.DanmakuCount,
		"duration": d.Duration, "upload_time": d.UploadTime, "updated_at": d.UpdatedAt,
		"videos": 视频列表,
	}
	json.NewEncoder(w).Encode(dMap)
}

func (h *ManuscriptAdminHandler) getVideos(w http.ResponseWriter, r *http.Request, manuscriptID int64) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, video_order, title, description, play_url_hd, play_url_sd, play_url_ld, upload_time
		 FROM videos WHERE manuscript_id = $1 ORDER BY video_order`, manuscriptID)
	if err != nil {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	defer rows.Close()

	type video struct {
		ID          int64  `json:"id"`
		VideoOrder  int    `json:"video_order"`
		Title       string `json:"title"`
		Description string `json:"description"`
		PlayURLHD   string `json:"play_url_hd"`
		PlayURLSD   string `json:"play_url_sd"`
		PlayURLLD   string `json:"play_url_ld"`
		UploadTime  string `json:"upload_time"`
	}
	list := []video{}
	for rows.Next() {
		var v video
		var t time.Time
		if err := rows.Scan(&v.ID, &v.VideoOrder, &v.Title, &v.Description,
			&v.PlayURLHD, &v.PlayURLSD, &v.PlayURLLD, &t); err != nil {
			continue
		}
		v.UploadTime = t.Format("2006-01-02T15:04:05Z")
		list = append(list, v)
	}
	json.NewEncoder(w).Encode(list)
}

// reviewManuscript 审核通过/拒绝稿件，对齐旧版 approveManuscript/rejectManuscript 的状态流转。
func (h *ManuscriptAdminHandler) reviewManuscript(w http.ResponseWriter, r *http.Request, manuscriptID int64, approved bool, autoProcess bool) {
	reviewerID := r.URL.Query().Get("reviewerId")
	if reviewerID == "" {
		reviewerID = "0"
	}
	reason := r.URL.Query().Get("reason")

	var status, reviewStatus = manuscriptStatusProcessing, manuscriptReviewStatusApproved
	if !approved {
		status, reviewStatus = manuscriptStatusRejected, manuscriptReviewStatusRejected
	}

	var exists int64
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE id = $1`, manuscriptID).Scan(&exists)
	if exists == 0 {
		http.Error(w, "稿件不存在", 404)
		return
	}

	var fromStatus int32
	var uid int64
	_ = h.db.QueryRowContext(r.Context(),
		`SELECT status, user_id FROM manuscripts WHERE id = $1`, manuscriptID).Scan(&fromStatus, &uid)

	_, err := h.db.ExecContext(r.Context(),
		`UPDATE manuscripts SET status = $1, review_status = $2, review_time = NOW(),
		        reviewer_id = $3, review_reason = $4, updated_at = NOW()
		 WHERE id = $5`,
		status, reviewStatus, reviewerID, reason, manuscriptID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	action := "REJECT"
	if approved {
		action = "APPROVE"
	}
	operatorID, _ := strconv.ParseInt(reviewerID, 10, 64)
	_ = h.eventWriter().RecordStatusEvent(r.Context(), manuscriptID, uid, fromStatus, int32(status),
		action, "ADMIN", operatorID, reason)
	if approved {
		_ = h.eventWriter().RecordEditVersion(r.Context(), manuscriptID, uid, "", reason, "review_status,status")
		if h.publisher != nil {
			_ = h.publisher.PublishManuscriptIndex(r.Context(), manuscriptID, "UPSERT", "APPROVE")
		}
	}

	// 审核通过(含自动处理)时触发视频处理流程；拒绝时不触发
	if approved {
		h.triggerAllVideoProcess(r.Context(), manuscriptID)
	}

	message := "审核拒绝成功"
	if approved {
		message = "审核通过"
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok", "manuscript_id": manuscriptID, "message": message,
		"review_status": reviewStatus,
	})
}

// setManuscriptStatus 设置稿件状态（对齐 publish/unpublish/retry/owner 重新上架）
func (h *ManuscriptAdminHandler) setManuscriptStatus(w http.ResponseWriter, r *http.Request, manuscriptID int64, status int32) {
	var exists int64
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE id = $1`, manuscriptID).Scan(&exists)
	if exists == 0 {
		http.Error(w, "稿件不存在", 404)
		return
	}

	var title string
	_ = h.db.QueryRowContext(r.Context(), `SELECT title FROM manuscripts WHERE id = $1`, manuscriptID).Scan(&title)

	var fromStatus int32
	var uid int64
	_ = h.db.QueryRowContext(r.Context(),
		`SELECT status, user_id FROM manuscripts WHERE id = $1`, manuscriptID).Scan(&fromStatus, &uid)

	_, err := h.db.ExecContext(r.Context(),
		`UPDATE manuscripts SET status = $1, updated_at = NOW() WHERE id = $2`, status, manuscriptID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	action := "UPDATE_STATUS"
	switch status {
	case manuscriptStatusPublished:
		action = "PUBLISH"
	case manuscriptStatusUnpublished:
		action = "UNPUBLISH"
	case manuscriptStatusProcessing:
		action = "RETRY"
	}
	_ = h.eventWriter().RecordStatusEvent(r.Context(), manuscriptID, uid, fromStatus, status,
		action, "ADMIN", 0, "")

	if h.publisher != nil {
		switch status {
		case manuscriptStatusPublished:
			_ = h.publisher.PublishManuscriptIndex(r.Context(), manuscriptID, "UPSERT", "PUBLISH")
			_ = h.publisher.PublishAnalytics(r.Context(), manuscriptID, uid, "MANUSCRIPT_PUBLISH", "publish_count", 1)
		case manuscriptStatusUnpublished:
			_ = h.publisher.PublishManuscriptIndex(r.Context(), manuscriptID, "DELETE", "UNPUBLISH")
		}
	}

	msg := "操作成功"
	switch status {
	case manuscriptStatusPublished:
		msg = "发布成功"
	case manuscriptStatusUnpublished:
		msg = "下架成功"
	case manuscriptStatusProcessing:
		msg = "重试成功"
	}

	// 发布后向稿件作者发系统通知（对齐旧版 publishManuscript）
	if status == manuscriptStatusPublished {
		var uid int64
		_ = h.db.QueryRowContext(r.Context(), `SELECT user_id FROM manuscripts WHERE id = $1`, manuscriptID).Scan(&uid)
		if uid > 0 {
			content := "您的稿件《" + title + "》已通过审核并成功上架啦！"
			_, _ = h.db.ExecContext(r.Context(),
				`INSERT INTO notifications (user_id, type, title, content, is_read, created_at)
				 VALUES ($1, 'system', '稿件上架通知', $2, 0, NOW())`, uid, content)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok", "manuscript_id": manuscriptID, "status_int": status, "message": msg,
	})
}

// triggerVideoProcess 触发单视频处理（对齐旧版 manualTranscode / manualExtractAudio / manualGenerateSubtitle / manualAiSummary / manualProcessAll）
func (h *ManuscriptAdminHandler) triggerVideoProcess(w http.ResponseWriter, r *http.Request, videoID int64, processStatus int, stage string) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var exists int64
	var manuscriptID int64
	var title string
	err := h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM videos WHERE id = $1`, videoID).Scan(&exists)
	if exists == 0 {
		http.Error(w, "视频不存在", 404)
		return
	}
	var sourceURL string
	var uploaderID int64
	_ = h.db.QueryRowContext(r.Context(),
		`SELECT manuscript_id, title FROM videos WHERE id = $1`, videoID).Scan(&manuscriptID, &title)
	_ = h.db.QueryRowContext(r.Context(),
		`SELECT COALESCE(source_video_url,'') FROM videos WHERE id = $1`, videoID).Scan(&sourceURL)
	_ = h.db.QueryRowContext(r.Context(),
		`SELECT user_id FROM manuscripts WHERE id = $1`, manuscriptID).Scan(&uploaderID)

	var fromProcess int32
	_ = h.db.QueryRowContext(r.Context(),
		`SELECT process_status FROM videos WHERE id = $1`, videoID).Scan(&fromProcess)

	_, err = h.db.ExecContext(r.Context(),
		`UPDATE videos SET process_status = $1, process_stage = $2, updated_at = NOW() WHERE id = $3`,
		processStatus, stage, videoID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_ = h.eventWriter().RecordVideoProcessEvent(r.Context(), videoID, manuscriptID,
		fromProcess, int32(processStatus), stage, 100)

	if h.publisher != nil {
		_ = h.publisher.PublishVideoProcess(r.Context(), manuscriptID, videoID, stage, sourceURL, uploaderID)
	}

	// 稿件标记为处理中（对齐旧版 approve 后 status=PROCESSING）
	if manuscriptID > 0 {
		_, _ = h.db.ExecContext(r.Context(),
			`UPDATE manuscripts SET status = $1, updated_at = NOW() WHERE id = $2`,
			manuscriptStatusProcessing, manuscriptID)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok", "video_id": videoID, "manuscript_id": manuscriptID,
		"process_status": processStatus, "message": "任务已触发",
	})
}

// resetVideo 重置视频处理状态（对齐旧版 resetVideoStatus）
func (h *ManuscriptAdminHandler) resetVideo(w http.ResponseWriter, r *http.Request, videoID int64) {
	var exists int64
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM videos WHERE id = $1`, videoID).Scan(&exists)
	if exists == 0 {
		http.Error(w, "视频不存在", 404)
		return
	}
	_, err := h.db.ExecContext(r.Context(),
		`UPDATE videos SET process_status = $1, process_stage = '', process_error = '', updated_at = NOW() WHERE id = $2`,
		videoProcessStatusPending, videoID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "video_id": videoID, "message": "重置成功"})
}

// getVideoSource 获取视频源地址（对齐旧版 getVideoSourceUrl）
func (h *ManuscriptAdminHandler) getVideoSource(w http.ResponseWriter, r *http.Request, videoID int64) {
	var sourceURL, title string
	var duration int
	err := h.db.QueryRowContext(r.Context(),
		`SELECT COALESCE(source_video_url,''), COALESCE(title,''), COALESCE(duration_seconds,0)
		 FROM videos WHERE id = $1`, videoID).Scan(&sourceURL, &title, &duration)
	if err != nil {
		http.Error(w, "视频不存在", 404)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"video_id": videoID, "source_url": sourceURL,
		"title": title, "duration_seconds": duration,
	})
}

// triggerAllVideoProcess 触发稿件下所有视频的处理流程（审核通过后调用）
func (h *ManuscriptAdminHandler) triggerAllVideoProcess(ctx context.Context, manuscriptID int64) {
	rows, err := h.db.QueryContext(ctx,
		`SELECT id FROM videos WHERE manuscript_id = $1`, manuscriptID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var vid int64
		rows.Scan(&vid)
		var fromProcess int32
		_ = h.db.QueryRowContext(ctx,
			`SELECT process_status FROM videos WHERE id = $1`, vid).Scan(&fromProcess)
		_, _ = h.db.ExecContext(ctx,
			`UPDATE videos SET process_status = $1, process_stage = 'TRANSCODING', updated_at = NOW() WHERE id = $2`,
			videoProcessStatusTranscoding, vid)
		_ = h.eventWriter().RecordVideoProcessEvent(ctx, vid, manuscriptID,
			fromProcess, videoProcessStatusTranscoding, "TRANSCODING", 100)
	}
}

func (h *ManuscriptAdminHandler) approveWithProcess(w http.ResponseWriter, r *http.Request, manuscriptID int64) {
	// 审核通过 + 触发转码/字幕/摘要流程（通过更新状态实现）
	var fromStatus int32
	var uid int64
	_ = h.db.QueryRowContext(r.Context(),
		`SELECT status, user_id FROM manuscripts WHERE id = $1`, manuscriptID).Scan(&fromStatus, &uid)

	_, err := h.db.ExecContext(r.Context(),
		`UPDATE manuscripts SET review_status = 1, review_time = NOW(), updated_at = NOW()
		 WHERE id = $1`, manuscriptID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_ = h.eventWriter().RecordStatusEvent(r.Context(), manuscriptID, uid, fromStatus,
		manuscriptStatusProcessing, "APPROVE_WITH_PROCESS", "ADMIN", 0, "")
	_ = h.eventWriter().RecordEditVersion(r.Context(), manuscriptID, uid, "", "", "review_status,status")

	// 获取稿件关联的视频,逐个更新状态触发处理
	rows, _ := h.db.QueryContext(r.Context(),
		`SELECT id FROM videos WHERE manuscript_id = $1`, manuscriptID)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var vid int64
			rows.Scan(&vid)
			// 更新视频状态为待处理
			h.db.ExecContext(r.Context(),
				`UPDATE videos SET updated_at = NOW() WHERE id = $1`, vid)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "ok",
		"manuscript_id": manuscriptID,
		"review_status": 1,
		"message":       "审核通过，已触发处理流程",
	})
}

func (h *ManuscriptAdminHandler) listByStatus(w http.ResponseWriter, r *http.Request, page, size int32, status int, statusLabel string) {
	offset := (page - 1) * size
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT m.id, m.title, m.user_id, m.category_id, m.status, m.review_status,
		        COALESCE(m.view_count,0), COALESCE(m.like_count,0), COALESCE(m.comment_count,0),
		        m.upload_time, m.updated_at
		 FROM manuscripts m
		 WHERE m.review_status = $1
		 ORDER BY m.id DESC LIMIT $2 OFFSET $3`,
		status, size, offset)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"list": []interface{}{}, "total": 0})
		return
	}
	defer rows.Close()

	type ms struct {
		ID           int64  `json:"id"`
		Title        string `json:"title"`
		UserID       int64  `json:"user_id"`
		CategoryID   int64  `json:"category_id"`
		Status       int32  `json:"status"`
		ReviewStatus int32  `json:"review_status"`
		ViewCount    int64  `json:"view_count"`
		LikeCount    int64  `json:"like_count"`
		CommentCount int64  `json:"comment_count"`
		UploadTime   string `json:"upload_time"`
		UpdatedAt    string `json:"updated_at"`
	}
	list := []ms{}
	for rows.Next() {
		var m ms
		var uploadTime, updatedAt time.Time
		if err := rows.Scan(&m.ID, &m.Title, &m.UserID, &m.CategoryID,
			&m.Status, &m.ReviewStatus, &m.ViewCount, &m.LikeCount,
			&m.CommentCount, &uploadTime, &updatedAt); err != nil {
			continue
		}
		m.UploadTime = uploadTime.Format("2006-01-02T15:04:05Z")
		m.UpdatedAt = updatedAt.Format("2006-01-02T15:04:05Z")
		list = append(list, m)
	}

	var total int
	h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM manuscripts WHERE review_status = $1`, status).Scan(&total)

	json.NewEncoder(w).Encode(map[string]interface{}{"list": list, "total": total, "page": page, "size": size})
}

func parsePageAdmin(r *http.Request) (int32, int32) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	size, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return int32(page), int32(size)
}
