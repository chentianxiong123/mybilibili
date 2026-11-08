package manuscript

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// TestInternalTakeDown_NoAuth 钉住 FINDING-022 现象 A：
// PUT /api/v1/manuscript/internal/{id}/take-down 在【无任何鉴权头】下
// 仍执行 UPDATE manuscripts SET status=-1, review_status=2 —— 未认证者
// 可下架任意稿件。AdminPathGuard 仅拦 /admin/ 前缀，/internal/ 不被保护。
func TestInternalTakeDown_NoAuth(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)
	m.db.ExpectExec(`UPDATE manuscripts SET status = -1, review_status = 2`).
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 无 X-User-Id / 无 Authorization —— 模拟未认证请求
	req := httptest.NewRequest(http.MethodPut, "/api/v1/manuscript/internal/5/take-down", nil)
	w := httptest.NewRecorder()
	m.h.handleRouter(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"status":"ok"}`, w.Body.String())
	assert.NoError(t, m.db.ExpectationsWereMet(), "FINDING-022: 无鉴权 take-down 仍执行 DB 更新（应被拒绝）")
}

// TestCommentCount_SetArbitrary_NoAuth 钉住 FINDING-022 现象 B：
// PUT /api/v1/manuscript/{id}/comment-count?count=N 在无鉴权下仍设置任意评论数。
func TestCommentCount_SetArbitrary_NoAuth(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)
	m.db.ExpectExec(`UPDATE manuscripts SET comment_count =`).
		WithArgs(int64(7), int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest(http.MethodPut, "/api/v1/manuscript/7/comment-count?count=999", nil)
	w := httptest.NewRecorder()
	m.h.handleRouter(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, m.db.ExpectationsWereMet(), "FINDING-022: 无鉴权仍可任意设置 comment_count")
}
