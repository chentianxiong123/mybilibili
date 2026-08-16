package profile

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/core-service/internal/httputil"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/profile/record/", h.handleRecord)
	mux.HandleFunc("/api/v1/profile/", h.handleProfileByPath)
}

func (h *Handler) handleProfileByPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/profile/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "user id required", 400)
		return
	}
	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", 400)
		return
	}

	switch {
	case len(parts) >= 2 && parts[1] == "init":
		if r.Method != "POST" {
			http.Error(w, "method not allowed", 405)
			return
		}
		var req struct {
			Tags []string `json:"tags"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		p, err := h.svc.Init(r.Context(), userID, req.Tags)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(p)
	case r.Method == "GET":
		p, err := h.svc.GetOrCreate(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(p)
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *Handler) handleRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	action := strings.TrimPrefix(r.URL.Path, "/api/v1/profile/record/")
	var req struct {
		CategoryID     int64    `json:"categoryId"`
		Tags           []string `json:"tags"`
		DurationSecond int64    `json:"durationSeconds"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var err error
	switch action {
	case "watch":
		err = h.svc.RecordWatch(r.Context(), userID, req.CategoryID, req.Tags, req.DurationSecond)
	case "like":
		err = h.svc.RecordLike(r.Context(), userID, req.CategoryID, req.Tags)
	case "collect":
		err = h.svc.RecordCollect(r.Context(), userID, req.CategoryID, req.Tags)
	default:
		http.Error(w, "unknown action", 400)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Write([]byte(`{"status":"ok"}`))
}


