package social

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/httputil"
)

type SocialHandler struct {
	followSvc  *FollowService
	dynamicSvc *DynamicService
	collectSvc *CollectionService
	shareRepo  *ShareRepository
	db         *sql.DB
}

func NewSocialHandler(followSvc *FollowService, dynamicSvc *DynamicService, collectSvc *CollectionService, shareRepo *ShareRepository, db *sql.DB) *SocialHandler {
	return &SocialHandler{followSvc: followSvc, dynamicSvc: dynamicSvc, collectSvc: collectSvc, shareRepo: shareRepo, db: db}
}

func (h *SocialHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/dynamic/", h.handleDynamic)
	mux.HandleFunc("/api/v1/dynamic/all", h.handleDynamicAll)
	mux.HandleFunc("/api/v1/dynamic/like/", h.handleDynamicLike)
	mux.HandleFunc("/api/v1/dynamic/comment/list", h.handleDynamicCommentList)
	mux.HandleFunc("/api/v1/dynamic/comment/add", h.handleDynamicCommentAdd)
	mux.HandleFunc("/api/v1/dynamic/comment/replies", h.handleDynamicCommentReplies)
	mux.HandleFunc("/api/v1/dynamic/comment/delete/", h.handleDynamicCommentDelete)
	mux.HandleFunc("/api/v1/dynamic/comment/like/", h.handleDynamicCommentLike)
	mux.HandleFunc("/api/v1/dynamic/share/", h.handleDynamicShare)
	mux.HandleFunc("/api/v1/dynamic/comment/increment/", h.handleDynamicCommentIncrement)
	mux.HandleFunc("/api/v1/collection/", h.handleCollection)
	mux.HandleFunc("/api/v1/share/", h.handleShare)
	mux.HandleFunc("/api/v1/watch-history/", h.handleWatchHistory)
	mux.HandleFunc("/api/v1/watch-history", h.handleWatchHistory)
}

func (h *SocialHandler) handleDynamicCommentList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	dynamicID, _ := strconv.ParseInt(r.URL.Query().Get("dynamicId"), 10, 64)
	page, limit := httputil.ParsePageParams(r)
	list, _ := h.dynamicSvc.ListComments(r.Context(), dynamicID, page, limit)
	httputil.WriteOK(w, h.enrichComments(r.Context(), list))
}

func (h *SocialHandler) handleDynamicCommentAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	dynamicID, _ := strconv.ParseInt(r.URL.Query().Get("dynamicId"), 10, 64)
	content := r.URL.Query().Get("content")
	parentID, _ := strconv.ParseInt(r.URL.Query().Get("parentId"), 10, 64)
	replyUserID, _ := strconv.ParseInt(r.URL.Query().Get("replyUserId"), 10, 64)
	if content == "" {
		var req struct {
			DynamicID   int64  `json:"dynamicId"`
			Content     string `json:"content"`
			ParentID    int64  `json:"parentId"`
			ReplyUserID int64  `json:"replyUserId"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		dynamicID, content, parentID, replyUserID = req.DynamicID, req.Content, req.ParentID, req.ReplyUserID
	}
	dc, err := h.dynamicSvc.AddComment(r.Context(), dynamicID, userID, content, parentID, replyUserID)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	httputil.WriteOK(w, h.enrichComments(r.Context(), []*DynamicComment{dc}))
}

func (h *SocialHandler) handleDynamicCommentReplies(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	commentID, _ := strconv.ParseInt(r.URL.Query().Get("commentId"), 10, 64)
	page, limit := httputil.ParsePageParams(r)
	list, _ := h.dynamicSvc.ListReplies(r.Context(), commentID, page, limit)
	json.NewEncoder(w).Encode(h.enrichComments(r.Context(), list))
}

func (h *SocialHandler) handleDynamicCommentDelete(w http.ResponseWriter, r *http.Request) {
	commentIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/dynamic/comment/delete/")
	commentID, _ := strconv.ParseInt(commentIDStr, 10, 64)
	if r.Method == "DELETE" && commentID > 0 {
		h.dynamicSvc.repo.DeleteComment(r.Context(), commentID)
	}
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *SocialHandler) handleDynamicCommentLike(w http.ResponseWriter, r *http.Request) {
	commentIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/dynamic/comment/like/")
	commentID, _ := strconv.ParseInt(commentIDStr, 10, 64)
	switch r.Method {
	case "POST":
		if commentID > 0 {
			h.dynamicSvc.repo.db.ExecContext(r.Context(),
				`UPDATE dynamic_comments SET like_count = like_count + 1 WHERE id = $1`, commentID)
		}
	case "DELETE":
		if commentID > 0 {
			h.dynamicSvc.repo.db.ExecContext(r.Context(),
				`UPDATE dynamic_comments SET like_count = GREATEST(like_count - 1, 0) WHERE id = $1`, commentID)
		}
	}
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *SocialHandler) handleDynamicAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	page, limit := httputil.ParsePageParams(r)
	list, _ := h.dynamicSvc.ListAll(r.Context(), page, limit)
	json.NewEncoder(w).Encode(h.enrichDynamics(r.Context(), list))
}

func (h *SocialHandler) handleDynamicLike(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/dynamic/like/"), 10, 64)
	userID := httputil.GetUserIDFromHeader(r)
	switch r.Method {
	case http.MethodPost:
		h.dynamicSvc.Like(r.Context(), id, userID)
		w.Write([]byte(`{"status":"ok"}`))
	case http.MethodDelete:
		liked, _ := h.dynamicSvc.IsLiked(r.Context(), id, userID)
		json.NewEncoder(w).Encode(map[string]interface{}{"liked": liked})
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *SocialHandler) handleDynamicShare(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/dynamic/share/"), 10, 64)
	if r.Method == "POST" && id > 0 {
		userID := httputil.GetUserIDFromHeader(r)
		h.dynamicSvc.ShareDynamic(r.Context(), id, userID)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *SocialHandler) handleDynamicCommentIncrement(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/dynamic/comment/increment/"), 10, 64)
	if r.Method == "POST" && id > 0 {
		h.dynamicSvc.IncrCommentCount(r.Context(), id, 1)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *SocialHandler) handleDynamic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/dynamic/")
	userID := httputil.GetUserIDFromHeader(r)
	parts := strings.Split(path, "/")

	switch {
	case len(parts) == 1 && parts[0] == "publish" && r.Method == "POST":
		content := r.URL.Query().Get("content")
		dynType, _ := strconv.ParseInt(r.URL.Query().Get("type"), 10, 32)
		refID, _ := strconv.ParseInt(r.URL.Query().Get("ref_manuscript_id"), 10, 64)
		d, err := h.dynamicSvc.Publish(r.Context(), userID, content, int32(dynType), "", refID)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(d)

	case len(parts) == 1 && parts[0] == "list" && r.Method == "GET":
		page, limit := httputil.ParsePageParams(r)
		list, _ := h.dynamicSvc.ListFollowing(r.Context(), userID, page, limit)
		json.NewEncoder(w).Encode(h.enrichDynamics(r.Context(), list))

	case len(parts) == 1 && parts[0] == "following" && r.Method == "GET":
		page, limit := httputil.ParsePageParams(r)
		list, _ := h.dynamicSvc.ListFollowing(r.Context(), userID, page, limit)
		json.NewEncoder(w).Encode(h.enrichDynamics(r.Context(), list))

	case len(parts) >= 2 && parts[0] == "user" && r.Method == "GET":
		uid, _ := strconv.ParseInt(parts[1], 10, 64)
		page, limit := httputil.ParsePageParams(r)
		list, err := h.dynamicSvc.ListByUser(r.Context(), uid, page, limit)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "database error")
			return
		}
		httputil.WriteOK(w, h.enrichDynamics(r.Context(), list))

	case len(parts) >= 2 && parts[0] == "like":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		if r.Method == "POST" {
			h.dynamicSvc.Like(r.Context(), id, userID)
		} else if r.Method == "DELETE" {
			h.dynamicSvc.Unlike(r.Context(), id, userID)
		}
		w.Write([]byte(`{"status":"ok"}`))

	case len(parts) >= 2 && parts[0] == "comment" && r.Method == "POST":
		dynamicID, _ := strconv.ParseInt(parts[1], 10, 64)
		content := r.URL.Query().Get("content")
		parentID, _ := strconv.ParseInt(r.URL.Query().Get("parent_id"), 10, 64)
		replyID, _ := strconv.ParseInt(r.URL.Query().Get("reply_user_id"), 10, 64)
		dc, err := h.dynamicSvc.AddComment(r.Context(), dynamicID, userID, content, parentID, replyID)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(dc)

	case len(parts) >= 2 && parts[0] == "comment" && r.Method == "GET":
		dynamicID, _ := strconv.ParseInt(parts[1], 10, 64)
		page, limit := httputil.ParsePageParams(r)
		list, _ := h.dynamicSvc.ListComments(r.Context(), dynamicID, page, limit)
		json.NewEncoder(w).Encode(h.enrichComments(r.Context(), list))

	case len(parts) >= 1 && r.Method == "DELETE":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		h.dynamicSvc.Delete(r.Context(), id, userID)
		w.Write([]byte(`{"status":"ok"}`))

	default:
		http.Error(w, "not found", 404)
	}
}

func (h *SocialHandler) handleCollection(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/collection/")
	userID := httputil.GetUserIDFromHeader(r)
	parts := strings.Split(path, "/")

	switch {
	case len(parts) == 1 && parts[0] == "" && r.Method == "POST":
		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			CoverURL    string `json:"cover_url"`
			Status      int32  `json:"status"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Title == "" {
			http.Error(w, "title required", 400)
			return
		}
		c, err := h.collectSvc.Create(r.Context(), userID, req.Title, req.Description, req.CoverURL, req.Status)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(c)

	case len(parts) >= 2 && parts[0] == "user" && r.Method == "GET":
		uid, _ := strconv.ParseInt(parts[1], 10, 64)
		list, _ := h.collectSvc.ListByUser(r.Context(), uid)
		if list == nil {
			list = []map[string]interface{}{}
		}
		httputil.WriteOK(w, list)

	case len(parts) >= 2 && parts[0] != "" && parts[1] == "manuscripts" && r.Method == "GET":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		page, limit := httputil.ParsePageParams(r)
		ids, err := h.collectSvc.ListManuscripts(r.Context(), id, page, limit)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "database error")
			return
		}
		if ids == nil {
			ids = []int64{}
		}
		httputil.WriteOK(w, ids)

	case len(parts) >= 1 && parts[0] != "" && r.Method == "GET" && !strings.Contains(parts[0], "manuscript"):
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		c, _ := h.collectSvc.GetByID(r.Context(), id)
		json.NewEncoder(w).Encode(c)

	case len(parts) >= 1 && parts[0] != "" && r.Method == "PUT":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Status      int32  `json:"status"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.collectSvc.Update(r.Context(), id, userID, req.Title, req.Description, req.Status)
		w.Write([]byte(`{"status":"ok"}`))

	case len(parts) >= 1 && parts[0] != "" && r.Method == "DELETE" && !strings.Contains(parts[0], "manuscript"):
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		h.collectSvc.Delete(r.Context(), id, userID)
		w.Write([]byte(`{"status":"ok"}`))

	case len(parts) >= 3 && parts[0] != "" && parts[1] == "manuscript" && r.Method == "POST":
		collectionID, _ := strconv.ParseInt(parts[0], 10, 64)
		manuscriptID, _ := strconv.ParseInt(parts[2], 10, 64)
		h.collectSvc.AddManuscript(r.Context(), collectionID, manuscriptID, userID)
		w.Write([]byte(`{"status":"ok"}`))

	case len(parts) >= 3 && parts[1] == "manuscript" && r.Method == "DELETE":
		collectionID, _ := strconv.ParseInt(parts[0], 10, 64)
		manuscriptID, _ := strconv.ParseInt(parts[2], 10, 64)
		h.collectSvc.RemoveManuscript(r.Context(), collectionID, manuscriptID, userID)
		w.Write([]byte(`{"status":"ok"}`))

	default:
		http.Error(w, "not found", 404)
	}
}

func (h *SocialHandler) handleShare(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/share/")
	if path == "statistics" && r.Method == "GET" {
		manuscriptID, _ := strconv.ParseInt(r.URL.Query().Get("manuscript_id"), 10, 64)
		if manuscriptID == 0 {
			http.Error(w, "manuscript_id required", 400)
			return
		}
		rows, err := h.shareRepo.db.QueryContext(r.Context(),
			`SELECT channel, COUNT(*) FROM shares WHERE manuscript_id = $1 GROUP BY channel`, manuscriptID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		stats := map[string]int64{}
		for rows.Next() {
			var ch string
			var n int64
			rows.Scan(&ch, &n)
			stats[ch] = n
		}
		json.NewEncoder(w).Encode(stats)
		return
	}

	manuscriptID, _ := strconv.ParseInt(path, 10, 64)
	if r.Method == "POST" && manuscriptID > 0 {
		userID := httputil.GetUserIDFromHeader(r)
		channel := r.URL.Query().Get("channel")
		ip := r.Header.Get("X-Forwarded-For")
		h.shareRepo.Record(r.Context(), userID, manuscriptID, channel, ip)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *SocialHandler) handleWatchHistory(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/watch-history/")
	if path == r.URL.Path {
		path = ""
	}

	switch r.Method {
	case "GET":
		page, limit := httputil.ParsePageParams(r)
		offset := (page - 1) * limit
		rows, err := h.shareRepo.db.QueryContext(r.Context(),
			`SELECT wh.id, wh.manuscript_id, wh.progress_seconds, wh.watched_at,
			        m.title, m.cover_url, u.nickname, u.avatar
			 FROM watch_history wh
			 JOIN manuscripts m ON m.id = wh.manuscript_id
			 JOIN users u ON u.id = m.user_id
			 WHERE wh.user_id = $1
			 ORDER BY wh.watched_at DESC LIMIT $2 OFFSET $3`,
			userID, limit, offset)
		if err != nil {
			httputil.WriteError(w, 500, err.Error())
			return
		}
		defer rows.Close()
		type WHItem struct {
			ID              int64  `json:"id"`
			VideoID         int64  `json:"videoId"`
			ManuscriptID    int64  `json:"manuscriptId"`
			ProgressSeconds int    `json:"progressSeconds"`
			WatchedAt       string `json:"watchedAt"`
			Video           struct {
				Title        string `json:"title"`
				CoverURL     string `json:"coverUrl"`
				ManuscriptID int64  `json:"manuscriptId"`
				Uploader     struct {
					Name   string `json:"name"`
					Avatar string `json:"avatar"`
				} `json:"uploader"`
			} `json:"video"`
		}
		var list []WHItem = []WHItem{}
		for rows.Next() {
			var item WHItem
			rows.Scan(&item.ID, &item.ManuscriptID, &item.ProgressSeconds, &item.WatchedAt,
				&item.Video.Title, &item.Video.CoverURL, &item.Video.Uploader.Name, &item.Video.Uploader.Avatar)
			item.VideoID = item.ManuscriptID
			item.Video.ManuscriptID = item.ManuscriptID
			list = append(list, item)
		}
		httputil.WriteOK(w, list)

	case "POST":
		manuscriptID, _ := strconv.ParseInt(r.URL.Query().Get("manuscript_id"), 10, 64)
		progress, _ := strconv.ParseInt(r.URL.Query().Get("progress_seconds"), 10, 32)
		h.shareRepo.db.ExecContext(r.Context(),
			`INSERT INTO watch_history (user_id, manuscript_id, progress_seconds) VALUES ($1,$2,$3)
			 ON CONFLICT (user_id, manuscript_id) DO UPDATE SET progress_seconds=$3, watched_at=NOW()`,
			userID, manuscriptID, progress)
		w.Write([]byte(`{"status":"ok"}`))

	case "DELETE":
		if path != "" {
			if id, err := strconv.ParseInt(path, 10, 64); err == nil && id > 0 {
				h.shareRepo.db.ExecContext(r.Context(), `DELETE FROM watch_history WHERE id = $1 AND user_id = $2`, id, userID)
				w.Write([]byte(`{"status":"ok"}`))
				return
			}
		}
		h.shareRepo.db.ExecContext(r.Context(), `DELETE FROM watch_history WHERE user_id = $1`, userID)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

// enrichDynamics 为动态列表批量补充 user 嵌套对象（avatar、username）、imageUrls 数组和 refVideo 引用稿件信息。
func (h *SocialHandler) enrichDynamics(ctx context.Context, list []*Dynamic) []map[string]interface{} {
	if len(list) == 0 {
		return []map[string]interface{}{}
	}
	uidSet := map[int64]struct{}{}
	refIds := map[int64]struct{}{}
	for _, d := range list {
		uidSet[d.UserID] = struct{}{}
		if d.RefManuscriptID > 0 {
			refIds[d.RefManuscriptID] = struct{}{}
		}
	}
	ids := make([]int64, 0, len(uidSet))
	for id := range uidSet {
		ids = append(ids, id)
	}
	users := h.loadUsers(ctx, ids)

	// 批量加载引用稿件信息
	refVideos := map[int64]map[string]interface{}{}
	for mid := range refIds {
		refVideos[mid] = h.loadManuscriptBrief(ctx, mid)
	}

	out := make([]map[string]interface{}, 0, len(list))
	for _, d := range list {
		u, ok := users[d.UserID]
		if !ok {
			u = map[string]string{"username": "", "avatar": ""}
		}
		item := map[string]interface{}{
			"id":              d.ID,
			"userId":          d.UserID,
			"content":         d.Content,
			"dynamicType":     d.DynamicType,
			"imageUrl":        d.ImageURL,
			"imageUrls":       []string{},
			"refManuscriptId": d.RefManuscriptID,
			"likeCount":       d.LikeCount,
			"commentCount":    d.CommentCount,
			"shareCount":      d.ShareCount,
			"createdAt":       d.CreatedAt,
			"user": map[string]interface{}{
				"id":       d.UserID,
				"username": u["username"],
				"avatar":   u["avatar"],
			},
		}
		if d.ImageURL != "" {
			item["imageUrls"] = []string{d.ImageURL}
		}
		if d.RefManuscriptID > 0 {
			if rv, ok := refVideos[d.RefManuscriptID]; ok {
				item["refVideo"] = rv
			}
		}
		out = append(out, item)
	}
	return out
}

// loadUsers 批量查询用户信息（username/nickname→name、avatar、level）。
func (h *SocialHandler) loadUsers(ctx context.Context, ids []int64) map[int64]map[string]string {
	users := map[int64]map[string]string{}
	if h.db == nil || len(ids) == 0 {
		return users
	}
	rows, err := h.db.QueryContext(ctx,
		`SELECT id, COALESCE(username,''), COALESCE(nickname,''), COALESCE(avatar,''), COALESCE(level,1) FROM users WHERE id = ANY($1)`, ids)
	if err != nil {
		return users
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var username, nickname, avatar string
		var level int
		rows.Scan(&id, &username, &nickname, &avatar, &level)
		name := nickname
		if name == "" {
			name = username
		}
		users[id] = map[string]string{"username": name, "avatar": avatar, "level": strconv.Itoa(level)}
	}
	return users
}

// loadManuscriptBrief 查询单个稿件简要信息（封面、标题、时长、播放数）。
func (h *SocialHandler) loadManuscriptBrief(ctx context.Context, mid int64) map[string]interface{} {
	if h.db == nil {
		return nil
	}
	var title, cover string
	var durationSecs, viewCount int64
	err := h.db.QueryRowContext(ctx,
		`SELECT COALESCE(title,''), COALESCE(cover_url,''), COALESCE(duration_seconds,0), COALESCE(view_count,0) FROM manuscripts WHERE id = $1`, mid).Scan(&title, &cover, &durationSecs, &viewCount)
	if err != nil {
		return nil
	}
	return map[string]interface{}{
		"id":        mid,
		"title":     title,
		"cover":     cover,
		"viewCount": viewCount,
		"duration":  formatDuration(durationSecs),
	}
}

// formatDuration 将秒数格式化为 mm:ss 或 hh:mm:ss。
func formatDuration(seconds int64) string {
	if seconds <= 0 {
		return ""
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if h > 0 {
		return strconv.FormatInt(h, 10) + ":" + pad2(m) + ":" + pad2(s)
	}
	return pad2(m) + ":" + pad2(s)
}

func pad2(n int64) string {
	if n < 10 {
		return "0" + strconv.FormatInt(n, 10)
	}
	return strconv.FormatInt(n, 10)
}

// enrichComments 为评论列表批量补充用户信息（userAvatar、userName、userLevel）和 replyCount。
func (h *SocialHandler) enrichComments(ctx context.Context, list []*DynamicComment) []map[string]interface{} {
	if len(list) == 0 {
		return []map[string]interface{}{}
	}
	uidSet := map[int64]struct{}{}
	for _, c := range list {
		uidSet[c.UserID] = struct{}{}
		if c.ReplyUserID > 0 {
			uidSet[c.ReplyUserID] = struct{}{}
		}
	}
	ids := make([]int64, 0, len(uidSet))
	for id := range uidSet {
		ids = append(ids, id)
	}
	users := h.loadUsers(ctx, ids)

	out := make([]map[string]interface{}, 0, len(list))
	for _, c := range list {
		u, ok := users[c.UserID]
		if !ok {
			u = map[string]string{"username": "", "avatar": "", "level": "1"}
		}
		item := map[string]interface{}{
			"id":          c.ID,
			"dynamicId":   c.DynamicID,
			"userId":      c.UserID,
			"content":     c.Content,
			"parentId":    c.ParentID,
			"replyUserId": c.ReplyUserID,
			"likeCount":   c.LikeCount,
			"createdAt":   c.CreatedAt,
			"userAvatar":  u["avatar"],
			"userName":    u["username"],
			"userLevel":   0,
			"replyCount":  0,
		}
		if lvl, err := strconv.Atoi(u["level"]); err == nil {
			item["userLevel"] = lvl
		}
		out = append(out, item)
	}
	return out
}
