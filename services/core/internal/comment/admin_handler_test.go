package comment

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAdminHandler(t *testing.T) (*AdminHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewAdminHandler(db), mock
}

func muxForAdmin(h *AdminHandler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestAdminHandler_handleRoute_List(t *testing.T) {
	h, mock := newAdminHandler(t)
	now := time.Now()

	mock.ExpectQuery(`FROM comments c WHERE 1=1 ORDER BY c.id DESC`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at"}).
			AddRow(1, 9, 100, "hello", 3, 1, 0, now))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments c WHERE 1=1`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	w := doReq(muxForAdmin(h), "GET", "/api/v1/admin/comments/list", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_handleRoute_List_WithFilters(t *testing.T) {
	h, mock := newAdminHandler(t)
	now := time.Now()

	mock.ExpectQuery(`AND c.status = \$1`).WithArgs("0", "%spam%", int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at"}).
			AddRow(1, 9, 100, "spam", 0, 0, 0, now))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments c WHERE 1=1 AND c.status = \$1 AND c.content LIKE \$2`).
		WithArgs("0", "%spam%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	w := doReq(muxForAdmin(h), "GET", "/api/v1/admin/comments/list?status=0&keyword=spam", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_handleRoute_List_QueryError(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectQuery(`FROM comments c WHERE 1=1`).WillReturnError(errorsNew("boom"))
	w := doReq(muxForAdmin(h), "GET", "/api/v1/admin/comments/list", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_handleRoute_Detail(t *testing.T) {
	h, mock := newAdminHandler(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, manuscript_id, user_id, content, like_count, reply_count, status, created_at`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at"}).
			AddRow(1, 9, 100, "hello", 3, 1, 0, now))
	w := doReq(muxForAdmin(h), "GET", "/api/v1/admin/comments/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectQuery(`SELECT id, manuscript_id, user_id, content`).WithArgs(int64(99)).WillReturnError(errorsNew("no rows"))
	w = doReq(muxForAdmin(h), "GET", "/api/v1/admin/comments/99", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_handleRoute_StatusUpdate(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectExec(`UPDATE comments SET status=\$1, updated_at=NOW\(\) WHERE id=\$2`).
		WithArgs(int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForAdmin(h), "PUT", "/api/v1/admin/comments/1/status", `{"status":1}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_handleRoute_Delete(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectExec(`DELETE FROM comments WHERE id=\$1`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForAdmin(h), "DELETE", "/api/v1/admin/comments/1/delete", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_handleRoute_SoftDelete(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectExec(`UPDATE comments SET status=1, updated_at=NOW\(\) WHERE id=\$1`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForAdmin(h), "DELETE", "/api/v1/admin/comments/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_exec_Error(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectExec(`UPDATE comments SET status=\$1`).WillReturnError(errorsNew("boom"))
	w := doReq(muxForAdmin(h), "PUT", "/api/v1/admin/comments/1/status", `{"status":1}`, nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mock.ExpectExec(`UPDATE comments SET status=\$1`).WithArgs(int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 0))
	w = doReq(muxForAdmin(h), "PUT", "/api/v1/admin/comments/1/status", `{"status":1}`, nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func errorsNew(msg string) error {
	return &simpleErr{msg}
}

type simpleErr struct{ msg string }

func (e *simpleErr) Error() string { return e.msg }
