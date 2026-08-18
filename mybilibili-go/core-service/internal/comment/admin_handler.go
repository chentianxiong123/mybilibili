package comment

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mybilibili/pkg/httputil"
)

type AdminHandler struct {
	db *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

func (h *AdminHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/admin/comments/", h.handleRoute)
}

func (h *AdminHandler) handleRoute(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/comments/"), "/")
	switch {
	case parts[0] == "list" && r.Method == "GET":
		status := r.URL.Query().Get("status")
		keyword := r.URL.Query().Get("keyword")
		page, size := httputil.ParsePageParams(r)
		offset := (page - 1) * size
		conds := ""
		args := []interface{}{}
		if status != "" {
			args = append(args, status)
			conds += fmt.Sprintf(" AND c.status = $%d", len(args))
		}
		if keyword != "" {
			args = append(args, "%"+keyword+"%")
			conds += fmt.Sprintf(" AND c.content LIKE $%d", len(args))
		}
		args = append(args, size, offset)
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT c.id, c.manuscript_id, c.user_id, c.content, c.like_count, c.reply_count, c.status, c.created_at
			 FROM comments c WHERE 1=1`+conds+` ORDER BY c.id DESC LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"list": []map[string]interface{}{}, "total": 0})
			return
		}
		defer rows.Close()
		list := []map[string]interface{}{}
		for rows.Next() {
			var id, msID, userID, likeCount, replyCount, st int64
			var content string
			var createdAt time.Time
			rows.Scan(&id, &msID, &userID, &content, &likeCount, &replyCount, &st, &createdAt)
			list = append(list, map[string]interface{}{
				"id": id, "manuscript_id": msID, "user_id": userID, "content": content,
				"like_count": likeCount, "reply_count": replyCount, "status": st,
				"created_at": createdAt.Format("2006-01-02T15:04:05Z"),
			})
		}
		var total int64
		countConds := ""
		countArgs := []interface{}{}
		if status != "" {
			countArgs = append(countArgs, status)
			countConds += fmt.Sprintf(" AND c.status = $%d", len(countArgs))
		}
		if keyword != "" {
			countArgs = append(countArgs, "%"+keyword+"%")
			countConds += fmt.Sprintf(" AND c.content LIKE $%d", len(countArgs))
		}
		h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM comments c WHERE 1=1`+countConds, countArgs...).Scan(&total)
		json.NewEncoder(w).Encode(map[string]interface{}{"list": list, "total": total})
	case len(parts) >= 1 && parts[0] != "" && r.Method == "GET":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var msID, userID, likeCount, replyCount, st int64
		var content string
		var createdAt time.Time
		err := h.db.QueryRowContext(r.Context(),
			`SELECT id, manuscript_id, user_id, content, like_count, reply_count, status, created_at
			 FROM comments WHERE id = $1`, id).Scan(&id, &msID, &userID, &content, &likeCount, &replyCount, &st, &createdAt)
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": id, "manuscript_id": msID, "user_id": userID, "content": content,
			"like_count": likeCount, "reply_count": replyCount, "status": st,
			"created_at": createdAt.Format("2006-01-02T15:04:05Z"),
		})
	case len(parts) >= 2 && parts[1] == "status" && r.Method == "PUT":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req struct {
			Status int `json:"status"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.exec(w, r, `UPDATE comments SET status=$1, updated_at=NOW() WHERE id=$2`, req.Status, id)
	case len(parts) >= 2 && parts[1] == "delete" && r.Method == "DELETE":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		h.exec(w, r, `DELETE FROM comments WHERE id=$1`, id)
	case len(parts) >= 1 && r.Method == "DELETE":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		h.exec(w, r, `UPDATE comments SET status=1, updated_at=NOW() WHERE id=$1`, id)
	}
}

func (h *AdminHandler) exec(w http.ResponseWriter, r *http.Request, query string, args ...interface{}) {
	_, err := h.db.ExecContext(r.Context(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte(`{"status":"ok"}`))
}
