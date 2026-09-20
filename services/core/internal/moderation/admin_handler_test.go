package moderation

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== AdminHandler ====================

func TestAdminHandler_Pending_EmptyResult(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectQuery(`FROM content_reviews WHERE status = 'pending'`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "user_id", "content", "status", "reviewed_at"}))
	w := doReq(muxForAdmin(h), "GET", "/api/v1/moderation/admin/pending", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"data":[]`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_Pending_DBError(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectQuery(`FROM content_reviews WHERE status = 'pending'`).WillReturnError(errors.New("boom"))
	w := doReq(muxForAdmin(h), "GET", "/api/v1/moderation/admin/pending", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_All_Basic(t *testing.T) {
	h, mock := newAdminHandler(t)
	now := time.Now()
	mock.ExpectQuery(`FROM content_reviews WHERE 1=1`).
		WithArgs("", int32(1), int32(20)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "user_id", "content", "status", "reviewed_at"}).
			AddRow(1, "comment", int64(2), "test content", "active", now))
	w := doReq(muxForAdmin(h), "GET", "/api/v1/moderation/admin/all", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":200`)
	assert.Contains(t, w.Body.String(), `"test content"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_All_Empty(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectQuery(`FROM content_reviews WHERE 1=1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "user_id", "content", "status", "reviewed_at"}))
	w := doReq(muxForAdmin(h), "GET", "/api/v1/moderation/admin/all", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"data":[]`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_All_DBError(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectQuery(`FROM content_reviews WHERE 1=1`).WillReturnError(errors.New("boom"))
	w := doReq(muxForAdmin(h), "GET", "/api/v1/moderation/admin/all", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_Batch_InvalidBody(t *testing.T) {
	h, _ := newAdminHandler(t)
	w := doReq(muxForAdmin(h), "POST", "/api/v1/moderation/admin/batch", `bad`, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"code":400`)
}

func TestAdminHandler_Batch_DeleteMultiple(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectExec(`UPDATE content_reviews SET status = 'deleted'`).WithArgs("comment", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE content_reviews SET status = 'deleted'`).WithArgs("comment", int64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForAdmin(h), "POST", "/api/v1/moderation/admin/batch",
		`{"action":"delete","items":[{"type":"comment","id":1},{"type":"comment","id":2}]}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_Batch_RestoreSingle(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectExec(`UPDATE content_reviews SET status = 'active'`).WithArgs("comment", int64(3)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForAdmin(h), "POST", "/api/v1/moderation/admin/batch",
		`{"action":"restore","items":[{"type":"comment","id":3}]}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_UnknownRoute404(t *testing.T) {
	h, _ := newAdminHandler(t)
	w := doReq(muxForAdmin(h), "GET", "/api/v1/moderation/admin/unknown", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), `"code":404`)
}

func TestAdminHandler_ExecFailure(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectExec(`UPDATE content_reviews SET status = 'active'`).WillReturnError(errors.New("boom"))
	w := doReq(muxForAdmin(h), "PUT", "/api/v1/moderation/admin/restore/comment/5", "", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"code":500`)
}

// ==================== AdminContentHandler ====================

func TestAdminContentHandler_Comments_POSTNotAllowed(t *testing.T) {
	h, _ := newContentAdminHandler(t)
	w := doReq(muxForContentAdmin(h), "POST", "/api/v1/moderation/admin/comments", "", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestAdminContentHandler_Comments_ReplyType(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	now := time.Now()
	mock.ExpectQuery(`FROM replies r JOIN users u`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "parent_id", "manuscript_id", "user_id", "nickname", "avatar", "manuscript_title", "parent_content", "content", "like_count", "status", "created_at"}).
			AddRow(1, int64(10), int64(9), int64(2), "nick", "a", "title", "parent", "reply text", 3, "NORMAL", now))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM replies`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w := doReq(muxForContentAdmin(h), "GET", "/api/v1/moderation/admin/comments?type=reply", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total":1`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_Comments_WithKeyword(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	now := time.Now()
	mock.ExpectQuery(`FROM comments c JOIN users u`).
		WithArgs("spam", int32(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "parent_id", "manuscript_id", "user_id", "nickname", "avatar", "manuscript_title", "parent_content", "content", "like_count", "status", "created_at"}).
			AddRow(1, nil, int64(9), int64(2), "nick", "a", "title", nil, "spam msg", 0, "0", now))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w := doReq(muxForContentAdmin(h), "GET", "/api/v1/moderation/admin/comments?keyword=spam", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_Comments_EmptyRows(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	mock.ExpectQuery(`FROM comments c JOIN users u`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "parent_id", "manuscript_id", "user_id", "nickname", "avatar", "manuscript_title", "parent_content", "content", "like_count", "status", "created_at"}))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w := doReq(muxForContentAdmin(h), "GET", "/api/v1/moderation/admin/comments", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total":0`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_CommentByID_DeleteReply200(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	mock.ExpectExec(`UPDATE replies SET status = 'REMOVED' WHERE id = \$1`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForContentAdmin(h), "DELETE", "/api/v1/moderation/admin/comments/reply/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":200`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_CommentByID_RestoreReply200(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	mock.ExpectExec(`UPDATE replies SET status = 'NORMAL' WHERE id = \$1`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForContentAdmin(h), "PUT", "/api/v1/moderation/admin/comments/reply/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":200`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_CommentByID_DeleteComment200(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	mock.ExpectExec(`UPDATE comments SET status = 1 WHERE id = \$1`).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForContentAdmin(h), "DELETE", "/api/v1/moderation/admin/comments/comment/5", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_CommentByID_RestoreComment200(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	mock.ExpectExec(`UPDATE comments SET status = 0 WHERE id = \$1`).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForContentAdmin(h), "PUT", "/api/v1/moderation/admin/comments/comment/5", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_CommentByID_BadPaths(t *testing.T) {
	h, _ := newContentAdminHandler(t)
	w := doReq(muxForContentAdmin(h), "DELETE", "/api/v1/moderation/admin/comments/comment", "", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = doReq(muxForContentAdmin(h), "DELETE", "/api/v1/moderation/admin/comments/comment/abc", "", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminContentHandler_CommentByID_PostNotAllowed(t *testing.T) {
	h, _ := newContentAdminHandler(t)
	w := doReq(muxForContentAdmin(h), "POST", "/api/v1/moderation/admin/comments/comment/5", "", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestAdminContentHandler_CommentByID_DeleteExecError(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	mock.ExpectExec(`UPDATE comments SET status = 1 WHERE id = \$1`).WillReturnError(errors.New("boom"))
	w := doReq(muxForContentAdmin(h), "DELETE", "/api/v1/moderation/admin/comments/comment/5", "", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAdminContentHandler_Danmaku_POSTNotAllowed(t *testing.T) {
	h, _ := newContentAdminHandler(t)
	w := doReq(muxForContentAdmin(h), "POST", "/api/v1/moderation/admin/danmaku", "", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestAdminContentHandler_Danmaku_WithKeyword(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	now := time.Now()
	mock.ExpectQuery(`FROM danmaku d JOIN users u`).
		WithArgs("%bad%", int32(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "manuscript_title", "user_id", "nickname", "avatar", "content", "time", "mode", "created_at"}).
			AddRow(1, 2, 3, "title", 4, "nick", "a", "bad", 1.5, 1, now))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku WHERE content ILIKE \$1`).WithArgs("%bad%").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w := doReq(muxForContentAdmin(h), "GET", "/api/v1/moderation/admin/danmaku?keyword=bad", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_Danmaku_EmptyResult(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	mock.ExpectQuery(`FROM danmaku d JOIN users u`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "manuscript_title", "user_id", "nickname", "avatar", "content", "time", "mode", "created_at"}))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku$`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w := doReq(muxForContentAdmin(h), "GET", "/api/v1/moderation/admin/danmaku", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total":0`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_Danmaku_DBError(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	mock.ExpectQuery(`FROM danmaku d JOIN users u`).WillReturnError(errors.New("boom"))
	w := doReq(muxForContentAdmin(h), "GET", "/api/v1/moderation/admin/danmaku", "", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAdminContentHandler_DanmakuByID_Delete200(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	mock.ExpectExec(`DELETE FROM danmaku WHERE id = \$1`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForContentAdmin(h), "DELETE", "/api/v1/moderation/admin/danmaku/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":200`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_DanmakuByID_InvalidID(t *testing.T) {
	h, _ := newContentAdminHandler(t)
	w := doReq(muxForContentAdmin(h), "DELETE", "/api/v1/moderation/admin/danmaku/abc", "", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminContentHandler_DanmakuByID_PutNotAllowed(t *testing.T) {
	h, _ := newContentAdminHandler(t)
	w := doReq(muxForContentAdmin(h), "PUT", "/api/v1/moderation/admin/danmaku/5", "", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
