//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Nuxt 代理链路：验证 Nuxt 反代到 Go 后端各路由可用
func TestProxyFlow_NuxtToBackend(t *testing.T) {
	// 健康检查
	resp, _ := doGet(t, nuxtURL+"/api/health", "")
	assert.Equal(t, 200, resp.StatusCode)

	// 推荐视频
	resp, _ = doGet(t, nuxtURL+"/api/video/random/visitor", "")
	assert.Equal(t, 200, resp.StatusCode)

	// 分类
	resp, _ = doGet(t, nuxtURL+"/api/category/getall", "")
	assert.Equal(t, 200, resp.StatusCode)

	// 搜索
	resp, _ = doGet(t, nuxtURL+"/api/search/video/only-pass?keyword=test&page=1", "")
	assert.Equal(t, 200, resp.StatusCode)

	// 登录（错误密码应被拒绝）
	resp, _ := doPost(t, nuxtURL+"/api/user/account/login", map[string]string{
		"username": "admin",
		"password": "wrong-password",
	})
	assert.Contains(t, []int{401, 404}, resp.StatusCode)
}