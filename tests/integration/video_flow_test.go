//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 视频链路：推荐 → 分类 → 单个视频 → 弹幕
func TestVideoFlow_RandomAndCategory(t *testing.T) {
	// 推荐视频
	resp, body := doGet(t, coreURL+"/api/v1/video/random/visitor", "")
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotEmpty(t, body.Data)

	// 分类列表
	resp, body = doGet(t, coreURL+"/api/v1/category", "")
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotEmpty(t, body.Data)
}

// 弹幕链路
func TestDanmakuFlow_SendAndGet(t *testing.T) {
	// 获取弹幕（vid=1）
	resp, _ := doGet(t, msgURL+"/api/v1/danmaku/video/1", "")
	assert.Equal(t, 200, resp.StatusCode)
}