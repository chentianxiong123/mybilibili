package article

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/httputil"
)

// PermChecker 校验当前请求是否拥有指定权限码（由 admin.Handler 实现）。
type PermChecker interface {
	CheckPermission(r *http.Request, permission string) (int64, bool)
}

// AdminHandler 专栏审核后台（对齐视频稿件的审核流程）。
// 公开可见条件：status=1 AND review_status=1 AND visibility=1。
type AdminHandler struct {
	db   *sql.DB
	perm PermChecker
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

func (h *AdminHandler) SetPermChecker(p PermChecker) {
	h.perm = p
}

func (h *AdminHandler) requirePerm(perm string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.perm == nil {
			httputil.WriteJSON(w, http.StatusForbidden, map[string]any{"code": 403, "message": "权限校验未配置", "data": nil})
			return
		}
		if _, ok := h.perm.CheckPermission(r, perm); !ok {
			if httputil.GetAdminIDFromHeader(r) == 0 {
				httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
			} else {
				httputil.WriteJSON(w, http.StatusForbidden, map[string]any{"code": 403, "message": "forbidden", "data": nil})
			}
			return
		}
		next(w, r)
	}
}

func (h *AdminHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/article/admin/pending", h.requirePerm("review:manage", h.handlePending))
	mux.HandleFunc("/api/v1/article/admin/all", h.requirePerm("review:manage", h.handleAll))
	mux.HandleFunc("/api/v1/article/admin/statistics", h.requirePerm("review:manage", h.handleStatistics))
	mux.HandleFunc("/api/v1/article/admin/review/", h.requirePerm("review:manage", h.handleReview))
}

type adminArticle struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	UserID       int64  `json:"user_id"`
	AuthorName   string `json:"author_name"`
	AuthorAvatar string `json:"author_avatar"`
	CoverURL     string `json:"cover_url"`
	Summary      string `json:"summary"`
	WordCount    int    `json:"word_count"`
	Status       int    `json:"status"`
	ReviewStatus int    `json:"review_status"`
	ReviewReason string `json:"review_reason"`
	PublishTime  string `json:"publish_time"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

const adminArticleCols = `a.id, a.title, a.user_id, COALESCE(u.nickname,''), COALESCE(u.avatar,''),
	COALESCE(a.cover_url,''), COALESCE(a.summary,''), COALESCE(a.word_count,0),
	a.status, COALESCE(a.review_status,0), COALESCE(a.review_reason,''),
	COALESCE(to_char(a.publish_time AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),''),
	COALESCE(to_char(a.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),''),
	COALESCE(to_char(a.updated_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),'')`

func (h *AdminHandler) list(w http.ResponseWriter, r *http.Request, onlyPending bool) {
	page, size := httputil.ParsePageParams(r)
	if size <= 0 || size > 100 {
		size = 20
	}
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
	statusStr := strings.TrimSpace(r.URL.Query().Get("status"))

	where := "a.status <> -1"
	args := []any{}
	if onlyPending {
		where = "a.status = 1 AND a.review_status = 0"
	} else if statusStr != "" {
		if st, err := strconv.Atoi(statusStr); err == nil {
			args = append(args, st)
			where = "a.status = $1"
		}
	}
	if keyword != "" {
		args = append(args, "%"+keyword+"%")
		where += " AND a.title ILIKE $" + strconv.Itoa(len(args))
	}
	offset := (page - 1) * size
	args = append(args, size, offset)
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT `+adminArticleCols+` FROM articles a LEFT JOIN users u ON u.id = a.user_id
		 WHERE `+where+` ORDER BY a.updated_at DESC LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)),
		args...)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "查询失败", "data": nil})
		return
	}
	defer rows.Close()

	list := []adminArticle{}
	for rows.Next() {
		var a adminArticle
		if err := rows.Scan(&a.ID, &a.Title, &a.UserID, &a.AuthorName, &a.AuthorAvatar,
			&a.CoverURL, &a.Summary, &a.WordCount, &a.Status, &a.ReviewStatus, &a.ReviewReason,
			&a.PublishTime, &a.CreatedAt, &a.UpdatedAt); err != nil {
			continue
		}
		list = append(list, a)
	}

	var total int
	countWhere := where
	countArgs := args[:len(args)-2]
	// count 查询用不到 LIMIT/OFFSET，占位符编号保持一致
	_ = h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM articles a WHERE `+countWhere, countArgs...).Scan(&total)

	httputil.WriteOK(w, map[string]any{"list": list, "total": total, "page": page, "size": size})
}

// GET /api/v1/article/admin/pending — 待审核专栏
func (h *AdminHandler) handlePending(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	h.list(w, r, true)
}

// GET /api/v1/article/admin/all — 全部专栏（可按 status 过滤）
func (h *AdminHandler) handleAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	h.list(w, r, false)
}

// GET /api/v1/article/admin/statistics — 审核概览
func (h *AdminHandler) handleStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	stats := map[string]any{}
	var total, pending, approved, rejected, draft int64
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM articles WHERE status <> -1`).Scan(&total)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM articles WHERE status = 1 AND review_status = 0`).Scan(&pending)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM articles WHERE status = 1 AND review_status = 1`).Scan(&approved)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM articles WHERE review_status = 2`).Scan(&rejected)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM articles WHERE status = 0`).Scan(&draft)
	stats["total"] = total
	stats["pending"] = pending
	stats["approved"] = approved
	stats["rejected"] = rejected
	stats["draft"] = draft
	httputil.WriteOK(w, stats)
}

// POST /api/v1/article/admin/review/{id} — {approved:bool, reason:string}
func (h *AdminHandler) handleReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	id, err := strconv.ParseInt(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/article/admin/review/"), "/"), 10, 64)
	if err != nil || id <= 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid id", "data": nil})
		return
	}
	var req struct {
		Approved bool   `json:"approved"`
		Reason   string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	reason := strings.TrimSpace(req.Reason)
	if !req.Approved && reason == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "请填写拒绝原因", "data": nil})
		return
	}
	if len([]rune(reason)) > 500 {
		reason = string([]rune(reason)[:500])
	}

	var exists int64
	if err := h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM articles WHERE id = $1`, id).Scan(&exists); err != nil || exists == 0 {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "专栏不存在", "data": nil})
		return
	}

	adminID, _ := h.perm.CheckPermission(r, "review:manage")
	if req.Approved {
		_, err = h.db.ExecContext(r.Context(),
			`UPDATE articles SET status = 1, review_status = 1, review_reason = '', review_time = NOW(),
			        reviewer_id = $1, publish_time = COALESCE(publish_time, NOW()), updated_at = NOW()
			 WHERE id = $2`, adminID, id)
	} else {
		_, err = h.db.ExecContext(r.Context(),
			`UPDATE articles SET status = 2, review_status = 2, review_reason = $1, review_time = NOW(),
			        reviewer_id = $2, updated_at = NOW()
			 WHERE id = $3`, reason, adminID, id)
	}
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "操作失败", "data": nil})
		return
	}
	msg := "审核通过"
	if !req.Approved {
		msg = "已拒绝"
	}
	httputil.WriteOK(w, map[string]any{"status": "ok", "article_id": id, "message": msg})
}
