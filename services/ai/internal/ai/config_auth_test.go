package ai

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfigCreate_NoAuth 钉住 FINDING-024（已修复）：
// POST /api/v1/ai/configs 无 admin 身份 → 401，DB 不执行 INSERT。
func TestConfigCreate_NoAuth(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/configs",
		`{"name":"test","type":"LLM","base_url":"http://evil.example","api_key":"sk-123","model":"gpt-4","max_tokens":1000,"temperature":0.7}`)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "FINDING-024 修复：无鉴权创建 AI 配置应 401")
}
