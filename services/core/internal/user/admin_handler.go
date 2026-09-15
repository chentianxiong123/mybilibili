package user

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/auth"
	"mybilibili/pkg/httputil"
)

type AuditRecorder interface {
	RecordAudit(ctx context.Context, operatorID int64, operatorName, module, action, targetType, targetID string, result int32, message, detail string) error
}

// PermChecker 校验当前请求是否拥有指定权限码。
// 返回 (adminID, true) 表示有权限；adminID 用于审计记录。
type PermChecker interface {
	CheckPermission(r *http.Request, permission string) (int64, bool)
}

type UserAdminHandler struct {
	db      *sql.DB
	auditor AuditRecorder
	perm    PermChecker
	jwt     *auth.JWT
}

func NewUserAdminHandler(db *sql.DB, auditor AuditRecorder, perm PermChecker, jwt *auth.JWT) *UserAdminHandler {
	return &UserAdminHandler{db: db, auditor: auditor, perm: perm, jwt: jwt}
}

// requirePerm 鉴权中间件：无权限返回 401/403，有则继续。
func (h *UserAdminHandler) requirePerm(perm string, next http.HandlerFunc) http.HandlerFunc {
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

func (h *UserAdminHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/user/admin/list", h.requirePerm("user:manage", h.handleList))
	mux.HandleFunc("/api/v1/user/admin/", h.handleRoute)
}

// handleRoute 分派 /api/v1/user/admin/{id} 下的子路径（除 list 外）。
func (h *UserAdminHandler) handleRoute(w http.ResponseWriter, r *http.Request) {
	// /api/v1/user/admin/list 由上面的精确路由处理，这里跳过
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/user/admin/")
	if path == "list" {
		h.handleList(w, r)
		return
	}
	parts := strings.Split(path, "/")
	if len(parts) >= 1 {
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid user id", "data": nil})
			return
		}
		if len(parts) == 1 {
			// 鉴权：查看/更新用户
			if r.Method == "GET" {
				h.withPerm(w, r, "user:manage", func() { h.handleGet(w, r, id) })
				return
			}
			if r.Method == "PUT" {
				h.withPerm(w, r, "user:manage", func() { h.handleUpdate(w, r, id) })
				return
			}
			httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
			return
		}
		if len(parts) == 2 {
			switch parts[1] {
			case "status":
				h.withPerm(w, r, "user:manage", func() { h.handleStatus(w, r, id) })
				return
			case "password":
				h.withPerm(w, r, "user:manage", func() { h.handlePassword(w, r, id) })
				return
			default:
				httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
				return
			}
		}
	}
	httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
}

// withPerm 鉴权后执行 next。
func (h *UserAdminHandler) withPerm(w http.ResponseWriter, r *http.Request, perm string, next func()) {
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
	next()
}

func (h *UserAdminHandler) handleList(w http.ResponseWriter, r *http.Request) {
	page, size := httputil.ParsePageParams(r)
	offset := (page - 1) * size
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, username, nickname, COALESCE(email,''), COALESCE(avatar,''), level, status, created_at,
		 COALESCE(phone, ''),
		 COALESCE(follower_count, 0),
		 COALESCE(following_count, 0),
		 (SELECT COUNT(*) FROM manuscripts WHERE user_id = u.id)
		 FROM users u ORDER BY id DESC LIMIT $1 OFFSET $2`, size, offset)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "查询失败", "data": nil})
		return
	}
	defer rows.Close()
	type userItem struct {
		ID              int64  `json:"id"`
		Username        string `json:"username"`
		Nickname        string `json:"nickname"`
		Email           string `json:"email"`
		Avatar          string `json:"avatar"`
		Level           int32  `json:"level"`
		Status          int32  `json:"status"`
		CreatedAt       string `json:"created_at"`
		Phone           string `json:"phone"`
		FollowerCount   int64  `json:"follower_count"`
		FollowingCount  int64  `json:"following_count"`
		ManuscriptCount int64  `json:"manuscript_count"`
	}
	list := []userItem{}
	for rows.Next() {
		var u userItem
		var createdAt string
		if scanErr := rows.Scan(&u.ID, &u.Username, &u.Nickname, &u.Email, &u.Avatar, &u.Level, &u.Status, &createdAt,
			&u.Phone, &u.FollowerCount, &u.FollowingCount, &u.ManuscriptCount); scanErr != nil {
			continue
		}
		u.CreatedAt = createdAt
		list = append(list, u)
	}
	var total int64
	_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM users`).Scan(&total)
	httputil.WriteOK(w, map[string]any{
		"list": list, "total": total, "page": page, "size": size,
	})
}

func (h *UserAdminHandler) handleGet(w http.ResponseWriter, r *http.Request, id int64) {
	var uid int64
	var level, status, gender int32
	var username, nickname, email, avatar, createdAt, phone, birthdate, bio, signature, announcement string
	var followerCount, followingCount, manuscriptCount int64
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, username, nickname, COALESCE(email,''), COALESCE(avatar,''), level, status, created_at,
		 COALESCE(phone,''), COALESCE(gender,0), COALESCE(birthdate::text,''),
		 COALESCE(bio,''), COALESCE(signature,''), COALESCE(announcement,''),
		 COALESCE(follower_count,0), COALESCE(following_count,0),
		 (SELECT COUNT(*) FROM manuscripts WHERE user_id = u.id)
		 FROM users u WHERE id=$1`, id).
		Scan(&uid, &username, &nickname, &email, &avatar, &level, &status, &createdAt,
			&phone, &gender, &birthdate, &bio, &signature, &announcement,
			&followerCount, &followingCount, &manuscriptCount)
	if err != nil {
		if err == sql.ErrNoRows {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "user not found", "data": nil})
		} else {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "数据库错误", "data": nil})
		}
		return
	}
	httputil.WriteOK(w, map[string]any{
		"id": uid, "username": username, "nickname": nickname, "email": email,
		"avatar": avatar, "level": level, "status": status, "created_at": createdAt,
		"phone": phone, "gender": gender, "birthdate": birthdate,
		"bio": bio, "signature": signature, "announcement": announcement,
		"follower_count": followerCount, "following_count": followingCount,
		"manuscript_count": manuscriptCount,
	})
}

func (h *UserAdminHandler) handleUpdate(w http.ResponseWriter, r *http.Request, id int64) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid body", "data": nil})
		return
	}
	for k, v := range body {
		switch k {
		case "nickname":
			if s, ok := v.(string); ok && s != "" {
				var exists int
				_ = h.db.QueryRowContext(r.Context(),
					`SELECT COUNT(*) FROM users WHERE nickname = $1 AND id != $2`, s, id).Scan(&exists)
				if exists > 0 {
					httputil.WriteJSON(w, http.StatusConflict, map[string]any{"code": 409, "message": "昵称已存在", "data": nil})
					return
				}
				_, _ = h.db.ExecContext(r.Context(),
					`UPDATE users SET nickname=$1, updated_at=NOW() WHERE id=$2`, s, id)
			}
		case "email":
			if s, ok := v.(string); ok {
				_, _ = h.db.ExecContext(r.Context(),
					`UPDATE users SET email=$1, updated_at=NOW() WHERE id=$2`, s, id)
			}
		case "level":
			if n, ok := v.(float64); ok && n > 0 {
				_, _ = h.db.ExecContext(r.Context(),
					`UPDATE users SET level=$1, updated_at=NOW() WHERE id=$2`, int32(n), id)
			}
		}
	}
	httputil.WriteOK(w, map[string]any{"status": "ok"})
}

func (h *UserAdminHandler) handleStatus(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != "PUT" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	var body struct {
		Status int32 `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid body", "data": nil})
		return
	}
	res, err := h.db.ExecContext(r.Context(),
		`UPDATE users SET status=$1, updated_at=NOW() WHERE id=$2`, body.Status, id)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "更新失败", "data": nil})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "user not found", "data": nil})
		return
	}
	if h.auditor != nil {
		h.auditor.RecordAudit(r.Context(), httputil.GetAdminIDFromHeader(r), "", "user", "UPDATE_USER_STATUS", "users", strconv.FormatInt(id, 10), 0, "更新用户状态", "")
	}
	httputil.WriteOK(w, map[string]any{"status": "ok"})
}

func (h *UserAdminHandler) handlePassword(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != "PUT" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	var body struct {
		NewPassword string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid body", "data": nil})
		return
	}
	if body.NewPassword == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "newPassword required", "data": nil})
		return
	}
	hash := sha256hex(body.NewPassword)
	res, err := h.db.ExecContext(r.Context(),
		`UPDATE users SET password=$1, updated_at=NOW() WHERE id=$2`, hash, id)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "更新失败", "data": nil})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "user not found", "data": nil})
		return
	}
	if h.auditor != nil {
		h.auditor.RecordAudit(r.Context(), httputil.GetAdminIDFromHeader(r), "", "user", "RESET_USER_PASSWORD", "users", strconv.FormatInt(id, 10), 0, "重置用户密码", "")
	}
	httputil.WriteOK(w, map[string]any{"status": "ok"})
}

func sha256hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}
