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

	"mybilibili/pkg/httputil"
)

type AuditRecorder interface {
	RecordAudit(ctx context.Context, operatorID int64, operatorName, module, action, targetType, targetID string, result int32, message, detail string) error
}

type UserAdminHandler struct {
	db      *sql.DB
	auditor AuditRecorder
}

func NewUserAdminHandler(db *sql.DB, auditor AuditRecorder) *UserAdminHandler {
	return &UserAdminHandler{db: db, auditor: auditor}
}

func (h *UserAdminHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/user/admin/", h.handleRoute)
}

// handleRoute 分派 /api/v1/user/admin/ 下的子路径
func (h *UserAdminHandler) handleRoute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/user/admin/")
	parts := strings.Split(path, "/")
	if len(parts) == 1 && parts[0] == "list" {
		h.handleList(w, r)
		return
	}
	if len(parts) >= 1 {
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			http.Error(w, "invalid user id", 400)
			return
		}
		if len(parts) == 1 {
			switch r.Method {
			case "GET":
				h.handleGet(w, r, id)
			case "PUT":
				h.handleUpdate(w, r, id)
			default:
				http.Error(w, "method not allowed", 405)
			}
			return
		}
		if len(parts) == 2 {
			switch parts[1] {
			case "status":
				h.handleStatus(w, r, id)
			case "password":
				h.handlePassword(w, r, id)
			default:
				http.Error(w, "not found", 404)
			}
			return
		}
	}
	http.Error(w, "not found", 404)
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
		http.Error(w, err.Error(), 500)
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
	json.NewEncoder(w).Encode(map[string]interface{}{
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
			http.Error(w, "user not found", 404)
		} else {
			http.Error(w, "db error: "+err.Error(), 500)
		}
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
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
		http.Error(w, "invalid body", 400)
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
					http.Error(w, "昵称已存在", 409)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *UserAdminHandler) handleStatus(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != "PUT" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var body struct {
		Status int32 `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	res, err := h.db.ExecContext(r.Context(),
		`UPDATE users SET status=$1, updated_at=NOW() WHERE id=$2`, body.Status, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "user not found", 404)
		return
	}
	if h.auditor != nil {
		h.auditor.RecordAudit(r.Context(), httputil.GetAdminIDFromHeader(r), "", "user", "UPDATE_USER_STATUS", "users", strconv.FormatInt(id, 10), 0, "更新用户状态", "")
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *UserAdminHandler) handlePassword(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != "PUT" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var body struct {
		NewPassword string `json:"newPassword"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if body.NewPassword == "" {
		http.Error(w, "newPassword required", 400)
		return
	}
	hash := sha256hex(body.NewPassword)
	res, err := h.db.ExecContext(r.Context(),
		`UPDATE users SET password=$1, updated_at=NOW() WHERE id=$2`, hash, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "user not found", 404)
		return
	}
	if h.auditor != nil {
		h.auditor.RecordAudit(r.Context(), httputil.GetAdminIDFromHeader(r), "", "user", "RESET_USER_PASSWORD", "users", strconv.FormatInt(id, 10), 0, "重置用户密码", "")
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func sha256hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}
