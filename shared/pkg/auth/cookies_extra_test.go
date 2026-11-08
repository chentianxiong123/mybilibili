package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddTokens_DisabledByDefault(t *testing.T) {
	body := map[string]interface{}{"code": 0}
	AddTokens(body, false, "tok", "rtok")
	_, hasTok := body["token"]
	_, hasRTok := body["refresh_token"]
	assert.False(t, hasTok, "默认不得把凭证写进响应体（防 XSS 整包偷走）")
	assert.False(t, hasRTok)
	assert.Equal(t, map[string]interface{}{"code": 0}, body)
}

func TestAddTokens_ExplicitNativeClient(t *testing.T) {
	body := map[string]interface{}{}
	AddTokens(body, true, "native-tok", "native-rtok")
	assert.Equal(t, "native-tok", body["token"])
	assert.Equal(t, "native-rtok", body["refresh_token"])
}

func TestAddTokens_OverwritesExistingKeys(t *testing.T) {
	body := map[string]interface{}{"token": "stale", "refresh_token": "stale-rt"}
	AddTokens(body, true, "new", "new-rt")
	assert.Equal(t, "new", body["token"])
	assert.Equal(t, "new-rt", body["refresh_token"])
}

func TestClearUserSessionCookies_ClearsAllThree(t *testing.T) {
	rec := httptest.NewRecorder()
	ClearUserSessionCookies(rec)

	cookies := cookiesOf(rec)
	for _, name := range []string{UserAccessTokenCookie, UserRefreshTokenCookie, UserInfoCookie} {
		c, ok := cookies[name]
		require.Truef(t, ok, "%s 必须被清除", name)
		assert.Equal(t, -1, c.MaxAge, "MaxAge<0 才会真正删除浏览器里的 cookie")
		assert.Empty(t, c.Value)
		assert.True(t, c.HttpOnly)
	}
}

func TestClearAdminSessionCookies_ClearsBoth(t *testing.T) {
	rec := httptest.NewRecorder()
	ClearAdminSessionCookies(rec)

	cookies := cookiesOf(rec)
	require.Contains(t, cookies, AdminAccessTokenCookie)
	require.Contains(t, cookies, AdminRefreshTokenCookie)
	assert.Equal(t, -1, cookies[AdminAccessTokenCookie].MaxAge)
	assert.Equal(t, -1, cookies[AdminRefreshTokenCookie].MaxAge)
	// 清除管理员登录态不得误伤普通用户 cookie
	assert.NotContains(t, cookies, UserAccessTokenCookie)
}

// ClearCookie 的 Secure 属性跟随 COOKIE_SECURE 环境变量。
func TestClearCookie_SecureFollowsEnv(t *testing.T) {
	t.Setenv("COOKIE_SECURE", "true")
	rec := httptest.NewRecorder()
	ClearCookie(rec, UserAccessTokenCookie)
	c, ok := cookiesOf(rec)[UserAccessTokenCookie]
	require.True(t, ok)
	assert.True(t, c.Secure, "线上 Secure 模式下清除 cookie 也必须带 Secure，否则浏览器拒绝")

	t.Setenv("COOKIE_SECURE", "")
	rec2 := httptest.NewRecorder()
	ClearCookie(rec2, UserAccessTokenCookie)
	c2, ok2 := cookiesOf(rec2)[UserAccessTokenCookie]
	require.True(t, ok2)
	assert.False(t, c2.Secure)
}

// 登录/登出响应里 cookie 的 SameSite 必须统一 Lax。
func TestSessionCookies_SameSiteLax(t *testing.T) {
	rec := httptest.NewRecorder()
	SetAccessCookie(rec, UserAccessTokenCookie, "a")
	SetRefreshCookie(rec, UserRefreshTokenCookie, "b")
	SetPlainCookie(rec, UserInfoCookie, "c")
	ClearUserSessionCookies(rec)

	for _, c := range rec.Result().Cookies() {
		assert.Equal(t, http.SameSiteLaxMode, c.SameSite, "cookie %s", c.Name)
	}
}
