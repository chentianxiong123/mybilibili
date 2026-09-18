package comment

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCreatorHandler(t *testing.T) (*CreatorCommentHTTPHandler, sqlmock.Sqlmock, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	repoDB, repoMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { repoDB.Close() })

	repo := NewCommentRepository(repoDB)
	svc := NewCommentService(repo)
	svc.SetDB(db)
	h := NewCreatorCommentHTTPHandler(repo, svc, db)
	return h, mock, repoMock
}

func TestCreatorCommentHTTPHandler_Register(t *testing.T) {
	h, _, _ := newCreatorHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	// Just verify it doesn't panic
	assert.NotNil(t, mux)
}

func TestCreatorCommentHTTPHandler_HandleList_MethodNotAllowed(t *testing.T) {
	h, _, _ := newCreatorHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/creator/comments", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleList(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestCreatorCommentHTTPHandler_HandleList_Unauthorized(t *testing.T) {
	h, _, _ := newCreatorHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/creator/comments", nil)
	w := httptest.NewRecorder()
	h.handleList(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreatorCommentHTTPHandler_HandleList_Success(t *testing.T) {
	h, _, repoMock := newCreatorHandler(t)

	// ListCommentsByCreator returns empty (commentType="all" -> queries both comments and replies)
	repoMock.ExpectQuery(`SELECT comments\.id.* FROM comments`).WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}))
	repoMock.ExpectQuery(`SELECT r\..* FROM replies r`).WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "manuscript_id", "user_id", "content", "like_count", "reply_to_user_id", "created_at"}))
	repoMock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	repoMock.ExpectQuery(`SELECT COUNT\(\*\) FROM replies`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	repoMock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users WHERE id = ANY`).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}))
	repoMock.ExpectQuery(`SELECT id, title, cover_url FROM manuscripts WHERE id = ANY`).WillReturnRows(sqlmock.NewRows([]string{"id", "title", "cover_url"}))

	req := httptest.NewRequest("GET", "/api/v1/creator/comments?sort=likes&commentType=all", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleList(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreatorCommentHTTPHandler_HandleByPath_Unauthorized(t *testing.T) {
	h, _, _ := newCreatorHandler(t)
	req := httptest.NewRequest("DELETE", "/api/v1/creator/comments/123", nil)
	w := httptest.NewRecorder()
	h.handleByPath(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreatorCommentHTTPHandler_HandleByPath_InvalidID(t *testing.T) {
	h, _, _ := newCreatorHandler(t)
	req := httptest.NewRequest("DELETE", "/api/v1/creator/comments/abc", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleByPath(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatorCommentHTTPHandler_HandleByPath_DELETE_NotFound(t *testing.T) {
	h, _, repoMock := newCreatorHandler(t)
	repoMock.ExpectExec(`DELETE FROM comments WHERE id = \$1 AND manuscript_id IN`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := httptest.NewRequest("DELETE", "/api/v1/creator/comments/123", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleByPath(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreatorCommentHTTPHandler_HandleByPath_DELETE_Success(t *testing.T) {
	h, _, repoMock := newCreatorHandler(t)
	repoMock.ExpectExec(`DELETE FROM comments WHERE id = \$1 AND manuscript_id IN`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("DELETE", "/api/v1/creator/comments/123", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleByPath(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreatorCommentHTTPHandler_HandleByPath_POST_EmptyContent(t *testing.T) {
	h, _, _ := newCreatorHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/creator/comments/123?content=", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleByPath(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatorCommentHTTPHandler_HandleByPath_POST_CommentNotOwned(t *testing.T) {
	h, _, repoMock := newCreatorHandler(t)
	repoMock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments WHERE id = \$1 AND user_id = \$2`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	req := httptest.NewRequest("POST", "/api/v1/creator/comments/123?content=hello", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleByPath(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreatorCommentHTTPHandler_HandleByPath_MethodNotAllowed(t *testing.T) {
	h, _, _ := newCreatorHandler(t)
	req := httptest.NewRequest("PUT", "/api/v1/creator/comments/123", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleByPath(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestCreatorCommentHTTPHandler_HandleDeleteReply_MethodNotAllowed(t *testing.T) {
	h, _, _ := newCreatorHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/creator/comments/reply/123", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleDeleteReply(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestCreatorCommentHTTPHandler_HandleDeleteReply_Unauthorized(t *testing.T) {
	h, _, _ := newCreatorHandler(t)
	req := httptest.NewRequest("DELETE", "/api/v1/creator/comments/reply/123", nil)
	w := httptest.NewRecorder()
	h.handleDeleteReply(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreatorCommentHTTPHandler_HandleDeleteReply_InvalidID(t *testing.T) {
	h, _, _ := newCreatorHandler(t)
	req := httptest.NewRequest("DELETE", "/api/v1/creator/comments/reply/abc", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleDeleteReply(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatorCommentHTTPHandler_HandleDeleteReply_NotFound(t *testing.T) {
	h, _, repoMock := newCreatorHandler(t)
	repoMock.ExpectExec(`DELETE FROM replies WHERE id = \$1 AND comment_id IN`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := httptest.NewRequest("DELETE", "/api/v1/creator/comments/reply/456", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleDeleteReply(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreatorCommentHTTPHandler_HandleDeleteReply_Success(t *testing.T) {
	h, _, repoMock := newCreatorHandler(t)
	repoMock.ExpectExec(`DELETE FROM replies WHERE id = \$1 AND comment_id IN`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("DELETE", "/api/v1/creator/comments/reply/456", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleDeleteReply(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// --- Unit tests for helper functions ---

func TestToInt64(t *testing.T) {
	tests := []struct {
		input    any
		expected int64
	}{
		{int64(42), 42},
		{int32(42), 42},
		{int(42), 42},
		{"string", 0},
		{nil, 0},
		{3.14, 0},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, toInt64(tt.input))
	}
}

func TestToString(t *testing.T) {
	assert.Equal(t, "hello", toString("hello"))
	assert.Equal(t, "", toString(42))
	assert.Equal(t, "", toString(nil))
}

func TestUserNicknameOf(t *testing.T) {
	assert.Equal(t, "", userNicknameOf(nil))
	assert.Equal(t, "Nick", userNicknameOf(&User{Nickname: "Nick", Username: "user1"}))
	assert.Equal(t, "user1", userNicknameOf(&User{Nickname: "", Username: "user1"}))
}

func TestUserAvatarOf(t *testing.T) {
	assert.Equal(t, "", userAvatarOf(nil))
	assert.Equal(t, "avatar.jpg", userAvatarOf(&User{Avatar: "avatar.jpg"}))
}

func TestManuscriptTitleOf(t *testing.T) {
	assert.Equal(t, "", manuscriptTitleOf(nil))
	assert.Equal(t, "Title", manuscriptTitleOf(&ManuscriptBrief{Title: "Title"}))
}

func TestManuscriptCoverOf(t *testing.T) {
	assert.Equal(t, "", manuscriptCoverOf(nil))
	assert.Equal(t, "cover.jpg", manuscriptCoverOf(&ManuscriptBrief{Cover: "cover.jpg"}))
}

func TestUserNameOf(t *testing.T) {
	assert.Equal(t, "", userNameOf(nil))
	assert.Equal(t, "user1", userNameOf(&User{Username: "user1"}))
}

func TestToNullInt64(t *testing.T) {
	assert.Equal(t, sql.NullInt64{}, toNullInt64(0))
	assert.Equal(t, sql.NullInt64{Int64: 42, Valid: true}, toNullInt64(42))
}

func TestSortCreatorItems(t *testing.T) {
	list := []map[string]any{
		{"likeCount": int64(5), "createTime": "2024-01-02"},
		{"likeCount": int64(10), "createTime": "2024-01-01"},
		{"likeCount": int64(3), "createTime": "2024-01-03"},
	}

	sortCreatorItems(list, "likes")
	assert.Equal(t, int64(10), toInt64(list[0]["likeCount"]))
	assert.Equal(t, int64(5), toInt64(list[1]["likeCount"]))
	assert.Equal(t, int64(3), toInt64(list[2]["likeCount"]))

	sortCreatorItems(list, "latest")
	// 2024-01-03 > 2024-01-02 > 2024-01-01
	assert.Equal(t, "2024-01-03", toString(list[0]["createTime"]))
	assert.Equal(t, "2024-01-02", toString(list[1]["createTime"]))
	assert.Equal(t, "2024-01-01", toString(list[2]["createTime"]))
}

func TestFindManuscriptsByIDs_Empty(t *testing.T) {
	h, _, _ := newCreatorHandler(t)
	result := h.findManuscriptsByIDs(context.Background(), []int64{})
	assert.Empty(t, result)
}

func TestToIDList(t *testing.T) {
	set := map[int64]struct{}{1: {}, 2: {}, 3: {}}
	result := toIDList(set)
	assert.Len(t, result, 3)
}
