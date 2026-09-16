package social

import (
	"context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/pkg/auth"
)

// argConverter 允许 sqlmock 处理 []int64 参数（生产环境由 lib/pq 转换为数组字面量）。
type argConverter struct{}

func (argConverter) ConvertValue(v interface{}) (driver.Value, error) {
	if arr, ok := v.([]int64); ok {
		s := "{"
		for i, x := range arr {
			if i > 0 {
				s += ","
			}
			s += strconv.FormatInt(x, 10)
		}
		return s + "}", nil
	}
	return driver.DefaultParameterConverter.ConvertValue(v)
}

func newMockFollowHandler(t *testing.T) (*FollowHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
		sqlmock.ValueConverterOption(argConverter{}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	svc := NewFollowService(NewFollowRepository(db))
	return NewFollowHandler(svc, db, auth.NewJWT("test-secret")), mock
}

func doReq(h http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func muxForFollow(h *FollowHandler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestFollowHandler_Register(t *testing.T) {
	h, mock := newMockFollowHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	w := doReq(mux, "GET", "/api/v1/follow/check/2", "", map[string]string{"X-User-Id": "1"})
	mock.ExpectQuery(`SELECT EXISTS`).WithArgs(int64(1), int64(2)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFollowHandler_handleFollow_NotFoundPath(t *testing.T) {
	h, _ := newMockFollowHandler(t)
	w := doReq(muxForFollow(h), "POST", "/api/v1/follow/", "", map[string]string{})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFollowHandler_handleFollow_MeRouteReturns(t *testing.T) {
	h, _ := newMockFollowHandler(t)
	// parts[0] == "me" -> handler returns without writing; status stays 200 from recorder default
	w := doReq(muxForFollow(h), "POST", "/api/v1/follow/me", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFollowHandler_handleFollow_Unauthorized(t *testing.T) {
	h, _ := newMockFollowHandler(t)
	w := doReq(muxForFollow(h), "POST", "/api/v1/follow/2", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestFollowHandler_handleFollow_InvalidTargetID(t *testing.T) {
	h, _ := newMockFollowHandler(t)
	w := doReq(muxForFollow(h), "POST", "/api/v1/follow/abc", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFollowHandler_handleFollow_Self(t *testing.T) {
	h, _ := newMockFollowHandler(t)
	w := doReq(muxForFollow(h), "POST", "/api/v1/follow/1", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFollowHandler_handleFollow_UserNotExists(t *testing.T) {
	h, mock := newMockFollowHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id = \$1\)`).WithArgs(int64(2)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	w := doReq(muxForFollow(h), "POST", "/api/v1/follow/2", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFollowHandler_handleFollow_Success(t *testing.T) {
	h, mock := newMockFollowHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id = \$1\)`).WithArgs(int64(2)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO follows`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET following_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET follower_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	w := doReq(muxForFollow(h), "POST", "/api/v1/follow/2", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowHandler_handleFollow_FollowError(t *testing.T) {
	h, mock := newMockFollowHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id = \$1\)`).WithArgs(int64(2)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO follows`).WillReturnError(errors.New("boom"))

	w := doReq(muxForFollow(h), "POST", "/api/v1/follow/2", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFollowHandler_handleFollow_Delete(t *testing.T) {
	h, mock := newMockFollowHandler(t)
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM follows`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET following_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET follower_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	w := doReq(muxForFollow(h), "DELETE", "/api/v1/follow/2", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowHandler_handleFollow_MethodNotAllowed(t *testing.T) {
	h, _ := newMockFollowHandler(t)
	w := doReq(muxForFollow(h), "PUT", "/api/v1/follow/2", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestFollowHandler_handleCheck(t *testing.T) {
	h, mock := newMockFollowHandler(t)

	w := doReq(muxForFollow(h), "GET", "/api/v1/follow/check/", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mock.ExpectQuery(`SELECT EXISTS`).WithArgs(int64(1), int64(2)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	w = doReq(muxForFollow(h), "GET", "/api/v1/follow/check/2", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowHandler_handleCheck_WithJWT(t *testing.T) {
	h, mock := newMockFollowHandler(t)
	token, err := h.jwt.Generate(7)
	require.NoError(t, err)

	mock.ExpectQuery(`SELECT EXISTS`).WithArgs(int64(7), int64(2)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	w := doReq(muxForFollow(h), "GET", "/api/v1/follow/check/2", "", map[string]string{"Authorization": "Bearer " + token})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowHandler_handleMyFollowers(t *testing.T) {
	h, mock := newMockFollowHandler(t)

	w := doReq(muxForFollow(h), "GET", "/api/v1/follow/me/followers", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	mock.ExpectQuery(`SELECT follower_id FROM follows`).WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"follower_id"}).AddRow(9))
	w = doReq(muxForFollow(h), "GET", "/api/v1/follow/me/followers?page=1&page_size=20", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowHandler_handleMyFollowing(t *testing.T) {
	h, mock := newMockFollowHandler(t)

	w := doReq(muxForFollow(h), "GET", "/api/v1/follow/me/following", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	mock.ExpectQuery(`SELECT following_id FROM follows`).WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"following_id"}).AddRow(9))
	w = doReq(muxForFollow(h), "GET", "/api/v1/follow/me/following", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowHandler_handleUserFollows(t *testing.T) {
	h, mock := newMockFollowHandler(t)

	w := doReq(muxForFollow(h), "GET", "/api/v1/follow/user/1", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)

	mock.ExpectQuery(`SELECT following_id FROM follows`).WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"following_id"}).AddRow(2))
	w = doReq(muxForFollow(h), "GET", "/api/v1/follow/user/1/following", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectQuery(`SELECT follower_id FROM follows`).WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"follower_id"}).AddRow(2))
	w = doReq(muxForFollow(h), "GET", "/api/v1/follow/user/1/followers", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	w = doReq(muxForFollow(h), "GET", "/api/v1/follow/user/1/bad", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowHandler_loadUserBriefs(t *testing.T) {
	h, mock := newMockFollowHandler(t)

	out := h.loadUserBriefs(context.Background(), nil)
	assert.Empty(t, out)

	mock.ExpectQuery(`SELECT id, COALESCE\(username,''\)`).WithArgs("{9}").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level", "signature"}).
			AddRow(9, "u", "nick", "http://a", 3, "sig"))
	out = h.loadUserBriefs(context.Background(), []int64{9})
	require.Len(t, out, 1)
	assert.Equal(t, "u", out[0]["username"])

	mock.ExpectQuery(`SELECT id, COALESCE\(username,''\)`).WillReturnError(errors.New("boom"))
	out = h.loadUserBriefs(context.Background(), []int64{1})
	assert.Empty(t, out)
}
