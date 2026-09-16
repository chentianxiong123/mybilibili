package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const mwSecret = "middleware-test"

func TestIdentityMiddleware_TraefikHeaders(t *testing.T) {
	j := NewJWT(mwSecret)
	mw := IdentityMiddleware(j)

	var seenID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenID = r.Header.Get("X-User-Id")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("X-User-Id", "555")
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)

	assert.Equal(t, "555", seenID, "Traefik 注入的头应直通")
}

func TestIdentityMiddleware_BearerFallback(t *testing.T) {
	j := NewJWT(mwSecret)
	mw := IdentityMiddleware(j)

	tok, err := j.Generate(777)
	assert.NoError(t, err)

	var seenUserID, seenRole, seenAdmin string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenUserID = r.Header.Get("X-User-Id")
		seenRole = r.Header.Get("X-User-Role")
		seenAdmin = r.Header.Get("X-Admin-Id")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)

	assert.Equal(t, "777", seenUserID)
	assert.Equal(t, RoleUser, seenRole)
	assert.Equal(t, "", seenAdmin, "普通用户不应注入 X-Admin-Id")
}

func TestIdentityMiddleware_BearerFallback_Admin(t *testing.T) {
	j := NewJWT(mwSecret)
	mw := IdentityMiddleware(j)

	tok, err := j.GenerateAdmin(99)
	assert.NoError(t, err)

	var seenAdmin, seenRole string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAdmin = r.Header.Get("X-Admin-Id")
		seenRole = r.Header.Get("X-User-Role")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)

	assert.Equal(t, "99", seenAdmin)
	assert.Equal(t, RoleAdmin, seenRole)
}

func TestIdentityMiddleware_NoAuth(t *testing.T) {
	j := NewJWT(mwSecret)
	mw := IdentityMiddleware(j)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := r.Header.Get("X-User-Id")
		role := r.Header.Get("X-User-Role")
		assert.Equal(t, "", uid, "匿名请求不应注入 X-User-Id")
		assert.Equal(t, "", role)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/x", nil)
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestIdentityMiddleware_InvalidBearer(t *testing.T) {
	j := NewJWT(mwSecret)
	mw := IdentityMiddleware(j)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 无效 token 不应注入任何身份头, 直接放行
		assert.Equal(t, "", r.Header.Get("X-User-Id"))
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUserIDFromHeader(t *testing.T) {
	// UserIDFromHeader 接受一个 getHeader 函数 (用于解耦 net/http)
	id := UserIDFromHeader(nil, func(key string) string {
		if key == "X-User-Id" {
			return "42"
		}
		return ""
	})
	assert.Equal(t, int64(42), id)
}

func TestUserIDFromHeader_Invalid(t *testing.T) {
	id := UserIDFromHeader(nil, func(key string) string {
		if key == "X-User-Id" {
			return "not-a-number"
		}
		return ""
	})
	assert.Equal(t, int64(0), id)
}

func TestUserIDFromHeader_Empty(t *testing.T) {
	id := UserIDFromHeader(nil, func(key string) string { return "" })
	assert.Equal(t, int64(0), id)
}

func TestAdminIDFromHeader(t *testing.T) {
	id := AdminIDFromHeader(func(key string) string {
		if key == "X-Admin-Id" {
			return "101"
		}
		return ""
	})
	assert.Equal(t, int64(101), id)
}