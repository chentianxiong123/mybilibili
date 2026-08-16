package comment

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/httputil"
)

type CreatorCommentHTTPHandler struct {
	repo    *CommentRepository
	service *CommentService
}

func NewCreatorCommentHTTPHandler(repo *CommentRepository, service *CommentService) *CreatorCommentHTTPHandler {
	return &CreatorCommentHTTPHandler{repo: repo, service: service}
}

func (h *CreatorCommentHTTPHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/creator/comments/reply/", h.handleDeleteReply)
	mux.HandleFunc("/api/v1/creator/comments/", h.handleByPath)
	mux.HandleFunc("/api/v1/creator/comments", h.handleList)
}

func (h *CreatorCommentHTTPHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		http.Error(w, "unauthorized", 401)
		return
	}
	page, size := httputil.ParsePageParams(r)
	var manuscriptID int64
	if v := r.URL.Query().Get("manuscriptId"); v != "" {
		manuscriptID, _ = strconv.ParseInt(v, 10, 64)
	}
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "latest"
	}
	commentType := r.URL.Query().Get("commentType")
	if commentType == "" {
		commentType = "all"
	}

	comments, err := h.repo.ListCommentsByCreator(r.Context(), userID, manuscriptID, page, size, sort, commentType)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	total, _ := h.repo.CountCommentsByCreator(r.Context(), userID, manuscriptID, commentType)

	list := make([]map[string]any, 0, len(comments))
	for _, c := range comments {
		user, _ := h.repo.FindUserByID(r.Context(), c.UserID)
		list = append(list, map[string]any{
			"id":           c.ID,
			"manuscriptId": c.ManuscriptID,
			"userId":       c.UserID,
			"userName":     userNameOf(user),
			"content":      c.Content,
			"likeCount":    c.LikeCount,
			"replyCount":   c.ReplyCount,
			"status":       c.Status,
			"createdAt":    c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"list": list, "total": total, "page": page, "size": size})
}

func (h *CreatorCommentHTTPHandler) handleByPath(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		http.Error(w, "unauthorized", 401)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/creator/comments/")
	parts := strings.Split(path, "/")
	commentID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid comment id", 400)
		return
	}

	switch r.Method {
	case "DELETE":
		if err := h.repo.DeleteCommentByCreator(r.Context(), commentID, userID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(`{"status":"ok"}`))
	case "POST":
		content := r.URL.Query().Get("content")
		if content == "" {
			http.Error(w, "content required", 400)
			return
		}
		var replyToUserID int64
		if v := r.URL.Query().Get("replyToUserId"); v != "" {
			replyToUserID, _ = strconv.ParseInt(v, 10, 64)
		}
		rep, err := h.repo.CreateReply(r.Context(), &Reply{
			CommentID: commentID, UserID: userID, ReplyToUserID: toNullInt64(replyToUserID), Content: content,
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_ = h.repo.IncrementReplyCount(r.Context(), commentID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"id": rep, "commentId": commentID, "content": content, "userId": userID})
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *CreatorCommentHTTPHandler) handleDeleteReply(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		http.Error(w, "unauthorized", 401)
		return
	}
	replyID, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/creator/comments/reply/"), 10, 64)
	if err != nil {
		http.Error(w, "invalid reply id", 400)
		return
	}
	if err := h.repo.DeleteReplyByCreator(r.Context(), replyID, userID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte(`{"status":"ok"}`))
}

func userNameOf(u *User) string {
	if u == nil {
		return ""
	}
	return u.Username
}

func toNullInt64(v int64) sql.NullInt64 {
	if v == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: v, Valid: true}
}
