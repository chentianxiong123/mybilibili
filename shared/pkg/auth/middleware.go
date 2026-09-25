package auth

import (
	"net/http"
	"strconv"
)

// HTTPMiddlewares 身份中间件（零信任第 1 层）。
//
// 契约：X-User-Id / X-User-Role / X-Admin-Id 是"身份声明"，**只允许由本中间件
// 从已验签的凭证推导**，绝不接受客户端直接传入。
//
// 为什么不能信客户端头：历史上这里短路放行"Traefik forwardAuth 注入的身份头"，
// 但本项目从未配置 forwardAuth（dev compose 与 deploy/k3s 均无），而 6 个 Go 服务
// 的端口都对宿主暴露（8080/8084/8086/8087/8088/8089），攻击者可绕过网关直接
// 伪造 X-User-Id 冒充任意用户（含只带伪造头、完全无凭证的情况）。
//
// 因此顺序固定为：先删除客户端身份头 → 再用 Bearer/cookie 验签结果重建。
// 各服务内网互调（如 core→search 的行为记录）必须改带 Bearer，而不是裸带头。
//
// 将来若真的启用 Traefik forwardAuth：必须先让网关剥离客户端身份头、或改为
// 仅接受可信网段的注入，否则放开会直接退回本漏洞。
func IdentityMiddleware(j *JWT) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1) 丢弃客户端带来的不可信身份头
			r.Header.Del("X-User-Id")
			r.Header.Del("X-Admin-Id")
			r.Header.Del("X-User-Role")

			// 2) 只在验签通过时按 claims 重建身份；无凭证/凭证无效 → 身份为空
			if j != nil {
				if tokenStr := TokenFromRequest(r); tokenStr != "" {
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
			}
			next.ServeHTTP(w, r)
		})
	}
}
