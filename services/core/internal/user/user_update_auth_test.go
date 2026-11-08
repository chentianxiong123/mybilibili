package user

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUserUpdate_NoAuth 钉住 FINDING-029（已修复）：
// PUT /api/v1/user/5 无用户身份 → 401（此前未认证者可覆写任意用户资料）。
func TestUserUpdate_NoAuth(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/user/5", strings.NewReader(`{"nickname":"hacked","signature":"x"}`))
	w := httptest.NewRecorder()
	h.handleUserByID(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "FINDING-029 修复：无鉴权更新用户资料应 401")
}
