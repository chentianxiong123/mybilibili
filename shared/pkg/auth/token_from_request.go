package auth

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

// TokenFromRequest 按优先级取访问令牌：
//  1. Authorization: Bearer <token>（app/小程序/脚本等非浏览器客户端）
//  2. Cookie —— 按接口类别挑选命名空间，见 cookieForPath
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
	// 后台接口先看 admin_token、用户接口先看 token；缺失时互相兜底，
	// 保证同时登录用户站与后台站的浏览器在任一侧都能通过。
	if IsAdminPath(r.URL.Path) {
		if v := cookieValue(r, AdminAccessTokenCookie); v != "" {
			return v
		}
		if v := cookieValue(r, UserAccessTokenCookie); v != "" {
			return v
		}
	} else {
		if v := cookieValue(r, UserAccessTokenCookie); v != "" {
			return v
		}
		if v := cookieValue(r, AdminAccessTokenCookie); v != "" {
			return v
		}
	}
	if t := r.URL.Query().Get("access_token"); t != "" {
		return t
	}
	return ""
}

// TokenSource 标识凭证来源，用于只对浏览器会话做滑动续期与 CSRF 校验。
type TokenSource int

const (
	SourceNone TokenSource = iota
	SourceBearer
	SourceCookie
	SourceQuery
)

// TokenFromRequestWithSource 与 TokenFromRequest 相同，但额外返回凭证来源。
func TokenFromRequestWithSource(r *http.Request) (string, TokenSource) {
	if t, ok := BearerToken(r.Header.Get("Authorization")); ok {
		return t, SourceBearer
	}
	if IsAdminPath(r.URL.Path) {
		if v := cookieValue(r, AdminAccessTokenCookie); v != "" {
			return v, SourceCookie
		}
		if v := cookieValue(r, UserAccessTokenCookie); v != "" {
			return v, SourceCookie
		}
	} else {
		if v := cookieValue(r, UserAccessTokenCookie); v != "" {
			return v, SourceCookie
		}
		if v := cookieValue(r, AdminAccessTokenCookie); v != "" {
			return v, SourceCookie
		}
	}
	if t := r.URL.Query().Get("access_token"); t != "" {
		return t, SourceQuery
	}
	return "", SourceNone
}

func cookieValue(r *http.Request, name string) string {
	if c, err := r.Cookie(name); err == nil && c.Value != "" {
		return c.Value
	}
	return ""
}

// unsafeMethods 需要 CSRF 防护的动词。
func unsafeMethod(m string) bool {
	switch strings.ToUpper(m) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return false
	}
	return true
}

// CheckSameOrigin 判断浏览器的跨站信息是否表明这是一次跨站写操作。
//
// 判定顺序：
//  1. Sec-Fetch-Site —— 浏览器自己算好的，且不受任何代理改写 Host 的影响，
//     只要存在就以它为准。
//  2. Origin/Referer 与目标主机比对 —— 兜给不发 Sec-Fetch-Site 的老浏览器。
//     注意 Nuxt 的 /api 代理会把 Host 改写成后端地址（localhost:8080），
//     而浏览器的 Origin 仍是页面地址（localhost:3200），因此全等不匹配时
//     退化为只比主机名，避免把同源的开发代理误判成跨站。
//
// curl / 服务间调用三个头都没有，视为非浏览器（本来也无法被诱导发起），
// 放行；只要任一头明确指向外部站点就拒绝。
func CheckSameOrigin(r *http.Request) bool {
	if s := r.Header.Get("Sec-Fetch-Site"); s != "" {
		return s != "cross-site"
	}

	origin := r.Header.Get("Origin")
	if strings.EqualFold(origin, "null") {
		return false
	}
	if origin == "" {
		origin = r.Header.Get("Referer")
	}
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}

	target := r.Header.Get("X-Forwarded-Host")
	if target == "" {
		target = r.Host
	}
	if strings.EqualFold(u.Host, target) {
		return true
	}
	// Host 被代理改写过端口，退化为主机名比较
	return hostOnly(u.Host) != "" && strings.EqualFold(hostOnly(u.Host), hostOnly(target))
}

// hostOnly 去掉主机名里的端口。
func hostOnly(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}
