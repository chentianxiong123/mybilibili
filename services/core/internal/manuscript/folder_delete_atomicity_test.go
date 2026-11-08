package manuscript

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// TestFavoriteFolderDelete_DestroysOtherUsersLinks pins FINDING-064（已修复）：
// handleFavoriteFolderByID DELETE 现先校验 folder 归属（SELECT user_id），
// 非本人 folder → 403，两个 DELETE 均 owner 作用域 + 检查 RowsAffected。
func TestFavoriteFolderDelete_DestroysOtherUsersLinks(t *testing.T) {
	h, mock := newHTTPHandler(t)

	// folder 5 属于其他用户（user_id=99）→ 403，不执行任何 DELETE。
	mock.ExpectQuery(`SELECT user_id FROM favorite_folders WHERE id = \$1`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(99)))

	req := newReq("DELETE", "/api/v1/manuscript/favorite/folders/5", "")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code, "FINDING-064 修复：删除他人收藏夹应 403")
}
