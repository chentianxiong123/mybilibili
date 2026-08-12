package core

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// ManuscriptHTTPHandler 提供 Java 侧 HTTP-only 的稿件接口兜底：
// upload-session 会话、internal take-down、fix-durations、comment-count 维护。
type ManuscriptHTTPHandler struct {
	db *sql.DB
}

func NewManuscriptHTTPHandler(db *sql.DB) *ManuscriptHTTPHandler {
	return &ManuscriptHTTPHandler{db: db}
}

func (h *ManuscriptHTTPHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/manuscript/upload-session", h.handleUploadSession)
	mux.HandleFunc("/api/v1/manuscript/upload-session/", h.handleUploadSessionByID)
	mux.HandleFunc("/api/v1/manuscript/fix-durations", h.handleFixDurations)
	mux.HandleFunc("/api/v1/manuscript/internal/", h.handleInternalByPath)
	mux.HandleFunc("/api/v1/manuscript/", h.handleManuscriptByID)
}

func (h *ManuscriptHTTPHandler) handleUploadSession(w http.ResponseWriter, r *http.Request) {
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
		Title       string `json:"title"`
		Description string `json:"description"`
		CategoryID  int64  `json:"category_id"`
		Tags        []string           `json:"tags"`
		Videos      []map[string]any   `json:"videos"`
		TotalChunks *int               `json:"total_chunks"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Title == "" && req.CategoryID == 0 && req.TotalChunks == nil {
		http.Error(w, "invalid upload session request", 400)
		return
	}
	sum := md5.Sum([]byte(fmt.Sprintf("%d-%d-%s", userID, req.CategoryID, req.Title)))
	uploadID := hex.EncodeToString(sum[:])
	total := 0
	if req.TotalChunks != nil {
		total = *req.TotalChunks
	}
	for _, v := range req.Videos {
		if tc, ok := v["total_chunks"].(float64); ok {
			total += int(tc)
		}
	}
	tags, _ := json.Marshal(req.Tags)
	videos, _ := json.Marshal(req.Videos)
	_, err := h.db.ExecContext(r.Context(),
		`INSERT INTO upload_sessions (id, user_id, title, description, category_id, tags, videos, total_chunks, status)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'pending')
		 ON CONFLICT (id) DO UPDATE SET title=$3, updated_at=NOW()`,
		uploadID, userID, req.Title, req.Description, req.CategoryID, string(tags), string(videos), total)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"upload_id": uploadID, "status": "pending", "total_chunks": total})
}

func (h *ManuscriptHTTPHandler) handleUploadSessionByID(w http.ResponseWriter, r *http.Request) {
	uploadID := strings.TrimPrefix(r.URL.Path, "/api/v1/manuscript/upload-session/")
	if uploadID == "" {
		http.Error(w, "not found", 404)
		return
	}
	userID := getUserIDFromHeader(r)
	switch r.Method {
	case "GET":
		row := h.db.QueryRowContext(r.Context(),
			`SELECT id, user_id, title, COALESCE(category_id,0), uploaded_chunks, total_chunks, status
			 FROM upload_sessions WHERE id = $1 AND user_id = $2`, uploadID, userID)
		var id string
		var uid int64
		var title string
		var catID int64
		var uploaded, total int
		var status string
		if err := row.Scan(&id, &uid, &title, &catID, &uploaded, &total, &status); err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"upload_id": id, "user_id": uid, "title": title, "category_id": catID,
			"uploaded_chunks": uploaded, "total_chunks": total, "status": status,
		})
	case "DELETE":
		_, err := h.db.ExecContext(r.Context(), `DELETE FROM upload_sessions WHERE id = $1 AND user_id = $2`, uploadID, userID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(`{"status":"ok"}`))
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *ManuscriptHTTPHandler) handleFixDurations(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	_, _ = h.db.ExecContext(r.Context(),
		`UPDATE manuscripts m SET duration_seconds = COALESCE((
		   SELECT SUM(v.duration_seconds) FROM videos v WHERE v.manuscript_id = m.id
		 ), 0), updated_at = NOW()`)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *ManuscriptHTTPHandler) handleInternalByPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/manuscript/internal/")
	parts := strings.Split(path, "/")
	if len(parts) >= 2 && parts[1] == "take-down" && r.Method == "PUT" {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		_, _ = h.db.ExecContext(r.Context(),
			`UPDATE manuscripts SET status = -1, review_status = 2, updated_at = NOW() WHERE id = $1`, id)
		w.Write([]byte(`{"status":"ok"}`))
		return
	}
	http.Error(w, "not found", 404)
}

func (h *ManuscriptHTTPHandler) handleManuscriptByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/manuscript/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "not found", 404)
		return
	}
	id, _ := strconv.ParseInt(parts[0], 10, 64)

	switch {
	case len(parts) >= 2 && parts[1] == "comment-count" && r.Method == "PUT":
		count, _ := strconv.ParseInt(r.URL.Query().Get("count"), 10, 64)
		_, _ = h.db.ExecContext(r.Context(), `UPDATE manuscripts SET comment_count = $2 WHERE id = $1`, id, count)
		w.Write([]byte(`{"status":"ok"}`))
	case len(parts) >= 2 && parts[1] == "increment-comment" && r.Method == "POST":
		_, _ = h.db.ExecContext(r.Context(), `UPDATE manuscripts SET comment_count = comment_count + 1 WHERE id = $1`, id)
		w.Write([]byte(`{"status":"ok"}`))
	case len(parts) >= 2 && parts[1] == "decrement-comment" && r.Method == "POST":
		_, _ = h.db.ExecContext(r.Context(), `UPDATE manuscripts SET comment_count = GREATEST(comment_count - 1, 0) WHERE id = $1`, id)
		w.Write([]byte(`{"status":"ok"}`))
	default:
		http.Error(w, "not found", 404)
	}
}