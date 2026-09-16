package comment

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "mybilibili/pkg/pb"
)

func doReq(h http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func muxForPublic(h *PublicAPIHandler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func newPublicHandler(t *testing.T) (*PublicAPIHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	svc := NewCommentService(NewCommentRepository(db))
	svc.SetDB(db)
	return NewPublicAPIHandler(svc, db), mock
}

func TestCommentToMap(t *testing.T) {
	c := &pb.CommentInfo{
		Id: 1, ManuscriptId: 9, UserId: 2, UserName: "nick", Content: "hello",
		LikeCount: 3, CreatedAt: "2026-01-01T00:00:00Z", Liked: true,
		Replies: []*pb.ReplyInfo{{Id: 5, UserId: 6, UserName: "r", Content: "x", CreatedAt: "2026-01-01T00:00:00Z"}},
	}
	m := commentToMap(c)
	assert.Equal(t, "nick", m["username"])
	assert.Equal(t, "2026-01-01T00:00:00Z", m["createTime"])
	reps, ok := m["replies"].([]interface{})
	require.True(t, ok)
	rm := reps[0].(map[string]interface{})
	assert.Equal(t, "r", rm["username"])
}

func TestCommentListToJSON(t *testing.T) {
	out := commentListToJSON(nil)
	assert.Empty(t, out)
	out = commentListToJSON([]*pb.CommentInfo{{Id: 1, UserName: "u", CreatedAt: "2026-01-01T00:00:00Z"}})
	require.Len(t, out, 1)
	assert.Equal(t, "u", out[0]["username"])
}

func TestReplyListToJSON_And_ReplyToMapJSON(t *testing.T) {
	info := &pb.ReplyInfo{Id: 1, UserId: 2, UserName: "u", Content: "c", CreatedAt: "2026-01-01T00:00:00Z"}
	m := replyToMapJSON(info)
	assert.Equal(t, "u", m["username"])
	assert.Equal(t, "2026-01-01T00:00:00Z", m["createTime"])

	m = replyToMapJSON(nil)
	assert.Empty(t, m)

	list := replyListToJSON([]*pb.ReplyInfo{info})
	require.Len(t, list, 1)
}

func TestDecodeCommentBody_JSON(t *testing.T) {
	r := httptest.NewRequest("POST", "/x", strings.NewReader(`{"manuscriptId":123,"content":"hi","num":42}`))
	r.Header.Set("Content-Type", "application/json")
	var msID, content, num string
	err := decodeCommentBody(r, map[string]*string{"manuscriptId": &msID, "content": &content, "num": &num})
	require.NoError(t, err)
	assert.Equal(t, "123", msID)
	assert.Equal(t, "hi", content)
	assert.Equal(t, "42", num)
}

func TestDecodeCommentBody_Form(t *testing.T) {
	r := httptest.NewRequest("POST", "/x", strings.NewReader("manuscriptId=7&content=ok"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var msID, content string
	err := decodeCommentBody(r, map[string]*string{"manuscriptId": &msID, "content": &content})
	require.NoError(t, err)
	assert.Equal(t, "7", msID)
	assert.Equal(t, "ok", content)
}

func TestDecodeCommentBody_InvalidJSON(t *testing.T) {
	r := httptest.NewRequest("POST", "/x", strings.NewReader(`bad`))
	r.Header.Set("Content-Type", "application/json")
	err := decodeCommentBody(r, map[string]*string{"content": new(string)})
	assert.Error(t, err)
}

func TestPublicAPIHandler_Register(t *testing.T) {
	h, _ := newPublicHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	// a route that short-circuits
	w := doReq(mux, "GET", "/api/v1/comment/batch-like-counts", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPublicAPIHandler_handleCommentList(t *testing.T) {
	h, mock := newPublicHandler(t)
	now := time.Now()

	mock.ExpectQuery(`FROM comments WHERE manuscript_id = \$1 AND status = 0 ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(int64(9), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}).
			AddRow(1, 9, 200, "hello", 3, 1, 0, now, now))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).WithArgs(int64(200)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(200, "u", "nick", "a", 1))

	w := doReq(muxForPublic(h), "GET", "/api/v1/comment/list?manuscriptId=9", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPublicAPIHandler_handleCommentAdd(t *testing.T) {
	h, mock := newPublicHandler(t)
	now := time.Now()

	w := doReq(muxForPublic(h), "GET", "/api/v1/comment/add", `{}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	w = doReq(muxForPublic(h), "POST", "/api/v1/comment/add", `{"manuscriptId":9,"content":"hello"}`, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = doReq(muxForPublic(h), "POST", "/api/v1/comment/add", `bad`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = doReq(muxForPublic(h), "POST", "/api/v1/comment/add", `{"manuscriptId":0,"content":""}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mock.ExpectQuery(`INSERT INTO comments`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WithArgs(int64(9), int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET comment_count = comment_count \+ 1`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO content_reviews`).WithArgs("comment", int64(1), "hello").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(1, "u", "nick", "a", 1))
	mock.ExpectQuery(`SELECT level, experience FROM users`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"level", "experience"}).AddRow(1, 0))
	mock.ExpectExec(`UPDATE users SET experience`).WithArgs(int64(5), int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	w = doReq(muxForPublic(h), "POST", "/api/v1/comment/add", `{"manuscriptId":9,"content":"hello"}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPublicAPIHandler_handleCommentReply(t *testing.T) {
	h, mock := newPublicHandler(t)
	now := time.Now()

	w := doReq(muxForPublic(h), "GET", "/api/v1/comment/reply", `{}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	w = doReq(muxForPublic(h), "POST", "/api/v1/comment/reply", `{"commentId":9,"content":"hi"}`, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = doReq(muxForPublic(h), "POST", "/api/v1/comment/reply", `{"commentId":0,"content":""}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mock.ExpectQuery(`INSERT INTO replies`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))
	mock.ExpectExec(`UPDATE comments SET reply_count = reply_count \+ 1`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id, manuscript_id, user_id, content`).WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}).
			AddRow(9, 5, 300, "parent", 1, 1, 0, now, now))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WithArgs(int64(5), int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET comment_count = comment_count \+ 1`).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(1, "u", "nick", "a", 1))
	mock.ExpectQuery(`SELECT level, experience FROM users`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"level", "experience"}).AddRow(1, 0))
	mock.ExpectExec(`UPDATE users SET experience`).WithArgs(int64(2), int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	w = doReq(muxForPublic(h), "POST", "/api/v1/comment/reply", `{"commentId":9,"content":"hi"}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPublicAPIHandler_handleCommentLike(t *testing.T) {
	h, mock := newPublicHandler(t)
	h.commentSvc.SetNotifier(&fakeNotifier{})

	w := doReq(muxForPublic(h), "POST", "/api/v1/comment/9/like", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = doReq(muxForPublic(h), "PUT", "/api/v1/comment/9/like", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO user_interactions`).WithArgs(int64(1), "COMMENT", int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE comments SET like_count = like_count \+ 1`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT user_id FROM comments`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(50))
	w = doReq(muxForPublic(h), "POST", "/api/v1/comment/9/like", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectExec(`DELETE FROM user_interactions`).WithArgs(int64(1), "COMMENT", int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE comments SET like_count = GREATEST`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doReq(muxForPublic(h), "DELETE", "/api/v1/comment/9/like", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPublicAPIHandler_handleCommentReplies(t *testing.T) {
	h, mock := newPublicHandler(t)
	now := time.Now()

	mock.ExpectQuery(`FROM replies WHERE comment_id = \$1 AND status = 'NORMAL'`).
		WithArgs(int64(9), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "user_id", "reply_to_user_id", "content", "like_count", "status", "created_at", "updated_at"}).
			AddRow(1, 9, 200, nil, "r", 1, "NORMAL", now, now))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).WithArgs(int64(200)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(200, "u", "nick", "a", 1))

	w := doReq(muxForPublic(h), "GET", "/api/v1/comment/9/replies", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPublicAPIHandler_handleReplyLike(t *testing.T) {
	h, mock := newPublicHandler(t)
	h.commentSvc.SetNotifier(&fakeNotifier{})

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM replies`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO user_interactions`).WithArgs(int64(1), "REPLY", int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE replies SET like_count = like_count \+ 1`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT user_id FROM replies`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(50))
	w := doReq(muxForPublic(h), "POST", "/api/v1/comment/reply/9/like", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectExec(`DELETE FROM user_interactions`).WithArgs(int64(1), "REPLY", int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE replies SET like_count = GREATEST`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doReq(muxForPublic(h), "DELETE", "/api/v1/comment/reply/9/like", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)

	w = doReq(muxForPublic(h), "PUT", "/api/v1/comment/reply/9/like", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPublicAPIHandler_handleBatchLikeCounts(t *testing.T) {
	h, mock := newPublicHandler(t)

	w := doReq(muxForPublic(h), "POST", "/api/v1/comment/batch-like-counts?ids=1", "", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	mock.ExpectQuery(`SELECT id, like_count FROM comments WHERE id IN \(\$1,\$2\)`).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "like_count"}).AddRow(1, 5).AddRow(2, 3))
	mock.ExpectQuery(`SELECT target_id FROM user_interactions`).
		WithArgs(int64(0), "COMMENT", int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"target_id"}).AddRow(1))
	w = doReq(muxForPublic(h), "GET", "/api/v1/comment/batch-like-counts?ids=1,2", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]interface{})
	item := data["1"].(map[string]interface{})
	assert.Equal(t, float64(5), item["like_count"])
	assert.Equal(t, true, item["is_liked"])
	require.NoError(t, mock.ExpectationsWereMet())
}
