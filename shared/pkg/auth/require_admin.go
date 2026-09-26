package auth

import (
	"net/http"
	"strings"
)

// adminPublicPaths 位于后台路径下、但必须保持匿名可访问的例外。
//
// 门卫不能要求先刷卡再刷卡：访问令牌过期时 X-Admin-Id 注入不进来，
// 若刷新口也被默认拒绝，就永远换不出新令牌，登录会变成一张单程票。
// 这两条路径自身就是凭证的签发处，身份由 handler 内部严格校验
// （refresh 要求 typ=admin_refresh + 一次性消费成功），所以公开是安全的。
var adminPublicPaths = []string{
	"/api/v1/admin/login",
	"/api/v1/admin/token/refresh",
}

// adminOnlyPrefixes 是后台专用、但路径里不含 /admin/ 的接口前缀。
// 这些路由原本同样没有任何鉴权（AI 渠道配置/技能/用量、运营统计看板）。
// 只列出「实测仅 admin 前端调用、且无服务间内部调用」的前缀。
//
// 明确排除：
//   - /api/v1/ai/summary/*        —— work 服务内部编排调用，不带凭证
//   - /api/v1/ai/customer/chat|history|transfer —— 普通用户客服入口
//   - /api/v1/search/hot/clean-expired —— core 定时任务内部调用，不带凭证
//   - /api/v1/banner*、/api/v1/category —— 同一路径下用户读、管理员写，需按方法区分
var adminOnlyPrefixes = []string{
	"/api/v1/ai/configs",
	"/api/v1/ai/config/test",
	"/api/v1/ai/bindings",
	"/api/v1/ai/skills",
	"/api/v1/ai/usage",
	"/api/v1/ai/assistant/send",
	"/api/v1/ai/customer/sessions",
	"/api/v1/statistics",
}

// IsAdminPath 判断请求路径是否属于必须持有管理员身份的后台接口。
//
// 规则（默认拒绝）：
//  1. 显式豁免的公开路径（adminPublicPaths）
//  2. 路径含 /admin/ 或以 /admin 结尾 —— 覆盖 live/moderation/video/support/
//     search/work/msg-danmaku/comment 等 9 组后台路由
//  3. 命中 adminOnlyPrefixes —— 覆盖路径里不带 /admin/ 的后台接口
//
// 用路径级默认拒绝，而不是逐条路由包装，是为了新增后台接口时不会漏挂鉴权
// （原始设计依赖的 Traefik forwardAuth 从未配置，逐条包装正是当初漏掉的原因）。
func IsAdminPath(path string) bool {
	for _, p := range adminPublicPaths {
		if path == p {
			return false
		}
	}
	if strings.Contains(path, "/admin/") || strings.HasSuffix(path, "/admin") {
		return true
	}
	for _, p := range adminOnlyPrefixes {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}

// RequireAdmin 要求请求已持有有效管理员身份，否则 401。
//
// 只认 X-Admin-Id：该头由 IdentityMiddleware 在 claims.IsAdmin 为真时注入，
// 客户端伪造的同名头会被 IdentityMiddleware 先删除，所以这里可以直接信任。
// 供需要在中间件之外单独加锁的场景使用。
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Admin-Id") == "" {
			writeUnauthorized(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// AdminPathGuard 后台接口默认拒绝：未持有管理员身份一律 401。
//
// 必须挂在 IdentityMiddleware 内侧，即
//
//	IdentityMiddleware(jwt)(AdminPathGuard(mux))
//
// 保证先验签注入 X-Admin-Id、再判断权限。
//
// 背景：这些后台接口的原始设计注释写着「鉴权由 Traefik forwardAuth 统一处理」，
// 但 dev compose 与 deploy/k3s 都没有配置 forwardAuth，等于门卫从未上岗——
// 实测不登录即可读写审核禁词、直播房间、客服工单、搜索索引、转码节点。
func AdminPathGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsAdminPath(r.URL.Path) && r.Header.Get("X-Admin-Id") == "" {
			writeUnauthorized(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"code":401,"message":"unauthorized","data":null}`))
}
