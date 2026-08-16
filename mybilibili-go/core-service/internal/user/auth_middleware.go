package user

import (
	"net/http"
	"strconv"

	"mybilibili/pkg/auth"
)

// AuthMiddleware 信任由 Traefik 注入的身份头（零信任第 1 层：网关全权验签）。
// - Traefik 已通过 forwardAuth 验签并注入 X-User-Id / X-User-Role / X-Admin-Id。
// - 本中间件不再自验签，仅透传已有请求头，保留下游 handler 无感兼容（读 X-User-Id）。
// - 开发环境（无 Traefik 直连 core）时，若请求带 Bearer token 则本地验签兜底注入，
//   保证本地跑通；生产流量走 Traefik 时该分支不会命中。
func AuthMiddleware(j *auth.JWT) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 已有身份头（Traefik 注入或 Flutter 双发）→ 直接放行
			if r.Header.Get("X-User-Id") != "" {
				next.ServeHTTP(w, r)
				return
			}
			// 无身份头但有 Bearer token → 本地验签兜底（仅开发/直连场景）
			if tokenStr, ok := auth.BearerToken(r.Header.Get("Authorization")); ok {
				if claims, err := j.Parse(tokenStr); err == nil && claims != nil {
					r.Header.Set("X-User-Id", itoa(claims.UserId))
					if claims.IsAdmin {
						r.Header.Set("X-Admin-Id", itoa(claims.UserId))
						r.Header.Set("X-User-Role", auth.RoleAdmin)
					} else {
						r.Header.Set("X-User-Role", auth.RoleUser)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func itoa(v int64) string {
	if v == 0 {
		return ""
	}
	return strconv.FormatInt(v, 10)
}