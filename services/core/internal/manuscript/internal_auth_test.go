package manuscript

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInternalTakeDown_NoAuth 钉住 FINDING-022 现象 A（已修复）：
// PUT /api/v1/manuscript/internal/{id}/take-down 在【无任何鉴权头】下
// 此前仍执行 UPDATE manuscripts SET status=-1 —— 未认证者可下架任意稿件。
// 修复后：无 admin 身份 → 401，DB 不执行。
func TestInternalTakeDown_NoAuth(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)
	// 修复后：不执行 DB UPDATE（未认证直接 401），故不设 ExpectExec。

	req := httptest.NewRequest(http.MethodPut, "/api/v1/manuscript/internal/5/take-down", nil)
	w := httptest.NewRecorder()
	m.h.handleRouter(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "FINDING-022 修复：无鉴权 take-down 应 401")
	assert.NoError(t, m.db.ExpectationsWereMet(), "未认证不应执行 DB 更新")
}

// TestCommentCount_SetArbitrary_NoAuth 钉住 FINDING-022 现象 B（已修复）：
// PUT /api/v1/manuscript/{id}/comment-count?count=N 在无鉴权下此前仍设置任意评论数。
// 修复后：无 admin 身份 → 401。
func TestCommentCount_SetArbitrary_NoAuth(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/manuscript/7/comment-count?count=999", nil)
	w := httptest.NewRecorder()
	m.h.handleRouter(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "FINDING-022 修复：无鉴权评论数篡改应 401")
	assert.NoError(t, m.db.ExpectationsWereMet(), "未认证不应执行 DB 更新")
}
