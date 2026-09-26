package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAdminPath_BackendPrefix(t *testing.T) {
	// 9 组曾无鉴权的后台路由，全部命中
	for _, p := range []string{
		"/api/v1/live/admin/rooms",
		"/api/v1/moderation/admin/prohibited-words",
		"/api/v1/video/admin/list",
		"/api/v1/video/process/admin/current",
		"/api/v1/support/admin/tickets",
		"/api/v1/search/admin/index/rebuild",
		"/api/v1/admin/transcoders/config",
		"/api/v1/admin/comments/list",
		"/api/v1/message/admin/system/broadcast",
		"/api/v1/manuscript/admin/pending",
		"/api/v1/user/admin/5/status",
		"/api/v1/admin/roles",
	} {
		assert.True(t, IsAdminPath(p), "%s 必须被后台门禁覆盖", p)
	}
}

func TestIsAdminPath_BackendPrefixesWithoutAdminInPath(t *testing.T) {
	for _, p := range []string{
		"/api/v1/ai/configs",
		"/api/v1/ai/configs/1/toggle",
		"/api/v1/ai/bindings/chat",
		"/api/v1/ai/skills/customer-service/route-test",
		"/api/v1/ai/usage/overview",
		"/api/v1/ai/config/test",
		"/api/v1/ai/assistant/send",
		"/api/v1/ai/customer/sessions",
		"/api/v1/ai/customer/sessions/pending/count",
		"/api/v1/statistics",
		"/api/v1/statistics/manuscript/status",
	} {
		assert.True(t, IsAdminPath(p), "%s 必须被后台门禁覆盖", p)
	}
}

func TestIsAdminPath_SearchHotManagement(t *testing.T) {
	// 热搜的 6 条管理口必须关门
	for _, p := range []string{
		"/api/v1/search/hot/keyword",
		"/api/v1/search/hot/rank",
		"/api/v1/search/hot/score",
		"/api/v1/search/hot/score-get",
		"/api/v1/search/hot/delete",
		"/api/v1/search/hot/get",
	} {
		assert.True(t, IsAdminPath(p), "%s 必须被后台门禁覆盖", p)
	}
	// 这 3 条有真实调用方，必须保持匿名
	for _, p := range []string{
		"/api/v1/search/hot",               // web 首页匿名读热搜
		"/api/v1/search/hot/increment",     // 搜索页写热度（普通用户）
		"/api/v1/search/hot/clean-expired", // core 定时清理
	} {
		assert.False(t, IsAdminPath(p), "%s 必须保持公开", p)
	}
}

func TestIsAdminPath_PublicAndUserPaths(t *testing.T) {
	for _, p := range []string{
		"/api/v1/admin/login", // 门卫自己不能要求刷卡
		"/api/v1/user/login",
		"/api/v1/user/me",
		"/api/v1/video/list",
		"/api/v1/search/videos",
		"/api/v1/health",
		"/api/v1/ai/health",
		"/api/v1/ai/customer/chat",         // 普通用户客服入口
		"/api/v1/ai/customer/history/4",    // 普通用户客服入口
		"/api/v1/ai/customer/transfer",     // 普通用户客服入口
		"/api/v1/ai/summary/generate",      // work 内部编排，不带凭证
		"/api/v1/search/hot/clean-expired", // core 定时任务，不带凭证
		"/api/v1/work/health",
	} {
		assert.False(t, IsAdminPath(p), "%s 不得被误伤", p)
	}
}

func TestIsAdminPath_PrefixDoesNotOverMatch(t *testing.T) {
	assert.True(t, IsAdminPath("/api/v1/ai/usage"), "精确前缀命中")
	assert.False(t, IsAdminPath("/api/v1/ai/usagex"), "前缀不能越过路径边界")
	assert.False(t, IsAdminPath("/api/v1/statistics-foo"), "前缀不能越界匹配")
	assert.False(t, IsAdminPath("/api/v1/ai/config"), "非完整前缀不匹配")
}

func TestAdminPathGuard_AnonymousRejected(t *testing.T) {
	var reached bool
	guard := AdminPathGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/live/admin/rooms", nil)
	rec := httptest.NewRecorder()
	guard.ServeHTTP(rec, req)

	assert.False(t, reached, "未授权请求不得进入业务 handler")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAdminPathGuard_AdminIdentityAllowed(t *testing.T) {
	var reached bool
	guard := AdminPathGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/live/admin/rooms", nil)
	req.Header.Set("X-Admin-Id", "1") // IdentityMiddleware 验签后注入
	rec := httptest.NewRecorder()
	guard.ServeHTTP(rec, req)

	assert.True(t, reached)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminPathGuard_UserPathPassesThrough(t *testing.T) {
	var reached bool
	guard := AdminPathGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/user/me", nil)
	rec := httptest.NewRecorder()
	guard.ServeHTTP(rec, req)

	assert.True(t, reached, "用户接口不得被后台门禁影响")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminPathGuard_AnonymousCanStillLogin(t *testing.T) {
	var reached bool
	guard := AdminPathGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/v1/admin/login", nil)
	rec := httptest.NewRecorder()
	guard.ServeHTTP(rec, req)

	assert.True(t, reached, "登录接口必须保持匿名可访问")
	assert.Equal(t, http.StatusOK, rec.Code)
}
