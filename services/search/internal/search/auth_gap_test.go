package search

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"mybilibili/pkg/auth"
)

// realMux builds the exact middleware stack from cmd/search/main.go:
// auth.IdentityMiddleware(jwt)(auth.AdminPathGuard(mux)), no bearer token.
func realMux(h *Handler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return auth.IdentityMiddleware(auth.NewJWT("dev-secret"))(auth.AdminPathGuard(mux))
}

func anonRequest(method, path, body string) *http.Request {
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	// no X-User-Id / X-Admin-Id / bearer token
	return req
}

// FINDING A3: POST /api/v1/search/hot/increment is an unauthenticated mutation.
// handleHotIncrement (search_handler.go:90-97) has no RequireUser / admin check
// and the path is not protected by AdminPathGuard (not /admin/, not in the
// adminOnlyPrefixes list; require_admin.go explicitly leaves it anonymous for
// the web search page). Any anonymous caller can inflate hot-search scores.
func TestAuthGap_HotIncrement_AnonymousReachable(t *testing.T) {
	h, _ := newTestHandler(t) // hotRepo nil => handler no-op but still 200
	rr := httptest.NewRecorder()

	realMux(h).ServeHTTP(rr, anonRequest(http.MethodPost, "/api/v1/search/hot/increment", `{"keyword":"golang"}`))

	// A properly enforced auth boundary would 401. Observed 200 => reachable.
	assert.Equal(t, http.StatusOK, rr.Code, "unauthenticated POST /search/hot/increment must NOT return 200")
}

// FINDING A4: POST /api/v1/search/hot/clean-expired is an unauthenticated
// mutation. handleCleanExpired (search_handler.go:178-185) has no auth check
// and the path is not protected by AdminPathGuard (require_admin.go documents it
// as an internal-cron path but does not enforce any token/secret). Any anonymous
// caller can prune hot-search entries (removes everything beyond the top-100).
func TestAuthGap_HotCleanExpired_AnonymousReachable(t *testing.T) {
	h, _ := newTestHandler(t)
	rr := httptest.NewRecorder()

	realMux(h).ServeHTTP(rr, anonRequest(http.MethodPost, "/api/v1/search/hot/clean-expired", ""))

	assert.Equal(t, http.StatusOK, rr.Code, "unauthenticated POST /search/hot/clean-expired must NOT return 200")
}

// NEGATIVE CONTROL: hot/keyword, hot/rank, hot/score and hot/delete ARE
// protected by AdminPathGuard via the adminOnlyPrefixes list in
// shared/pkg/auth/require_admin.go, even though their paths do not contain
// "/admin/". These are therefore NOT auth gaps. The task's stated premise
// ("AdminPathGuard only intercepts /admin/") is stale: the guard is broader.
func TestAdminGuard_ProtectsHotAdminMutations(t *testing.T) {
	cases := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/api/v1/search/hot/keyword", `{"keyword":"g","score":1,"rank":1}`},
		{http.MethodPut, "/api/v1/search/hot/rank", `{"keyword":"g","rank":1}`},
		{http.MethodPut, "/api/v1/search/hot/score", `{"keyword":"g","score":1}`},
		{http.MethodDelete, "/api/v1/search/hot/delete", `{"keyword":"g"}`},
		{http.MethodGet, "/api/v1/search/hot/get?keyword=g", ""},
		{http.MethodGet, "/api/v1/search/hot/score-get?keyword=g", ""},
	}
	h, _ := newTestHandler(t)
	for _, c := range cases {
		rr := httptest.NewRecorder()
		realMux(h).ServeHTTP(rr, anonRequest(c.method, c.path, c.body))
		assert.Equal(t, http.StatusUnauthorized, rr.Code,
			"anonymous %s %s should be blocked by AdminPathGuard (X-Admin-Id required)", c.method, c.path)
	}
}
