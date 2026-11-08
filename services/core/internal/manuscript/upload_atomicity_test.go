package manuscript

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUploadComplete_VideoInsertFailure_LeavesOrphanManuscript pins FINDING-058（已修复）：
// handleUploadComplete 现把稿件+视频+标签+upload_session 包进同一事务——
// 视频插入失败 → ROLLBACK，稿件不持久（不再产生孤儿稿件）。
func TestUploadComplete_VideoInsertFailure_LeavesOrphanManuscript(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id=\$1`).WithArgs("sess1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT title, description, category_id, tags, videos FROM upload_sessions WHERE id = \$1`).
		WithArgs("sess1").
		WillReturnRows(sqlmock.NewRows([]string{"title", "description", "category_id", "tags", "videos"}).
			AddRow("t", "d", int64(1), `[]`, `[{"url":"/uploads/a.mp4"}]`))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1\)`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// FINDING-058 修复：事务内稿件 INSERT 成功后视频 INSERT 失败 → ROLLBACK。
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(42)))
	mock.ExpectQuery(`INSERT INTO videos`).WillReturnError(errors.New("video insert failed"))
	mock.ExpectRollback()

	req := newReq("POST", "/api/v1/manuscript/upload-session/sess1/complete", "")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
