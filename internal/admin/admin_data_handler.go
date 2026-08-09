package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type AdminDataHandler struct {
	db *sql.DB
}

func NewAdminDataHandler(db *sql.DB) *AdminDataHandler {
	return &AdminDataHandler{db: db}
}

func (h *AdminDataHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/manuscript/admin/", h.handleManuscriptAdmin)
	mux.HandleFunc("/api/v1/video/admin/", h.handleVideoAdmin)
	mux.HandleFunc("/api/v1/comment/admin/", h.handleCommentAdmin)
	mux.HandleFunc("/api/v1/admin/live/", h.handleLiveAdmin)
	mux.HandleFunc("/api/v1/admin/meeting/", h.handleMeetingAdmin)
	mux.HandleFunc("/api/v1/admin/security-settings", h.handleSecuritySettings)
	mux.HandleFunc("/api/v1/admin/content-review/", h.handleContentReview)
}

func (h *AdminDataHandler) handleManuscriptAdmin(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/manuscript/admin/")
	parts := strings.Split(path, "/")
	switch {
	case parts[0] == "pending" && r.Method == "GET":
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	case parts[0] == "all" && r.Method == "GET":
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	case parts[0] == "statistics" && r.Method == "GET":
		json.NewEncoder(w).Encode(map[string]interface{}{"total": 0, "pending": 0, "published": 0, "rejected": 0})
	case len(parts) >= 2 && parts[0] == "approve" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.exec(w, r, `UPDATE manuscripts SET review_status=1, status=2, review_time=NOW() WHERE id=$1`, id)
	case len(parts) >= 2 && parts[0] == "reject" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.exec(w, r, `UPDATE manuscripts SET review_status=2, status=4, review_time=NOW() WHERE id=$1`, id)
	case len(parts) >= 2 && parts[0] == "publish" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.exec(w, r, `UPDATE manuscripts SET status=3 WHERE id=$1`, id)
	case len(parts) >= 2 && parts[0] == "unpublish" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.exec(w, r, `UPDATE manuscripts SET status=-1 WHERE id=$1`, id)
	case len(parts) >= 2 && parts[0] == "retry" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.exec(w, r, `UPDATE manuscripts SET status=2, process_status=0 WHERE id=$1`, id)
	}
}

func (h *AdminDataHandler) handleVideoAdmin(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/video/admin/"), "/")
	if parts[0] == "list" && r.Method == "GET" {
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	} else if len(parts) >= 1 && r.Method == "DELETE" {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		h.exec(w, r, `DELETE FROM videos WHERE id=$1`, id)
	}
}

func (h *AdminDataHandler) handleCommentAdmin(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/comment/admin/"), "/")
	switch {
	case parts[0] == "list" && r.Method == "GET":
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	case len(parts) >= 1 && r.Method == "DELETE":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		h.exec(w, r, `UPDATE comments SET status=1 WHERE id=$1`, id)
	}
}

func (h *AdminDataHandler) handleLiveAdmin(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/live/"), "/")
	if parts[0] == "rooms" && r.Method == "GET" {
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	} else if parts[0] == "stats" && r.Method == "GET" {
		json.NewEncoder(w).Encode(map[string]interface{}{"live_count": 0, "total_viewers": 0})
	}
}

func (h *AdminDataHandler) handleMeetingAdmin(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/meeting/"), "/")
	switch {
	case parts[0] == "rooms" && r.Method == "GET":
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	case parts[0] == "pending" && r.Method == "GET":
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	case len(parts) >= 2 && parts[0] == "approve" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.exec(w, r, `UPDATE meeting_room SET status=0 WHERE id=$1`, id)
	case len(parts) >= 2 && parts[0] == "reject" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.exec(w, r, `UPDATE meeting_room SET status=4 WHERE id=$1`, id)
	}
}

func (h *AdminDataHandler) handleSecuritySettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		json.NewEncoder(w).Encode(map[string]interface{}{"rate_limit": 100, "captcha_enabled": false})
	} else {
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *AdminDataHandler) handleContentReview(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/content-review/"), "/")
	switch {
	case parts[0] == "pending" && r.Method == "GET":
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	case parts[0] == "all" && r.Method == "GET":
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	default:
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *AdminDataHandler) exec(w http.ResponseWriter, r *http.Request, query string, args ...interface{}) {
	_, err := h.db.ExecContext(r.Context(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte(`{"status":"ok"}`))
}
