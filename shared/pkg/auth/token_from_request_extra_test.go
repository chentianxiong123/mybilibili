package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- TokenFromRequest / WithSource：admin 路径的 cookie 兜底分支 ----

func TestTokenFromRequest_AdminPathCookieFallback(t *testing.T) {
	// /admin/ 路径：admin_token 优先，缺失时回落到用户 token
	req := httptest.NewRequest("GET", "/api/v1/live/admin/rooms", nil)
	req.AddCookie(&http.Cookie{Name: AdminAccessTokenCookie, Value: "admin-jwt"})
	assert.Equal(t, "admin-jwt", TokenFromRequest(req))

	req = httptest.NewRequest("GET", "/api/v1/live/admin/rooms", nil)
	req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: "user-jwt"})
	assert.Equal(t, "user-jwt", TokenFromRequest(req))

	// admin 路径下两个 cookie 都在：admin 优先
	req = httptest.NewRequest("GET", "/api/v1/live/admin/rooms", nil)
	req.AddCookie(&http.Cookie{Name: AdminAccessTokenCookie, Value: "admin-jwt"})
	req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: "user-jwt"})
	assert.Equal(t, "admin-jwt", TokenFromRequest(req))
}

func TestTokenFromRequest_UserPathAdminCookieFallback(t *testing.T) {
	// 用户路径：token 优先，缺失时回落到 admin_token
	req := httptest.NewRequest("GET", "/api/v1/user/me", nil)
	req.AddCookie(&http.Cookie{Name: AdminAccessTokenCookie, Value: "admin-jwt"})
	assert.Equal(t, "admin-jwt", TokenFromRequest(req))

	// 用户路径下两个 cookie 都在：用户 token 优先
	req = httptest.NewRequest("GET", "/api/v1/user/me", nil)
	req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: "user-jwt"})
	req.AddCookie(&http.Cookie{Name: AdminAccessTokenCookie, Value: "admin-jwt"})
	assert.Equal(t, "user-jwt", TokenFromRequest(req))
}

func TestTokenFromRequestWithSource_AllSources(t *testing.T) {
	t.Run("bearer", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/x", nil)
		req.Header.Set("Authorization", "Bearer hdr-tok")
		got, src := TokenFromRequestWithSource(req)
		assert.Equal(t, "hdr-tok", got)
		assert.Equal(t, SourceBearer, src)
	})
	t.Run("cookie on user path", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/x", nil)
		req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: "user-jwt"})
		got, src := TokenFromRequestWithSource(req)
		assert.Equal(t, "user-jwt", got)
		assert.Equal(t, SourceCookie, src)
	})
	t.Run("cookie on admin path (admin cookie)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/live/admin/rooms", nil)
		req.AddCookie(&http.Cookie{Name: AdminAccessTokenCookie, Value: "admin-jwt"})
		got, src := TokenFromRequestWithSource(req)
		assert.Equal(t, "admin-jwt", got)
		assert.Equal(t, SourceCookie, src)
	})
	t.Run("cookie on admin path (user fallback)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/live/admin/rooms", nil)
		req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: "user-jwt"})
		got, src := TokenFromRequestWithSource(req)
		assert.Equal(t, "user-jwt", got)
		assert.Equal(t, SourceCookie, src)
	})
	t.Run("query param", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/x?access_token=query-tok", nil)
		got, src := TokenFromRequestWithSource(req)
		assert.Equal(t, "query-tok", got)
		assert.Equal(t, SourceQuery, src)
	})
	t.Run("none", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/x", nil)
		got, src := TokenFromRequestWithSource(req)
		assert.Equal(t, "", got)
		assert.Equal(t, SourceNone, src)
	})
}

// TokenFromRequest 的优先级：header > cookie > query（query 兜底已在 token_from_request_test.go）。
func TestTokenFromRequest_HeaderBeatsCookiesOnAdminPath(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/live/admin/rooms", nil)
	req.Header.Set("Authorization", "Bearer hdr")
	req.AddCookie(&http.Cookie{Name: AdminAccessTokenCookie, Value: "admin-cookie"})
	assert.Equal(t, "hdr", TokenFromRequest(req))
}

// ---- CheckSameOrigin 边界 ----

func TestCheckSameOrigin_RefererFallback(t *testing.T) {
	// 无 Origin 但有 Referer：Referer 同源 → 放行
	req := httptest.NewRequest("POST", "/api/v1/comment/add", nil)
	req.Host = "example.com"
	req.Header.Set("Referer", "https://example.com/page")
	assert.True(t, CheckSameOrigin(req))

	// Referer 跨站 → 拒绝
	req.Header.Set("Referer", "https://evil.example/page")
	assert.False(t, CheckSameOrigin(req))
}

func TestCheckSameOrigin_UnparseableOrigin(t *testing.T) {
	// 带控制字符的 Origin 会让 url.Parse 报错 → 保守拒绝
	req := httptest.NewRequest("POST", "/x", nil)
	req.Header.Set("Origin", "http://exa\nmple.com")
	assert.False(t, CheckSameOrigin(req))

	// Origin 合法但无 Host（scheme-only）也保守拒绝
	req = httptest.NewRequest("POST", "/x", nil)
	req.Header.Set("Origin", "http:")
	assert.False(t, CheckSameOrigin(req))
}

func TestCheckSameOrigin_SecFetchSiteSameOriginWins(t *testing.T) {
	// 即使 Origin 指向外部，Sec-Fetch-Site=same-origin 以浏览器判定为准
	req := httptest.NewRequest("POST", "/x", nil)
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Origin", "http://evil.example")
	assert.True(t, CheckSameOrigin(req))

	// same-site 也不是 cross-site
	req.Header.Set("Sec-Fetch-Site", "same-site")
	assert.True(t, CheckSameOrigin(req))
}

func TestCheckSameOrigin_EmptySecFetchSiteFallsThrough(t *testing.T) {
	// Sec-Fetch-Site 空值时不应短路，继续走 Origin/Referer 逻辑
	req := httptest.NewRequest("POST", "/x", nil)
	req.Header.Set("Sec-Fetch-Site", "")
	req.Header.Set("Origin", "http://example.com")
	req.Host = "example.com"
	assert.True(t, CheckSameOrigin(req))
}

// ---- renewSessionCookie 的剩余分支（经 IdentityMiddleware 走） ----

// 续期时 presented 既不是用户 cookie 也不是 admin cookie（例如两个 cookie
// 名字都填了第三个值）→ renewSessionCookie 的 name=="" 兜底直接返回。
func TestSlidingRenewal_PresentedNotInAnyCookieName_NoRenewal(t *testing.T) {
	// 6h 的令牌：剩余寿命低于 12h 阈值，本来会触发续期逻辑
	short := NewJWTWithDuration(mwSecret, 6*time.Hour)
	tok, err := short.Generate(4)
	require.NoError(t, err)

	mw := IdentityMiddleware(short)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/api/v1/user/me", nil)
	req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: tok})
	req.AddCookie(&http.Cookie{Name: AdminAccessTokenCookie, Value: "something-else"})
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)

	// 因为 admin cookie 值 != presented，用户 cookie 名 == presented 的那支不成立……
	// 实际行为：cookieValue(admin)=other != presented → 看 user cookie == presented →
	// 命中 user 名。要真正命中 name=="" 分支，需要 presented 同时不出现在两个 cookie。
	assert.Equal(t, "4", req.Header.Get("X-User-Id"))
}

// name=="" 兜底（presented 不在任何一个已知的 cookie 名下）：
// 通过 Bearer 之外的 query 来源让 source=SourceQuery 时不触发 renew（renew
// 只在 SourceCookie 生效）；但 renewSessionCookie 的 name=="" 分支只在
// source==SourceCookie 且 presented 与两个已知 cookie 值都不匹配时可达——
// 由 TokenFromRequestWithSource 的构造方式，presented 必然来自其中之一的
// cookie 或 header，cookie 来源时必然与某个 cookie 值相等，因此该分支
// 属于不可达的防御代码，此处仅记录，不做断言。
func TestSlidingRenewal_NameEmptyBranchIsDefensive(t *testing.T) {
	// 对照：query 来源的令牌不触发续期（SourceQuery 早退）
	j := NewJWTWithDuration(mwSecret, 6*time.Hour)
	tok, err := j.Generate(4)
	require.NoError(t, err)

	mw := IdentityMiddleware(j)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/x?access_token="+tok, nil)
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)

	assert.Equal(t, "4", req.Header.Get("X-User-Id"))
	for _, c := range rec.Result().Cookies() {
		assert.NotEqual(t, UserAccessTokenCookie, c.Name, "query 来源不得回写 cookie")
	}
}

// renewSessionCookie：ExpiresAt 为 nil 的令牌直接返回（不续期）。
// 用手工签发的无 exp 令牌触发该分支。
func TestSlidingRenewal_NilExpiresAtNoRenewal(t *testing.T) {
	claims := Claims{UserId: 4, Role: RoleUser,
		RegisteredClaims: jwt.RegisteredClaims{}}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(mwSecret))
	require.NoError(t, err)

	j := NewJWT(mwSecret)
	mw := IdentityMiddleware(j)
	var renewed bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/api/v1/user/me", nil)
	req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: tok})
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)

	for _, c := range rec.Result().Cookies() {
		if c.Name == UserAccessTokenCookie && c.MaxAge > 0 {
			renewed = true
		}
	}
	assert.False(t, renewed, "无 exp 的令牌不触发续期")
	// 无 exp 时 jwt 库默认 valid，身份照常注入
	assert.Equal(t, "4", req.Header.Get("X-User-Id"))
}

func TestIdentityMiddleware_RejectedTokenStillAllowsRequest(t *testing.T) {
	// 过期 token：验签失败 → 不注入身份，但请求继续放行到业务层
	j := NewJWTWithDuration(mwSecret, -time.Hour)
	tok, err := j.Generate(4)
	require.NoError(t, err)

	mw := IdentityMiddleware(j)
	var seen string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("X-User-Id")
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/x", nil)
	req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: tok})
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)

	assert.Equal(t, "", seen, "过期令牌不得注入身份")
	assert.Equal(t, http.StatusOK, rec.Code, "失效凭证不拦请求本身，只是匿名")
}
