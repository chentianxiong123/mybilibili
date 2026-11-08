package ai

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigCreate_NoAuth 钉住 FINDING-024：
// POST /api/v1/ai/configs 在【无任何鉴权头】下仍执行 INSERT ai_api_configs
// （含 APIKey）—— 未认证者可创建 AI 配置。路径不含 /admin/，
// AdminPathGuard 不保护；handleConfigs 不调用 getAdminID/RequireUser。
func TestConfigCreate_NoAuth(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`INSERT INTO ai_api_configs`).
		WithArgs("test", "LLM", "http://evil.example", "sk-123", "gpt-4", int32(1000), float64(0.7)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mux := http.NewServeMux()
	h.Register(mux)

	// 无 X-Admin-Id / Authorization —— 模拟未认证请求
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/configs",
		`{"name":"test","type":"LLM","base_url":"http://evil.example","api_key":"sk-123","model":"gpt-4","max_tokens":1000,"temperature":0.7}`)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet(), "FINDING-024: 无鉴权仍执行 INSERT ai_api_configs（应被拒绝）")
}
