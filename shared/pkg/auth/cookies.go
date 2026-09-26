package auth

import (
	"net/http"
	"os"
	"strings"
	"time"
)

// Cookie 名称。普通用户与管理员必须分开：
// users.id 与 admin_users.id 是两套独立自增 ID，同名 cookie 会让两种登录态
// 在同一浏览器里互相覆盖，并让管理员 token 被用户路由按 users.id 解释。
const (
	UserAccessTokenCookie  = "token"
	UserRefreshTokenCookie = "refresh_token"
	UserInfoCookie         = "user_info"

	AdminAccessTokenCookie  = "admin_token"
	AdminRefreshTokenCookie = "admin_refresh"
)

// 访问/刷新 cookie 的存活期，与 JWT 有效期对齐。
const (
	accessTokenMaxAge  = 24 * time.Hour
	refreshTokenMaxAge = 30 * 24 * time.Hour
	// user_info 是 JS 能读到的登录态信号，必须活得和 refresh_token 一样久，
	// 否则会话还有效、前端却已经"看着未登录"。
	userInfoMaxAge = 30 * 24 * time.Hour
)

// AddTokens 按需把凭证写进登录/注册/刷新的响应体。
//
// 默认**不写**：响应体是浏览器 JS 读得到的，读得到就能被 XSS 整包偷走，
// 那 HttpOnly 就白做了。凭证一律走 HttpOnly cookie，浏览器三端不传这个开关。
// 原生客户端（flutter 等）没有 cookie jar，需要显式传 {"includeTokens": true}
// 才拿得到——这是一个显式的、知道自己在做什么的取舍，而不是默认就把钥匙递出去。
func AddTokens(body map[string]interface{}, include bool, token, refreshToken string) {
	if !include {
		return
	}
	body["token"] = token
	body["refresh_token"] = refreshToken
}

// CookieSecure 由 COOKIE_SECURE=true 开启。
// 线上经 Traefik websecure(443) 部署时必须开启；dev 走 http，开启会导致
// 浏览器直接丢弃 cookie，所以默认关。
func CookieSecure() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SECURE")))
	return v == "true" || v == "1" || v == "yes"
}

// SetAccessCookie 写 HttpOnly 访问令牌 cookie。
func SetAccessCookie(w http.ResponseWriter, name, token string) {
	setCookie(w, name, token, int(accessTokenMaxAge/time.Second), true)
}

// SetRefreshCookie 写 HttpOnly 刷新令牌 cookie。
func SetRefreshCookie(w http.ResponseWriter, name, token string) {
	setCookie(w, name, token, int(refreshTokenMaxAge/time.Second), true)
}

// SetPlainCookie 写可被 JS 读取的非凭证 cookie（仅用户展示信息）。
func SetPlainCookie(w http.ResponseWriter, name, value string) {
	setCookie(w, name, value, int(userInfoMaxAge/time.Second), false)
}

// ClearCookie 删除指定 cookie（登出用）。
func ClearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   CookieSecure(),
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	})
}

func setCookie(w http.ResponseWriter, name, value string, maxAge int, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: httpOnly,
		Secure:   CookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearUserSessionCookies 清空普通用户登录态。
func ClearUserSessionCookies(w http.ResponseWriter) {
	ClearCookie(w, UserAccessTokenCookie)
	ClearCookie(w, UserRefreshTokenCookie)
	ClearCookie(w, UserInfoCookie)
}

// ClearAdminSessionCookies 清空管理员登录态。
func ClearAdminSessionCookies(w http.ResponseWriter) {
	ClearCookie(w, AdminAccessTokenCookie)
	ClearCookie(w, AdminRefreshTokenCookie)
}
