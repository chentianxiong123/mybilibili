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

func TestIsAdminPath_SubtitleAndAI(t *testing.T) {
	// 字幕 admin 操作：approve/reject/set-default/pending/videos/scan/import-*
	for _, p := range []string{
		"/api/v1/subtitle",
		"/api/v1/subtitle/pending",
		"/api/v1/subtitle/videos",
		"/api/v1/subtitle/set-default",
		"/api/v1/subtitle/import-srt",
		"/api/v1/subtitle/import-system",
		"/api/v1/subtitle/scan/11",
		"/api/v1/subtitle/123/approve",
		"/api/v1/subtitle/456/reject",
		"/api/v1/subtitle/789/preview",
		"/api/v1/subtitle/321/set-default",
		"/api/v1/subtitle/1", // DELETE
	} {
		assert.True(t, IsAdminPath(p), "%s 必须被后台门禁覆盖", p)
	}
	// 公开例外：播放器读 + 用户上传 + work 内部 generate
	for _, p := range []string{
		"/api/v1/subtitle/video",
		"/api/v1/subtitle/video/123",
		"/api/v1/subtitle/video/10/zh-CN",
		"/api/v1/subtitle/upload",
		"/api/v1/subtitle/upload-srt",
		"/api/v1/subtitle/generate",
	} {
		assert.False(t, IsAdminPath(p), "%s 必须保持公开", p)
	}
	// AI 客服转接 + AI 审核（dead 但暴露）：必须关门
	assert.True(t, IsAdminPath("/api/v1/ai/customer/transfer"))
	for _, p := range []string{
		"/api/v1/ai/review",
		"/api/v1/ai/review/content",
		"/api/v1/ai/review/comment",
		"/api/v1/ai/review/reply",
		"/api/v1/ai/review/report",
	} {
		assert.True(t, IsAdminPath(p), "%s 必须被后台门禁覆盖", p)
	}
	// 临近的 user-self 端点不能误伤
	for _, p := range []string{
		"/api/v1/ai/customer/chat",
		"/api/v1/ai/customer/history/4",
		"/api/v1/ai/summary/generate",
		"/api/v1/ai/summary/check/10",
	} {
		assert.False(t, IsAdminPath(p), "%s 不能误判为后台", p)
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
		"/api/v1/ai/customer/chat",      // 普通用户客服入口
		"/api/v1/ai/customer/history/4", // 普通用户客服入口
		// 注意：/api/v1/ai/customer/transfer 不在这里——是 admin 转接人工坐席，adminOnlyPrefixes 已收
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
