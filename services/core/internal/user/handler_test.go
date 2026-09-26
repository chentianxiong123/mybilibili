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

	"mybilibili/pkg/auth"
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
		"status", "coin_count", "gender", "created_at", "updated_at",
	}).AddRow(
		uid, username, password, nickname, "", "",
		1, 0, "", "",
		0, 0, 0,
		status, 0, 0, time.Now(), time.Now(),
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
	// 默认不把凭证回进响应体：浏览器读得到 body，读得到就能被 XSS 偷走
	assert.NotContains(t, rec.Body.String(), `"token"`)
	assert.NotContains(t, rec.Body.String(), `"refresh_token"`)
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
	assert.NotContains(t, rec.Body.String(), `"token"`)
	assert.NotContains(t, rec.Body.String(), `"refresh_token"`)
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
	assert.NotContains(t, rec.Body.String(), `"token"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 显式带 includeTokens:true 时才把凭证回进 body（原生客户端用）。
func TestHandleLogin_200_IncludeTokens(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	pwd := "secret123"
	mock.ExpectQuery(`SELECT.*FROM users.*WHERE username`).
		WithArgs("alice").
		WillReturnRows(userRow(5, "alice", sha256Hex(pwd), "Alice", 1))
	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO login_logs`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET coin_count`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(5)).
		WillReturnRows(userRow(5, "alice", sha256Hex(pwd), "Alice", 1))

	body := map[string]interface{}{"username": "alice", "password": pwd, "includeTokens": true}
	rec := doRequest(t, mux, "POST", "/api/v1/user/login", body, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"token"`)
	assert.Contains(t, rec.Body.String(), `"refresh_token"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 用户刷新口必须拒收管理员刷新令牌：两套 ID 命名空间不同，混用即越权。
func TestHandleRefresh_401_AdminRefreshTokenRejected(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	tok, err := h.svc.jwt.GenerateAdminRefresh(7)
	require.NoError(t, err)

	body := map[string]string{"refreshToken": tok}
	rec := doRequest(t, mux, "POST", "/api/v1/user/token/refresh", body, nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.NotContains(t, rec.Body.String(), `"token"`)
}

// 同一张刷新令牌重放第二次 → 401。
func TestHandleRefresh_401_Replay(t *testing.T) {
	// 一次性消费依赖 TokenStore；测试环境默认没有，装一个内存实现
	prev := auth.CurrentTokenStore()
	auth.SetTokenStore(auth.NewMemoryTokenStore())
	defer auth.SetTokenStore(prev)

	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	tok, err := h.svc.jwt.GenerateRefresh(7)
	require.NoError(t, err)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(7)).
		WillReturnRows(userRow(7, "u7", "h", "Nick7", 1))

	first := doRequest(t, mux, "POST", "/api/v1/user/token/refresh", map[string]string{"refreshToken": tok}, nil)
	require.Equal(t, http.StatusOK, first.Code)

	second := doRequest(t, mux, "POST", "/api/v1/user/token/refresh", map[string]string{"refreshToken": tok}, nil)
	assert.Equal(t, http.StatusUnauthorized, second.Code)
	assert.Contains(t, second.Body.String(), "already used")
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
	// dynamic count
	mock.ExpectQuery(`SELECT COUNT.*FROM user_dynamics WHERE user_id`).
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

func TestHandleGetUser_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(1)).
		WillReturnRows(userRow(1, "alice", "hash", "Alice", 1))
	mock.ExpectQuery(`SELECT COUNT.*FROM follows WHERE following_id`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(`SELECT COUNT.*FROM follows WHERE follower_id`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(like_count\)`).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(10))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(views\)`).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(100))
	mock.ExpectQuery(`SELECT COUNT.*FROM user_dynamics`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`SELECT COALESCE\(signature.*FROM users WHERE id`).
		WillReturnRows(sqlmock.NewRows([]string{"signature", "announcement", "bio", "gender", "birthdate"}).
			AddRow("hello", "new post", "bio text", 1, nil))
	mock.ExpectQuery(`SELECT tag_name FROM user_tags`).
		WillReturnRows(sqlmock.NewRows([]string{"tag_name"}).AddRow("golang").AddRow("tech"))

	rec := doRequest(t, mux, "GET", "/api/v1/user/1", nil, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"username":"alice"`)
	assert.Contains(t, rec.Body.String(), `"followerCount":5`)
	assert.Contains(t, rec.Body.String(), `"followingCount":3`)
	assert.Contains(t, rec.Body.String(), `"totalLikeCount":10`)
	assert.Contains(t, rec.Body.String(), `"totalViewCount":100`)
	assert.Contains(t, rec.Body.String(), `"dynamicCount":2`)
	assert.Contains(t, rec.Body.String(), `"tags"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetUser_404(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(9999)).
		WillReturnError(sql.ErrNoRows)

	rec := doRequest(t, mux, "GET", "/api/v1/user/9999", nil, nil)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "user not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUpdateUser_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectExec(`UPDATE users SET signature`).
		WithArgs("new signature", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := map[string]interface{}{"signature": "new signature"}
	rec := doRequest(t, mux, "PUT", "/api/v1/user/1", body, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUpdateUser_401(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	body := map[string]interface{}{"nickname": "new nickname"}
	rec := doRequest(t, mux, "PUT", "/api/v1/user/me", body, nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleGetFollowers_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT id, user_id, ip, user_agent, status, login_time FROM login_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "ip", "user_agent", "status", "login_time"}).
			AddRow(1, 99, "127.0.0.1", "Mozilla/5.0", 1, "2025-01-01T00:00:00Z"))
	mock.ExpectQuery(`SELECT COUNT.*FROM login_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rec := doRequest(t, mux, "GET", "/api/v1/user/login-logs", nil, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"list"`)
	assert.Contains(t, rec.Body.String(), `"total"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetFollowing_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users WHERE id IN`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).
			AddRow(1, "alice", "Alice", "", 1))

	body := []int64{1}
	rec := doRequest(t, mux, "POST", "/api/v1/user/batch", body, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"username":"alice"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCheckFollow_200(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := doRequest(t, mux, "GET", "/api/v1/user/default-avatar?name=Alice", nil, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `<text`)
	assert.Contains(t, rec.Body.String(), `A`)
}

func TestHandleTokenRefresh_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	svc := h.svc
	tok, err := svc.jwt.GenerateRefresh(7)
	require.NoError(t, err)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(7)).
		WillReturnRows(userRow(7, "u7", "h", "Nick7", 1))

	body := map[string]string{"refreshToken": tok}
	rec := doRequest(t, mux, "POST", "/api/v1/user/token/refresh", body, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotContains(t, rec.Body.String(), `"token"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleTokenRefresh_401(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	body := map[string]string{"refreshToken": "invalid.token.value"}
	rec := doRequest(t, mux, "POST", "/api/v1/user/token/refresh", body, nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleMeAvatar_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectExec(`UPDATE users SET avatar`).
		WithArgs("https://example.com/avatar.png", int64(50)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := map[string]string{"avatar": "https://example.com/avatar.png"}
	rec := doRequest(t, mux, "POST", "/api/v1/user/me/avatar", body, map[string]string{
		"X-User-Id": "50",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMeAvatar_400_Empty(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	body := map[string]string{"avatar": ""}
	rec := doRequest(t, mux, "POST", "/api/v1/user/me/avatar", body, map[string]string{
		"X-User-Id": "50",
	})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleMeAvatar_405_Method(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := doRequest(t, mux, "GET", "/api/v1/user/me/avatar", nil, map[string]string{
		"X-User-Id": "50",
	})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestHandleMePut_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT COUNT.*FROM users WHERE nickname`).
		WithArgs("new nick", int64(50)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`UPDATE users SET nickname`).
		WithArgs("new nick", int64(50)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := map[string]interface{}{"nickname": "new nick"}
	rec := doRequest(t, mux, "PUT", "/api/v1/user/me", body, map[string]string{
		"X-User-Id": "50",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDefaultAvatar_200(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := doRequest(t, mux, "GET", "/api/v1/user/default-avatar?name=Bob", nil, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `<svg`)
	assert.Contains(t, rec.Body.String(), `B`)
}

func TestHandleBatch_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users WHERE id IN`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).
			AddRow(1, "alice", "Alice", "", 1).
			AddRow(2, "bob", "Bob", "", 2))

	body := []int64{1, 2}
	rec := doRequest(t, mux, "POST", "/api/v1/user/batch", body, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"username"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBatch_Empty(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	body := []int64{}
	rec := doRequest(t, mux, "POST", "/api/v1/user/batch", body, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandleLoginLogs_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT id, user_id, ip, user_agent, status, login_time FROM login_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "ip", "user_agent", "status", "login_time"}).
			AddRow(1, 99, "127.0.0.1", "Mozilla/5.0", 1, "2025-01-01T00:00:00Z").
			AddRow(2, 99, "192.168.1.1", "Chrome", 1, "2025-01-02T00:00:00Z"))
	mock.ExpectQuery(`SELECT COUNT.*FROM login_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	rec := doRequest(t, mux, "GET", "/api/v1/user/login-logs?page=1&size=10", nil, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"list"`)
	assert.Contains(t, rec.Body.String(), `"total":2`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLoginLogCount_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT COUNT.*FROM login_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	rec := doRequest(t, mux, "GET", "/api/v1/user/login-logs/count?status=1", nil, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"total":5`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandlePrivacy_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT public_collection.*FROM user_privacy_settings`).
		WillReturnRows(sqlmock.NewRows([]string{
			"public_collection", "public_birthday_tags", "public_coin_videos",
			"public_like_videos", "public_following_list", "public_followers_list",
		}).AddRow(1, 0, 0, 0, 0, 0))

	rec := doRequest(t, mux, "GET", "/api/v1/user/privacy/test", nil, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"public_collection"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleTags_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT tag_name FROM user_tags`).
		WillReturnRows(sqlmock.NewRows([]string{"tag_name"}).AddRow("golang").AddRow("tech"))

	rec := doRequest(t, mux, "GET", "/api/v1/user/tags", nil, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"golang"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleTags_POST_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectExec(`INSERT INTO user_tags`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doRequest(t, mux, "POST", "/api/v1/user/tags?tagName=golang", nil, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleTags_DELETE_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectExec(`DELETE FROM user_tags`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doRequest(t, mux, "DELETE", "/api/v1/user/tags?tagName=golang", nil, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessageSettings_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT private_message_notification.*FROM message_settings`).
		WillReturnRows(sqlmock.NewRows([]string{
			"private_message_notification", "reply_notification", "at_notification",
			"like_notification", "system_notification",
		}).AddRow(1, 1, 1, 1, 1))

	rec := doRequest(t, mux, "GET", "/api/v1/user/settings/message", nil, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"private_message_notification"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessageSettings_PUT_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectExec(`INSERT INTO message_settings`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE message_settings SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := map[string]interface{}{"like_notification": 0}
	rec := doRequest(t, mux, "PUT", "/api/v1/user/settings/message", body, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCreatorSettings_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT COALESCE\(default_category_id.*FROM creator_settings`).
		WillReturnRows(sqlmock.NewRows([]string{
			"default_category_id", "auto_publish", "comment_notify",
			"like_notify", "follow_notify",
		}).AddRow(1, 0, 1, 1, 1))

	rec := doRequest(t, mux, "GET", "/api/v1/user/settings/creator", nil, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"default_category_id"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCreatorSettings_PUT_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectExec(`INSERT INTO creator_settings`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE creator_settings SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := map[string]interface{}{"auto_publish": 1}
	rec := doRequest(t, mux, "PUT", "/api/v1/user/settings/creator", body, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCaptcha_New_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectExec(`INSERT INTO verification_codes`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doRequest(t, mux, "GET", "/api/v1/captcha/new", nil, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"captchaId"`)
	assert.Contains(t, rec.Body.String(), `"question"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCaptcha_Verify_Invalid(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	body := map[string]string{"captchaId": "", "answer": ""}
	rec := doRequest(t, mux, "POST", "/api/v1/captcha/verify", body, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `false`)
}

func TestHandleCaptcha_Verify_NotFound(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT id FROM verification_codes`).
		WillReturnError(sql.ErrNoRows)

	body := map[string]string{"captchaId": "123", "answer": "42"}
	rec := doRequest(t, mux, "POST", "/api/v1/captcha/verify", body, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `false`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleForgotPassword_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectExec(`UPDATE users SET password`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := map[string]string{"email": "a@b.com", "code": "123456", "newPassword": "newpwd"}
	rec := doRequest(t, mux, "POST", "/api/v1/user/password/forgot", body, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleEmailCode_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectExec(`INSERT INTO verification_codes`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := map[string]string{"email": "test@example.com"}
	rec := doRequest(t, mux, "POST", "/api/v1/user/email/code", body, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"code_sent"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleEmailVerify_Invalid(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	body := map[string]string{"email": "", "code": ""}
	rec := doRequest(t, mux, "POST", "/api/v1/user/email/verify", body, nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleEmailVerify_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT id FROM verification_codes`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec(`UPDATE verification_codes SET used = 1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := map[string]string{"email": "test@example.com", "code": "123456"}
	rec := doRequest(t, mux, "POST", "/api/v1/user/email/verify", body, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"verified"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandlePinnedVideo_GET_Empty(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT COALESCE\(pinned_video_id`).
		WillReturnRows(sqlmock.NewRows([]string{"pinned_video_id"}).AddRow(0))

	rec := doRequest(t, mux, "GET", "/api/v1/user/pinned-video", nil, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandlePinnedVideo_POST_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectExec(`UPDATE users SET pinned_video_id`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := map[string]interface{}{"videoId": 42}
	rec := doRequest(t, mux, "POST", "/api/v1/user/pinned-video", body, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandlePinnedVideo_DELETE_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectExec(`UPDATE users SET pinned_video_id = NULL`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doRequest(t, mux, "DELETE", "/api/v1/user/pinned-video", nil, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandlePrivacy_PUT_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectExec(`INSERT INTO user_privacy_settings`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE user_privacy_settings SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := map[string]interface{}{"public_collection": 0}
	rec := doRequest(t, mux, "PUT", "/api/v1/user/privacy/test", body, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUserByID_InvalidID(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := doRequest(t, mux, "GET", "/api/v1/user/abc", nil, nil)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "not found")
}

func TestHandleUserByID_PutInvalidBody(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest("PUT", "/api/v1/user/1", strings.NewReader("not json"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleUserByID_PutNoFields(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	body := map[string]interface{}{"unknown_field": "value"}
	rec := doRequest(t, mux, "PUT", "/api/v1/user/1", body, nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "no updatable fields")
}

func TestHandleMe_405_Method(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := doRequest(t, mux, "DELETE", "/api/v1/user/me", nil, map[string]string{
		"X-User-Id": "99",
	})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

// 防 unused import 警告
var _ = context.Background
var _ = (*pb.LoginResponse)(nil)
var _ = strings.HasPrefix
