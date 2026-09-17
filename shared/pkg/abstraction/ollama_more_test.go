package abstraction

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOllamaCaller_doCall_NonMapReq(t *testing.T) {
	stub := newOllamaStub(t)
	c := newOllamaCallerWithURL(t, stub.URL)

	// req 用结构体而不是 map, 触发 json.Marshal/Unmarshal 路径
	type reqStruct struct {
		VideoID float64 `json:"video_id"`
	}
	resp := map[string]interface{}{}
	err := c.Call(context.Background(), "ai", "Summary", reqStruct{VideoID: 99}, &resp)
	require.NoError(t, err)
	assert.Equal(t, "AI 摘要内容", resp["summary"])
}

func TestOllamaCaller_Moderate_JSONUnparseable(t *testing.T) {
	stub := newOllamaStub(t)
	// 覆盖 stub 让 chat 返回无法解析为 JSON 的字符串
	stub.Server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(200)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "非 JSON 字符串"}},
			},
		})
	})

	c := newOllamaCallerWithURL(t, stub.URL)
	resp := map[string]interface{}{}
	err := c.Call(context.Background(), "ai", "Moderate", map[string]interface{}{
		"content": "x", "scene": "comment",
	}, &resp)
	require.NoError(t, err)
	// 解析失败时 passed 默认 true
	assert.Equal(t, true, resp["passed"])
	assert.Equal(t, "", resp["reason"])
}

func TestOllamaCaller_Summary_ViaOtherRespTypes(t *testing.T) {
	stub := newOllamaStub(t)
	c := newOllamaCallerWithURL(t, stub.URL)

	// resp 既不是 struct 也不是 *map[string]interface{} — 走 default 路径
	resp := "ignored-string"
	err := c.Call(context.Background(), "ai", "Summary", map[string]interface{}{"video_id": 1}, &resp)
	assert.NoError(t, err)
}

func TestOllamaCaller_ChatFallbackTimeout(t *testing.T) {
	// 启动一个永不响应的 server, 验证 chat 的 ctx 超时路径
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// hang
		<-r.Context().Done()
	}))
	defer srv.Close()
	c := newOllamaCallerWithURL(t, srv.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.chat(ctx, "sys", "user")
	assert.Error(t, err)
}

func TestExtractJSON_BracketEdgeCases(t *testing.T) {
	// 嵌套花括号
	assert.Equal(t, `{"a":{"b":1}}`, extractJSON(`prefix {"a":{"b":1}} suffix`))
	// 引号转义
	assert.Equal(t, `{"a":"\""}`, extractJSON(`p {"a":"\""} s`))
	// 多个花括号取最后一个合法 JSON
	got := extractJSON(`{"x":1} junk {"y":2}`)
	assert.Contains(t, got, `"y"`)
}