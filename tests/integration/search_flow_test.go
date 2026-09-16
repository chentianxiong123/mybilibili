//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 搜索链路：热搜 → 视频搜索 → 用户搜索
func TestSearchFlow_HotAndSearch(t *testing.T) {
	// 热搜
	resp, _ := doGet(t, searchURL+"/api/v1/search/hot", "")
	assert.Equal(t, 200, resp.StatusCode)

	// 搜索视频
	resp, _ = doGet(t, searchURL+"/api/v1/search/videos?keyword=test&page=1&pageSize=5", "")
	assert.Equal(t, 200, resp.StatusCode)

	// 搜索建议
	resp, _ = doGet(t, searchURL+"/api/v1/search/suggest?keyword=test&size=5", "")
	assert.Equal(t, 200, resp.StatusCode)
}