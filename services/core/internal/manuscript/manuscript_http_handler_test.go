package manuscript

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mybilibili/core/internal/comment"
	"mybilibili/core/internal/social"
	"mybilibili/core/internal/user"
	pb "mybilibili/pkg/pb"
)

func newHTTPHandler(t *testing.T) (*ManuscriptHTTPHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	userDB, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { userDB.Close() })

	commentDB, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { commentDB.Close() })

	interactDB, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { interactDB.Close() })

	svc := NewManuscriptService(NewManuscriptRepository(db), user.NewRepository(userDB))
	commentSvc := comment.NewCommentService(comment.NewCommentRepository(commentDB))
	interactSvc := social.NewInteractionService(social.NewInteractionRepository(interactDB))
	interactSvc.SetDB(interactDB)

	h := NewManuscriptHTTPHandler(db, svc, commentSvc, interactSvc)
	return h, mock
}

func newReq(method, path string, body ...string) *http.Request {
	var r *http.Request
	if len(body) > 0 && body[0] != "" {
		r = httptest.NewRequest(method, path, strings.NewReader(body[0]))
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	r.Header.Set("X-User-Id", "10")
	return r
}

type mocks struct {
	h    *ManuscriptHTTPHandler
	db   sqlmock.Sqlmock
	interact sqlmock.Sqlmock
}

func newHTTPHandlerWithInteract(t *testing.T) mocks {
	t.Helper()
	db, dbMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	userDB, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { userDB.Close() })

	commentDB, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { commentDB.Close() })

	interactDB, interactMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { interactDB.Close() })

	svc := NewManuscriptService(NewManuscriptRepository(db), user.NewRepository(userDB))
	commentSvc := comment.NewCommentService(comment.NewCommentRepository(commentDB))
	interactSvc := social.NewInteractionService(social.NewInteractionRepository(interactDB))
	interactSvc.SetDB(interactDB)

	h := NewManuscriptHTTPHandler(db, svc, commentSvc, interactSvc)
	return mocks{h: h, db: dbMock, interact: interactMock}
}

func TestHTTPHandler_Router_EmptyPath(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestHTTPHandler_Router_UnknownRoute(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/other/route", nil)
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestHTTPHandler_FixDurations(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`UPDATE manuscripts m SET duration_seconds`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("POST", "/api/v1/manuscript/fix-durations")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_FixDurations_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/fix-durations")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_InternalByPath(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`UPDATE manuscripts SET status = -1`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("PUT", "/api/v1/manuscript/internal/1/take-down")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_InternalByPath_NotFound(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/internal/foo")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestHTTPHandler_CommentCount(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`UPDATE manuscripts SET comment_count`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("PUT", "/api/v1/manuscript/1/comment-count?count=5")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_CommentCount_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/1/comment-count")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_IncrementComment(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`UPDATE manuscripts SET comment_count = comment_count \+ 1`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("POST", "/api/v1/manuscript/1/increment-comment")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_IncrementComment_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/1/increment-comment")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_DecrementComment(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`UPDATE manuscripts SET comment_count = GREATEST`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("POST", "/api/v1/manuscript/1/decrement-comment")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_DecrementComment_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/1/decrement-comment")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_Recommended(t *testing.T) {
	h, mock := newHTTPHandler(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE status = 3 ORDER BY upload_time`).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				now, int64(0), now, now, "", int32(0), "local"))
	mock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))
	mock.ExpectQuery(`SELECT is_vertical FROM videos`).WillReturnRows(sqlmock.NewRows([]string{"is_vertical"}).AddRow(0))

	req := newReq("GET", "/api/v1/manuscript/recommended")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_Hot(t *testing.T) {
	h, mock := newHTTPHandler(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE status = 3 ORDER BY view_count`).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
				int64(100), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				now, int64(0), now, now, "", int32(0), "local"))
	mock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))
	mock.ExpectQuery(`SELECT is_vertical FROM videos`).WillReturnRows(sqlmock.NewRows([]string{"is_vertical"}).AddRow(0))

	req := newReq("GET", "/api/v1/manuscript/hot")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_ManuscriptDetail(t *testing.T) {
	h, mock := newHTTPHandler(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				now, int64(0), now, now, "", int32(0), "local"))
	mock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))
	mock.ExpectExec(`UPDATE manuscripts SET view_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("GET", "/api/v1/manuscript/1")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_ManuscriptDetail_InvalidID(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/abc")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHTTPHandler_Category(t *testing.T) {
	h, mock := newHTTPHandler(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE category_id`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE category_id`).WillReturnRows(sqlmock.NewRows(manuscriptCols).
		AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
			int64(0), int64(0), int64(0), int64(0), int64(0),
			int64(0), int64(0), int32(3), int32(1), "",
			now, int64(0), now, now, "", int32(0), "local"))
	mock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))
	mock.ExpectQuery(`SELECT is_vertical FROM videos`).WillReturnRows(sqlmock.NewRows([]string{"is_vertical"}).AddRow(0))

	req := newReq("GET", "/api/v1/manuscript/category/1")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_UserManuscripts(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id`).WillReturnRows(sqlmock.NewRows(manuscriptCols))
	mock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))

	req := newReq("GET", "/api/v1/manuscript/user/10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_UserSearch(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id = .+ AND title ILIKE`).WillReturnRows(sqlmock.NewRows(manuscriptCols))
	mock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))

	req := newReq("GET", "/api/v1/manuscript/user/10/search?keyword=test")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_UserManuscriptStats(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = 0`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = 1`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = 3`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = 4`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = -1`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := newReq("GET", "/api/v1/manuscript/user/10/stats")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_PublishManuscript(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("POST", "/api/v1/manuscript/1/publish")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_PublishManuscript_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/1/publish")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_PublishManuscript_InvalidID(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("POST", "/api/v1/manuscript/abc/publish")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHTTPHandler_UnpublishManuscript(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("POST", "/api/v1/manuscript/1/unpublish")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_UnpublishManuscript_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/1/unpublish")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_DeleteManuscript(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`DELETE FROM manuscripts WHERE id`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("DELETE", "/api/v1/manuscript/1")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleDeleteManuscript(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_DeleteManuscript_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/1")
	w := httptest.NewRecorder()
	h.handleDeleteManuscript(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_DeleteManuscript_InvalidID(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("DELETE", "/api/v1/manuscript/abc")
	w := httptest.NewRecorder()
	h.handleDeleteManuscript(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHTTPHandler_UpdateManuscript(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET title`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM manuscript_tags WHERE manuscript_id`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO manuscript_tags`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"title":"new title","description":"new desc","category_id":2,"tags":["go"]}`
	req := newReq("PUT", "/api/v1/manuscript/1", body)
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleUpdateManuscript(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_UpdateManuscript_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/1")
	w := httptest.NewRecorder()
	h.handleUpdateManuscript(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_UpdateManuscript_InvalidID(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("PUT", "/api/v1/manuscript/abc")
	w := httptest.NewRecorder()
	h.handleUpdateManuscript(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHTTPHandler_UpdateManuscript_NotFound(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id`).WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	req := newReq("PUT", "/api/v1/manuscript/999", `{"title":"new"}`)
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()
	h.handleUpdateManuscript(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestHTTPHandler_UpdateManuscript_Forbidden(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(99)))

	req := newReq("PUT", "/api/v1/manuscript/1", `{"title":"new"}`)
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleUpdateManuscript(w, req)
	assert.Equal(t, 403, w.Code)
}

func TestHTTPHandler_UploadSession(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`INSERT INTO upload_sessions`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"client_id":"abc","title":"my video","description":"desc","category_id":1,"tags":["go"],"videos":[]}`
	req := newReq("POST", "/api/v1/manuscript/upload-session", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_UploadSession_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/upload-session")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_UploadSession_Unauthorized(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-session", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestHTTPHandler_UploadSessionByID_GET(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT id, user_id, title, COALESCE\(category_id,0\), uploaded_chunks, total_chunks, status`).
		WithArgs("abc", int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "category_id", "uploaded_chunks", "total_chunks", "status"}).
			AddRow("abc", int64(10), "title", int64(1), 0, 10, "pending"))

	req := newReq("GET", "/api/v1/manuscript/upload-session/abc")
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_UploadSessionByID_Delete(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`DELETE FROM upload_sessions WHERE id`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("DELETE", "/api/v1/manuscript/upload-session/abc")
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_UploadSessionByID_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("PATCH", "/api/v1/manuscript/upload-session/abc")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_UploadSessionByID_NotFound(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT id, user_id, title, COALESCE\(category_id,0\), uploaded_chunks, total_chunks, status`).
		WillReturnError(sql.ErrNoRows)

	req := newReq("GET", "/api/v1/manuscript/upload-session/notexist")
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestHTTPHandler_MeList(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id`).WillReturnRows(sqlmock.NewRows(manuscriptCols))

	req := newReq("GET", "/api/v1/manuscript/me/list")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_MeStats(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = 0`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = 1`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = 3`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = 4`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = -1`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := newReq("GET", "/api/v1/manuscript/me/stats")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_FavoriteFolders_Get(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT f.id, f.name, COALESCE`).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "cnt"}).
		AddRow(int64(1), "my folder", int64(5)))

	req := newReq("GET", "/api/v1/manuscript/favorite/folders")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_FavoriteFolders_Post(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`INSERT INTO favorite_folders`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	body := `{"name":"new folder"}`
	req := newReq("POST", "/api/v1/manuscript/favorite/folders", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_FavoriteFolders_Post_EmptyName(t *testing.T) {
	h, _ := newHTTPHandler(t)

	body := `{"name":"  "}`
	req := newReq("POST", "/api/v1/manuscript/favorite/folders", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHTTPHandler_FavoriteFolders_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("DELETE", "/api/v1/manuscript/favorite/folders")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_FavoriteFolderByID_Put(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`UPDATE favorite_folders SET name`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"name":"renamed"}`
	req := newReq("PUT", "/api/v1/manuscript/favorite/folders/1", body)
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_FavoriteFolderByID_Delete(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`DELETE FROM favorite_folder_videos WHERE folder_id`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM favorite_folders WHERE id`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("DELETE", "/api/v1/manuscript/favorite/folders/1")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_FavoriteFolderByID_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/favorite/folders/1")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_VideoFavoriteFolders_Get(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT ff.id, ff.name FROM favorite_folders ff`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
		AddRow(int64(1), "folder1"))

	req := newReq("GET", "/api/v1/manuscript/1/favorite/folders")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_VideoFavoriteFolders_Put(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM favorite_folder_videos`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO favorite_folder_videos`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO favorite_folder_videos`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	body := `{"folderIds":[1,2]}`
	req := newReq("PUT", "/api/v1/manuscript/1/favorite/folders", body)
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_VideoFavoriteFolders_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("POST", "/api/v1/manuscript/1/favorite/folders")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_VideoFavorite(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`INSERT INTO favorite_folder_videos`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"folderIds":[1]}`
	req := newReq("POST", "/api/v1/manuscript/1/favorite", body)
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_VideoFavorite_InvalidID(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("POST", "/api/v1/manuscript/abc/favorite")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHTTPHandler_VideoFavorite_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/1/favorite")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHTTPHandler_VideoFavoriteFolderByID_Delete(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`DELETE FROM favorite_folder_videos ffv`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("DELETE", "/api/v1/manuscript/1/favorite/folders/2")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_VideoFavoriteFolderByID_Invalid(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("DELETE", "/api/v1/manuscript/abc/favorite/folders/2")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHTTPHandler_VideoFavoriteFolderByID_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/1/favorite/folders/2")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHTTPHandler_FavoriteFolderVideos(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM favorite_folders WHERE id`).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT manuscript_id, created_at FROM favorite_folder_videos`).WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "created_at"}).
		AddRow(int64(1), time.Now()))

	req := newReq("GET", "/api/v1/manuscript/favorite/folders/1/videos")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_FavoriteFolderVideos_NotFound(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM favorite_folders WHERE id`).WillReturnError(sql.ErrNoRows)

	req := newReq("GET", "/api/v1/manuscript/favorite/folders/999/videos")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestHTTPHandler_FavoriteFolderVideos_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("POST", "/api/v1/manuscript/favorite/folders/1/videos")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_Like(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	m.interact.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE user_id`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	m.interact.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`UPDATE manuscripts SET like_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE target_type`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	req := newReq("POST", "/api/v1/manuscript/1/like")
	w := httptest.NewRecorder()
	m.h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_Like_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/1/like")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_Coin(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	m.interact.ExpectQuery(`SELECT coin_count FROM users WHERE id`).WillReturnRows(sqlmock.NewRows([]string{"coin_count"}).AddRow(int64(100)))
	m.interact.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`UPDATE manuscripts SET coin_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`UPDATE users SET coin_count`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("POST", "/api/v1/manuscript/1/coin?coinCount=2")
	w := httptest.NewRecorder()
	m.h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_Coin_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/1/coin")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_Collect(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	m.interact.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE user_id`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	m.interact.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`UPDATE manuscripts SET collect_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"folderId":1}`
	req := newReq("POST", "/api/v1/manuscript/1/collect", body)
	w := httptest.NewRecorder()
	m.h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_Share(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	m.interact.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`UPDATE manuscripts SET share_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"channel":"wechat"}`
	req := newReq("POST", "/api/v1/manuscript/1/share", body)
	w := httptest.NewRecorder()
	m.h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_UploadComplete(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id`).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT title, description, category_id, tags, videos FROM upload_sessions`).WillReturnRows(sqlmock.NewRows([]string{"title", "description", "category_id", "tags", "videos"}).
		AddRow("t", "d", int64(1), "[]", `[{"url":"http://video.mp4"}]`))
	mock.ExpectQuery(`INSERT INTO manuscripts`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectQuery(`INSERT INTO videos`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(100)))
	mock.ExpectExec(`UPDATE upload_sessions SET status`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET experience`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("POST", "/api/v1/manuscript/upload-session/abc/complete")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestUploadWorkDir(t *testing.T) {
	dir := uploadWorkDir()
	assert.NotEmpty(t, dir)
}

func TestUploadWorkDir_Env(t *testing.T) {
	os.Setenv("MYBILIBILI_UPLOAD_DIR", "/tmp/test-uploads")
	defer os.Unsetenv("MYBILIBILI_UPLOAD_DIR")

	dir := uploadWorkDir()
	assert.Equal(t, "/tmp/test-uploads", dir)
}

func TestHTTPHandler_Like_Delete(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	m.interact.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE user_id`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	m.interact.ExpectExec(`DELETE FROM user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`UPDATE manuscripts SET like_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE target_type`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := newReq("DELETE", "/api/v1/manuscript/1/like")
	w := httptest.NewRecorder()
	m.h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_Like_UnknownMethod(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("PATCH", "/api/v1/manuscript/1/like")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_Collect_Delete(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	m.interact.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE user_id`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	m.interact.ExpectExec(`DELETE FROM user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`UPDATE manuscripts SET collect_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE target_type`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := newReq("DELETE", "/api/v1/manuscript/1/collect")
	w := httptest.NewRecorder()
	m.h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_Collect_UnknownMethod(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("PATCH", "/api/v1/manuscript/1/collect")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_Coin_Body(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	m.interact.ExpectQuery(`SELECT coin_count FROM users WHERE id`).WillReturnRows(sqlmock.NewRows([]string{"coin_count"}).AddRow(int64(100)))
	m.interact.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`UPDATE manuscripts SET coin_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`UPDATE users SET coin_count`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"count":3}`
	req := newReq("POST", "/api/v1/manuscript/1/coin", body)
	w := httptest.NewRecorder()
	m.h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_Share_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/1/share")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_Coin_CoinCountFromQuery(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	m.interact.ExpectQuery(`SELECT coin_count FROM users WHERE id`).WillReturnRows(sqlmock.NewRows([]string{"coin_count"}).AddRow(int64(100)))
	m.interact.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`UPDATE manuscripts SET coin_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`UPDATE users SET coin_count`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("POST", "/api/v1/manuscript/1/coin?coinCount=5")
	w := httptest.NewRecorder()
	m.h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_VideoFavoriteFolders_Put_BadJSON(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM favorite_folder_videos`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	req := newReq("PUT", "/api/v1/manuscript/1/favorite/folders", `not json`)
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_FavoriteFolders_Get_Error(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT f.id, f.name, COALESCE`).WillReturnError(sql.ErrConnDone)

	req := newReq("GET", "/api/v1/manuscript/favorite/folders")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHTTPHandler_FavoriteFolders_Post_InternalError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`INSERT INTO favorite_folders`).WillReturnError(sql.ErrConnDone)

	body := `{"name":"new folder"}`
	req := newReq("POST", "/api/v1/manuscript/favorite/folders", body)
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHTTPHandler_FavoriteFolderByID_Put_EmptyName(t *testing.T) {
	h, _ := newHTTPHandler(t)

	body := `{"name":"  "}`
	req := newReq("PUT", "/api/v1/manuscript/favorite/folders/1", body)
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHTTPHandler_FavoriteFolderByID_Delete_Error(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`DELETE FROM favorite_folder_videos WHERE folder_id`).WillReturnError(sql.ErrConnDone)

	req := newReq("DELETE", "/api/v1/manuscript/favorite/folders/1")
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHTTPHandler_FavoriteFolderVideos_Forbidden(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM favorite_folders WHERE id`).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(99)))

	req := newReq("GET", "/api/v1/manuscript/favorite/folders/1/videos")
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 403, w.Code)
}

func TestHTTPHandler_FavoriteFolderVideos_DBError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM favorite_folders WHERE id`).WillReturnError(sql.ErrConnDone)

	req := newReq("GET", "/api/v1/manuscript/favorite/folders/1/videos")
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHTTPHandler_FavoriteFolderVideos_QueryError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM favorite_folders WHERE id`).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT manuscript_id, created_at FROM favorite_folder_videos`).WillReturnError(sql.ErrConnDone)

	req := newReq("GET", "/api/v1/manuscript/favorite/folders/1/videos")
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHTTPHandler_VideoFavoriteFolders_Put_Error(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM favorite_folder_videos`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO favorite_folder_videos`).WillReturnError(sql.ErrConnDone)

	body := `{"folderIds":[1]}`
	req := newReq("PUT", "/api/v1/manuscript/1/favorite/folders", body)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHTTPHandler_VideoFavoriteFolders_Put_CommitError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM favorite_folder_videos`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO favorite_folder_videos`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(sql.ErrConnDone)

	body := `{"folderIds":[1]}`
	req := newReq("PUT", "/api/v1/manuscript/1/favorite/folders", body)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHTTPHandler_VideoFavoriteFolders_Get_Error(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT ff.id, ff.name FROM favorite_folders ff`).WillReturnError(sql.ErrConnDone)

	req := newReq("GET", "/api/v1/manuscript/1/favorite/folders")
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHTTPHandler_UploadSessionByID_GET_Error(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT id, user_id, title, COALESCE\(category_id,0\), uploaded_chunks, total_chunks, status`).
		WillReturnError(sql.ErrConnDone)

	req := newReq("GET", "/api/v1/manuscript/upload-session/abc")
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestHTTPHandler_UploadSessionByID_Delete_Error(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`DELETE FROM upload_sessions WHERE id`).WillReturnError(sql.ErrConnDone)

	req := newReq("DELETE", "/api/v1/manuscript/upload-session/abc")
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHTTPHandler_UploadSession_DBError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`INSERT INTO upload_sessions`).WillReturnError(sql.ErrConnDone)

	body := `{"client_id":"abc","title":"my video","description":"desc","category_id":1}`
	req := newReq("POST", "/api/v1/manuscript/upload-session", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHTTPHandler_Register(t *testing.T) {
	h, _ := newHTTPHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	assert.NotNil(t, mux)
}



func TestHTTPHandler_Like_Coin_Collect_Share_NoUser(t *testing.T) {
	h, _ := newHTTPHandler(t)

	for _, path := range []string{
		"/api/v1/manuscript/1/collect",
		"/api/v1/manuscript/1/share",
		"/api/v1/manuscript/1/like",
		"/api/v1/manuscript/1/coin",
	} {
		req := httptest.NewRequest("POST", path, nil)
		w := httptest.NewRecorder()
		h.handleRouter(w, req)
		assert.Equal(t, 401, w.Code, "path: %s", path)
	}
}

func TestHTTPHandler_UploadSession_WithTotalChunks(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`INSERT INTO upload_sessions`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"title":"video","total_chunks":5}`
	req := newReq("POST", "/api/v1/manuscript/upload-session", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHTTPHandler_UpdateManuscript_BadBody(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("PUT", "/api/v1/manuscript/1", `not json`)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUpdateManuscript(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHTTPHandler_UploadSessionByID_EmptyID(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/upload-session/")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestManuscriptListPageJSON2(t *testing.T) {
	resp := &pb.ListUserManuscriptsResponse{
		Manuscripts: []*pb.ManuscriptInfo{
			{Id: 1, Title: "a"},
			{Id: 2, Title: "b"},
		},
		Total:    2,
		Page:     1,
		PageSize: 10,
	}

	result := manuscriptListPageJSON(resp)
	list := result["list"].([]map[string]interface{})
	assert.Len(t, list, 2)
	assert.Equal(t, int64(2), result["total"])
}

func TestManuscriptsByIDs(t *testing.T) {
	h, mock := newHTTPHandler(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				now, int64(0), now, now, "", int32(0), "local"))
	mock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))
	mock.ExpectQuery(`SELECT is_vertical FROM videos`).WillReturnRows(sqlmock.NewRows([]string{"is_vertical"}).AddRow(0))

	ctx := context.Background()
	result := h.manuscriptsByIDs(ctx, []int64{1})
	assert.Len(t, result, 1)
}

func TestManuscriptsByIDs_Error(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	result := h.manuscriptsByIDs(ctx, []int64{999})
	assert.Empty(t, result)
}
