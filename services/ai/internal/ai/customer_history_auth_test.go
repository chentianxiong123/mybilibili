package ai

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCustomerHistory_PathIDOR 钉住 FINDING-038：
// GET /api/v1/ai/customer/history/{id} 用【路径 userID】优先于 header（且无
// RequireUser）—— 任意未认证者可读【任意用户】的客服聊天历史（IDOR）。
// handleCustomerHistory（ai_http_handler.go:117）`userID` 取自 path，
// 仅当 path==0 才回退 header。
func TestCustomerHistory_PathIDOR(t *testing.T) {
	customer := NewCustomerService(&fakeCaller{customerHist: []map[string]interface{}{{"content": "secret-history"}}})
	h := NewAIChatHandler(nil, customer)
	mux := http.NewServeMux()
	h.Register(mux)

	// 路径 victim=777，无鉴权头
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/customer/history/777", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "FINDING-038: 无鉴权读任意用户聊天历史（应被拒绝）")
	assert.Contains(t, w.Body.String(), "secret-history")
}
