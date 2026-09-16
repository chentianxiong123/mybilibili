package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

func TestHandleSummary_400(t *testing.T) {
	h, _ := newTestHandler(t)
	h.WithSummary(NewSummaryService(&mockCaller{summary: "摘要内容"}))

	mux := http.NewServeMux()
	h.Register(mux)

	// 缺少 video_id → 400
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/summary/generate", `{"manuscript_id":1}`)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(400), resp["code"])
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

// ==================== configs POST ====================

func TestHandleConfigs_POST_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`INSERT INTO ai_api_configs`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/configs",
		`{"name":"test","type":"LLM","base_url":"http://localhost","api_key":"sk-xxx","model":"gpt-4","max_tokens":2048,"temperature":0.5}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, "ok", resp.Data["status"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== configs/{id} GET/PUT/DELETE ====================

func TestHandleConfigByID_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT id, name, COALESCE\(type,'LLM'\), base_url, api_key`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "type", "base_url", "api_key", "model", "max_tokens", "temperature", "enabled", "extra_config", "created_at",
		}).AddRow(1, "cfg1", "LLM", "http://localhost", "sk-1234567890abcdef", "gpt-4", 4096, 0.7, 1, "", now))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/configs/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int        `json:"code"`
		Data *ApiConfig `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "cfg1", resp.Data.Name)
	assert.NotEqual(t, "sk-1234567890abcdef", resp.Data.APIKey)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleConfigByID_PUT_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE ai_api_configs SET`).WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPut, "/api/v1/ai/configs/1",
		`{"name":"updated","type":"LLM","model":"gpt-4o"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp.Data["status"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleConfigByID_DELETE_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`DELETE FROM ai_api_configs WHERE id`).WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodDelete, "/api/v1/ai/configs/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp.Data["status"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== configs/{id}/toggle ====================

func TestHandleConfigByID_Toggle(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE ai_api_configs SET enabled`).WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPut, "/api/v1/ai/configs/1/toggle", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp.Data["status"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== configs/{id}/bind ====================

func TestHandleConfigByID_Bind(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`INSERT INTO ai_bindings`).WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/configs/chat/bind",
		`{"config_id":1}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp.Data["status"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== configs/types ====================

func TestHandleConfigTypes(t *testing.T) {
	h, _ := newTestHandler(t)

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/configs/types", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int      `json:"code"`
		Data []string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Contains(t, resp.Data, "LLM")
	assert.Contains(t, resp.Data, "ASR")
}

// ==================== configs/features ====================

func TestHandleConfigFeatures(t *testing.T) {
	h, _ := newTestHandler(t)

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/configs/features", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                      `json:"code"`
		Data []map[string]string      `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Greater(t, len(resp.Data), 0)
	assert.Equal(t, "chat", resp.Data[0]["feature"])
}

// ==================== bindings GET/POST ====================

func TestHandleBindings_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT feature, api_config_id FROM ai_bindings`).
		WillReturnRows(sqlmock.NewRows([]string{"feature", "api_config_id"}).
			AddRow("chat", 1).AddRow("summary", 2))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/bindings", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int             `json:"code"`
		Data map[string]int64 `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, int64(1), resp.Data["chat"])
	assert.Equal(t, int64(2), resp.Data["summary"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBindings_POST_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`INSERT INTO ai_bindings`).WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/bindings",
		`{"feature":"chat","config_id":1}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp.Data["status"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== bindings/{feature} GET/POST ====================

func TestHandleBindingsByPath_GET(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT api_config_id FROM ai_bindings WHERE feature`).
		WillReturnRows(sqlmock.NewRows([]string{"api_config_id"}).AddRow(3))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/bindings/chat", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "chat", resp.Data["feature"])
	assert.Equal(t, float64(3), resp.Data["config_id"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBindingsByPath_POST(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`INSERT INTO ai_bindings`).WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/bindings/chat",
		`{"configId":5}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp.Data["status"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== skills POST / GET by type ====================

func TestHandleSkills_POST_CreateMissingDefaults(t *testing.T) {
	h, mock := newTestHandler(t)
	// CreateMissingCustomerServiceDefaults 循环 6 个默认技能，
	// 每次先 CountSkillByName（返回 0），再 CreateSkill。
	// sqlmock 按注册顺序匹配，所以必须交错注册。
	for i := 0; i < 6; i++ {
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM ai_skills WHERE name`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery(`INSERT INTO ai_skills`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(i + 1)))
	}

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/skills",
		`{"customer_service_defaults":true}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(6), resp.Data["created"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSkills_GET_ByType(t *testing.T) {
	h, mock := newTestHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT id, name, description, system_prompt`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "system_prompt", "few_shot_examples", "type", "enabled", "created_at",
		}).AddRow(1, "账号问题", "处理账号问题", "你是客服", "", "CUSTOMER_SERVICE", 1, now))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/skills?type=CUSTOMER_SERVICE", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int     `json:"code"`
		Data []*Skill `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "账号问题", resp.Data[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== skills/{id}/toggle ====================

func TestHandleSkillByPath_Toggle(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE ai_skills SET enabled`).WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPut, "/api/v1/ai/skills/1/toggle", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp.Data["status"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== skills/{id} GET 404 ====================

func TestHandleSkillByPath_GET_404(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, name, description, system_prompt`).
		WillReturnError(sql.ErrNoRows)

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/skills/999", "")
	assert.Equal(t, http.StatusNotFound, rr.Code)
	var resp struct {
		Code int    `json:"code"`
		Data *Skill `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 404, resp.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== skills/{id} PUT ====================

func TestHandleSkillByPath_PUT_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE ai_skills SET name`).WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPut, "/api/v1/ai/skills/1",
		`{"name":"updated","description":"desc","type":"LLM"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp.Data["status"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== skills/{id} DELETE ====================

func TestHandleSkillByPath_DELETE_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`DELETE FROM ai_skills WHERE id`).WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodDelete, "/api/v1/ai/skills/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp.Data["status"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== usage/overview ====================

func TestHandleUsage_Overview(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM ai_usage_logs`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(token_count\),0\) FROM ai_usage_logs`).WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(5000))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(duration_ms\),0\) FROM ai_usage_logs`).WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(123.45))
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT feature\) FROM ai_usage_logs`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT user_id\) FROM ai_usage_logs WHERE user_id IS NOT NULL`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/usage/overview", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(10), resp.Data["total_calls"])
	assert.Equal(t, float64(5000), resp.Data["total_tokens"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== usage/features ====================

func TestHandleUsage_Features(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT feature, COUNT\(\*\)`).
		WillReturnRows(sqlmock.NewRows([]string{"feature", "calls", "tokens", "avg_duration_ms"}).
			AddRow("chat", 100, 3000, 150.5))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/usage/features", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "chat", resp.Data[0]["feature"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== usage/daily ====================

func TestHandleUsage_Daily(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT TO_CHAR\(created_at`).
		WillReturnRows(sqlmock.NewRows([]string{"day", "calls", "tokens"}).
			AddRow("2026-01-01", 50, 1500))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/usage/daily?days=7", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "2026-01-01", resp.Data[0]["date"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== usage unknown path ====================

func TestHandleUsage_404(t *testing.T) {
	h, _ := newTestHandler(t)

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/usage/unknown", "")
	assert.Equal(t, http.StatusNotFound, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(404), resp["code"])
}

// ==================== customer messages ====================

func TestHandleCustomer_Messages(t *testing.T) {
	h, mock := newTestHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT id, role, content, token_count, created_at FROM ai_chat_messages WHERE conversation_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "token_count", "created_at"}).
			AddRow(1, "user", "你好", 10, now))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/customer/1/messages", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "user", resp.Data[0]["role"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== customer reply ====================

func TestHandleCustomer_Reply(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`INSERT INTO ai_chat_messages`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE ai_conversations SET updated_at`).WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/customer/1/reply",
		`{"content":"您好，请问有什么可以帮助您？"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp.Data["status"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== customer reply empty content ====================

func TestHandleCustomer_Reply_EmptyContent(t *testing.T) {
	h, _ := newTestHandler(t)

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/customer/1/reply",
		`{"content":""}`)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(400), resp["code"])
}

// ==================== customer resolve ====================

func TestHandleCustomer_Resolve(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE ai_conversations SET status = 1`).WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/customer/1/resolve", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp.Data["status"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== customer sessions ====================

func TestHandleCustomer_Sessions(t *testing.T) {
	h, mock := newTestHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT id, user_id, title, status, created_at FROM ai_conversations WHERE status = 0`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "status", "created_at"}).
			AddRow(1, 100, "测试会话", 0, now))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/customer/sessions", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "测试会话", resp.Data[0]["title"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== customer pending count ====================

func TestHandleCustomer_PendingCount(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM ai_conversations WHERE status = 0`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/customer/sessions/pending/count", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(5), resp.Data["count"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== customer 404 ====================

func TestHandleCustomer_404(t *testing.T) {
	h, _ := newTestHandler(t)

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/customer/unknown", "")
	assert.Equal(t, http.StatusNotFound, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(404), resp["code"])
}

// ==================== summary/check ====================

func TestHandleSummary_Check_200(t *testing.T) {
	h, _ := newTestHandler(t)
	h.WithSummary(NewSummaryService(&mockCaller{hasSum: true}))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/summary/check/123", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(200), resp.Data["code"])
	assert.Equal(t, true, resp.Data["data"])
}

func TestHandleSummary_Check_400(t *testing.T) {
	h, _ := newTestHandler(t)
	h.WithSummary(NewSummaryService(&mockCaller{hasSum: false}))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/summary/check/0", "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(400), resp["code"])
}

func TestHandleSummary_Check_NilSvc(t *testing.T) {
	h, _ := newTestHandler(t)

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/summary/check/123", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, false, resp.Data["data"])
}

// ==================== summary/generate ====================

func TestHandleSummary_Generate_405(t *testing.T) {
	h, _ := newTestHandler(t)
	h.WithSummary(NewSummaryService(&mockCaller{}))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/summary/generate", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(405), resp["code"])
}

func TestHandleSummary_Generate_503(t *testing.T) {
	h, _ := newTestHandler(t)

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/summary/generate",
		`{"manuscript_id":1,"video_id":123}`)
	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(503), resp["code"])
}

// TestHandleSummary_Generate_200: GenerateSummary 依赖 MinioStorageService（具体类型，无法 mock），
// storage 为 nil 时返回 503，故此处验证 nil 路径即可。

// ==================== summary/stream ====================

func TestHandleSummary_Stream_400(t *testing.T) {
	h, _ := newTestHandler(t)
	h.WithSummary(NewSummaryService(&mockCaller{summary: "stream data"}))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/summary/stream/0", "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(400), resp["code"])
}

// ==================== summary/{videoId} GET ====================

func TestHandleSummary_Get_200(t *testing.T) {
	h, _ := newTestHandler(t)
	h.WithSummary(NewSummaryService(&mockCaller{summary: "这是一段AI生成的摘要"}))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/summary/456", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int               `json:"code"`
		Data map[string]string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Contains(t, resp.Data["summary"], "AI生成")
}

func TestHandleSummary_Get_NilSvc(t *testing.T) {
	h, _ := newTestHandler(t)

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/summary/456", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(t, body, "AI summary not available")
}

func TestHandleSummary_Get_Error(t *testing.T) {
	h, _ := newTestHandler(t)
	svc := &SummaryService{caller: &mockCallerErr{err: errors.New("downstream error")}}
	h.WithSummary(svc)

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/summary/456", "")
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(500), resp["code"])
}

// ==================== customer-service defaults ====================

func TestHandleCustomerDefaults_405(t *testing.T) {
	h, _ := newTestHandler(t)

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/skills/customer-service/defaults", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(405), resp["code"])
}

// ==================== config/test ====================

func TestHandleConfigTest_200(t *testing.T) {
	h, _ := newTestHandler(t)

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/config/test",
		`{"provider":"openai","model":"gpt-4o"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, true, resp.Data["success"])
	assert.Equal(t, "openai", resp.Data["provider"])
}

// ==================== assistant/send ====================

func TestHandleAssistantSend_200(t *testing.T) {
	h, _ := newTestHandler(t)

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/assistant/send",
		`{"message":"帮我查一下订单"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Contains(t, resp.Data["reply"], "帮我查一下订单")
}

func TestHandleAssistantSend_Content(t *testing.T) {
	h, _ := newTestHandler(t)

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/assistant/send",
		`{"content":"content字段测试"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Contains(t, resp.Data["reply"], "content字段测试")
}

// ==================== route-test ====================

func TestHandleRouteTest_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, name, description, system_prompt`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "system_prompt", "few_shot_examples", "type", "enabled", "created_at",
		}).AddRow(1, "账号与登录问题", "账号注册、登录、密码找回", "", "", "CUSTOMER_SERVICE", 1, time.Now()))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/skills/customer-service/route-test",
		`{"content":"账号"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	// "账号" 包含在 "账号与登录问题" 中，MatchCustomerServiceSkill 用 strings.Contains 匹配
	assert.Equal(t, true, resp.Data["matched"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRouteTest_NilSkill(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, name, description, system_prompt`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "system_prompt", "few_shot_examples", "type", "enabled", "created_at",
		}).AddRow(1, "完全不匹配的内容", "无关描述", "", "", "CUSTOMER_SERVICE", 1, time.Now()))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/skills/customer-service/route-test",
		`{"content":"直播设置"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	// 没有匹配的技能，返回 default
	assert.Equal(t, false, resp.Data["matched"])
	assert.Equal(t, "default", resp.Data["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== route (skills) ====================

func TestHandleRouteSkills_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, name, description, system_prompt`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "system_prompt", "few_shot_examples", "type", "enabled", "created_at",
		}).AddRow(1, "账号与登录问题", "账号注册、登录、密码找回", "", "", "CUSTOMER_SERVICE", 1, time.Now()).
			AddRow(2, "直播与互动", "开播、直播间设置", "", "", "CUSTOMER_SERVICE", 1, time.Now()))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/skills/customer-service/route",
		`{"content":"账号","limit":3}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	// "账号" 包含在 "账号与登录问题" 中，RouteSkills 用 strings.Contains 匹配 → score > 0
	assert.GreaterOrEqual(t, len(resp.Data), 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRouteSkills_Empty(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, name, description, system_prompt`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "system_prompt", "few_shot_examples", "type", "enabled", "created_at",
		}).AddRow(1, "完全不匹配的内容", "无关描述", "", "", "CUSTOMER_SERVICE", 1, time.Now()))

	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/skills/customer-service/route",
		`{"content":"完全无关的内容XYZ","limit":3}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	// 无匹配，返回空数组
	assert.Empty(t, resp.Data)
	require.NoError(t, mock.ExpectationsWereMet())
}

// mockCallerErr 用于模拟返回错误的 ServiceCaller
type mockCallerErr struct {
	err error
}

func (m *mockCallerErr) Call(ctx context.Context, target, method string, req, resp interface{}) error {
	return m.err
}

func (m *mockCallerErr) CallStream(ctx context.Context, target, method string, req interface{}) (<-chan []byte, error) {
	return nil, m.err
}

func (m *mockCallerErr) Close() error { return nil }