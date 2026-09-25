package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cookiesOf(rec *httptest.ResponseRecorder) map[string]*http.Cookie {
	out := map[string]*http.Cookie{}
	for _, c := range rec.Result().Cookies() {
		out[c.Name] = c
	}
	return out
}

func TestSetAccessCookie_IsHttpOnly(t *testing.T) {
	rec := httptest.NewRecorder()
	SetAccessCookie(rec, UserAccessTokenCookie, "jwt-value")

	c, ok := cookiesOf(rec)[UserAccessTokenCookie]
	require.True(t, ok, "必须下发 token cookie")
	assert.True(t, c.HttpOnly, "访问令牌必须 HttpOnly，JS/XSS 读不到")
	assert.Equal(t, "/", c.Path)
	assert.Equal(t, http.SameSiteLaxMode, c.SameSite)
	assert.Equal(t, "jwt-value", c.Value)
	assert.Equal(t, 24*3600, c.MaxAge, "存活期应与访问令牌 24h 对齐")
}

func TestSetRefreshCookie_IsHttpOnly(t *testing.T) {
	rec := httptest.NewRecorder()
	SetRefreshCookie(rec, UserRefreshTokenCookie, "rj")

	c, ok := cookiesOf(rec)[UserRefreshTokenCookie]
	require.True(t, ok)
	assert.True(t, c.HttpOnly)
	assert.Equal(t, 30*24*3600, c.MaxAge)
}

func TestAdminSessionCookies_AreSeparateNamespace(t *testing.T) {
	rec := httptest.NewRecorder()
	SetAccessCookie(rec, AdminAccessTokenCookie, "aj")
	SetRefreshCookie(rec, AdminRefreshTokenCookie, "arj")

	cookies := cookiesOf(rec)
	require.Contains(t, cookies, AdminAccessTokenCookie)
	require.Contains(t, cookies, AdminRefreshTokenCookie)
	assert.NotContains(t, cookies, UserAccessTokenCookie, "管理员登录不得写用户的 token cookie")
	assert.True(t, cookies[AdminAccessTokenCookie].HttpOnly)
}

func TestSetPlainCookie_UsableByJS(t *testing.T) {
	rec := httptest.NewRecorder()
	SetPlainCookie(rec, UserInfoCookie, "%7B%22id%22%3A4%7D")

	c, ok := cookiesOf(rec)[UserInfoCookie]
	require.True(t, ok)
	assert.False(t, c.HttpOnly, "user_info 只是展示数据，前端要能读")
}

func TestCookieSecure_IsEnvGated(t *testing.T) {
	t.Setenv("COOKIE_SECURE", "")
	assert.False(t, CookieSecure(), "dev 走 http，默认不能开 Secure，否则浏览器直接丢弃")

	t.Setenv("COOKIE_SECURE", "true")
	assert.True(t, CookieSecure())

	t.Setenv("COOKIE_SECURE", "1")
	assert.True(t, CookieSecure())

	rec := httptest.NewRecorder()
	SetAccessCookie(rec, UserAccessTokenCookie, "jwt")
	assert.True(t, cookiesOf(rec)[UserAccessTokenCookie].Secure)
}

func TestClearCookie_DeletesIt(t *testing.T) {
	rec := httptest.NewRecorder()
	ClearCookie(rec, UserAccessTokenCookie)

	c, ok := cookiesOf(rec)[UserAccessTokenCookie]
	require.True(t, ok)
	assert.Equal(t, -1, c.MaxAge, "MaxAge<0 才会真正删除浏览器里的 cookie")
	assert.Empty(t, c.Value)
}
