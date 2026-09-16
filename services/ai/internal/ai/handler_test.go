package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCaller 实现 abstraction.ServiceCaller，用于注入 Summary/Review 服务。
type mockCaller struct {
	summary  string
	hasSum   bool
	moderate map[string]interface{}
}

func (m *mockCaller) Call(ctx context.Context, target, method string, req, resp interface{}) error {
	switch method {
	case "Summary":
		if s, ok := resp.(*struct {
			Summary string `json:"summary"`
		}); ok {
			s.Summary = m.summary
		}
	case "CheckSummary":
		if s, ok := resp.(*struct {
			HasSummary bool `json:"has_summary"`
		}); ok {
			s.HasSummary = m.hasSum
		}
	case "Moderate":
		if s, ok := resp.(*map[string]interface{}); ok {
			*s = m.moderate
		}
	}
	return nil
}

func (m *mockCaller) CallStream(ctx context.Context, target, method string, req interface{}) (<-chan []byte, error) {
	ch := make(chan []byte, 1)
	ch <- []byte(m.summary)
	close(ch)
	return ch, nil
}

func (m *mockCaller) Close() error { return nil }

// newTestHandler 构造 AI Handler + sqlmock DB。
func newTestHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)
	t.Cleanup(func() { db.Close() })
	return h, mock
}

func doAI(t *testing.T, mux http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestHandleSummary_200(t *testing.T) {
	h, _ := newTestHandler(t)
	h.WithSummary(NewSummaryService(&mockCaller{summary: "这是一段AI生成的中文摘要。"}))

	mux := http.NewServeMux()
	h.Register(mux)

	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/summary/123", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code    int               `json:"code"`
		Message string            `json:"message"`
		Data    map[string]string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 200, resp.Code)
	assert.Contains(t, resp.Data["summary"], "AI生成")
}

func TestHandleReviewContent_200(t *testing.T) {
	_, _ = newTestHandler(t)
	review := NewReviewService(&mockCaller{moderate: map[string]interface{}{"passed": true, "reason": ""}})
	chatH := NewAIChatHandler(review, nil)

	mux := http.NewServeMux()
	chatH.Register(mux)

	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/review/content?content=helloworld", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var tags []string
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &tags))
	assert.Empty(t, tags)
}

func TestHandleReviewComment_200(t *testing.T) {
	_, _ = newTestHandler(t)
	review := NewReviewService(&mockCaller{moderate: map[string]interface{}{"passed": true, "reason": ""}})
	chatH := NewAIChatHandler(review, nil)

	mux := http.NewServeMux()
	chatH.Register(mux)

	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/review/comment?content=正常评论内容", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Passed bool   `json:"passed"`
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp.Passed)
}

func TestHandleConfigs_200(t *testing.T) {
	h, mock := newTestHandler(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT id, name, COALESCE\(type,'LLM'\)`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "type", "base_url", "api_key", "model", "max_tokens", "temperature", "enabled", "extra_config", "created_at",
		}).AddRow(1, "默认LLM", "LLM", "http://localhost", "sk-1234567890abcdef", "gpt-4o", 4096, 0.7, 1, "", now))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/configs", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int          `json:"code"`
		Data []*ApiConfig `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "默认LLM", resp.Data[0].Name)
	// APIKey 应被掩码
	assert.NotEqual(t, "sk-1234567890abcdef", resp.Data[0].APIKey)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSkills_200(t *testing.T) {
	h, mock := newTestHandler(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT id, name, description, system_prompt`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "system_prompt", "few_shot_examples", "type", "enabled", "created_at",
		}).AddRow(1, "内容审核", "审核用户内容", "你是审核员", "", "MODERATION", 1, now))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/skills", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int     `json:"code"`
		Data []*Skill `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "内容审核", resp.Data[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}