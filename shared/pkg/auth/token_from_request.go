package auth

import "net/http"

// TokenFromRequest 按优先级取访问令牌：
//  1. Authorization: Bearer <token>（app/小程序/脚本等非浏览器客户端）
//  2. Cookie: token=<jwt>（同源浏览器自动携带，Nuxt SSR 与 /api 代理均透传）
//
// 两者都没有时返回空串。统一收口点，避免各 handler 各自解析导致遗漏。
func TokenFromRequest(r *http.Request) string {
	if t, ok := BearerToken(r.Header.Get("Authorization")); ok {
		return t
	}
	if c, err := r.Cookie("token"); err == nil && c.Value != "" {
		return c.Value
	}
	return ""
}
