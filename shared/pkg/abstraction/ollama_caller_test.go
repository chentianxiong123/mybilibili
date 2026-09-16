package abstraction

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ollamaStub struct {
	*httptest.Server
	chatCalls int
}

func newOllamaStub(t *testing.T) *ollamaStub {
	t.Helper()
	s := &ollamaStub{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(200)
			return
		}
		if r.URL.Path == "/v1/chat/completions" {
			s.chatCalls++
			body := map[string]any{
				"choices": []map[string]any{
					{"message": map[string]string{"content": "  AI 摘要内容  "}},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(body)
			return
		}
		w.WriteHeader(404)
	}))
	t.Cleanup(s.Close)
	return s
}

func newOllamaCallerWithURL(t *testing.T, url string) *ollamaCaller {
	t.Helper()
	t.Setenv("OLLAMA_URL", url)
	t.Setenv("OLLAMA_MODEL", "test-model")
	ollamaFallback = false
	t.Cleanup(func() { ollamaFallback = false })
	return newOllamaCaller()
}

func TestOllamaCaller_CheckOllama(t *testing.T) {
	stub := newOllamaStub(t)
	c := newOllamaCallerWithURL(t, stub.URL)
	assert.True(t, c.checkOllama())

	c2 := newOllamaCallerWithURL(t, "http://127.0.0.1:1")
	assert.False(t, c2.checkOllama())
}

func TestOllamaCaller_Summary(t *testing.T) {
	stub := newOllamaStub(t)
	c := newOllamaCallerWithURL(t, stub.URL)
	ctx := context.Background()

	var resp struct{ Summary string }
	err := c.Call(ctx, "ai", "Summary", map[string]interface{}{"video_id": 42.0}, &resp)
	require.NoError(t, err)
	assert.Equal(t, "AI 摘要内容", resp.Summary)
	assert.Equal(t, 1, stub.chatCalls)
}

func TestOllamaCaller_SummaryViaMap(t *testing.T) {
	stub := newOllamaStub(t)
	c := newOllamaCallerWithURL(t, stub.URL)

	resp := map[string]interface{}{}
	err := c.Call(context.Background(), "ai", "Summary", map[string]interface{}{"video_id": 1.0}, &resp)
	require.NoError(t, err)
	assert.Equal(t, "AI 摘要内容", resp["summary"])
}

func TestOllamaCaller_FallbackWhenUnreachable(t *testing.T) {
	c := newOllamaCallerWithURL(t, "http://127.0.0.1:1")

	var resp struct{ Summary string }
	err := c.Call(context.Background(), "ai", "Summary", map[string]interface{}{"video_id": 7.0}, &resp)
	require.NoError(t, err)
	assert.Contains(t, resp.Summary, "placeholder")
	assert.True(t, ollamaFallback)
}

func TestOllamaCaller_ModerateParsed(t *testing.T) {
	stub := newOllamaStub(t)
	ollamaFallback = false
	t.Cleanup(func() { ollamaFallback = false })

	// 覆盖 stub 的 chat 返回，改为违规 JSON
	stub.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(200)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": `好的，判断如下：{"passed":false,"reason":"含敏感词"}`}},
			},
		})
	})

	c := newOllamaCallerWithURL(t, stub.URL)
	resp := map[string]interface{}{}
	err := c.Call(context.Background(), "ai", "Moderate", map[string]interface{}{
		"content": "测试内容", "scene": "comment",
	}, &resp)
	require.NoError(t, err)
	assert.Equal(t, false, resp["passed"])
	assert.Equal(t, "含敏感词", resp["reason"])
}

func TestOllamaCaller_CustomerChat(t *testing.T) {
	stub := newOllamaStub(t)
	c := newOllamaCallerWithURL(t, stub.URL)

	var resp struct{ Reply string }
	err := c.Call(context.Background(), "ai", "CustomerChat", map[string]interface{}{"content": "你好"}, &resp)
	require.NoError(t, err)
	assert.Equal(t, "AI 摘要内容", resp.Reply)
}

func TestOllamaCaller_CheckSummaryAndHistory(t *testing.T) {
	stub := newOllamaStub(t)
	c := newOllamaCallerWithURL(t, stub.URL)

	var chk struct{ HasSummary bool }
	err := c.Call(context.Background(), "ai", "CheckSummary", map[string]interface{}{}, &chk)
	require.NoError(t, err)
	assert.False(t, chk.HasSummary)

	hist := []map[string]interface{}{{"old": "x"}}
	err = c.Call(context.Background(), "ai", "CustomerHistory", map[string]interface{}{}, &hist)
	require.NoError(t, err)
	assert.Len(t, hist, 0)

	err = c.Call(context.Background(), "ai", "CustomerTransfer", map[string]interface{}{}, nil)
	require.NoError(t, err)
}

func TestOllamaCaller_UnknownMethod(t *testing.T) {
	stub := newOllamaStub(t)
	c := newOllamaCallerWithURL(t, stub.URL)
	err := c.Call(context.Background(), "ai", "NoSuchMethod", map[string]interface{}{}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown method")
}

func TestOllamaCaller_ChatUnreachable(t *testing.T) {
	c := newOllamaCallerWithURL(t, "http://127.0.0.1:1")
	ollamaFallback = true
	t.Cleanup(func() { ollamaFallback = false })

	_, err := c.chat(context.Background(), "sys", "user")
	assert.Error(t, err)
}

func TestOllamaCaller_ChatBadJSON(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(200)
			return
		}
		_, _ = w.Write([]byte("not-json"))
	}))
	defer stub.Close()
	c := newOllamaCallerWithURL(t, stub.URL)

	_, err := c.chat(context.Background(), "sys", "user")
	assert.Error(t, err)
}

func TestOllamaCaller_ChatNoChoices(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(200)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{}})
	}))
	defer stub.Close()
	c := newOllamaCallerWithURL(t, stub.URL)

	_, err := c.chat(context.Background(), "sys", "user")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no choices")
}

func TestOllamaCaller_CallStream(t *testing.T) {
	c := newOllamaCallerWithURL(t, "http://127.0.0.1:1")
	ch, err := c.CallStream(context.Background(), "ai", "m", nil)
	require.NoError(t, err)
	msg, ok := <-ch
	assert.True(t, ok)
	assert.NotEmpty(t, msg)
	assert.NoError(t, c.Close())
}

func TestExtractJSON(t *testing.T) {
	assert.Equal(t, `{"a":1}`, extractJSON(`前言 {"a":1} 结尾`))
	assert.Equal(t, `{"a":1}`, extractJSON(`{"a":1}`))
	assert.Equal(t, "no-brace", extractJSON("no-brace"))
	assert.Equal(t, `{"a":"}"}`, extractJSON(`前 {"a":"}"} 尾`))
	assert.Equal(t, "}", extractJSON("}"))
}

func TestOllamaCaller_TimeoutSet(t *testing.T) {
	c := newOllamaCallerWithURL(t, "http://127.0.0.1:1")
	assert.NotNil(t, c.httpClient)
	assert.Equal(t, 120*time.Second, c.httpClient.Timeout)
}

func TestOllamaCaller_DefaultEnv(t *testing.T) {
	os.Unsetenv("OLLAMA_URL")
	os.Unsetenv("OLLAMA_MODEL")
	ollamaFallback = false
	t.Cleanup(func() { ollamaFallback = false })

	c := newOllamaCaller()
	assert.Equal(t, "http://localhost:11434", c.baseURL)
	assert.Equal(t, "deepseek-r1:8b", c.model)
	assert.True(t, strings.HasSuffix(c.baseURL, ":11434"))
}
