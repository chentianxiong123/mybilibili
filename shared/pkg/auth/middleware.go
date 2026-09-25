package auth

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// renewThreshold 剩余有效期低于该值时滑动续期一次。
// 访问令牌 24h，续期后重新回到 24h，因此实际每 12h 才会触发一次。
const renewThreshold = 12 * time.Hour

// IdentityMiddleware 身份中间件（零信任第 1 层）。
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
//
// 另外承担两件与会话相关的事，都只在凭证来自 Cookie（即浏览器会话）时生效：
//   - CSRF：跨站发起的写操作直接 403（Bearer 无法被跨站设置，不受此影响）
//   - 滑动续期：临近过期时换发新令牌并回写 Set-Cookie，让"活跃即不失效"
func IdentityMiddleware(j *JWT) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1) 丢弃客户端带来的不可信身份头
			r.Header.Del("X-User-Id")
			r.Header.Del("X-Admin-Id")
			r.Header.Del("X-User-Role")

			// 2) 只在验签通过时按 claims 重建身份；无凭证/凭证无效 → 身份为空
			if j != nil {
				if tok, source := TokenFromRequestWithSource(r); tok != "" {
					if claims, err := j.Parse(tok); err == nil && claims != nil && claims.IsAccess() &&
						!isRevoked(r.Context(), claims) {
						// 2a) 浏览器会话的跨站写操作：凭证本身没问题，但请求不该被允许
						if source == SourceCookie && unsafeMethod(r.Method) && !CheckSameOrigin(r) {
							writeJSONError(w, http.StatusForbidden, "cross-site request rejected")
							return
						}

						r.Header.Set("X-User-Id", strconv.FormatInt(claims.UserId, 10))
						if claims.IsAdmin {
							r.Header.Set("X-Admin-Id", strconv.FormatInt(claims.UserId, 10))
							r.Header.Set("X-User-Role", RoleAdmin)
						} else {
							r.Header.Set("X-User-Role", RoleUser)
						}

						// 2b) 滑动续期：活跃会话不会因为 24h 到期而被踢下线
						if source == SourceCookie {
							renewSessionCookie(w, r, j, tok, claims)
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// renewSessionCookie 临近过期时换发访问令牌，写回原来那一个 cookie。
// 只在剩余有效期不足 renewThreshold 时触发；换发后有效期回到 24h，
// 不会每个请求都下发 Set-Cookie。
func renewSessionCookie(w http.ResponseWriter, r *http.Request, j *JWT, presented string, claims *Claims) {
	if claims.ExpiresAt == nil {
		return
	}
	if time.Until(claims.ExpiresAt.Time) > renewThreshold {
		return
	}
	// 精确判断凭证来自哪个 cookie，只续期它，避免把用户/管理员的另一个会话改写掉
	name := ""
	if v := cookieValue(r, AdminAccessTokenCookie); v != presented {
		if v := cookieValue(r, UserAccessTokenCookie); v == presented {
			name = UserAccessTokenCookie
		}
	} else {
		name = AdminAccessTokenCookie
	}
	if name == "" {
		return
	}

	fresh, err := j.GenerateWithRole(claims.UserId, claims.AccessRole())
	if err != nil || fresh == "" {
		return
	}
	SetAccessCookie(w, name, fresh)
}

// isRevoked 该令牌是否已被吊销（登出 / 改密 / 封号时写入的名单）。
//
// 名单读取出错一律**放行**：Redis 抖动不该把全站在线用户踢下线。
// 吊销是尽力而为的安全增强，可用性优先；等恢复后新写入自然生效。
func isRevoked(ctx context.Context, claims *Claims) bool {
	if claims == nil || claims.Jti == "" {
		return false // 存量令牌没有 Jti，无法按张吊销，保持兼容
	}
	st := CurrentRevocationStore()
	if st == nil {
		return false
	}
	revoked, err := st.IsRevoked(ctx, claims.Jti)
	if err != nil {
		return false
	}
	return revoked
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(`{"code":` + strconv.Itoa(code) + `,"message":"` + msg + `","data":null}`))
}
