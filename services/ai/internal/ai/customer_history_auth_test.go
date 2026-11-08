package ai

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCustomerHistory_PathIDOR 钉住 FINDING-038（已修复）：
// GET /api/v1/ai/customer/history/777 无鉴权 → 401（此前路径 userID 优先，可读任意用户历史）。
func TestCustomerHistory_PathIDOR(t *testing.T) {
	customer := NewCustomerService(&fakeCaller{customerHist: []map[string]interface{}{{"content": "secret-history"}}})
	h := NewAIChatHandler(nil, customer)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/customer/history/777", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "FINDING-038 修复：无鉴权读任意用户历史应 401")
}
