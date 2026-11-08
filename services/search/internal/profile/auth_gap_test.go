package profile

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"mybilibili/pkg/auth"
)

// realMux builds the exact middleware stack used in cmd/search/main.go:
//   auth.IdentityMiddleware(jwt)(auth.AdminPathGuard(mux))
// with no bearer token presented. IdentityMiddleware strips client-supplied
// identity headers and, with no valid token, leaves identity EMPTY and passes
// through. AdminPathGuard only blocks paths that IsAdminPath marks as admin
// (paths containing /admin/ plus a fixed adminOnlyPrefixes list). Profile
// routes are NOT in that list, so they fall through to the handler.
func realMux(h *Handler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return auth.IdentityMiddleware(auth.NewJWT("dev-secret"))(auth.AdminPathGuard(mux))
}

// FINDING A1 (已修复): POST /api/v1/profile/{id}/init 是未认证写操作。
// handleProfileByPath (profile_handler.go:39-53) 曾直接用 URL 路径 userID，
// 未读 X-User-Id / RequireUser / admin 身份 → 任意匿名者可覆写任意用户画像。
// 修复后：匿名 → 401。
func TestAuthGap_ProfileInit_AnonymousReachable(t *testing.T) {
	h := newTestHandler(t) // memory store
	body := bytes.NewBufferString(`{"tags":["pwned","by-anon"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profile/42/init", body)
	// deliberately NO X-User-Id header and no bearer token
	rr := httptest.NewRecorder()

	realMux(h).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "unauthenticated POST /profile/42/init must be 401")
}
