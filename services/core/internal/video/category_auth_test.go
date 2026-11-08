package video

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// TestCategoryDelete_NoAuth 钉住 FINDING-023：
// DELETE /api/v1/category/{id} 在【无任何鉴权头】下仍执行分类删除 ——
// 未认证者可删除任意分类。路径不含 /admin/（IsAdminPath 仅 Contains /admin/），
// AdminPathGuard 不保护；video_handler.go 无 RequireUser/GetUserIDFromHeader。
func TestCategoryDelete_NoAuth(t *testing.T) {
	h, mock := newTestHandler(t)
	// handleCategoryByID DELETE 先 SELECT 分类（存在检查）再 DeleteCategory 事务
	mock.ExpectQuery(`SELECT id, name, parent_id, sort_order FROM categories WHERE id =`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "parent_id", "sort_order"}).AddRow(5, "动画", 0, 5))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE categories SET parent_id = 0 WHERE parent_id =`).
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM categories WHERE id =`).
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// 无 X-User-Id / Authorization —— 模拟未认证请求
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/category/5", nil)
	w := httptest.NewRecorder()
	h.handleCategoryByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet(), "FINDING-023: 无鉴权仍执行分类删除（应被拒绝）")
}

// TestCategoryUpdate_NoAuth 钉住 FINDING-023：
// PUT /api/v1/category/{id} 在无鉴权下仍执行分类更新（改名/改父级/排序）。
func TestCategoryUpdate_NoAuth(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT .* FROM categories WHERE id =`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "parent_id", "sort_order"}).AddRow(5, "动画", 0, 5))
	mock.ExpectExec(`UPDATE categories`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest(http.MethodPut, "/api/v1/category/5", nil)
	req.Body = http.NoBody
	w := httptest.NewRecorder()
	h.handleCategoryByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet(), "FINDING-023: 无鉴权仍执行分类更新（应被拒绝）")
}
