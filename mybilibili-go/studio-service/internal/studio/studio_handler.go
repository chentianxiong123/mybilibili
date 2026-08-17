package studio

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/httputil"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/studio/export-tasks", h.handleCreateTask)
	mux.HandleFunc("/api/v1/studio/export-tasks/", h.handleTaskByID)
	mux.HandleFunc("/api/v1/studio/assets/upload", h.handleAssetUpload)
}

func (h *Handler) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	var req struct {
		ProjectID string `json:"projectId"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.ProjectID == "" {
		http.Error(w, "projectId required", 400)
		return
	}
	t, err := h.svc.CreateTask(r.Context(), userID, req.ProjectID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(t)
}

func (h *Handler) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/api/v1/studio/export-tasks/")
	switch {
	case r.Method == "GET":
		t, err := h.svc.GetTask(r.Context(), taskID)
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(t)
	case strings.HasSuffix(r.URL.Path, "/cancel") && r.Method == "POST":
		h.svc.CancelTask(r.Context(), taskID)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleAssetUpload(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "parse form: "+err.Error(), 400)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required", 400)
		return
	}
	defer file.Close()
	assetType := r.FormValue("type")
	taskID, _ := strconv.ParseInt(r.FormValue("task_id"), 10, 64)
	url, err := h.svc.UploadAsset(r.Context(), userID, taskID, assetType, header.Filename, file)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "uploaded", "url": url})
}


