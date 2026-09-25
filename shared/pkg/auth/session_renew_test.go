package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 用 6h 有效期的 JWT 构造"临近过期"（滑动续期阈值是 12h）
func newShortLivedJWT() *JWT { return NewJWTWithDuration(mwSecret, 6*time.Hour) }

func runWithCookie(t *testing.T, j *JWT, cookieName, tok, method, path string, hdr map[string]string) (*httptest.ResponseRecorder, bool, string) {
	t.Helper()
	mw := IdentityMiddleware(j)
	var seenID string
	var reached bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		seenID = r.Header.Get("X-User-Id")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(method, path, nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: tok})
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)
	return rec, reached, seenID
}

func TestSlidingRenewal_NearExpiryIssuesNewCookie(t *testing.T) {
	j := newShortLivedJWT()
	tok, err := j.Generate(4)
	require.NoError(t, err)

	rec, reached, seenID := runWithCookie(t, j, UserAccessTokenCookie, tok, "GET", "/api/v1/user/me", nil)

	assert.True(t, reached)
	assert.Equal(t, "4", seenID)

	cookies := cookiesOf(rec)
	require.Contains(t, cookies, UserAccessTokenCookie, "临近过期必须回写 Set-Cookie")
	assert.True(t, cookies[UserAccessTokenCookie].HttpOnly, "续期下发的 cookie 同样必须 HttpOnly")

	// 同一秒内重签得到的令牌可能字节相同，这里只验证身份不丢
	c2, err := j.Parse(cookies[UserAccessTokenCookie].Value)
	require.NoError(t, err)
	assert.Equal(t, int64(4), c2.UserId)
	assert.True(t, c2.IsAccess())
}

// 有效期确实被延长了：用 2h 的存量令牌、24h 的签发器走续期
func TestSlidingRenewal_ExtendsExpiry(t *testing.T) {
	issuer := NewJWTWithDuration(mwSecret, 2*time.Hour)
	mw := IdentityMiddleware(NewJWTWithDuration(mwSecret, 24*time.Hour))

	tok, err := issuer.Generate(4)
	require.NoError(t, err)
	oldClaims, err := issuer.Parse(tok)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/user/me", nil)
	req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: tok})
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	c := cookiesOf(rec)[UserAccessTokenCookie]
	require.NotNil(t, c, "必须回写 Set-Cookie")
	newClaims, err := issuer.Parse(c.Value)
	require.NoError(t, err)
	assert.True(t, newClaims.ExpiresAt.Time.After(oldClaims.ExpiresAt.Time),
		"续期后过期时间必须更晚，否则等于没续")
}

func TestSlidingRenewal_FreshTokenNotRenewed(t *testing.T) {
	j := NewJWT(mwSecret) // 默认 24h > 12h 阈值
	tok, err := j.Generate(4)
	require.NoError(t, err)

	rec, _, _ := runWithCookie(t, j, UserAccessTokenCookie, tok, "GET", "/api/v1/user/me", nil)

	assert.NotContains(t, cookiesOf(rec), UserAccessTokenCookie, "还很新鲜时不该每个请求都下发 Set-Cookie")
}

func TestSlidingRenewal_BearerSessionsAreNotRenewed(t *testing.T) {
	j := newShortLivedJWT()
	tok, err := j.Generate(4)
	require.NoError(t, err)

	mw := IdentityMiddleware(j)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req := httptest.NewRequest("GET", "/api/v1/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)

	assert.NotContains(t, cookiesOf(rec), UserAccessTokenCookie,
		"app 等非浏览器客户端不该被塞 cookie")
}

func TestSlidingRenewal_ReusesTheSameCookieName(t *testing.T) {
	j := newShortLivedJWT()
	tok, err := j.GenerateAdmin(1)
	require.NoError(t, err)

	rec, _, seenAdmin := runWithCookie(t, j, AdminAccessTokenCookie, tok, "GET", "/api/v1/admin/list", nil)

	cookies := cookiesOf(rec)
	require.Contains(t, cookies, AdminAccessTokenCookie)
	assert.NotContains(t, cookies, UserAccessTokenCookie, "不得顺手把用户 cookie 也改写掉")
	assert.Equal(t, "1", seenAdmin)
}

// ---- 刷新令牌不能当访问令牌用 ----

func TestRefreshTokenRejectedWhenUsedAsAccessToken(t *testing.T) {
	j := NewJWT(mwSecret)
	refresh, err := j.GenerateRefresh(4)
	require.NoError(t, err)

	_, reached, seenID := runWithCookie(t, j, UserAccessTokenCookie, refresh, "GET", "/api/v1/user/me", nil)

	assert.True(t, reached, "请求仍应放行到业务层")
	assert.Equal(t, "", seenID, "7 天的刷新令牌不得被当成 24h 的访问令牌")
}

func TestRefreshTokenRejectedOverBearer(t *testing.T) {
	j := NewJWT(mwSecret)
	refresh, err := j.GenerateRefresh(4)
	require.NoError(t, err)

	mw := IdentityMiddleware(j)
	var seenID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenID = r.Header.Get("X-User-Id")
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer "+refresh)
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)

	assert.Equal(t, "", seenID)
}

func TestAccessTokenTypStillAccepted(t *testing.T) {
	j := NewJWT(mwSecret)
	tok, err := j.Generate(9)
	require.NoError(t, err)

	claims, err := j.Parse(tok)
	require.NoError(t, err)
	assert.True(t, claims.IsAccess())
	assert.Equal(t, RoleUser, claims.AccessRole())
}

func TestLegacyTokenWithoutTypStillAccepted(t *testing.T) {
	// 兼容：线上还有没有 typ 字段的存量令牌
	c := &Claims{UserId: 4}
	assert.True(t, c.IsAccess())
	assert.Equal(t, RoleUser, c.AccessRole())

	admin := &Claims{UserId: 1, IsAdmin: true}
	assert.True(t, admin.IsAccess())
	assert.Equal(t, RoleAdmin, admin.AccessRole())
}
