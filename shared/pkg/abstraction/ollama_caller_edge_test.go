package abstraction

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 覆盖 ollama_caller.go 的 CustomerChat 分支:
// - AI 不可达时的 fallback 文案
// - resp 为 *map[string]interface{} 时的回填

func TestOllamaCaller_CustomerChat_FallbackAndMapResp(t *testing.T) {
	c := newOllamaCallerWithURL(t, "http://127.0.0.1:1")

	resp := map[string]interface{}{}
	err := c.Call(context.Background(), "ai", "CustomerChat", map[string]interface{}{"content": "你好"}, &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["reply"], "客服助手")
}
// baseURL 非法 → http.NewRequestWithContext 失败
func TestOllamaCaller_Chat_InvalidBaseURL(t *testing.T) {
	c := &ollamaCaller{baseURL: "://bad-url", httpClient: &http.Client{Timeout: time.Second}}
	_, err := c.chat(context.Background(), "sys", "user")
	require.Error(t, err)
}
