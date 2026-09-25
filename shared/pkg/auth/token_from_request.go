package auth

import "net/http"

// TokenFromRequest 按优先级取访问令牌：
//  1. Authorization: Bearer <token>（app/小程序/脚本等非浏览器客户端）
//  2. Cookie: token=<jwt>（同源浏览器自动携带，Nuxt SSR 与 /api 代理均透传）
//  3. ?access_token=<jwt>（仅 EventSource 等无法自定义请求头的场景）
//
// 两者都没有时返回空串。统一收口点，避免各 handler 各自解析导致遗漏。
//
// 关于第 3 条：EventSource 不能设置请求头，后台的转码实时流
// （/api/v1/video/process/admin/stream）只能靠 URL 携带凭证。代价是 token 会进入
// 访问日志/浏览器历史，因此仅在确有需要时使用；阶段 1 落地 HttpOnly Cookie 后
// EventSource 会自动带 Cookie，届时应删除这一回退。
func TokenFromRequest(r *http.Request) string {
	if t, ok := BearerToken(r.Header.Get("Authorization")); ok {
		return t
	}
	if c, err := r.Cookie("token"); err == nil && c.Value != "" {
		return c.Value
	}
	if t := r.URL.Query().Get("access_token"); t != "" {
		return t
	}
	return ""
}
