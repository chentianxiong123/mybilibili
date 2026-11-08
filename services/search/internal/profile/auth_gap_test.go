package profile

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

// FINDING A1: POST /api/v1/profile/{id}/init is an unauthenticated mutation.
// handleProfileByPath (profile_handler.go:39-53) never reads X-User-Id /
// RequireUser / admin identity. userID comes straight from the URL path, so ANY
// anonymous caller can initialize (overwrite) the tags of ARBITRARY users.
func TestAuthGap_ProfileInit_AnonymousReachable(t *testing.T) {
	h := newTestHandler(t) // memory store
	body := bytes.NewBufferString(`{"tags":["pwned","by-anon"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profile/42/init", body)
	// deliberately NO X-User-Id header and no bearer token
	rr := httptest.NewRecorder()

	realMux(h).ServeHTTP(rr, req)

	// A correctly-enforced auth boundary returns 401. Observed: 200 => reachable.
	assert.Equal(t, http.StatusOK, rr.Code, "unauthenticated POST /profile/42/init must NOT return 200")

	// Confirm the mutation actually happened for user 42 (IDOR: arbitrary user).
	p, err := h.svc.Get(context.Background(), 42)
	require.NoError(t, err)
	require.NotNil(t, p, "profile for user 42 was created by an anonymous caller")
	assert.Equal(t, []string{"pwned", "by-anon"}, p.Tags)
}

// FINDING A2: POST /api/v1/profile/record/{watch|like|collect} is reachable
// unauthenticated. handleRecord (profile_handler.go:66-97) calls
// httputil.GetUserIDFromHeader but does NOT reject userID==0, so an anonymous
// caller (no token) proceeds with userID=0 and writes to a shared "user 0"
// profile (and pollutes the record store) instead of getting 401.
func TestAuthGap_ProfileRecord_AnonymousReachable(t *testing.T) {
	h := newTestHandler(t)
	body := bytes.NewBufferString(`{"categoryId":5,"tags":["anon"],"durationSeconds":10}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profile/record/watch", body)
	// deliberately NO X-User-Id header / no bearer token
	rr := httptest.NewRecorder()

	realMux(h).ServeHTTP(rr, req)

	// FINDING-018 修复后：未认证 → 401（此前 200 + user0 污染）。
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "unauthenticated POST /profile/record/watch must return 401")
}

// sanity: an authenticated caller (valid JWT) still gets 200 (regression guard).
func TestAuthGap_ProfileRecord_AuthedStillWorks(t *testing.T) {
	h := newTestHandler(t)
	body := bytes.NewBufferString(`{"categoryId":5,"tags":["user"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profile/record/like", body)
	// IdentityMiddleware 删除客户端身份头，仅从有效 JWT 重建 → 需真 token。
	tok, err := auth.NewJWT("dev-secret").Generate(7)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()

	realMux(h).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	p, err := h.svc.Get(context.Background(), 7)
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, int64(1), p.LikeCount)
}
