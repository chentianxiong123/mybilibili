package httputil

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func GetUserIDFromHeader(r *http.Request) int64 {
	idStr := r.Header.Get("X-User-Id")
	if idStr == "" {
		return 0
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// GetAdminIDFromHeader 读取管理员身份。
//
// 只认 X-Admin-Id，绝不能回退到 X-User-Id：users.id 与 admin_users.id 是两套
// 独立自增 ID，一旦回退，任何普通用户都会被当成"同号管理员"提权
// （实测 users.id=4 的 string 拿到了 admin_users.id=4 system_admin 的全部权限）。
// X-Admin-Id 由 IdentityMiddleware 在 claims.IsAdmin 为真时才注入，故可信。
func GetAdminIDFromHeader(r *http.Request) int64 {
	idStr := r.Header.Get("X-Admin-Id")
	if idStr == "" {
		return 0
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func ParsePageParams(r *http.Request) (int32, int32) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	sizeStr := r.URL.Query().Get("page_size")
	if sizeStr == "" {
		sizeStr = r.URL.Query().Get("pageSize")
	}
	if sizeStr == "" {
		sizeStr = r.URL.Query().Get("size")
	}
	if sizeStr == "" {
		sizeStr = r.URL.Query().Get("limit")
	}
	size, _ := strconv.ParseInt(sizeStr, 10, 32)
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 20
	}
	return int32(page), int32(size)
}

func WriteOK(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, map[string]interface{}{"code": 200, "data": data, "message": "ok"})
}

func WriteJSON(w http.ResponseWriter, httpStatus int, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(body)
}

func WriteError(w http.ResponseWriter, httpStatus int, msg string) {
	WriteJSON(w, httpStatus, map[string]interface{}{"code": httpStatus, "message": msg, "data": nil})
}

func RequireUser(w http.ResponseWriter, r *http.Request) (int64, bool) {
	uid := GetUserIDFromHeader(r)
	if uid == 0 {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "unauthorized", "data": nil})
		return 0, false
	}
	return uid, true
}

func PathValue(r *http.Request, key string) string {
	v := r.PathValue(key)
	if v != "" {
		return v
	}
	return ""
}
