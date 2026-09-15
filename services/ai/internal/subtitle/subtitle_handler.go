package subtitle

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/httputil"
)

type Handler struct {
	svc *Service
	gen *WhisperGenerator
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// SetGenerator 注入 whisper 字幕生成器（由 main 注入 MinIO 存储）。
func (h *Handler) SetGenerator(gen *WhisperGenerator) *Handler {
	h.gen = gen
	return h
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/subtitle/videos", h.handleAllVideos)
	mux.HandleFunc("/api/v1/subtitle/import-srt", h.handleImport)
	mux.HandleFunc("/api/v1/subtitle/scan/", h.handleScan)
	mux.HandleFunc("/api/v1/subtitle/import-system", h.handleImportSystem)
	mux.HandleFunc("/api/v1/subtitle/set-default", h.handleSetDefault)
	mux.HandleFunc("/api/v1/subtitle/generate", h.handleGenerate)
	mux.HandleFunc("/api/v1/subtitle/video/", h.handleVideoSubtitle)
	mux.HandleFunc("/api/v1/subtitle/pending", h.handlePending)
	mux.HandleFunc("/api/v1/subtitle/upload", h.handleUpload)
	mux.HandleFunc("/api/v1/subtitle/upload-srt", h.handleUploadSRT)
	mux.HandleFunc("/api/v1/subtitle/", h.handleSubtitleByID)
}

func (h *Handler) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed"})
		return
	}
	if h.gen == nil {
		httputil.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{"code": 503, "message": "whisper generator not configured"})
		return
	}
	var req struct {
		ManuscriptID int64 `json:"manuscript_id"`
		VideoID      int64 `json:"video_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "bad request"})
		return
	}
	if req.VideoID <= 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "video_id required"})
		return
	}
	id, cues, err := h.gen.GenerateFromAudio(r.Context(), req.ManuscriptID, req.VideoID)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": err.Error()})
		return
	}
	httputil.WriteOK(w, map[string]any{"subtitle_id": id, "cues": cues})
}

func subtitleToMap(sub *Subtitle) map[string]interface{} {
	m := map[string]interface{}{
		"id":            sub.ID,
		"video_id":      sub.VideoID,
		"videoId":       sub.VideoID,
		"language":      sub.Language,
		"language_name": sub.LanguageName,
		"languageName":  sub.LanguageName,
		"format":        sub.Format,
		"is_default":    sub.IsDefault,
		"isDefault":     sub.IsDefault,
		"uploaded_by":   sub.UploadedBy,
		"status":        sub.Status,
		"source":        sub.Source,
		"upload_time":   sub.UploadTime,
	}
	var cues []map[string]interface{}
	if json.Unmarshal([]byte(sub.Content), &cues) == nil {
		m["content"] = cues
		m["cues"] = cues
	} else if parsed, err := ParseSRT(sub.Content); err == nil {
		cues = make([]map[string]interface{}, 0, len(parsed))
		for _, c := range parsed {
			cues = append(cues, c.ToCueMap())
		}
		m["content"] = cues
		m["cues"] = cues
	} else {
		m["content"] = []interface{}{}
		m["cues"] = []interface{}{}
	}
	return m
}

func (h *Handler) handleVideoSubtitle(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/subtitle/video/")
	parts := strings.Split(path, "/")
	videoID, _ := strconv.ParseInt(parts[0], 10, 64)

	if len(parts) >= 2 && parts[1] != "" {
		lang := parts[1]
		sub, err := h.svc.GetByLanguage(r.Context(), videoID, lang)
		if err != nil || sub == nil {
			httputil.WriteOK(w, map[string]interface{}{})
			return
		}
		httputil.WriteOK(w, subtitleToMap(sub))
		return
	}

	list, _ := h.svc.ListByVideo(r.Context(), videoID)
	if list == nil {
		list = []*Subtitle{}
	}
	out := make([]map[string]interface{}, 0, len(list))
	for _, sub := range list {
		out = append(out, subtitleToMap(sub))
	}
	httputil.WriteOK(w, out)
}

func (h *Handler) handlePending(w http.ResponseWriter, r *http.Request) {
	list, _ := h.svc.ListPending(r.Context())
	if list == nil {
		list = []*Subtitle{}
	}
	out := make([]map[string]interface{}, 0, len(list))
	for _, sub := range list {
		out = append(out, subtitleToMap(sub))
	}
	httputil.WriteOK(w, out)
}

func (h *Handler) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}

	userID := httputil.GetUserIDFromHeader(r)
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

	storeContent := content
	if req.Content == "" && req.SRTContent != "" {
		if cues, err := ParseSRT(req.SRTContent); err == nil {
			storeContent = SRTCuesToJSON(cues)
		}
	}

	sub, err := h.svc.Upload(r.Context(), req.VideoID, userID, req.Language, req.LanguageName, storeContent)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if req.IsDefault {
		h.svc.SetDefault(r.Context(), req.VideoID, sub.ID)
	}

	httputil.WriteOK(w, subtitleToMap(sub))
}

func (h *Handler) handleUploadSRT(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
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
	storeContent := string(content)
	if cues, err := ParseSRT(storeContent); err == nil {
		storeContent = SRTCuesToJSON(cues)
	}
	sub, err := h.svc.Upload(r.Context(), videoID, userID, language, languageName, storeContent)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	httputil.WriteOK(w, subtitleToMap(sub))
}

func (h *Handler) handleAllVideos(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	list, err := h.svc.ListAll(r.Context())
	if err != nil {
		httputil.WriteOK(w, []map[string]interface{}{})
		return
	}
	out := make([]map[string]interface{}, 0, len(list))
	for _, s := range list {
		m := subtitleToMap(s)
		m["created_at"] = s.UploadTime.Format("2006-01-02 15:04:05")
		out = append(out, m)
	}
	httputil.WriteOK(w, out)
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
	cues, err := ParseSRT(req.Srt)
	if err != nil || len(cues) == 0 {
		http.Error(w, "invalid srt content", 400)
		return
	}
	storeContent := SRTCuesToJSON(cues)
	sub, err := h.svc.Upload(r.Context(), req.VideoID, httputil.GetUserIDFromHeader(r), "zh-CN", "中文", storeContent)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	httputil.WriteOK(w, subtitleToMap(sub))
}

func (h *Handler) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	videoID, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/subtitle/scan/"), 10, 64)
	list, err := h.svc.ListByVideoForScan(r.Context(), videoID)
	if err != nil {
		httputil.WriteOK(w, []map[string]interface{}{})
		return
	}
	out := make([]map[string]interface{}, 0, len(list))
	for _, s := range list {
		out = append(out, subtitleToMap(s))
	}
	httputil.WriteOK(w, out)
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
	storeContent := req.Srt
	if cues, err := ParseSRT(req.Srt); err == nil {
		storeContent = SRTCuesToJSON(cues)
	}
	sub, err := h.svc.Upload(r.Context(), req.VideoID, 0, "zh-CN", "中文", storeContent)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	h.svc.Approve(r.Context(), sub.ID)
	httputil.WriteOK(w, subtitleToMap(sub))
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
	httputil.WriteOK(w, map[string]string{"status": "ok"})
}

func (h *Handler) handleSubtitleByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/subtitle/")
	parts := strings.Split(path, "/")
	id := parts[0]

	switch {
	case len(parts) >= 2 && parts[1] == "approve" && r.Method == "POST":
		h.svc.Approve(r.Context(), id)
		httputil.WriteOK(w, map[string]string{"status": "ok"})
	case len(parts) >= 2 && parts[1] == "reject" && r.Method == "POST":
		h.svc.Reject(r.Context(), id)
		httputil.WriteOK(w, map[string]string{"status": "ok"})
	case len(parts) >= 2 && parts[1] == "preview" && r.Method == "GET":
		cues, _ := h.svc.Preview(r.Context(), id)
		httputil.WriteOK(w, cues)
	case len(parts) >= 2 && parts[1] == "set-default" && r.Method == "POST":
		videoID, _ := strconv.ParseInt(r.URL.Query().Get("video_id"), 10, 64)
		h.svc.SetDefault(r.Context(), videoID, id)
		httputil.WriteOK(w, map[string]string{"status": "ok"})
	case r.Method == "GET":
		sub, err := h.svc.repo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		httputil.WriteOK(w, subtitleToMap(sub))
	case r.Method == "DELETE":
		h.svc.Delete(r.Context(), id)
		httputil.WriteOK(w, map[string]string{"status": "ok"})
	}
}


