package core

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HTTPHandler struct {
	danmakuSvc  *DanmakuService
	messageRepo *MessageRepository
	notifBroadcaster *NotificationBroadcaster
}

func NewHTTPHandler(danmakuSvc *DanmakuService, messageRepo *MessageRepository, notifBroadcaster *NotificationBroadcaster) *HTTPHandler {
	return &HTTPHandler{
		danmakuSvc:       danmakuSvc,
		messageRepo:      messageRepo,
		notifBroadcaster: notifBroadcaster,
	}
}

func (h *HTTPHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/danmaku/send", h.handleSendDanmaku)
	mux.HandleFunc("/api/v1/danmaku/video/", h.handleGetDanmaku)
	mux.HandleFunc("/sse/danmaku", h.handleSSEDanmaku)
	mux.HandleFunc("/api/v1/message/send", h.handleSendMessage)
	mux.HandleFunc("/api/v1/message/conversations", h.handleGetConversations)
	mux.HandleFunc("/sse/notify", h.handleSSENotify)
	mux.HandleFunc("/api/v1/health", h.handleHealth)
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *HTTPHandler) handleSendDanmaku(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}

	userID := getUserIDFromHeader(r)
	if userID == 0 {
		http.Error(w, "unauthorized", 401)
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
		http.Error(w, "invalid request", 400)
		return
	}
	if req.Content == "" {
		http.Error(w, "content required", 400)
		return
	}
	if req.Color == "" {
		req.Color = "#ffffff"
	}

	event, err := h.danmakuSvc.Send(r.Context(), req.VideoID, req.ManuscriptID, userID, req.Content, req.Time, req.Color, req.Mode)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func (h *HTTPHandler) handleGetDanmaku(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/danmaku/video/"), "/")
	videoID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid video id", 400)
		return
	}

	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	var events []*DanmakuEvent
	if startTimeStr != "" && endTimeStr != "" {
		startTime, _ := strconv.ParseFloat(startTimeStr, 64)
		endTime, _ := strconv.ParseFloat(endTimeStr, 64)
		events, _ = h.danmakuSvc.ListByTimeRange(r.Context(), videoID, startTime, endTime)
	} else {
		events, _ = h.danmakuSvc.ListByVideo(r.Context(), videoID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (h *HTTPHandler) handleSSEDanmaku(w http.ResponseWriter, r *http.Request) {
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

	ch := h.danmakuSvc.broadcaster.Subscribe(videoID)
	defer h.danmakuSvc.broadcaster.Unsubscribe(videoID, ch)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-ch:
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (h *HTTPHandler) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}

	userID := getUserIDFromHeader(r)
	if userID == 0 {
		http.Error(w, "unauthorized", 401)
		return
	}

	var req struct {
		ReceiverID int64  `json:"receiver_id"`
		Content    string `json:"content"`
		MessageType int32 `json:"message_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	if req.Content == "" {
		http.Error(w, "content required", 400)
		return
	}
	if req.MessageType == 0 {
		req.MessageType = 1
	}

	msg, err := h.messageRepo.SendMessage(r.Context(), userID, req.ReceiverID, req.Content, req.MessageType)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	h.notifBroadcaster.Send(req.ReceiverID, &NotificationEvent{
		Type: "message", Content: req.Content, FromUID: userID,
		CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func (h *HTTPHandler) handleGetConversations(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	if userID == 0 {
		http.Error(w, "unauthorized", 401)
		return
	}

	convs, err := h.messageRepo.GetConversations(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(convs)
}

func (h *HTTPHandler) handleSSENotify(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user_id", 400)
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

	ch := h.notifBroadcaster.Subscribe(userID)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-ch:
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-time.After(30 * time.Second):
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func getUserIDFromHeader(r *http.Request) int64 {
	idStr := r.Header.Get("X-User-Id")
	if idStr == "" {
		return 0
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func StartHTTPServer(addr string, h *HTTPHandler) {
	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}