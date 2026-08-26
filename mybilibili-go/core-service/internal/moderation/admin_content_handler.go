package moderation

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/httputil"
)

// AdminContentHandler 提供内容审核中心所需的评论区/弹幕管理接口。
type AdminContentHandler struct {
	db *sql.DB
}

func NewAdminContentHandler(db *sql.DB) *AdminContentHandler {
	return &AdminContentHandler{db: db}
}

func (h *AdminContentHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/moderation/admin/comments", h.handleCommentList)
	mux.HandleFunc("/api/v1/moderation/admin/comments/", h.handleCommentByID)
	mux.HandleFunc("/api/v1/moderation/admin/danmaku", h.handleDanmakuList)
	mux.HandleFunc("/api/v1/moderation/admin/danmaku/", h.handleDanmakuByID)
}

func (h *AdminContentHandler) handleCommentList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	page, size := httputil.ParsePageParams(r)
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
	targetType := r.URL.Query().Get("type")
	if targetType == "" {
		targetType = "comment"
	}
	status := r.URL.Query().Get("status")

	var rows *sql.Rows
	var err error
	if targetType == "reply" {
		query := `SELECT r.id, r.comment_id, c.manuscript_id, r.user_id,
		          COALESCE(u.nickname, u.username), u.avatar,
		          COALESCE(m.title, ''), c.content, r.content,
		          r.like_count, r.status, r.created_at
		          FROM replies r
		          JOIN users u ON u.id = r.user_id
		          JOIN comments c ON c.id = r.comment_id
		          LEFT JOIN manuscripts m ON m.id = c.manuscript_id
		          WHERE 1=1`
		args := []interface{}{}
		query, args = appendFilter(query, args, keyword, "r.content")
		if status != "" {
			query, args = appendFilter(query, args, status, "r.status")
		}
		args = append(args, size, (page-1)*size)
		rows, err = h.db.QueryContext(r.Context(), query+fmt.Sprintf(" ORDER BY r.id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args...)
	} else {
		query := `SELECT c.id, NULL::bigint, c.manuscript_id, c.user_id,
		          COALESCE(u.nickname, u.username), u.avatar,
		          COALESCE(m.title, ''), NULL::text, c.content,
		          c.like_count, c.status::text, c.created_at
		          FROM comments c
		          JOIN users u ON u.id = c.user_id
		          LEFT JOIN manuscripts m ON m.id = c.manuscript_id
		          WHERE 1=1`
		args := []interface{}{}
		query, args = appendFilter(query, args, keyword, "c.content")
		if status != "" {
			args = append(args, status)
			query += fmt.Sprintf(" AND c.status = $%d::int", len(args))
		}
		args = append(args, size, (page-1)*size)
		rows, err = h.db.QueryContext(r.Context(), query+fmt.Sprintf(" ORDER BY c.id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args...)
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	for rows.Next() {
		var id int64
		var parentID sql.NullInt64
		var manuscriptID sql.NullInt64
		var userID int64
		var nickname, avatar, manuscriptTitle string
		var parentContent sql.NullString
		var content string
		var likeCount int
		var status sql.NullString
		var createdAt sql.NullTime
		if err := rows.Scan(&id, &parentID, &manuscriptID, &userID, &nickname, &avatar, &manuscriptTitle, &parentContent, &content, &likeCount, &status, &createdAt); err != nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"id":              id,
			"type":            targetType,
			"parentId":        parentID.Int64,
			"manuscriptId":    manuscriptID.Int64,
			"manuscriptTitle": manuscriptTitle,
			"userId":          userID,
			"userName":        nickname,
			"userAvatar":      avatar,
			"parentContent":   parentContent.String,
			"content":         content,
			"likeCount":       likeCount,
			"status":          status.String,
			"createdAt":       createdAt.Time,
		})
	}

	var total int64
	if targetType == "reply" {
		_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM replies`).Scan(&total)
	} else {
		_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM comments`).Scan(&total)
	}
	httputil.WriteOK(w, map[string]interface{}{"list": list, "total": total})
}

func appendFilter(query string, args []interface{}, value, column string) (string, []interface{}) {
	if value == "" {
		return query, args
	}
	args = append(args, value)
	return query + fmt.Sprintf(" AND %s = $%d", column, len(args)), args
}

func (h *AdminContentHandler) handleCommentByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/moderation/admin/comments/"), "/")
	if len(parts) < 2 {
		http.Error(w, "invalid path", 400)
		return
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", 400)
		return
	}
	switch r.Method {
	case http.MethodDelete:
		if parts[0] == "reply" {
			h.exec(w, r, `UPDATE replies SET status = 'REMOVED' WHERE id = $1`, id)
		} else {
			h.exec(w, r, `UPDATE comments SET status = 1 WHERE id = $1`, id)
		}
	case http.MethodPut:
		if parts[0] == "reply" {
			h.exec(w, r, `UPDATE replies SET status = 'NORMAL' WHERE id = $1`, id)
		} else {
			h.exec(w, r, `UPDATE comments SET status = 0 WHERE id = $1`, id)
		}
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *AdminContentHandler) handleDanmakuList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	page, size := httputil.ParsePageParams(r)
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))

	query := `SELECT d.id, d.video_id, d.manuscript_id, COALESCE(m.title, '(稿件已删除)'),
	          d.user_id, COALESCE(u.nickname, u.username), u.avatar,
	          d.content, d.time, d.mode, d.created_at
	          FROM danmaku d
	          JOIN users u ON u.id = d.user_id
	          LEFT JOIN manuscripts m ON m.id = d.manuscript_id
	          WHERE 1=1`
	args := []interface{}{}
	if keyword != "" {
		args = append(args, "%"+keyword+"%")
		query += fmt.Sprintf(" AND d.content ILIKE $%d", len(args))
	}
	args = append(args, size, (page-1)*size)
	rows, err := h.db.QueryContext(r.Context(), query+fmt.Sprintf(" ORDER BY d.id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args...)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	for rows.Next() {
		var id, videoID, manuscriptID, userID int64
		var nickname, avatar, manuscriptTitle, content string
		var time float64
		var mode int
		var createdAt sql.NullTime
		if err := rows.Scan(&id, &videoID, &manuscriptID, &manuscriptTitle, &userID, &nickname, &avatar, &content, &time, &mode, &createdAt); err != nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"id":              id,
			"videoId":         videoID,
			"manuscriptId":    manuscriptID,
			"manuscriptTitle": manuscriptTitle,
			"userId":          userID,
			"userName":        nickname,
			"userAvatar":      avatar,
			"content":         content,
			"time":            time,
			"mode":            mode,
			"createdAt":       createdAt.Time,
		})
	}

	var total int64
	totalArgs := []interface{}{}
	if keyword != "" {
		totalArgs = append(totalArgs, "%"+keyword+"%")
		_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM danmaku WHERE content ILIKE $1`, totalArgs...).Scan(&total)
	} else {
		_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM danmaku`).Scan(&total)
	}
	httputil.WriteOK(w, map[string]interface{}{"list": list, "total": total})
}

func (h *AdminContentHandler) handleDanmakuByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/moderation/admin/danmaku/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id == 0 {
		http.Error(w, "invalid id", 400)
		return
	}
	switch r.Method {
	case http.MethodDelete:
		h.exec(w, r, `DELETE FROM danmaku WHERE id = $1`, id)
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *AdminContentHandler) exec(w http.ResponseWriter, r *http.Request, query string, args ...interface{}) {
	res, err := h.db.ExecContext(r.Context(), query, args...)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "database error")
		return
	}
	n, _ := res.RowsAffected()
	httputil.WriteOK(w, map[string]interface{}{"affected": n})
}
