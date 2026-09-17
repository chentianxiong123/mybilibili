package social

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/pkg/auth"
)

func newMockSocialHandler(t *testing.T) (*SocialHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp), sqlmock.ValueConverterOption(argConverter{}))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	followSvc := NewFollowService(NewFollowRepository(db))
	dynamicSvc := NewDynamicService(NewDynamicRepository(db))
	collectSvc := NewCollectionService(NewCollectionRepository(db))
	shareRepo := NewShareRepository(db)
	h := NewSocialHandler(followSvc, dynamicSvc, collectSvc, shareRepo, db, auth.NewJWT("test-secret"))
	return h, mock
}

func muxForSocial(h *SocialHandler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestFormatDuration(t *testing.T) {
	assert.Equal(t, "", formatDuration(0))
	assert.Equal(t, "", formatDuration(-5))
	assert.Equal(t, "01:30", formatDuration(90))
	assert.Equal(t, "1:02:03", formatDuration(3723))
}

func TestPad2(t *testing.T) {
	assert.Equal(t, "00", pad2(0))
	assert.Equal(t, "09", pad2(9))
	assert.Equal(t, "10", pad2(10))
	assert.Equal(t, "59", pad2(59))
}

func TestHandleDynamicAll_GET(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`SELECT id, user_id, content, dynamic_type`).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}))
	w := doReq(muxForSocial(h), "GET", "/api/v1/dynamic/all", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDynamicAll_MethodNotAllowed(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/all", "", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleDynamicLike_Post(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`INSERT INTO dynamic_likes`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`like_count = GREATEST`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT like_count FROM user_dynamics`).WillReturnRows(sqlmock.NewRows([]string{"like_count"}).AddRow(1))
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/like/1", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDynamicLike_Delete(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`DELETE FROM dynamic_likes`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`like_count = GREATEST`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT like_count FROM user_dynamics`).WillReturnRows(sqlmock.NewRows([]string{"like_count"}).AddRow(0))
	w := doReq(muxForSocial(h), "DELETE", "/api/v1/dynamic/like/1", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDynamicLike_Unauthorized(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/like/1", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleDynamicLike_InvalidID(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/like/0", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleDynamicLike_MethodNotAllowed(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "PUT", "/api/v1/dynamic/like/1", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleDynamicLike_DBErrorOnInsert(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`INSERT INTO dynamic_likes`).WillReturnError(errors.New("db error"))
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/like/1", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleDynamicLike_DeleteDBError(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`DELETE FROM dynamic_likes`).WillReturnError(errors.New("db error"))
	w := doReq(muxForSocial(h), "DELETE", "/api/v1/dynamic/like/1", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleDynamicShare_Post(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/share/1", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestHandleDynamicShare_Unauthorized(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/share/1", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleDynamicShare_NonPost(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "GET", "/api/v1/dynamic/share/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamicCommentIncrement_Post(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`comment_count = GREATEST`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/comment/increment/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamicCommentIncrement_NonPost(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "GET", "/api/v1/dynamic/comment/increment/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamicCommentList_GET(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	now := time.Now()
	mock.ExpectQuery(`ORDER BY created_at DESC`).WillReturnRows(sqlmock.NewRows([]string{"id", "dynamic_id", "user_id", "content", "parent_id", "reply_user_id", "like_count", "status", "created_at", "updated_at"}).
		AddRow(1, 1, 100, "nice", 0, 0, 1, 0, now, now))
	mock.ExpectQuery(`SELECT parent_id, COUNT`).WillReturnRows(sqlmock.NewRows([]string{"parent_id", "count"}))
	w := doReq(muxForSocial(h), "GET", "/api/v1/dynamic/comment/list?dynamicId=1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDynamicCommentList_MethodNotAllowed(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/comment/list", "", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleDynamicCommentAdd_Post(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`INSERT INTO dynamic_comments`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectExec(`comment_count = GREATEST`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT level, experience FROM users`).WillReturnRows(sqlmock.NewRows([]string{"level", "experience"}).AddRow(1, 0))
	mock.ExpectExec(`UPDATE users SET experience`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT parent_id, COUNT`).WillReturnRows(sqlmock.NewRows([]string{"parent_id", "count"}))
	mock.ExpectQuery(`SELECT target_id FROM user_interactions`).WillReturnRows(sqlmock.NewRows([]string{"target_id"}))
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/comment/add?dynamicId=1&content=hello", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDynamicCommentAdd_PostJSON(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`INSERT INTO dynamic_comments`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectExec(`comment_count = GREATEST`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT level, experience FROM users`).WillReturnRows(sqlmock.NewRows([]string{"level", "experience"}).AddRow(1, 0))
	mock.ExpectExec(`UPDATE users SET experience`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT parent_id, COUNT`).WillReturnRows(sqlmock.NewRows([]string{"parent_id", "count"}))
	mock.ExpectQuery(`SELECT target_id FROM user_interactions`).WillReturnRows(sqlmock.NewRows([]string{"target_id"}))
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/comment/add", `{"dynamicId":1,"content":"hello"}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamicCommentAdd_MethodNotAllowed(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "GET", "/api/v1/dynamic/comment/add", "", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleDynamicCommentReplies_GET(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	now := time.Now()
	mock.ExpectQuery(`FROM dynamic_comments WHERE parent_id`).WillReturnRows(sqlmock.NewRows([]string{"id", "dynamic_id", "user_id", "content", "parent_id", "reply_user_id", "like_count", "status", "created_at", "updated_at"}).
		AddRow(2, 1, 100, "reply", 1, 200, 0, 0, now, now))
	mock.ExpectQuery(`SELECT parent_id, COUNT`).WillReturnRows(sqlmock.NewRows([]string{"parent_id", "count"}))
	mock.ExpectQuery(`SELECT target_id FROM user_interactions`).WillReturnRows(sqlmock.NewRows([]string{"target_id"}))
	w := doReq(muxForSocial(h), "GET", "/api/v1/dynamic/comment/replies?commentId=1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamicCommentReplies_MethodNotAllowed(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/comment/replies", "", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleDynamicCommentDelete_DELETE(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`UPDATE dynamic_comments SET status = 1`).WithArgs(int64(5), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "DELETE", "/api/v1/dynamic/comment/delete/5", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamicCommentDelete_Unauthorized(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "DELETE", "/api/v1/dynamic/comment/delete/5", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleDynamicCommentLike_Post(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE dynamic_comments SET like_count = like_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/comment/like/5", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamicCommentLike_Delete(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`DELETE FROM user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE dynamic_comments SET like_count = GREATEST`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "DELETE", "/api/v1/dynamic/comment/like/5", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamicCommentLike_Unauthorized(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/comment/like/5", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleDynamicCommentLike_InvalidID(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/comment/like/0", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleDynamicCommentLike_MethodNotAllowed(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "GET", "/api/v1/dynamic/comment/like/5", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleDynamicCommentLike_PostDBError(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`INSERT INTO user_interactions`).WillReturnError(errors.New("db err"))
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/comment/like/5", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleDynamicCommentLike_DeleteDBError(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`DELETE FROM user_interactions`).WillReturnError(errors.New("db err"))
	w := doReq(muxForSocial(h), "DELETE", "/api/v1/dynamic/comment/like/5", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleDynamic_Publish(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`INSERT INTO user_dynamics`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	now := time.Now()
	mock.ExpectQuery(`SELECT id, user_id, content, dynamic_type`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}).
			AddRow(1, 1, "hello", 1, "", 0, 0, 0, 0, 0, now))
	mock.ExpectQuery(`SELECT level, experience FROM users`).WillReturnRows(sqlmock.NewRows([]string{"level", "experience"}).AddRow(1, 0))
	mock.ExpectExec(`UPDATE users SET experience`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/publish?content=hello&type=1", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamic_Publish_Unauthorized(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/publish", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleDynamic_List(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`FROM user_dynamics WHERE status = 0 ORDER BY created_at DESC`).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}))
	w := doReq(muxForSocial(h), "GET", "/api/v1/dynamic/list", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamic_Following(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`FROM user_dynamics d WHERE d.status = 0`).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}))
	w := doReq(muxForSocial(h), "GET", "/api/v1/dynamic/following", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamic_User(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`FROM user_dynamics WHERE user_id`).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}))
	w := doReq(muxForSocial(h), "GET", "/api/v1/dynamic/user/100", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamic_CommentPost(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`INSERT INTO dynamic_comments`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectExec(`comment_count = GREATEST`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/comment/1?content=hello", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamic_CommentGet(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	now := time.Now()
	mock.ExpectQuery(`ORDER BY created_at DESC`).WillReturnRows(sqlmock.NewRows([]string{"id", "dynamic_id", "user_id", "content", "parent_id", "reply_user_id", "like_count", "status", "created_at", "updated_at"}).
		AddRow(1, 1, 100, "c", 0, 0, 0, 0, now, now))
	mock.ExpectQuery(`SELECT parent_id, COUNT`).WillReturnRows(sqlmock.NewRows([]string{"parent_id", "count"}))
	w := doReq(muxForSocial(h), "GET", "/api/v1/dynamic/comment/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamic_Delete(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`UPDATE user_dynamics SET status = 1`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "DELETE", "/api/v1/dynamic/1", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamic_Default(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "GET", "/api/v1/dynamic/badroute", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleDynamic_User_DBError(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`FROM user_dynamics WHERE user_id`).WillReturnError(errors.New("db err"))
	w := doReq(muxForSocial(h), "GET", "/api/v1/dynamic/user/100", "", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleDynamic_CommentPost_Error(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`INSERT INTO dynamic_comments`).WillReturnError(errors.New("db err"))
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/comment/1?content=hello", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCollection_Create_POST(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`INSERT INTO manuscript_collections`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	now := time.Now()
	mock.ExpectQuery(`SELECT id, title, description, cover_url`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "cover_url", "user_id", "manuscript_count", "view_count", "status", "created_at", "updated_at"}).
			AddRow(1, "t", "d", "c", 1, 0, 0, 0, now, now))
	w := doReq(muxForSocial(h), "POST", "/api/v1/collection", `{"title":"t","description":"d","cover_url":"c"}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleCollection_Create_EmptyTitle(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "POST", "/api/v1/collection", `{}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCollection_ListUser(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`FROM manuscript_collections WHERE user_id`).WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "cover_url", "user_id", "manuscript_count", "view_count", "status", "created_at", "updated_at"}))
	w := doReq(muxForSocial(h), "GET", "/api/v1/collection/user/100", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleCollection_ListManuscripts(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`JOIN manuscripts m ON m.id = cr.manuscript_id`).WillReturnRows(sqlmock.NewRows([]string{"id", "title", "cover_url", "view_count", "upload_time", "duration", "user_id", "status"}))
	mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w := doReq(muxForSocial(h), "GET", "/api/v1/collection/1/manuscripts", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleCollection_GetByID(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT id, title, description, cover_url`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "cover_url", "user_id", "manuscript_count", "view_count", "status", "created_at", "updated_at"}).
			AddRow(1, "t", "d", "c", 100, 0, 0, 0, now, now))
	w := doReq(muxForSocial(h), "GET", "/api/v1/collection/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleCollection_GetByID_NotFound(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`SELECT id, title, description, cover_url`).WillReturnError(errors.New("no rows"))
	w := doReq(muxForSocial(h), "GET", "/api/v1/collection/1", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleCollection_Update_FormData(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`UPDATE manuscript_collections SET title`).WillReturnResult(sqlmock.NewResult(0, 1))
	r := httptest.NewRequest("PUT", "/api/v1/collection/1", strings.NewReader(""))
	r.Header.Set("X-User-Id", "1")
	r.ParseForm()
	r.Form.Set("name", "nt")
	r.Form.Set("description", "nd")
	w := httptest.NewRecorder()
	muxForSocial(h).ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleCollection_Update_JSON(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`UPDATE manuscript_collections SET title`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "PUT", "/api/v1/collection/1", `{"title":"nt","description":"nd"}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleCollection_Delete(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`DELETE FROM manuscript_collections`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "DELETE", "/api/v1/collection/1", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleCollection_AddManuscript(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`INSERT INTO manuscript_collection_relations`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscript_collections SET manuscript_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "POST", "/api/v1/collection/1/manuscript/9", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleCollection_RemoveManuscript(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`DELETE FROM manuscript_collection_relations`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscript_collections SET manuscript_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "DELETE", "/api/v1/collection/1/manuscript/9", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleCollection_Default(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "PATCH", "/api/v1/collection/1", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleCollection_Create_DBError(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`INSERT INTO manuscript_collections`).WillReturnError(errors.New("boom"))
	w := doReq(muxForSocial(h), "POST", "/api/v1/collection", `{"title":"t"}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleShare_Statistics(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`SELECT channel, COUNT`).WillReturnRows(sqlmock.NewRows([]string{"channel", "count"}).AddRow("wechat", 5))
	w := doReq(muxForSocial(h), "GET", "/api/v1/share/statistics?manuscript_id=1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleShare_Statistics_MissingID(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "GET", "/api/v1/share/statistics", "", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleShare_Statistics_DBError(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`SELECT channel, COUNT`).WillReturnError(errors.New("db err"))
	w := doReq(muxForSocial(h), "GET", "/api/v1/share/statistics?manuscript_id=1", "", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleShare_Record(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`INSERT INTO shares`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET share_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "POST", "/api/v1/share/1?channel=wechat", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleShare_NonPostNonStats(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "GET", "/api/v1/share/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleWatchHistory_Get(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`SELECT wh.id, wh.manuscript_id`).WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "progress_seconds", "watched_at", "title", "cover_url", "nickname", "avatar"}))
	w := doReq(muxForSocial(h), "GET", "/api/v1/watch-history", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleWatchHistory_Get_DBError(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`SELECT wh.id, wh.manuscript_id`).WillReturnError(errors.New("db err"))
	w := doReq(muxForSocial(h), "GET", "/api/v1/watch-history", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleWatchHistory_Post(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`INSERT INTO watch_history`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "POST", "/api/v1/watch-history?manuscript_id=9&progress_seconds=30", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleWatchHistory_Delete_ByID(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`DELETE FROM watch_history WHERE id`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "DELETE", "/api/v1/watch-history/123", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleWatchHistory_Delete_All(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`DELETE FROM watch_history WHERE user_id`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "DELETE", "/api/v1/watch-history", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleWatchHistory_Unauthorized(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	w := doReq(muxForSocial(h), "GET", "/api/v1/watch-history", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleWatchHistory_Delete_InvalidID(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`DELETE FROM watch_history WHERE user_id`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "DELETE", "/api/v1/watch-history/abc", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEnrichDynamics_Empty(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	out := h.enrichDynamics(context.Background(), 1, nil)
	assert.Empty(t, out)
}

func TestEnrichDynamics_WithData(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	now := time.Now()
	list := []*Dynamic{
		{ID: 1, UserID: 100, Content: "hi", DynamicType: 1, ImageURL: "http://img", RefManuscriptID: 9, LikeCount: 2, CreatedAt: now},
	}
	mock.ExpectQuery(`SELECT id, COALESCE`).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(100, "alice", "Alice", "http://a", 3))
	mock.ExpectQuery(`SELECT dynamic_id FROM dynamic_likes`).WillReturnRows(sqlmock.NewRows([]string{"dynamic_id"}))
	mock.ExpectQuery(`SELECT COALESCE`).WillReturnRows(sqlmock.NewRows([]string{"title", "cover_url", "duration_seconds", "view_count"}).AddRow("mtitle", "http://c", 100, 50))
	out := h.enrichDynamics(context.Background(), 1, list)
	require.Len(t, out, 1)
	assert.Equal(t, "hi", out[0]["content"])
	assert.Equal(t, []string{"http://img"}, out[0]["imageUrls"])
	assert.NotNil(t, out[0]["refVideo"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEnrichDynamics_NoUserMatch(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	now := time.Now()
	list := []*Dynamic{
		{ID: 1, UserID: 100, Content: "hi", DynamicType: 1, CreatedAt: now},
	}
	mock.ExpectQuery(`SELECT id, COALESCE`).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}))
	mock.ExpectQuery(`SELECT dynamic_id FROM dynamic_likes`).WillReturnRows(sqlmock.NewRows([]string{"dynamic_id"}))
	out := h.enrichDynamics(context.Background(), 0, list)
	require.Len(t, out, 1)
}

func TestEnrichDynamics_QueryError(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	now := time.Now()
	list := []*Dynamic{
		{ID: 1, UserID: 100, Content: "hi", DynamicType: 1, CreatedAt: now},
	}
	mock.ExpectQuery(`SELECT id, COALESCE`).WillReturnError(errors.New("db err"))
	mock.ExpectQuery(`SELECT dynamic_id FROM dynamic_likes`).WillReturnError(errors.New("db err"))
	out := h.enrichDynamics(context.Background(), 1, list)
	require.Len(t, out, 1)
}

func TestLoadUsers_Empty(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	users := h.loadUsers(context.Background(), nil)
	assert.Empty(t, users)
}

func TestLoadUsers_NilDB(t *testing.T) {
	h := &SocialHandler{db: nil}
	users := h.loadUsers(context.Background(), []int64{1})
	assert.Empty(t, users)
}

func TestLoadUsers_QueryError(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`SELECT id, COALESCE`).WillReturnError(errors.New("db err"))
	users := h.loadUsers(context.Background(), []int64{1})
	assert.Empty(t, users)
}

func TestLoadUsers_OK(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`SELECT id, COALESCE`).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).
		AddRow(1, "u", "n", "http://a", 5))
	users := h.loadUsers(context.Background(), []int64{1})
	require.Len(t, users, 1)
	assert.Equal(t, "n", users[1]["username"])
}

func TestLoadUsers_NicknameEmpty(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`SELECT id, COALESCE`).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).
		AddRow(1, "u", "", "http://a", 2))
	users := h.loadUsers(context.Background(), []int64{1})
	require.Len(t, users, 1)
	assert.Equal(t, "u", users[1]["username"])
}

func TestLoadManuscriptBrief_NilDB(t *testing.T) {
	h := &SocialHandler{db: nil}
	result := h.loadManuscriptBrief(context.Background(), 1)
	assert.Nil(t, result)
}

func TestLoadManuscriptBrief_Error(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`SELECT COALESCE`).WillReturnError(errors.New("no rows"))
	result := h.loadManuscriptBrief(context.Background(), 1)
	assert.Nil(t, result)
}

func TestLoadManuscriptBrief_OK(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`SELECT COALESCE`).WillReturnRows(sqlmock.NewRows([]string{"title", "cover_url", "duration_seconds", "view_count"}).
		AddRow("t", "http://c", 90, 100))
	result := h.loadManuscriptBrief(context.Background(), 1)
	require.NotNil(t, result)
	assert.Equal(t, "t", result["title"])
	assert.Equal(t, "01:30", result["duration"])
}

func TestLoadManuscriptBrief_HourFormat(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`SELECT COALESCE`).WillReturnRows(sqlmock.NewRows([]string{"title", "cover_url", "duration_seconds", "view_count"}).
		AddRow("t", "http://c", 3723, 100))
	result := h.loadManuscriptBrief(context.Background(), 1)
	require.NotNil(t, result)
	assert.Equal(t, "1:02:03", result["duration"])
}

func TestEnrichComments_Empty(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	out := h.enrichComments(context.Background(), 1, nil)
	assert.Empty(t, out)
}

func TestEnrichComments_WithData(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	now := time.Now()
	list := []*DynamicComment{
		{ID: 1, DynamicID: 10, UserID: 100, Content: "c", ReplyUserID: 200, LikeCount: 1, CreatedAt: now},
	}
	mock.ExpectQuery(`SELECT id, COALESCE`).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(100, "alice", "Alice", "http://a", 3))
	mock.ExpectQuery(`SELECT parent_id, COUNT`).WillReturnRows(sqlmock.NewRows([]string{"parent_id", "count"}))
	mock.ExpectQuery(`SELECT target_id FROM user_interactions`).WillReturnRows(sqlmock.NewRows([]string{"target_id"}))
	out := h.enrichComments(context.Background(), 1, list)
	require.Len(t, out, 1)
	assert.Equal(t, "c", out[0]["content"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEnrichComments_QueryErrors(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	now := time.Now()
	list := []*DynamicComment{
		{ID: 1, DynamicID: 10, UserID: 100, Content: "c", CreatedAt: now},
	}
	mock.ExpectQuery(`SELECT id, COALESCE`).WillReturnError(errors.New("db err"))
	mock.ExpectQuery(`SELECT parent_id, COUNT`).WillReturnError(errors.New("db err"))
	mock.ExpectQuery(`SELECT target_id FROM user_interactions`).WillReturnError(errors.New("db err"))
	out := h.enrichComments(context.Background(), 1, list)
	require.Len(t, out, 1)
}

func TestEnrichComments_UserNotFound(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	now := time.Now()
	list := []*DynamicComment{
		{ID: 1, DynamicID: 10, UserID: 100, Content: "c", ParentID: 5, ReplyUserID: 200, CreatedAt: now},
	}
	mock.ExpectQuery(`SELECT id, COALESCE`).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}))
	mock.ExpectQuery(`SELECT parent_id, COUNT`).WillReturnRows(sqlmock.NewRows([]string{"parent_id", "count"}))
	mock.ExpectQuery(`SELECT target_id FROM user_interactions`).WillReturnRows(sqlmock.NewRows([]string{"target_id"}))
	out := h.enrichComments(context.Background(), 0, list)
	require.Len(t, out, 1)
}

func TestNewSocialHandler(t *testing.T) {
	h, _ := newMockSocialHandler(t)
	assert.NotNil(t, h)
	assert.NotNil(t, h.followSvc)
	assert.NotNil(t, h.dynamicSvc)
	assert.NotNil(t, h.collectSvc)
	assert.NotNil(t, h.shareRepo)
}

func TestHandleDynamicShare_ShareDynamic(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`share_count = GREATEST`).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/share/1", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleCollection_Update_IsPublicFalse(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`UPDATE manuscript_collections SET title`).WillReturnResult(sqlmock.NewResult(0, 1))
	r := httptest.NewRequest("PUT", "/api/v1/collection/1", strings.NewReader(""))
	r.Header.Set("X-User-Id", "1")
	r.ParseForm()
	r.Form.Set("isPublic", "false")
	w := httptest.NewRecorder()
	muxForSocial(h).ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDynamicPublish_PublishError(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectQuery(`INSERT INTO user_dynamics`).WillReturnError(errors.New("publish fail"))
	w := doReq(muxForSocial(h), "POST", "/api/v1/dynamic/publish?content=hello&type=1", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleDynamicCommentDelete_NonDelete(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	_ = mock
	w := doReq(muxForSocial(h), "GET", "/api/v1/dynamic/comment/delete/5", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}
