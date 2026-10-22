package article

import (
	"context"
	"database/sql"
	"encoding/json"
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"mybilibili/pkg/httputil"
)

// Handler 专栏（图文）HTTP 接口。挂在 core 服务 /api/v1/article/* 下，
// 由 traefik 默认 /api/v1/ 路由转发（无需新增路由标签）。
type Handler struct {
	db       *sql.DB
	notifier MessageNotifier
}

// MessageNotifier 写站内信的通知出口。生产环境由 msg-danmaku gRPC 客户端实现，
// 负责落库、失效未读缓存并推送 SSE；测试环境未注入时退回直写 messages 表。
type MessageNotifier interface {
	SendNotification(ctx context.Context, senderID, receiverID int64, content string, msgType int32, targetID, commentID int64)
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) SetNotifier(n MessageNotifier) {
	h.notifier = n
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/article/drafts", h.handleDrafts)
	mux.HandleFunc("/api/v1/article/mine", h.handleMine)
	mux.HandleFunc("/api/v1/article/feed", h.handleFeed)
	mux.HandleFunc("/api/v1/article/save", h.handleSave)
	mux.HandleFunc("/api/v1/article/folders", h.handleFolders)
	mux.HandleFunc("/api/v1/article/mention-search", h.handleMentionSearch)
	mux.HandleFunc("/api/v1/article/comment/", h.handleCommentSub)
	mux.HandleFunc("/api/v1/article/user/", h.handleUserArticles)
	mux.HandleFunc("/api/v1/article/", h.handleByID)
}

// articleRow 列表/详情统一结构（snake_case，前端直接消费）
type articleRow struct {
	ID              int64  `json:"id"`
	UserID          int64  `json:"user_id"`
	Title           string `json:"title"`
	Content         string `json:"content,omitempty"`
	Summary         string `json:"summary"`
	CoverURL        string `json:"cover_url"`
	Topics          string `json:"topics"`
	CollectionName  string `json:"collection_name"`
	Status          int    `json:"status"`
	Visibility      int    `json:"visibility"`
	CommentEnabled  int    `json:"comment_enabled"`
	FeaturedComment int    `json:"featured_comment"`
	IsOriginal      int    `json:"is_original"`
	AIDeclaration   int    `json:"ai_declaration"`
	WordCount       int    `json:"word_count"`
	ViewCount       int64  `json:"view_count"`
	LikeCount       int64  `json:"like_count"`
	FavoriteCount   int64  `json:"favorite_count"`
	CommentCount    int64  `json:"comment_count"`
	CoinCount       int64  `json:"coin_count"`
	ShareCount      int64  `json:"share_count"`
	PublishTime     string `json:"publish_time"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	AuthorNickname  string `json:"author_name"`
	AuthorAvatar    string `json:"author_avatar"`
	// 仅详情接口返回：当前登录用户的互动状态
	MyLiked     bool `json:"my_liked"`
	MyFavorited bool `json:"my_favorited"`
	MyCoin      int  `json:"my_coin"`
}

var tagRe = regexp.MustCompile(`<[^>]+>`)
var spaceRe = regexp.MustCompile(`\s+`)

// htmlToText 去除标签并反转义，用于生成摘要与字数统计。
func htmlToText(s string) string {
	if s == "" {
		return ""
	}
	t := tagRe.ReplaceAllString(s, " ")
	t = html.UnescapeString(t)
	t = spaceRe.ReplaceAllString(t, " ")
	return strings.TrimSpace(t)
}

func runeCount(s string) int {
	return utf8.RuneCountInString(htmlToText(s))
}

func excerpt(s string, n int) string {
	t := htmlToText(s)
	r := []rune(t)
	if len(r) <= n {
		return t
	}
	return string(r[:n])
}

func fmtTime(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format("2006-01-02T15:04:05Z")
}

const articleCols = `a.id, a.user_id, a.title, a.content, a.summary, a.cover_url, a.topics,
	a.collection_name, a.status, a.visibility, a.comment_enabled, a.featured_comment,
	a.is_original, a.ai_declaration, a.word_count, a.view_count, a.like_count,
	a.favorite_count, a.comment_count, a.coin_count, a.share_count, a.scheduled_at, a.publish_time,
	a.created_at, a.updated_at, COALESCE(u.nickname,''), COALESCE(u.avatar,'')`

func scanArticle(rows interface{ Scan(...any) error }, withContent bool) (articleRow, error) {
	var a articleRow
	var content string
	var scheduled sql.NullTime
	var publish, created, updated sql.NullTime
	var err error
	if withContent {
		err = rows.Scan(&a.ID, &a.UserID, &a.Title, &content, &a.Summary, &a.CoverURL, &a.Topics,
			&a.CollectionName, &a.Status, &a.Visibility, &a.CommentEnabled, &a.FeaturedComment,
			&a.IsOriginal, &a.AIDeclaration, &a.WordCount, &a.ViewCount, &a.LikeCount,
			&a.FavoriteCount, &a.CommentCount, &a.CoinCount, &a.ShareCount, &scheduled, &publish,
			&created, &updated, &a.AuthorNickname, &a.AuthorAvatar)
		a.Content = content
	} else {
		err = rows.Scan(&a.ID, &a.UserID, &a.Title, &content, &a.Summary, &a.CoverURL, &a.Topics,
			&a.CollectionName, &a.Status, &a.Visibility, &a.CommentEnabled, &a.FeaturedComment,
			&a.IsOriginal, &a.AIDeclaration, &a.WordCount, &a.ViewCount, &a.LikeCount,
			&a.FavoriteCount, &a.CommentCount, &a.CoinCount, &a.ShareCount, &scheduled, &publish,
			&created, &updated, &a.AuthorNickname, &a.AuthorAvatar)
	}
	a.PublishTime = fmtTime(publish)
	a.CreatedAt = fmtTime(created)
	a.UpdatedAt = fmtTime(updated)
	return a, err
}

// GET /api/v1/article/drafts — 当前用户草稿箱
func (h *Handler) handleDrafts(w http.ResponseWriter, r *http.Request) {
	uid := httputil.GetUserIDFromHeader(r)
	if uid == 0 {
		httputil.WriteJSON(w, http.StatusOK, map[string]any{"code": 200, "message": "ok", "data": []any{}})
		return
	}
	kw := strings.TrimSpace(r.URL.Query().Get("keyword"))
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT `+articleCols+` FROM articles a LEFT JOIN users u ON u.id = a.user_id
		 WHERE a.user_id = $1 AND a.status = 0 AND ($2 = '' OR a.title ILIKE '%'||$2||'%')
		 ORDER BY a.updated_at DESC`, uid, kw)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "查询失败", "data": nil})
		return
	}
	defer rows.Close()
	list := h.collect(rows, false)
	httputil.WriteOK(w, list)
}

// GET /api/v1/article/mine?status=0|1|all — 当前用户稿件（草稿/已发布）
func (h *Handler) handleMine(w http.ResponseWriter, r *http.Request) {
	uid := httputil.GetUserIDFromHeader(r)
	if uid == 0 {
		httputil.WriteJSON(w, http.StatusOK, map[string]any{"code": 200, "message": "ok", "data": []any{}})
		return
	}
	status := r.URL.Query().Get("status")
	clause := "a.status >= 0"
	args := []any{uid}
	if status == "0" {
		clause = "a.status = 0"
	} else if status == "1" {
		clause = "a.status = 1"
	}
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT `+articleCols+` FROM articles a LEFT JOIN users u ON u.id = a.user_id
		 WHERE a.user_id = $1 AND `+clause+` ORDER BY a.updated_at DESC`, args...)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "查询失败", "data": nil})
		return
	}
	defer rows.Close()
	httputil.WriteOK(w, h.collect(rows, false))
}

// GET /api/v1/article/user/{uid}?page&size — 某用户已发布且公开的专栏
func (h *Handler) handleUserArticles(w http.ResponseWriter, r *http.Request) {
	uidStr := strings.TrimPrefix(r.URL.Path, "/api/v1/article/user/")
	uid, err := strconv.ParseInt(strings.Trim(uidStr, "/"), 10, 64)
	if err != nil || uid <= 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid uid", "data": nil})
		return
	}
	viewer := httputil.GetUserIDFromHeader(r)
	page, size := httputil.ParsePageParams(r)
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT `+articleCols+` FROM articles a LEFT JOIN users u ON u.id = a.user_id
		 WHERE a.user_id = $1 AND a.status = 1 AND (a.visibility = 1 OR $2 = a.user_id)
		 ORDER BY a.publish_time DESC NULLS LAST, a.id DESC LIMIT $3 OFFSET $4`,
		uid, viewer, size, (page-1)*size)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "查询失败", "data": nil})
		return
	}
	defer rows.Close()
	httputil.WriteOK(w, map[string]any{"list": h.collect(rows, false), "page": page, "size": size})
}

// GET /api/v1/article/feed?page&size — 专栏首页信息流（已发布且公开）
func (h *Handler) handleFeed(w http.ResponseWriter, r *http.Request) {
	page, size := httputil.ParsePageParams(r)
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT `+articleCols+` FROM articles a LEFT JOIN users u ON u.id = a.user_id
		 WHERE a.status = 1 AND a.visibility = 1
		 ORDER BY a.publish_time DESC NULLS LAST, a.id DESC LIMIT $1 OFFSET $2`,
		size, (page-1)*size)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "查询失败", "data": nil})
		return
	}
	defer rows.Close()
	httputil.WriteOK(w, map[string]any{"list": h.collect(rows, false), "page": page, "size": size})
}

func (h *Handler) collect(rows *sql.Rows, withContent bool) []articleRow {
	list := []articleRow{}
	for rows.Next() {
		a, err := scanArticle(rows, withContent)
		if err != nil {
			continue
		}
		list = append(list, a)
	}
	return list
}

// saveReq 保存请求体。publish=true 走发布（status=1），否则存草稿（status=0）。
type saveReq struct {
	ID              int64  `json:"id"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	CoverURL        string `json:"cover_url"`
	Topics          string `json:"topics"`
	CollectionName  string `json:"collection_name"`
	Visibility      *int   `json:"visibility"`
	CommentEnabled  *int   `json:"comment_enabled"`
	FeaturedComment *int   `json:"featured_comment"`
	IsOriginal      *int   `json:"is_original"`
	AIDeclaration   *int   `json:"ai_declaration"`
	Publish         bool   `json:"publish"`
	WordCount       *int   `json:"word_count"`
}

// POST /api/v1/article/save — 新建/更新草稿，或发布专栏
func (h *Handler) handleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	var req saveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid body", "data": nil})
		return
	}
	title := strings.TrimSpace(req.Title)
	text := htmlToText(req.Content)
	if req.Publish && title == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "标题不能为空", "data": nil})
		return
	}
	if req.Publish && text == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "正文不能为空", "data": nil})
		return
	}
	visibility := 1
	if req.Visibility != nil && (*req.Visibility == 0 || *req.Visibility == 1) {
		visibility = *req.Visibility
	}
	commentEnabled := 1
	if req.CommentEnabled != nil {
		commentEnabled = boolInt(*req.CommentEnabled != 0)
	}
	featured := 0
	if req.FeaturedComment != nil && *req.FeaturedComment != 0 {
		featured = 1
	}
	isOriginal := 0
	if req.IsOriginal != nil && *req.IsOriginal != 0 {
		isOriginal = 1
	}
	aiDecl := 0
	if req.AIDeclaration != nil && *req.AIDeclaration != 0 {
		aiDecl = 1
	}
	summary := excerpt(req.Content, 120)
	wordCount := runeCount(req.Content)
	if req.WordCount != nil && *req.WordCount > 0 {
		wordCount = *req.WordCount
	}

	if req.ID > 0 {
		// 校验归属
		var owner int64
		if err := h.db.QueryRowContext(r.Context(), `SELECT user_id FROM articles WHERE id = $1`, req.ID).Scan(&owner); err != nil {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "专栏不存在", "data": nil})
			return
		}
		if owner != uid {
			httputil.WriteJSON(w, http.StatusForbidden, map[string]any{"code": 403, "message": "无权操作", "data": nil})
			return
		}
		setPublish := ""
		if req.Publish {
			setPublish = ", status = 1, publish_time = COALESCE(publish_time, NOW())"
		}
		_, err := h.db.ExecContext(r.Context(),
			`UPDATE articles SET title=$1, content=$2, summary=$3, cover_url=$4, topics=$5,
				collection_name=$6, visibility=$7, comment_enabled=$8, featured_comment=$9,
				is_original=$10, ai_declaration=$11, word_count=$12`+setPublish+`, updated_at=NOW()
			 WHERE id=$13 AND user_id=$14`,
			title, req.Content, summary, req.CoverURL, req.Topics, req.CollectionName,
			visibility, commentEnabled, featured, isOriginal, aiDecl, wordCount, req.ID, uid)
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "保存失败", "data": nil})
			return
		}
		httputil.WriteOK(w, map[string]any{"id": req.ID, "published": req.Publish})
		return
	}

	status := 0
	var publishTime any = nil
	if req.Publish {
		status = 1
		publishTime = time.Now()
	}
	var id int64
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO articles (user_id, title, content, summary, cover_url, topics, collection_name,
			status, visibility, comment_enabled, featured_comment, is_original, ai_declaration,
			word_count, publish_time)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING id`,
		uid, title, req.Content, summary, req.CoverURL, req.Topics, req.CollectionName,
		status, visibility, commentEnabled, featured, isOriginal, aiDecl, wordCount, publishTime).Scan(&id)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "创建失败", "data": nil})
		return
	}
	httputil.WriteOK(w, map[string]any{"id": id, "published": req.Publish})
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// GET/DELETE /api/v1/article/{id}  及 /{id}/view
func (h *Handler) handleByID(w http.ResponseWriter, r *http.Request) {
	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/article/"), "/")
	if rest == "" {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
		return
	}
	parts := strings.Split(rest, "/")
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid id", "data": nil})
		return
	}

	if len(parts) >= 2 {
		switch parts[1] {
		case "view":
			h.db.ExecContext(r.Context(), `UPDATE articles SET view_count = view_count + 1 WHERE id = $1 AND status = 1`, id)
			httputil.WriteOK(w, map[string]any{"status": "ok"})
			return
		case "like":
			h.toggleInteraction(w, r, id, "like", "like_count")
			return
		case "favorite":
			h.toggleInteraction(w, r, id, "favorite", "favorite_count")
			return
		case "collect":
			h.collectFolders(w, r, id)
			return
		case "coin":
			h.addCoin(w, r, id)
			return
		case "share":
			h.addShare(w, r, id)
			return
		case "interaction":
			h.getInteraction(w, r, id)
			return
		case "comments":
			h.handleComments(w, r, id)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		viewer := httputil.GetUserIDFromHeader(r)
		row := h.db.QueryRowContext(r.Context(),
			`SELECT `+articleCols+` FROM articles a LEFT JOIN users u ON u.id = a.user_id WHERE a.id = $1`, id)
		a, err := scanArticle(row, true)
		if err != nil {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "专栏不存在", "data": nil})
			return
		}
		isOwner := viewer != 0 && viewer == a.UserID
		if !isOwner && (a.Status != 1 || a.Visibility != 1) {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "专栏不存在或无权查看", "data": nil})
			return
		}
		if viewer != 0 {
			_ = h.db.QueryRowContext(r.Context(),
				`SELECT
					EXISTS(SELECT 1 FROM user_interactions WHERE user_id=$1 AND target_type='article' AND target_id=$2 AND interaction_type='like'),
					EXISTS(SELECT 1 FROM user_interactions WHERE user_id=$1 AND target_type='article' AND target_id=$2 AND interaction_type='favorite')`,
				viewer, id).Scan(&a.MyLiked, &a.MyFavorited)
			_ = h.db.QueryRowContext(r.Context(),
				`SELECT COALESCE(SUM(count),0) FROM article_coins WHERE article_id=$1 AND user_id=$2`, id, viewer).Scan(&a.MyCoin)
		}
		httputil.WriteOK(w, a)
	case http.MethodDelete:
		uid, ok := httputil.RequireUser(w, r)
		if !ok {
			return
		}
		res, err := h.db.ExecContext(r.Context(),
			`UPDATE articles SET status = -1, updated_at = NOW() WHERE id = $1 AND user_id = $2`, id, uid)
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "删除失败", "data": nil})
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "专栏不存在或无权操作", "data": nil})
			return
		}
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
	}
}

// toggleInteraction 切换点赞/收藏（target_type='article'），并维护 articles 计数列。
func (h *Handler) toggleInteraction(w http.ResponseWriter, r *http.Request, id int64, itype, counterCol string) {
	if r.Method != http.MethodPost {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	var exists bool
	_ = h.db.QueryRowContext(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM user_interactions WHERE user_id=$1 AND target_type='article' AND target_id=$2 AND interaction_type=$3)`,
		uid, id, itype).Scan(&exists)
	if exists {
		h.db.ExecContext(r.Context(),
			`DELETE FROM user_interactions WHERE user_id=$1 AND target_type='article' AND target_id=$2 AND interaction_type=$3`,
			uid, id, itype)
		h.db.ExecContext(r.Context(),
			`UPDATE articles SET `+counterCol+` = GREATEST(`+counterCol+` - 1, 0) WHERE id=$1`, id)
	} else {
		h.db.ExecContext(r.Context(),
			`INSERT INTO user_interactions (user_id, target_type, target_id, interaction_type)
			 VALUES ($1, 'article', $2, $3) ON CONFLICT DO NOTHING`, uid, id, itype)
		h.db.ExecContext(r.Context(),
			`UPDATE articles SET `+counterCol+` = `+counterCol+` + 1 WHERE id=$1`, id)
		if itype == "like" {
			h.notifyArticleLike(r, uid, id)
		}
	}
	var count int64
	h.db.QueryRowContext(r.Context(), `SELECT `+counterCol+` FROM articles WHERE id=$1`, id).Scan(&count)
	httputil.WriteOK(w, map[string]any{"active": !exists, "count": count})
}

// addCoin 投币（单用户每篇最多 2 枚）。
func (h *Handler) addCoin(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != http.MethodPost {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	var body struct {
		Count int `json:"count"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	add := body.Count
	if add <= 0 {
		add = 1
	}
	if add > 2 {
		add = 2
	}
	var cur int
	_ = h.db.QueryRowContext(r.Context(),
		`SELECT COALESCE(count,0) FROM article_coins WHERE article_id=$1 AND user_id=$2`, id, uid).Scan(&cur)
	if cur >= 2 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "最多投 2 个币", "data": map[string]any{"my_coin": cur}})
		return
	}
	if cur+add > 2 {
		add = 2 - cur
	}
	h.db.ExecContext(r.Context(),
		`INSERT INTO article_coins (article_id, user_id, count) VALUES ($1,$2,$3)
		 ON CONFLICT (article_id, user_id) DO UPDATE SET count = article_coins.count + EXCLUDED.count, updated_at = NOW()`,
		id, uid, add)
	h.db.ExecContext(r.Context(), `UPDATE articles SET coin_count = coin_count + $1 WHERE id=$2`, add, id)
	var total int64
	h.db.QueryRowContext(r.Context(), `SELECT coin_count FROM articles WHERE id=$1`, id).Scan(&total)
	httputil.WriteOK(w, map[string]any{"my_coin": cur + add, "count": total})
}

// addShare 分享计数 +1。
func (h *Handler) addShare(w http.ResponseWriter, r *http.Request, id int64) {
	h.db.ExecContext(r.Context(), `UPDATE articles SET share_count = share_count + 1 WHERE id=$1`, id)
	var count int64
	h.db.QueryRowContext(r.Context(), `SELECT share_count FROM articles WHERE id=$1`, id).Scan(&count)
	httputil.WriteOK(w, map[string]any{"count": count})
}

// getInteraction 返回当前用户互动状态与文章各计数。
func (h *Handler) getInteraction(w http.ResponseWriter, r *http.Request, id int64) {
	viewer := httputil.GetUserIDFromHeader(r)
	out := map[string]any{"liked": false, "favorited": false, "coin": 0}
	var like, fav, coin, comment, share int64
	if err := h.db.QueryRowContext(r.Context(),
		`SELECT like_count, favorite_count, coin_count, comment_count, share_count FROM articles WHERE id=$1`, id).
		Scan(&like, &fav, &coin, &comment, &share); err != nil {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "专栏不存在", "data": nil})
		return
	}
	out["like_count"] = like
	out["favorite_count"] = fav
	out["coin_count"] = coin
	out["comment_count"] = comment
	out["share_count"] = share
	if viewer != 0 {
		var liked, favorited bool
		var myCoin int
		_ = h.db.QueryRowContext(r.Context(),
			`SELECT
				EXISTS(SELECT 1 FROM user_interactions WHERE user_id=$1 AND target_type='article' AND target_id=$2 AND interaction_type='like'),
				EXISTS(SELECT 1 FROM user_interactions WHERE user_id=$1 AND target_type='article' AND target_id=$2 AND interaction_type='favorite')`,
			viewer, id).Scan(&liked, &favorited)
		_ = h.db.QueryRowContext(r.Context(),
			`SELECT COALESCE(SUM(count),0) FROM article_coins WHERE article_id=$1 AND user_id=$2`, id, viewer).Scan(&myCoin)
		out["liked"] = liked
		out["favorited"] = favorited
		out["coin"] = myCoin
	}
	httputil.WriteOK(w, out)
}

type commentNode struct {
	ID            int64         `json:"id"`
	ArticleID     int64         `json:"article_id"`
	UserID        int64         `json:"user_id"`
	ParentID      int64         `json:"parent_id"`
	ReplyToUserID int64         `json:"reply_to_user_id"`
	Content       string        `json:"content"`
	LikeCount     int           `json:"like_count"`
	CreatedAt     string        `json:"created_at"`
	AuthorName    string        `json:"author_name"`
	AuthorAvatar  string        `json:"author_avatar"`
	ReplyToName   string        `json:"reply_to_name"`
	MyLiked       bool          `json:"my_liked"`
	Replies       []commentNode `json:"replies"`
}

// GET /api/v1/article/{id}/comments?page&size  评论（含一级回复）
// POST /api/v1/article/{id}/comments           发表评论/回复
func (h *Handler) handleComments(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		page, size := httputil.ParsePageParams(r)
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT c.id, c.article_id, c.user_id, COALESCE(c.parent_id,0), c.content, c.like_count, c.created_at,
				COALESCE(u.nickname,''), COALESCE(u.avatar,'')
			 FROM article_comments c LEFT JOIN users u ON u.id=c.user_id
			 WHERE c.article_id=$1 AND c.status=0 AND c.parent_id IS NULL
			 ORDER BY c.created_at DESC LIMIT $2 OFFSET $3`, id, size, (page-1)*size)
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "查询失败", "data": nil})
			return
		}
		roots := []commentNode{}
		ids := []int64{}
		for rows.Next() {
			var c commentNode
			var created sql.NullTime
			if err := rows.Scan(&c.ID, &c.ArticleID, &c.UserID, &c.ParentID, &c.Content, &c.LikeCount, &created, &c.AuthorName, &c.AuthorAvatar); err != nil {
				continue
			}
			c.CreatedAt = fmtTime(created)
			c.Replies = []commentNode{}
			roots = append(roots, c)
			ids = append(ids, c.ID)
		}
		rows.Close()
		if len(ids) > 0 {
			replyRows, err := h.db.QueryContext(r.Context(),
				`SELECT c.id, c.article_id, c.user_id, COALESCE(c.parent_id,0), COALESCE(c.reply_to_user_id,0),
					c.content, c.like_count, c.created_at, COALESCE(u.nickname,''), COALESCE(u.avatar,''), COALESCE(ru.nickname,'')
				 FROM article_comments c
				 LEFT JOIN users u ON u.id=c.user_id
				 LEFT JOIN users ru ON ru.id=c.reply_to_user_id
				 WHERE c.status=0 AND c.parent_id = ANY($1) ORDER BY c.created_at ASC`, ids)
			if err == nil {
				byParent := map[int64]int{}
				for i := range roots {
					byParent[roots[i].ID] = i
				}
				for replyRows.Next() {
					var c commentNode
					var created sql.NullTime
					if err := replyRows.Scan(&c.ID, &c.ArticleID, &c.UserID, &c.ParentID, &c.ReplyToUserID, &c.Content, &c.LikeCount, &created, &c.AuthorName, &c.AuthorAvatar, &c.ReplyToName); err != nil {
						continue
					}
					c.CreatedAt = fmtTime(created)
					c.Replies = []commentNode{}
					if idx, ok := byParent[c.ParentID]; ok {
						roots[idx].Replies = append(roots[idx].Replies, c)
					}
				}
				replyRows.Close()
			}
		}
		if viewer := httputil.GetUserIDFromHeader(r); viewer != 0 {
			liked := map[int64]bool{}
			if lr, err := h.db.QueryContext(r.Context(),
				`SELECT l.comment_id FROM article_comment_likes l
				 JOIN article_comments c ON c.id = l.comment_id
				 WHERE l.user_id=$1 AND c.article_id=$2`, viewer, id); err == nil {
				for lr.Next() {
					var cid int64
					lr.Scan(&cid)
					liked[cid] = true
				}
				lr.Close()
			}
			for i := range roots {
				if liked[roots[i].ID] {
					roots[i].MyLiked = true
				}
				for j := range roots[i].Replies {
					if liked[roots[i].Replies[j].ID] {
						roots[i].Replies[j].MyLiked = true
					}
				}
			}
		}
		var total int
		_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM article_comments WHERE article_id=$1 AND status=0`, id).Scan(&total)
		httputil.WriteOK(w, map[string]any{"list": roots, "total": total, "page": page, "size": size})

	case http.MethodPost:
		uid, ok := httputil.RequireUser(w, r)
		if !ok {
			return
		}
		var body struct {
			Content       string `json:"content"`
			ParentID      int64  `json:"parent_id"`
			ReplyToUserID int64  `json:"reply_to_user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid body", "data": nil})
			return
		}
		content := strings.TrimSpace(body.Content)
		if content == "" {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "评论内容不能为空", "data": nil})
			return
		}
		if len([]rune(content)) > 1000 {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "评论过长", "data": nil})
			return
		}
		var enabled int
		if err := h.db.QueryRowContext(r.Context(),
			`SELECT comment_enabled FROM articles WHERE id=$1 AND status=1`, id).Scan(&enabled); err != nil {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "专栏不存在", "data": nil})
			return
		}
		if enabled == 0 {
			httputil.WriteJSON(w, http.StatusForbidden, map[string]any{"code": 403, "message": "该专栏已关闭评论", "data": nil})
			return
		}
		var parentID any
		replyTo := body.ReplyToUserID
		if body.ParentID > 0 {
			var owner int64
			if err := h.db.QueryRowContext(r.Context(),
				`SELECT user_id FROM article_comments WHERE id=$1 AND article_id=$2 AND status=0`, body.ParentID, id).Scan(&owner); err != nil {
				httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "回复的评论不存在", "data": nil})
				return
			}
			parentID = body.ParentID
			if replyTo == 0 {
				replyTo = owner
			}
		} else {
			parentID = nil
		}
		var replyToArg any
		if replyTo > 0 {
			replyToArg = replyTo
		} else {
			replyToArg = nil
		}
		var newID int64
		var created time.Time
		err := h.db.QueryRowContext(r.Context(),
			`INSERT INTO article_comments (article_id, user_id, parent_id, reply_to_user_id, content)
			 VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
			id, uid, parentID, replyToArg, content).Scan(&newID, &created)
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "评论失败", "data": nil})
			return
		}
		h.db.ExecContext(r.Context(), `UPDATE articles SET comment_count = comment_count + 1 WHERE id=$1`, id)
		h.notifyArticleMentions(r, content, uid, id, newID)
		h.notifyArticleReply(r, content, uid, replyTo, id, newID)
		var nickname, avatar, replyName string
		_ = h.db.QueryRowContext(r.Context(), `SELECT COALESCE(nickname,''), COALESCE(avatar,'') FROM users WHERE id=$1`, uid).Scan(&nickname, &avatar)
		if replyTo > 0 {
			_ = h.db.QueryRowContext(r.Context(), `SELECT COALESCE(nickname,'') FROM users WHERE id=$1`, replyTo).Scan(&replyName)
		}
		var pid int64
		if body.ParentID > 0 {
			pid = body.ParentID
		}
		httputil.WriteOK(w, commentNode{
			ID: newID, ArticleID: id, UserID: uid, ParentID: pid, ReplyToUserID: replyTo,
			Content: content, CreatedAt: created.Format("2006-01-02T15:04:05Z"),
			AuthorName: nickname, AuthorAvatar: avatar, ReplyToName: replyName, Replies: []commentNode{},
		})

	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
	}
}

// DELETE /api/v1/article/comment/{cid}        删除自己的评论
// POST   /api/v1/article/comment/{cid}/like   评论点赞/取消
func (h *Handler) handleCommentSub(w http.ResponseWriter, r *http.Request) {
	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/article/comment/"), "/")
	parts := strings.Split(rest, "/")
	cid, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || cid <= 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid id", "data": nil})
		return
	}

	// 评论点赞
	if len(parts) >= 2 && parts[1] == "like" {
		if r.Method != http.MethodPost {
			httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
			return
		}
		uid, ok := httputil.RequireUser(w, r)
		if !ok {
			return
		}
		var exists bool
		_ = h.db.QueryRowContext(r.Context(),
			`SELECT EXISTS(SELECT 1 FROM article_comment_likes WHERE comment_id=$1 AND user_id=$2)`, cid, uid).Scan(&exists)
		if exists {
			h.db.ExecContext(r.Context(), `DELETE FROM article_comment_likes WHERE comment_id=$1 AND user_id=$2`, cid, uid)
			h.db.ExecContext(r.Context(), `UPDATE article_comments SET like_count = GREATEST(like_count - 1, 0) WHERE id=$1`, cid)
		} else {
			h.db.ExecContext(r.Context(),
				`INSERT INTO article_comment_likes (comment_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, cid, uid)
			h.db.ExecContext(r.Context(), `UPDATE article_comments SET like_count = like_count + 1 WHERE id=$1`, cid)
			h.notifyArticleCommentLike(r, uid, cid)
		}
		var count int
		h.db.QueryRowContext(r.Context(), `SELECT like_count FROM article_comments WHERE id=$1`, cid).Scan(&count)
		httputil.WriteOK(w, map[string]any{"active": !exists, "count": count})
		return
	}

	// 删除评论
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	var owner, articleID int64
	if err := h.db.QueryRowContext(r.Context(),
		`SELECT user_id, article_id FROM article_comments WHERE id=$1`, cid).Scan(&owner, &articleID); err != nil {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "评论不存在", "data": nil})
		return
	}
	if owner != uid {
		httputil.WriteJSON(w, http.StatusForbidden, map[string]any{"code": 403, "message": "无权删除", "data": nil})
		return
	}
	h.db.ExecContext(r.Context(), `UPDATE article_comments SET status=1, updated_at=NOW() WHERE id=$1`, cid)
	h.db.ExecContext(r.Context(),
		`UPDATE articles SET comment_count = (SELECT COUNT(*) FROM article_comments WHERE article_id=$1 AND status=0) WHERE id=$1`, articleID)
	var count int64
	_ = h.db.QueryRowContext(r.Context(), `SELECT comment_count FROM articles WHERE id=$1`, articleID).Scan(&count)
	httputil.WriteOK(w, map[string]any{"status": "ok", "count": count})
}

// GET /api/v1/article/folders?article_id=xx — 当前用户收藏夹（含是否已收录该专栏）
// POST /api/v1/article/folders {name}        — 新建收藏夹
func (h *Handler) handleFolders(w http.ResponseWriter, r *http.Request) {
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		articleID, _ := strconv.ParseInt(r.URL.Query().Get("article_id"), 10, 64)
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT f.id, f.name, f.visible,
			        (SELECT COUNT(*) FROM favorite_folder_articles x WHERE x.folder_id=f.id) AS cnt,
			        EXISTS(SELECT 1 FROM favorite_folder_articles y WHERE y.folder_id=f.id AND y.article_id=$2) AS checked
			 FROM favorite_folders f WHERE f.user_id=$1 ORDER BY f.created_at ASC`, uid, articleID)
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "查询失败", "data": nil})
			return
		}
		defer rows.Close()
		type folder struct {
			ID      int64  `json:"id"`
			Name    string `json:"name"`
			Visible int    `json:"visible"`
			Count   int64  `json:"count"`
			Checked bool   `json:"checked"`
		}
		list := []folder{}
		for rows.Next() {
			var f folder
			if err := rows.Scan(&f.ID, &f.Name, &f.Visible, &f.Count, &f.Checked); err != nil {
				continue
			}
			list = append(list, f)
		}
		httputil.WriteOK(w, list)
	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "请输入收藏夹名称", "data": nil})
			return
		}
		name := strings.TrimSpace(req.Name)
		if len([]rune(name)) > 20 {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "名称最多20个字", "data": nil})
			return
		}
		var id int64
		err := h.db.QueryRowContext(r.Context(),
			`INSERT INTO favorite_folders (user_id, name, visible) VALUES ($1,$2,1) RETURNING id`, uid, name).Scan(&id)
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "创建失败（可能已存在同名收藏夹）", "data": nil})
			return
		}
		httputil.WriteOK(w, map[string]any{"id": id, "name": name, "visible": 1, "count": 0, "checked": false})
	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
	}
}

// POST /api/v1/article/{id}/collect {adds:[fid], removes:[fid]} — 收藏到指定收藏夹
func (h *Handler) collectFolders(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != http.MethodPost {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	var body struct {
		Adds    []int64 `json:"adds"`
		Removes []int64 `json:"removes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid body", "data": nil})
		return
	}
	var status int
	if err := h.db.QueryRowContext(r.Context(), `SELECT status FROM articles WHERE id=$1`, id).Scan(&status); err != nil || status != 1 {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "专栏不存在", "data": nil})
		return
	}
	for _, fid := range body.Adds {
		var cnt int
		h.db.QueryRowContext(r.Context(),
			`SELECT COUNT(*) FROM favorite_folders WHERE id=$1 AND user_id=$2`, fid, uid).Scan(&cnt)
		if cnt == 0 {
			continue
		}
		h.db.ExecContext(r.Context(),
			`INSERT INTO favorite_folder_articles (folder_id, article_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, fid, id)
	}
	for _, fid := range body.Removes {
		h.db.ExecContext(r.Context(),
			`DELETE FROM favorite_folder_articles WHERE folder_id=$1 AND article_id=$2
			 AND folder_id IN (SELECT id FROM favorite_folders WHERE user_id=$3)`, fid, id, uid)
	}
	// 该用户是否仍收藏着本专栏 → 维护 article.favorite_count 与 user_interactions
	var mine int
	h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM favorite_folder_articles ffa
		 JOIN favorite_folders f ON f.id = ffa.folder_id
		 WHERE ffa.article_id=$1 AND f.user_id=$2`, id, uid).Scan(&mine)
	var hadLike bool
	_ = h.db.QueryRowContext(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM user_interactions WHERE user_id=$1 AND target_type='article' AND target_id=$2 AND interaction_type='favorite')`,
		uid, id).Scan(&hadLike)
	has := mine > 0
	if has && !hadLike {
		h.db.ExecContext(r.Context(),
			`INSERT INTO user_interactions (user_id, target_type, target_id, interaction_type)
			 VALUES ($1,'article',$2,'favorite') ON CONFLICT DO NOTHING`, uid, id)
		h.db.ExecContext(r.Context(), `UPDATE articles SET favorite_count = favorite_count + 1 WHERE id=$1`, id)
	} else if !has && hadLike {
		h.db.ExecContext(r.Context(),
			`DELETE FROM user_interactions WHERE user_id=$1 AND target_type='article' AND target_id=$2 AND interaction_type='favorite'`, uid, id)
		h.db.ExecContext(r.Context(), `UPDATE articles SET favorite_count = GREATEST(favorite_count - 1, 0) WHERE id=$1`, id)
	}
	var total int64
	h.db.QueryRowContext(r.Context(), `SELECT favorite_count FROM articles WHERE id=$1`, id).Scan(&total)
	httputil.WriteOK(w, map[string]any{"favorited": has, "count": total, "folders": mine})
}

// GET /api/v1/article/mention-search?keyword=xx — @ 提及用户搜索
func (h *Handler) handleMentionSearch(w http.ResponseWriter, r *http.Request) {
	kw := strings.TrimSpace(r.URL.Query().Get("keyword"))
	if kw == "" {
		httputil.WriteOK(w, []any{})
		return
	}
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, COALESCE(nickname,''), COALESCE(avatar,'') FROM users
		 WHERE (nickname ILIKE '%'||$1||'%' OR username ILIKE '%'||$1||'%')
		 ORDER BY id ASC LIMIT 10`, kw)
	if err != nil {
		httputil.WriteOK(w, []any{})
		return
	}
	defer rows.Close()
	type u struct {
		UID      int64  `json:"uid"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}
	list := []u{}
	for rows.Next() {
		var it u
		if err := rows.Scan(&it.UID, &it.Nickname, &it.Avatar); err != nil {
			continue
		}
		list = append(list, it)
	}
	httputil.WriteOK(w, list)
}

var mentionRe = regexp.MustCompile(`@([^\s@]+)`)

// notifyArticleReply 回复评论时给被回复人写 messages(type=2)。
// 与视频评论的 sendReplyNotification 对齐：自己回自己不通知。
func (h *Handler) notifyArticleReply(r *http.Request, content string, senderID, receiverID, articleID, commentID int64) {
	if receiverID <= 0 || receiverID == senderID {
		return
	}
	text := "回复了你的评论：" + excerpt(content, 140)
	if h.notifier != nil {
		h.notifier.SendNotification(r.Context(), senderID, receiverID, text, 2, articleID, commentID)
		return
	}
	h.db.ExecContext(r.Context(),
		`INSERT INTO messages (sender_id, receiver_id, content, message_type, target_id, comment_id, is_read, created_at)
		 VALUES ($1,$2,$3,2,$4,$5,0,NOW())`,
		senderID, receiverID, text, articleID, commentID)
}

// notifyArticleLike 点赞专栏时给作者写 messages(type=4)，自己赞自己不通知。
func (h *Handler) notifyArticleLike(r *http.Request, senderID, articleID int64) {
	var owner int64
	var title string
	if err := h.db.QueryRowContext(r.Context(),
		`SELECT user_id, COALESCE(title,'') FROM articles WHERE id=$1`, articleID).Scan(&owner, &title); err != nil {
		return
	}
	if owner <= 0 || owner == senderID {
		return
	}
	text := "赞了你的专栏《" + title + "》"
	if h.notifier != nil {
		h.notifier.SendNotification(r.Context(), senderID, owner, text, 4, articleID, 0)
		return
	}
	h.db.ExecContext(r.Context(),
		`INSERT INTO messages (sender_id, receiver_id, content, message_type, target_id, is_read, created_at)
		 VALUES ($1,$2,$3,4,$4,0,NOW())`,
		senderID, owner, text, articleID)
}

// notifyArticleCommentLike 点赞专栏评论时给评论作者写 messages(type=6)。
func (h *Handler) notifyArticleCommentLike(r *http.Request, senderID, commentID int64) {
	var owner, articleID int64
	var content string
	if err := h.db.QueryRowContext(r.Context(),
		`SELECT user_id, article_id, COALESCE(content,'') FROM article_comments WHERE id=$1`, commentID).
		Scan(&owner, &articleID, &content); err != nil {
		return
	}
	if owner <= 0 || owner == senderID {
		return
	}
	text := "赞了你的评论\"" + excerpt(content, 100) + "\""
	if h.notifier != nil {
		h.notifier.SendNotification(r.Context(), senderID, owner, text, 6, articleID, commentID)
		return
	}
	h.db.ExecContext(r.Context(),
		`INSERT INTO messages (sender_id, receiver_id, content, message_type, target_id, comment_id, is_read, created_at)
		 VALUES ($1,$2,$3,6,$4,$5,0,NOW())`,
		senderID, owner, text, articleID, commentID)
}

// notifyArticleMentions 解析评论里的 @昵称 并写入 messages(type=3)。
func (h *Handler) notifyArticleMentions(r *http.Request, content string, senderID, articleID, commentID int64) {
	names := mentionRe.FindAllStringSubmatch(content, -1)
	if len(names) == 0 {
		return
	}
	seen := map[int64]bool{}
	snippet := excerpt(content, 140)
	for _, m := range names {
		name := strings.TrimRight(m[1], ",.;:!?，。；：！？、）】")
		if name == "" {
			continue
		}
		var uid int64
		if err := h.db.QueryRowContext(r.Context(),
			`SELECT id FROM users WHERE nickname=$1 OR username=$1 LIMIT 1`, name).Scan(&uid); err != nil || uid == 0 {
			continue
		}
		if uid == senderID || seen[uid] {
			continue
		}
		seen[uid] = true
		text := "在评论中@了你：" + snippet
		if h.notifier != nil {
			h.notifier.SendNotification(r.Context(), senderID, uid, text, 3, articleID, commentID)
			continue
		}
		h.db.ExecContext(r.Context(),
			`INSERT INTO messages (sender_id, receiver_id, content, message_type, target_id, comment_id, is_read, created_at)
			 VALUES ($1,$2,$3,3,$4,$5,0,NOW())`,
			senderID, uid, text, articleID, commentID)
	}
}
