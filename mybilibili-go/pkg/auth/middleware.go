package auth

import (
	"net/http"
	"strconv"
)

// HTTPMiddlewares 网关验签兼容中间件。
//
// 设计目标（零信任第 1 层，网关全权验签）：
//   - 生产流量：Traefik 已通过 forwardAuth 验签并注入 X-User-Id / X-User-Role / X-Admin-Id，
//     下游各服务直接信任这些头，不做重复验签（单签发点，secret 只在网关层）。
//   - 开发/直连（绕过 Traefik）：若请求带 Bearer token，则本地验签兜底注入身份头，
//     保证本地直连每个服务都能拿到身份。
//
// IdentityMiddleware 应挂到所有需要用户身份的 HTTP 服务上，保证无论流量来自
// 网关还是本地直连，handler 读到 X-User-Id 都有意义。
func IdentityMiddleware(j *JWT) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 已有身份头（Traefik 注入或上层已验签）→ 直接放行
			if r.Header.Get("X-User-Id") != "" {
				next.ServeHTTP(w, r)
				return
			}
			// 无身份头但有 Bearer token → 本地验签兜底（仅开发/直连场景）
			if tokenStr, ok := BearerToken(r.Header.Get("Authorization")); ok {
				if claims, err := j.Parse(tokenStr); err == nil && claims != nil {
					r.Header.Set("X-User-Id", strconv.FormatInt(claims.UserId, 10))
					if claims.IsAdmin {
						r.Header.Set("X-Admin-Id", strconv.FormatInt(claims.UserId, 10))
						r.Header.Set("X-User-Role", RoleAdmin)
					} else {
						r.Header.Set("X-User-Role", RoleUser)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}