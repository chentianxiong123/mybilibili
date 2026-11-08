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
	// FINDING-082 修复：负 ID 视为无效（此前 -5 作为有效身份通过）。
	if id <= 0 {
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
	// FINDING-082 修复：负 ID 视为无效。
	if id <= 0 {
		return 0
	}
	return id
}

func ParsePageParams(r *http.Request) (int32, int32) {
	// FINDING-083 修复：溢出/解析失败 → page/size 归零 → SafePage 重置 page=1/size=20
	// （此前 ParseInt 溢出返回钳制 max 值 → 荒谬 offset）。
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	if err != nil {
		page = 0
	}
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
	size, err := strconv.ParseInt(sizeStr, 10, 32)
	if err != nil {
		size = 0
	}
	page32, size32 := SafePage(int32(page), int32(size))
	return page32, size32
}

// SafePage 钳制分页参数：page∈[1,1e6]、size∈[1,50]（默认 20）。
// FINDING-012 修复：page 此前可高达 2^31-1，使 (page-1)*size int32 溢出为负
// （切片 panic / SQL 负 OFFSET / 500）。
func SafePage(page, size int32) (int32, int32) {
	const (
		maxPage = 1_000_000
		maxSize = 50
	)
	if page < 1 {
		page = 1
	}
	if page > maxPage {
		page = maxPage
	}
	if size < 1 || size > maxSize {
		size = 20
	}
	return page, size
}

// PageOffset 计算安全偏移量：page∈[1,1e6]、size∈[1,50] → offset≤5e7，int32 不溢出。
func PageOffset(page, size int32) int32 {
	page, size = SafePage(page, size)
	return (page - 1) * size
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

// RequireAdmin 校验管理员身份（X-Admin-Id），未认证时 401（FINDING-022 修复：
// 此前 IdentityMiddleware 缺失/无效 token 时不 401，导致未认证者可调敏感端点）。
func RequireAdmin(w http.ResponseWriter, r *http.Request) (int64, bool) {
	aid := GetAdminIDFromHeader(r)
	if aid == 0 {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "unauthorized", "data": nil})
		return 0, false
	}
	return aid, true
}

func PathValue(r *http.Request, key string) string {
	v := r.PathValue(key)
	if v != "" {
		return v
	}
	return ""
}
