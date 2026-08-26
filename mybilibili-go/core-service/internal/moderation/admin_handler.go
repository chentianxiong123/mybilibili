package moderation

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/httputil"
)

type AdminHandler struct {
	db *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

func (h *AdminHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/moderation/admin/", h.handleRoute)
}

func (h *AdminHandler) handleRoute(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/moderation/admin/"), "/")
	contentType := r.URL.Query().Get("contentType")
	status := r.URL.Query().Get("status")
	page, size := httputil.ParsePageParams(r)
	conds := ""
	args := []interface{}{}
	if contentType != "" {
		args = append(args, contentType)
		conds += fmt.Sprintf(" AND type = $%d", len(args))
	}
	args = append(args, status)
	conds += fmt.Sprintf(" AND ($%d::text = '' OR status = $%d::text)", len(args), len(args))
	args = append(args, page, size)
	switch {
	case parts[0] == "pending" && r.Method == "GET":
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT id, type, user_id, content, status, reviewed_at FROM content_reviews
			 WHERE status = 'pending'`+conds+` ORDER BY id DESC LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
		if err != nil {
			httputil.WriteOK(w, []map[string]any{})
			return
		}
		defer rows.Close()
		list := []map[string]any{}
		for rows.Next() {
			var id int64
			var typ, content string
			var userID sql.NullInt64
			var st sql.NullString
			var reviewedAt sql.NullTime
			rows.Scan(&id, &typ, &userID, &content, &st, &reviewedAt)
			list = append(list, map[string]any{
				"id": id, "type": typ, "user_id": userID.Int64, "content": content,
				"status": st.String, "reviewed_at": reviewedAt,
			})
		}
		httputil.WriteOK(w, list)
	case parts[0] == "all" && r.Method == "GET":
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT id, type, user_id, content, status, reviewed_at FROM content_reviews
			 WHERE 1=1`+conds+` ORDER BY id DESC LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
		if err != nil {
			httputil.WriteOK(w, []map[string]any{})
			return
		}
		defer rows.Close()
		list := []map[string]any{}
		for rows.Next() {
			var id int64
			var typ, content string
			var userID sql.NullInt64
			var st sql.NullString
			var reviewedAt sql.NullTime
			rows.Scan(&id, &typ, &userID, &content, &st, &reviewedAt)
			list = append(list, map[string]any{
				"id": id, "type": typ, "user_id": userID.Int64, "content": content,
				"status": st.String, "reviewed_at": reviewedAt,
			})
		}
		httputil.WriteOK(w, list)
	case parts[0] == "restore" && len(parts) >= 3 && r.Method == "PUT":
		h.exec(w, r, `UPDATE content_reviews SET status = 'active', reviewed_at = NOW() WHERE type = $1 AND id = $2`, parts[1], parts[2])
	case len(parts) >= 2 && r.Method == "DELETE":
		h.exec(w, r, `UPDATE content_reviews SET status = 'deleted', reviewed_at = NOW() WHERE type = $1 AND id = $2`, parts[0], parts[1])
	case parts[0] == "batch" && r.Method == "POST":
		var req struct {
			Action string           `json:"action"`
			Items  []map[string]any `json:"items"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid body", "data": nil})
			return
		}
		for _, item := range req.Items {
			typ, _ := item["type"].(string)
			id, _ := item["id"].(float64)
			if req.Action == "restore" {
				h.exec(w, r, `UPDATE content_reviews SET status = 'active', reviewed_at = NOW() WHERE type = $1 AND id = $2`, typ, int64(id))
			} else {
				h.exec(w, r, `UPDATE content_reviews SET status = 'deleted', reviewed_at = NOW() WHERE type = $1 AND id = $2`, typ, int64(id))
			}
		}
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	default:
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
	}
}

func (h *AdminHandler) exec(w http.ResponseWriter, r *http.Request, query string, args ...interface{}) {
	_, err := h.db.ExecContext(r.Context(), query, args...)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "操作失败", "data": nil})
		return
	}
	httputil.WriteOK(w, map[string]any{"status": "ok"})
}
