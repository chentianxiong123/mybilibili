package social

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDynamicCommentIncrement_NoAuth pins FINDING-030（已修复）：
// POST /api/v1/dynamic/comment/increment/{id} 无鉴权现返回 401，不执行 UPDATE。
func TestDynamicCommentIncrement_NoAuth(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mux := muxForSocial(h)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dynamic/comment/increment/7", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "FINDING-030 修复: 未认证应 401")
	_ = mock
}
