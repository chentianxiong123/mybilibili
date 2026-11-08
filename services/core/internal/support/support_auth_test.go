package support

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSessionProcess_NoAuth 钉住 FINDING-031（已修复）：
// PUT /api/v1/operation/internal/tickets/session/{sessionID} 无 admin → 401。
func TestSessionProcess_NoAuth(t *testing.T) {
	h, _ := newMockSupport(t)
	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/operation/internal/tickets/session/9", strings.NewReader(`{"adminReply":"hacked reply"}`))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "FINDING-031 修复：无鉴权处理工单应 401")
}

// TestCustomerSession_SpoofUserID 钉住 FINDING-031（已修复）：
// POST /api/v1/operation/internal/tickets/customer-session 无鉴权 → 401（body userId 不再被信任）。
func TestCustomerSession_SpoofUserID(t *testing.T) {
	h, _ := newMockSupport(t)
	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/operation/internal/tickets/customer-session", strings.NewReader(`{"userId":777,"title":"t","content":"c"}`))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "FINDING-031 修复：无鉴权建工单应 401")
}
