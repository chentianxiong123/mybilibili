package user

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "mybilibili/pkg/pb"
)

// failingIncrCache 让 Incr 一定失败，覆盖 recordLoginFailure 的 Incr 错误分支。
type failingIncrCache struct {
	*fakeCache
	err error
}

func (f *failingIncrCache) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	return 0, f.err
}

// ---- admin_handler.go: withPerm 错误路径 ----

func TestUserAdminHandler_WithPerm_NoPermChecker_Subpath(t *testing.T) {
	// 子路径（非 /list）在 perm==nil 时必须 403。
	h, _ := newUserAdminHandler(t, nil, nil)
	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/1", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "权限校验未配置")
}

func TestUserAdminHandler_WithPerm_Denied_NoAdminHeader(t *testing.T) {
	// 子路径：无权限且无 X-Admin-Id → 401。
	h, _ := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{}}, nil)
	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/1", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestUserAdminHandler_WithPerm_Denied_WithAdminHeader(t *testing.T) {
	// 子路径：无权限但有 X-Admin-Id → 403。
	h, _ := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{}}, nil)
	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/1", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "forbidden")
}

func TestUserAdminHandler_WithPerm_Denied_PasswordSubpath(t *testing.T) {
	// 覆盖 /{id}/password 的 withPerm 拒绝分支。
	h, _ := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{}}, nil)
	w := doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1/password", map[string]interface{}{"newPassword": "x"}, map[string]string{"X-Admin-Id": "9"})
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---- admin_handler.go: handleRoute 的 "list" 兜底与末尾 404 ----

func TestUserAdminHandler_HandleRoute_ListFallback(t *testing.T) {
	// handleRoute 里 path=="list" 的防御分支（精确路由 /list 已单独注册，这里直接调函数覆盖）。
	h, mock := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{"user:manage": true}, adminID: 1}, nil)
	now := time.Now()
	mock.ExpectQuery(`FROM users u ORDER BY id DESC`).
		WithArgs(int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "email", "avatar", "level", "status", "created_at", "phone", "follower_count", "following_count", "manuscript_count"}).
			AddRow(1, "u", "nick", "e", "a", 1, 1, now.Format("2006-01-02T15:04:05Z"), "138", 1, 2, 3))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	req := httptest.NewRequest("GET", "/api/v1/user/admin/list", nil)
	rec := httptest.NewRecorder()
	h.handleRoute(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserAdminHandler_HandleRoute_TooManyParts(t *testing.T) {
	// /{id}/status/extra → 3 段，落到末尾 404。
	h, _ := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{"user:manage": true}}, nil)
	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/1/status/extra", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "not found")
}

// ---- admin_handler.go: handleList 扫描失败 continue ----

func TestUserAdminHandler_List_ScanErrorContinue(t *testing.T) {
	h, mock := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{"user:manage": true}, adminID: 1}, nil)
	// 返回少一列，Scan 失败 → continue，列表为空。
	mock.ExpectQuery(`FROM users u ORDER BY id DESC`).
		WithArgs(int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname"}).AddRow(1, "u", "n"))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/list", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"list":[]`)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- service.go: Login 锁定与 recordLoginFailure Incr 错误 ----

func TestLogin_AccountLocked(t *testing.T) {
	svc, mock, cache := newTestService(t)

	// 预置锁定 key。
	require.NoError(t, cache.Set(context.Background(), "login:fail:1:lock", []byte("1"), time.Minute))

	username := "alice"
	mock.ExpectQuery(`SELECT.*FROM users.*WHERE username`).
		WithArgs(username).
		WillReturnRows(userRow(1, username, sha256Hex("right"), "Alice", 1))

	_, err := svc.Login(context.Background(), &pb.LoginRequest{Username: username, Password: "right"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "locked")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordLoginFailure_IncrError(t *testing.T) {
	svc, mock, _ := newTestService(t)
	svc.SetCacheStore(&failingIncrCache{newFakeCache(), errors.New("redis boom")})

	username := "alice"
	mock.ExpectQuery(`SELECT.*FROM users.*WHERE username`).
		WithArgs(username).
		WillReturnRows(userRow(1, username, sha256Hex("right"), "Alice", 1))

	// 密码错误 → recordLoginFailure → Incr 报错（被忽略，仍返回未认证）。
	_, err := svc.Login(context.Background(), &pb.LoginRequest{Username: username, Password: "wrong"})
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

var _ = sql.ErrNoRows
