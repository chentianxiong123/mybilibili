package subtitle

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/subtitle/video/", h.handleVideoSubtitle)
	mux.HandleFunc("/api/v1/subtitle/pending", h.handlePending)
	mux.HandleFunc("/api/v1/subtitle/upload", h.handleUpload)
	mux.HandleFunc("/api/v1/subtitle/", h.handleSubtitleByID)
}

func (h *Handler) handleVideoSubtitle(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/subtitle/video/")
	parts := strings.Split(path, "/")
	videoID, _ := strconv.ParseInt(parts[0], 10, 64)

	if len(parts) >= 2 && parts[1] != "" {
		lang := parts[1]
		sub, err := h.svc.GetByLanguage(r.Context(), videoID, lang)
		if err != nil || sub == nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(sub)
		return
	}

	list, _ := h.svc.ListByVideo(r.Context(), videoID)
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handlePending(w http.ResponseWriter, r *http.Request) {
	list, _ := h.svc.ListPending(r.Context())
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}

	userID := getUserID(r)
	var req struct {
		VideoID      int64  `json:"video_id"`
		Language     string `json:"language"`
		LanguageName string `json:"language_name"`
		Content      string `json:"content"`
		SRTContent   string `json:"srt_content"`
		IsDefault    bool   `json:"is_default"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	content := req.Content
	if content == "" && req.SRTContent != "" {
		content = req.SRTContent
	}
	if content == "" {
		http.Error(w, "content required", 400)
		return
	}
	if req.Language == "" {
		req.Language = "zh-CN"
	}
	if req.LanguageName == "" {
		req.LanguageName = "中文"
	}

	sub, err := h.svc.Upload(r.Context(), req.VideoID, userID, req.Language, req.LanguageName, content)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if req.IsDefault {
		h.svc.SetDefault(r.Context(), req.VideoID, sub.ID)
	}

	json.NewEncoder(w).Encode(sub)
}

func (h *Handler) handleSubtitleByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/subtitle/")
	parts := strings.Split(path, "/")
	id := parts[0]

	switch {
	case len(parts) >= 2 && parts[1] == "approve" && r.Method == "POST":
		h.svc.Approve(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	case len(parts) >= 2 && parts[1] == "reject" && r.Method == "POST":
		h.svc.Reject(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	case len(parts) >= 2 && parts[1] == "preview" && r.Method == "GET":
		cues, _ := h.svc.Preview(r.Context(), id)
		json.NewEncoder(w).Encode(cues)
	case len(parts) >= 2 && parts[1] == "set-default" && r.Method == "POST":
		videoID, _ := strconv.ParseInt(r.URL.Query().Get("video_id"), 10, 64)
		h.svc.SetDefault(r.Context(), videoID, id)
		w.Write([]byte(`{"status":"ok"}`))
	case r.Method == "GET":
		sub, err := h.svc.repo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(sub)
	case r.Method == "DELETE":
		h.svc.Delete(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func getUserID(r *http.Request) int64 {
	idStr := r.Header.Get("X-User-Id")
	if idStr == "" {
		return 0
	}
	id, _ := strconv.ParseInt(idStr, 10, 64)
	return id
}
