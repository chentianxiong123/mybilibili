package subtitle

import (
	"encoding/json"
	"io"
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
	mux.HandleFunc("/api/v1/subtitle/videos", h.handleAllVideos)
	mux.HandleFunc("/api/v1/subtitle/import-srt", h.handleImport)
	mux.HandleFunc("/api/v1/subtitle/scan/", h.handleScan)
	mux.HandleFunc("/api/v1/subtitle/import-system", h.handleImportSystem)
	mux.HandleFunc("/api/v1/subtitle/set-default", h.handleSetDefault)
	mux.HandleFunc("/api/v1/subtitle/video/", h.handleVideoSubtitle)
	mux.HandleFunc("/api/v1/subtitle/pending", h.handlePending)
	mux.HandleFunc("/api/v1/subtitle/upload", h.handleUpload)
	mux.HandleFunc("/api/v1/subtitle/upload-srt", h.handleUploadSRT)
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

func (h *Handler) handleUploadSRT(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := getUserID(r)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "parse form: "+err.Error(), 400)
		return
	}
	videoID, _ := strconv.ParseInt(r.FormValue("video_id"), 10, 64)
	language := r.FormValue("language")
	if language == "" {
		language = "zh-CN"
	}
	languageName := r.FormValue("language_name")
	if languageName == "" {
		languageName = "中文"
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required", 400)
		return
	}
	defer file.Close()
	content, _ := io.ReadAll(file)
	if len(content) == 0 {
		http.Error(w, "empty file", 400)
		return
	}
	sub, err := h.svc.Upload(r.Context(), videoID, userID, language, languageName, string(content))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(sub)
}

func (h *Handler) handleAllVideos(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	list, err := h.svc.ListAll(r.Context())
	if err != nil {
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}
	out := make([]map[string]interface{}, 0, len(list))
	for _, s := range list {
		out = append(out, map[string]interface{}{
			"id": s.ID, "video_id": s.VideoID, "language": s.Language,
			"language_name": s.LanguageName, "status": s.Status,
			"created_at": s.UploadTime.Format("2006-01-02 15:04:05"),
		})
	}
	json.NewEncoder(w).Encode(out)
}

func (h *Handler) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		VideoID int64  `json:"video_id"`
		Srt     string `json:"srt"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.VideoID == 0 || req.Srt == "" {
		http.Error(w, "video_id and srt required", 400)
		return
	}
	sub, err := h.svc.Upload(r.Context(), req.VideoID, getUserID(r), "zh-CN", "中文", req.Srt)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(sub)
}

func (h *Handler) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	videoID, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/subtitle/scan/"), 10, 64)
	list, err := h.svc.ListByVideoForScan(r.Context(), videoID)
	if err != nil {
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}
	out := make([]map[string]interface{}, 0, len(list))
	for _, s := range list {
		out = append(out, map[string]interface{}{
			"id": s.ID, "video_id": s.VideoID, "language": s.Language,
			"language_name": s.LanguageName, "status": s.Status,
		})
	}
	json.NewEncoder(w).Encode(out)
}

func (h *Handler) handleImportSystem(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		VideoID int64  `json:"video_id"`
		Srt     string `json:"srt"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.VideoID == 0 || req.Srt == "" {
		http.Error(w, "video_id and srt required", 400)
		return
	}
	sub, err := h.svc.Upload(r.Context(), req.VideoID, 0, "zh-CN", "中文", req.Srt)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	h.svc.Approve(r.Context(), sub.ID)
	json.NewEncoder(w).Encode(sub)
}

func (h *Handler) handleSetDefault(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		VideoID int64  `json:"video_id"`
		ID      string `json:"id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	h.svc.SetDefault(r.Context(), req.VideoID, req.ID)
	w.Write([]byte(`{"status":"ok"}`))
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
