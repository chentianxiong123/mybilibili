package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func csrfRun(t *testing.T, method, path string, hdr map[string]string) (*httptest.ResponseRecorder, bool) {
	t.Helper()
	j := NewJWT(mwSecret)
	mw := IdentityMiddleware(j)
	tok, err := j.Generate(4)
	require.NoError(t, err)

	var reached bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(method, path, nil) // Host: example.com
	req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: tok})
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)
	return rec, reached
}

func TestCSRF_CrossSiteCookieWriteRejected(t *testing.T) {
	for _, tc := range []struct {
		name string
		hdr  map[string]string
	}{
		{"Origin 指向外部站点", map[string]string{"Origin": "http://evil.example"}},
		{"Sec-Fetch-Site 为 cross-site", map[string]string{"Sec-Fetch-Site": "cross-site"}},
		{"Origin 为 null（沙箱 iframe）", map[string]string{"Origin": "null"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec, reached := csrfRun(t, "POST", "/api/v1/comment/add", tc.hdr)
			assert.False(t, reached, "跨站写操作不得进入业务 handler")
			assert.Equal(t, http.StatusForbidden, rec.Code)
		})
	}
}

func TestCSRF_SameOriginCookieWriteAllowed(t *testing.T) {
	rec, reached := csrfRun(t, "POST", "/api/v1/comment/add",
		map[string]string{"Origin": "http://example.com"})
	assert.True(t, reached)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCSRF_CrossSiteGetStillAllowed(t *testing.T) {
	// GET 不改状态，拦了只会误伤（比如带 Origin 的普通读请求）
	rec, reached := csrfRun(t, "GET", "/api/v1/video/list",
		map[string]string{"Origin": "http://evil.example", "Sec-Fetch-Site": "cross-site"})
	assert.True(t, reached)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCSRF_NonBrowserClientUnaffected(t *testing.T) {
	// curl / 服务间调用既不带 Origin 也不带 Sec-Fetch-Site，放行
	rec, reached := csrfRun(t, "POST", "/api/v1/comment/add", nil)
	assert.True(t, reached)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCSRF_BearerSessionsUnaffected(t *testing.T) {
	// Bearer 头无法被跨站设置，CSRF 对它本就不成立，不该误伤
	j := NewJWT(mwSecret)
	mw := IdentityMiddleware(j)
	tok, err := j.Generate(4)
	require.NoError(t, err)

	var reached bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("POST", "/api/v1/comment/add", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)

	assert.True(t, reached)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCSRF_OriginSurvivesProxyHostRewrite(t *testing.T) {
	// Nuxt /api 代理会把 Host 换成后端地址，浏览器 Origin 仍是页面地址；
	// 这种同源请求绝不能被误判成跨站
	j := NewJWT(mwSecret)
	mw := IdentityMiddleware(j)
	tok, err := j.Generate(4)
	require.NoError(t, err)

	var reached bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("POST", "/api/v1/comment/add", nil)
	req.Host = "localhost:8080" // 代理改写后的 Host
	req.Header.Set("Origin", "http://localhost:3200")
	req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: tok})
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)

	assert.True(t, reached, "同源但端口被代理改写，不能误伤")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCSRF_XForwardedHostPreferred(t *testing.T) {
	j := NewJWT(mwSecret)
	mw := IdentityMiddleware(j)
	tok, err := j.Generate(4)
	require.NoError(t, err)

	var reached bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("POST", "/api/v1/comment/add", nil)
	req.Host = "mybilibili-core:8080"
	req.Header.Set("X-Forwarded-Host", "localhost")
	req.Header.Set("Origin", "http://localhost")
	req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: tok})
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)

	assert.True(t, reached)
	assert.Equal(t, http.StatusOK, rec.Code)
}
