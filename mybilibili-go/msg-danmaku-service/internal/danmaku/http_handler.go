package danmaku

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mybilibili/pkg/httputil"
)

type HTTPHandler struct {
	svc        *DanmakuService
	broadcaster *DanmakuBroadcaster
}

func NewHTTPHandler(svc *DanmakuService, broadcaster *DanmakuBroadcaster) *HTTPHandler {
	return &HTTPHandler{svc: svc, broadcaster: broadcaster}
}

func (h *HTTPHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/danmaku/send", h.handleSend)
	mux.HandleFunc("/api/v1/danmaku/video/", h.handleListByVideo)
	mux.HandleFunc("/api/v1/danmaku/batch-count", h.handleBatchCount)
	mux.HandleFunc("/api/v1/danmaku/trend", h.handleTrend)
	mux.HandleFunc("/api/v1/danmaku/", h.handleByPath)
	mux.HandleFunc("/api/v1/creator/danmaku/list", h.handleCreatorList)
	mux.HandleFunc("/api/v1/creator/danmaku/", h.handleCreatorByPath)
	mux.HandleFunc("/sse/danmaku", h.handleSSE)
}

func (h *HTTPHandler) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var req struct {
		VideoID      int64   `json:"video_id"`
		ManuscriptID int64   `json:"manuscript_id"`
		Content      string  `json:"content"`
		Time         float64 `json:"time"`
		Color        string  `json:"color"`
		Mode         int32   `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid body", "data": nil})
		return
	}
	if req.VideoID <= 0 || strings.TrimSpace(req.Content) == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "video_id and content required", "data": nil})
		return
	}
	event, err := h.svc.Send(r.Context(), req.VideoID, req.ManuscriptID, userID, req.Content, req.Time, req.Color, req.Mode)
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "发送弹幕失败", "data": nil})
		return
	}
	h.broadcaster.Broadcast(req.VideoID, event)
	httputil.WriteOK(w, event)
}

func (h *HTTPHandler) handleListByVideo(w http.ResponseWriter, r *http.Request) {
	videoIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/danmaku/video/")
	videoID, _ := strconv.ParseInt(videoIDStr, 10, 64)
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	if startTimeStr != "" && endTimeStr != "" {
		startTime, _ := strconv.ParseFloat(startTimeStr, 64)
		endTime, _ := strconv.ParseFloat(endTimeStr, 64)
		events, _ := h.svc.ListByTimeRange(r.Context(), videoID, startTime, endTime)
		httputil.WriteOK(w, events)
		return
	}
	events, _ := h.svc.ListByVideo(r.Context(), videoID)
	httputil.WriteOK(w, events)
}

func (h *HTTPHandler) handleBatchCount(w http.ResponseWriter, r *http.Request) {
	idsStr := r.URL.Query().Get("ids")
	if idsStr == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "ids required", "data": nil})
		return
	}
	parts := strings.Split(idsStr, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		id, _ := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if id > 0 {
			ids = append(ids, id)
		}
	}
	counts, _ := h.svc.CountByManuscriptIDs(r.Context(), ids)
	httputil.WriteOK(w, counts)
}

func (h *HTTPHandler) handleTrend(w http.ResponseWriter, r *http.Request) {
	idsStr := r.URL.Query().Get("ids")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	if idsStr == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "ids required", "data": nil})
		return
	}
	parts := strings.Split(idsStr, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		id, _ := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if id > 0 {
			ids = append(ids, id)
		}
	}
	trend, _ := h.svc.Trend(r.Context(), ids, startDate, endDate)
	httputil.WriteOK(w, trend)
}

func (h *HTTPHandler) handleByPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/danmaku/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid id", "data": nil})
		return
	}
	if r.Method == http.MethodDelete {
		userID := httputil.GetUserIDFromHeader(r)
		if userID == 0 {
			httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
			return
		}
		if err := h.svc.Delete(r.Context(), id, userID); err != nil {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "删除失败", "data": nil})
			return
		}
		httputil.WriteOK(w, map[string]any{"status": "ok"})
		return
	}
	httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
}

func (h *HTTPHandler) handleCreatorList(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	videoIDStr := r.URL.Query().Get("video_id")
	page, size := httputil.ParsePageParams(r)
	var videoID int64
	if videoIDStr != "" {
		videoID, _ = strconv.ParseInt(videoIDStr, 10, 64)
	}
	list, total, _ := h.svc.CreatorList(r.Context(), userID, videoID, page, size)
	httputil.WriteOK(w, map[string]any{"list": list, "total": total})
}

func (h *HTTPHandler) handleCreatorByPath(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/creator/danmaku/")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if r.Method == http.MethodDelete {
		userID := httputil.GetUserIDFromHeader(r)
		if userID == 0 {
			httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
			return
		}
		if err := h.svc.CreatorDelete(r.Context(), id, userID); err != nil {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "删除失败", "data": nil})
			return
		}
		httputil.WriteOK(w, map[string]any{"status": "ok"})
		return
	}
	httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
}

func (h *HTTPHandler) handleSSE(w http.ResponseWriter, r *http.Request) {
	videoIDStr := r.URL.Query().Get("video_id")
	videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid video_id", 400)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", 500)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := h.broadcaster.Subscribe(videoID)
	defer h.broadcaster.Unsubscribe(videoID, ch)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-time.After(30 * time.Second):
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

