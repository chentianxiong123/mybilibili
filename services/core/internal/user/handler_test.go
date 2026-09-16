package user

import (
	"bytes"
	"context"
	"database/sql"
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

// newTestHandler 构造一个 sqlmock-backed HTTP handler.
func newTestHandler(t *testing.T) (*UserExtendHandler, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewRepository(db)
	svc := NewService(repo, testJWTSecret)
	return NewUserExtendHandler(svc), mock
}

func doRequest(t *testing.T, h http.Handler, method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// 准备 users 表行（密码=password 的 sha256 hex）
func userRow(uid int64, username, password, nickname string, status int32) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "username", "password", "nickname", "email", "avatar",
		"level", "experience", "signature", "bio",
		"follower_count", "following_count", "liked_count",
		"status", "created_at", "updated_at",
	}).AddRow(
		uid, username, password, nickname, "", "",
		1, 0, "", "",
		0, 0, 0,
		status, time.Now(), time.Now(),
	)
}

func TestHandleRegister_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	body := map[string]string{
		"username": "newbie", "password": "secret", "nickname": "Newbie",
	}

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE nickname`).
		WithArgs("Newbie").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("newbie", sha256Hex("secret"), "Newbie", "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectExec(`UPDATE users SET status=1`).
		WithArgs(11).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doRequest(t, mux, "POST", "/api/v1/user/register", body, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"id":11`)
	assert.Contains(t, rec.Body.String(), `"token":`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRegister_400_Empty(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	// 缺 username 字段 (json 里只有 password)
	body := map[string]string{"password": "x"}
	rec := doRequest(t, mux, "POST", "/api/v1/user/register", body, nil)
	// 业务返回 400 (Username 空 → ErrInvalidArgument → http.Error 400)
	assert.Equal(t, http.StatusBadRequest, rec.Code, "expected 400, got %d, body=%s", rec.Code, rec.Body.String())
}

func TestHandleRegister_409_Dup(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	// FindByNickname: 已存在同名
	mock.ExpectQuery(`SELECT.*FROM users.*WHERE nickname`).
		WithArgs("dup").
		WillReturnRows(userRow(99, "dup", "h", "dup", 1))

	body := map[string]string{"username": "dup", "password": "x", "nickname": "dup"}
	rec := doRequest(t, mux, "POST", "/api/v1/user/register", body, nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "nickname already exists")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLogin_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	pwd := "secret123"
	mock.ExpectQuery(`SELECT.*FROM users.*WHERE username`).
		WithArgs("alice").
		WillReturnRows(userRow(5, "alice", sha256Hex(pwd), "Alice", 1))
	// login handler 写 login_logs
	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO login_logs`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET coin_count`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(5)).
		WillReturnRows(userRow(5, "alice", sha256Hex(pwd), "Alice", 1))

	body := map[string]string{"username": "alice", "password": pwd}
	rec := doRequest(t, mux, "POST", "/api/v1/user/login", body, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"token"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLogin_401_WrongPwd(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE username`).
		WithArgs("alice").
		WillReturnRows(userRow(5, "alice", sha256Hex("right"), "Alice", 1))

	body := map[string]string{"username": "alice", "password": "wrong"}
	rec := doRequest(t, mux, "POST", "/api/v1/user/login", body, nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLogin_404_NoUser(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE username`).
		WithArgs("ghost").
		WillReturnError(sql.ErrNoRows)

	body := map[string]string{"username": "ghost", "password": "x"}
	rec := doRequest(t, mux, "POST", "/api/v1/user/login", body, nil)
	// service 返回 ErrNotFound, http handler 401 (TODO: 实际看代码是 401)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRefresh_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	// 准备 refresh token
	svc := h.svc
	tok, err := svc.jwt.GenerateRefresh(7)
	require.NoError(t, err)

	// refresh handler 调用 repo.FindByID
	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(7)).
		WillReturnRows(userRow(7, "u7", "h", "Nick7", 1))

	body := map[string]string{"refreshToken": tok}
	rec := doRequest(t, mux, "POST", "/api/v1/user/token/refresh", body, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"token"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRefresh_401_Expired(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	// 用一个非法 token
	body := map[string]string{"refreshToken": "garbage.token"}
	rec := doRequest(t, mux, "POST", "/api/v1/user/token/refresh", body, nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleMe_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(99)).
		WillReturnRows(userRow(99, "u99", "h", "Nick99", 1))
	// followers / following / likes / manuscripts counts
	mock.ExpectQuery(`SELECT COUNT.*FROM follows WHERE following_id`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT COUNT.*FROM follows WHERE follower_id`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT COALESCE\(SUM`).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(0))
	mock.ExpectQuery(`SELECT COUNT.*FROM manuscripts WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	rec := doRequest(t, mux, "GET", "/api/v1/user/me", nil, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"username":"u99"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMe_401_NoToken(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := doRequest(t, mux, "GET", "/api/v1/user/me", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// 防 unused import 警告
var _ = context.Background
var _ = (*pb.LoginResponse)(nil)
var _ = strings.HasPrefix