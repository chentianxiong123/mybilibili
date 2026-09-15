package comment

import (
	"context"
	"database/sql"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"mybilibili/core-service/internal/user"
	"mybilibili/pkg/httputil"
)

type CreatorCommentHTTPHandler struct {
	repo    *CommentRepository
	service *CommentService
	db      *sql.DB
}

func NewCreatorCommentHTTPHandler(repo *CommentRepository, service *CommentService, db *sql.DB) *CreatorCommentHTTPHandler {
	return &CreatorCommentHTTPHandler{repo: repo, service: service, db: db}
}

func (h *CreatorCommentHTTPHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/creator/comments/reply/", h.handleDeleteReply)
	mux.HandleFunc("/api/v1/creator/comments/", h.handleByPath)
	mux.HandleFunc("/api/v1/creator/comments", h.handleList)
}

func (h *CreatorCommentHTTPHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	page, size := httputil.ParsePageParams(r)
	var manuscriptID int64
	if v := r.URL.Query().Get("manuscriptId"); v != "" {
		manuscriptID, _ = strconv.ParseInt(v, 10, 64)
	}
	sortParam := r.URL.Query().Get("sort")
	if sortParam == "" {
		sortParam = "latest"
	}
	commentType := r.URL.Query().Get("commentType")
	if commentType == "" {
		commentType = "all"
	}

	comments, err := h.repo.ListCommentsByCreator(r.Context(), userID, manuscriptID, page, size, sortParam, commentType)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "查询失败", "data": nil})
		return
	}
	var replies []*Reply
	if commentType == "reply" || commentType == "all" {
		replies, _ = h.repo.ListRepliesByCreator(r.Context(), userID, manuscriptID, page, size, sortParam)
	}
	total, _ := h.repo.CountCommentsByCreator(r.Context(), userID, manuscriptID, commentType)
	// all 模式：total 应为 comments + replies 之和，否则前端分页错乱
	if commentType == "all" {
		repliesTotal, _ := h.repo.CountCommentsByCreator(r.Context(), userID, manuscriptID, "reply")
		total += repliesTotal
	}

	// 批量收集 userIds 和 manuscriptIds
	uidSet := map[int64]struct{}{}
	midSet := map[int64]struct{}{}
	for _, c := range comments {
		uidSet[c.UserID] = struct{}{}
		midSet[c.ManuscriptID] = struct{}{}
	}
	for _, rep := range replies {
		uidSet[rep.UserID] = struct{}{}
		if rep.ReplyToUserID.Valid {
			uidSet[rep.ReplyToUserID.Int64] = struct{}{}
		}
		midSet[rep.ManuscriptID] = struct{}{}
	}
	users := h.repo.FindUsersByIDs(r.Context(), toIDList(uidSet))
	manuscripts := h.findManuscriptsByIDs(r.Context(), toIDList(midSet))

	list := make([]map[string]any, 0, len(comments)+len(replies))
	for _, c := range comments {
		u := users[c.UserID]
		m := manuscripts[c.ManuscriptID]
		liked, _ := h.repo.IsCommentLiked(r.Context(), c.ID, userID)
		list = append(list, map[string]any{
			"id":              c.ID,
			"manuscriptId":    c.ManuscriptID,
			"manuscriptTitle":  manuscriptTitleOf(m),
			"manuscriptCover": manuscriptCoverOf(m),
			"userId":          c.UserID,
			"userName":        userNicknameOf(u),
			"userAvatar":       userAvatarOf(u),
			"content":         c.Content,
			"likeCount":       c.LikeCount,
			"replyCount":      c.ReplyCount,
			"status":          c.Status,
			"createTime":      c.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"liked":           liked,
			"commentType":     "comment",
		})
	}
	for _, rep := range replies {
		u := users[rep.UserID]
		m := manuscripts[rep.ManuscriptID]
		liked, _ := h.repo.IsReplyLiked(r.Context(), rep.ID, userID)
		var replyToName string
		if rep.ReplyToUserID.Valid {
			if ru, ok := users[rep.ReplyToUserID.Int64]; ok {
				replyToName = userNicknameOf(ru)
			}
		}
		list = append(list, map[string]any{
			"id":              rep.ID,
			"manuscriptId":    rep.ManuscriptID,
			"manuscriptTitle":  manuscriptTitleOf(m),
			"manuscriptCover": manuscriptCoverOf(m),
			"userId":          rep.UserID,
			"userName":        userNicknameOf(u),
			"userAvatar":       userAvatarOf(u),
			"content":         rep.Content,
			"likeCount":       rep.LikeCount,
			"replyCount":      0,
			"status":          0,
			"createTime":      rep.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"liked":           liked,
			"commentType":     "reply",
			"parentCommentId": rep.CommentID,
			"replyToUserName": replyToName,
		})
	}

	// all 模式下统一排序
	if commentType == "all" {
		sortCreatorItems(list, sortParam)
	}

	httputil.WriteOK(w, map[string]any{"list": list, "total": total, "page": page, "size": size})
}

func (h *CreatorCommentHTTPHandler) findManuscriptsByIDs(ctx context.Context, ids []int64) map[int64]*ManuscriptBrief {
	out := map[int64]*ManuscriptBrief{}
	if len(ids) == 0 {
		return out
	}
	rows, err := h.db.QueryContext(ctx, `SELECT id, title, cover_url FROM manuscripts WHERE id = ANY($1)`, ids)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var title, cover sql.NullString
		rows.Scan(&id, &title, &cover)
		out[id] = &ManuscriptBrief{Title: title.String, Cover: cover.String}
	}
	return out
}

func toIDList(set map[int64]struct{}) []int64 {
	out := make([]int64, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	return out
}

func sortCreatorItems(list []map[string]any, sortKey string) {
	if sortKey == "likes" {
		// 按点赞降序
		sort.Slice(list, func(i, j int) bool {
			return toInt64(list[i]["likeCount"]) > toInt64(list[j]["likeCount"])
		})
		return
	}
	// 默认按时间降序
	sort.Slice(list, func(i, j int) bool {
		return toString(list[i]["createTime"]) > toString(list[j]["createTime"])
	})
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int32:
		return int64(n)
	case int:
		return int64(n)
	}
	return 0
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func userNicknameOf(u *User) string {
	if u == nil {
		return ""
	}
	if u.Nickname != "" {
		return u.Nickname
	}
	return u.Username
}

func userAvatarOf(u *User) string {
	if u == nil {
		return ""
	}
	return u.Avatar
}

type ManuscriptBrief struct {
	Title string
	Cover string
}

func manuscriptTitleOf(m *ManuscriptBrief) string {
	if m == nil {
		return ""
	}
	return m.Title
}

func manuscriptCoverOf(m *ManuscriptBrief) string {
	if m == nil {
		return ""
	}
	return m.Cover
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

func (h *CreatorCommentHTTPHandler) handleByPath(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/creator/comments/")
	parts := strings.Split(path, "/")
	commentID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid comment id", "data": nil})
		return
	}

	switch r.Method {
	case "DELETE":
		n, err := h.repo.DeleteCommentByCreator(r.Context(), commentID, userID)
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "删除失败", "data": nil})
			return
		}
		if n == 0 {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "评论不存在或无权删除", "data": nil})
			return
		}
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	case "POST":
		content := r.URL.Query().Get("content")
		if content == "" {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "回复内容不能为空", "data": nil})
			return
		}
		var replyToUserID int64
		if v := r.URL.Query().Get("replyToUserId"); v != "" {
			replyToUserID, _ = strconv.ParseInt(v, 10, 64)
		}
		// 先校验评论是否存在且属于当前创作者稿件
		exists, _ := h.repo.IsCommentOwnedByCreator(r.Context(), commentID, userID)
		if !exists {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "评论不存在或无权回复", "data": nil})
			return
		}
		rep, err := h.repo.CreateReply(r.Context(), &Reply{
			CommentID: commentID, UserID: userID, ReplyToUserID: toNullInt64(replyToUserID), Content: content,
		})
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "回复失败", "data": nil})
			return
		}
		_ = h.repo.IncrementReplyCount(r.Context(), commentID)
		user.AwardExperience(r.Context(), h.db, userID, 2)
		httputil.WriteOK(w, map[string]any{"id": rep, "commentId": commentID, "content": content, "userId": userID})
	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
	}
}

func (h *CreatorCommentHTTPHandler) handleDeleteReply(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	replyID, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/creator/comments/reply/"), 10, 64)
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid reply id", "data": nil})
		return
	}
	n, err := h.repo.DeleteReplyByCreator(r.Context(), replyID, userID)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "删除失败", "data": nil})
		return
	}
	if n == 0 {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "回复不存在或无权删除", "data": nil})
		return
	}
	httputil.WriteOK(w, map[string]any{"status": "ok"})
}

