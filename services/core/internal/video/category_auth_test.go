package video

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCategoryDelete_NoAuth 钉住 FINDING-023（已修复）：
// DELETE /api/v1/category/{id} 无 admin 身份 → 401，DB 不执行。
func TestCategoryDelete_NoAuth(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/category/5", nil)
	w := httptest.NewRecorder()
	h.handleCategoryByID(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "FINDING-023 修复：无鉴权删除分类应 401")
}

// TestCategoryUpdate_NoAuth 钉住 FINDING-023（已修复）：
// PUT /api/v1/category/{id} 无 admin 身份 → 401。
func TestCategoryUpdate_NoAuth(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/category/5", nil)
	req.Body = http.NoBody
	w := httptest.NewRecorder()
	h.handleCategoryByID(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "FINDING-023 修复：无鉴权更新分类应 401")
}
