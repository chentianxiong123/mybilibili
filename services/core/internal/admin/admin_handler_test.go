package admin

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/pkg/auth"
)

const testJWTSecret = "test-secret-key-for-admin"

func newTestHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewRepository(db)
	svc := NewService(repo)
	jwt := auth.NewJWT(testJWTSecret)
	return NewHandler(svc, jwt), mock
}

func doReq(t *testing.T, h http.Handler, method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
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

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

func permRows(codes ...string) *sqlmock.Rows {
	r := sqlmock.NewRows([]string{"code"})
	for _, c := range codes {
		r.AddRow(c)
	}
	return r
}

func expectPermQuery(mock sqlmock.Sqlmock, adminID int64, codes []string) {
	mock.ExpectQuery(`SELECT DISTINCT p.code`).WithArgs(adminID).WillReturnRows(permRows(codes...))
}

func expectRolesQuery(mock sqlmock.Sqlmock, adminID int64) {
	mock.ExpectQuery(`SELECT r.id, r.name`).WithArgs(adminID).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "description", "create_time"}))
}

// ==================== handleLogin ====================

func TestHandleLogin_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	pwd := "admin123"
	mock.ExpectQuery(`SELECT id, username, password`).WithArgs("admin", sha256Hex(pwd)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "admin_level"}).
			AddRow(1, "admin", sha256Hex(pwd), 1))
	mock.ExpectQuery(`SELECT r.id, r.name`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "create_time"}).
			AddRow(1, "超级管理员", "super", ""))
	mock.ExpectExec(`INSERT INTO login_logs`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT DISTINCT p.code`).WithArgs(int64(1)).
		WillReturnRows(permRows("admin:manage"))

	rec := doReq(t, mux, "POST", "/api/v1/admin/login",
		map[string]string{"username": "admin", "password": pwd}, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"超级管理员"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLogin_401_WrongPassword(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT id, username, password`).WithArgs("admin", sha256Hex("wrong")).
		WillReturnError(sql.ErrNoRows)

	rec := doReq(t, mux, "POST", "/api/v1/admin/login",
		map[string]string{"username": "admin", "password": "wrong"}, nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLogin_401_UserNotExist(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT id, username, password`).WithArgs("ghost", sha256Hex("x")).
		WillReturnError(sql.ErrNoRows)

	rec := doReq(t, mux, "POST", "/api/v1/admin/login",
		map[string]string{"username": "ghost", "password": "x"}, nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLogin_405_Method(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := doReq(t, mux, "GET", "/api/v1/admin/login", nil, nil)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestHandleLogin_WithForwardedFor(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	pwd := "admin123"
	mock.ExpectQuery(`SELECT id, username, password`).WithArgs("admin", sha256Hex(pwd)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "admin_level"}).
			AddRow(1, "admin", sha256Hex(pwd), 1))
	mock.ExpectQuery(`SELECT r.id, r.name`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "create_time"}))
	mock.ExpectExec(`INSERT INTO login_logs`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT DISTINCT p.code`).WithArgs(int64(1)).
		WillReturnRows(permRows())

	body := map[string]string{"username": "admin", "password": pwd}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest("POST", "/api/v1/admin/login", &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLogin_NoRoles(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	pwd := "admin123"
	mock.ExpectQuery(`SELECT id, username, password`).WithArgs("admin", sha256Hex(pwd)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "admin_level"}).
			AddRow(1, "admin", sha256Hex(pwd), 1))
	mock.ExpectQuery(`SELECT r.id, r.name`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "create_time"}))
	mock.ExpectExec(`INSERT INTO login_logs`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT DISTINCT p.code`).WithArgs(int64(1)).
		WillReturnRows(permRows("user:manage"))

	rec := doReq(t, mux, "POST", "/api/v1/admin/login",
		map[string]string{"username": "admin", "password": pwd}, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"role":"管理员"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLogin_NullPermissions(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	pwd := "admin123"
	mock.ExpectQuery(`SELECT id, username, password`).WithArgs("admin", sha256Hex(pwd)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "admin_level"}).
			AddRow(1, "admin", sha256Hex(pwd), 1))
	mock.ExpectQuery(`SELECT r.id, r.name`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "create_time"}))
	mock.ExpectExec(`INSERT INTO login_logs`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT DISTINCT p.code`).WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("db error"))

	rec := doReq(t, mux, "POST", "/api/v1/admin/login",
		map[string]string{"username": "admin", "password": pwd}, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"permissions":[]`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleRegister ====================

func TestHandleRegister_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"admin:manage"})
	mock.ExpectQuery(`INSERT INTO admin_users`).WithArgs("newadmin", sha256Hex("pass123"), int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectExec(`INSERT INTO audit_logs`).WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doReq(t, mux, "POST", "/api/v1/admin/register",
		map[string]interface{}{"username": "newadmin", "password": "pass123", "level": 1},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRegister_405_Method(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"admin:manage"})

	rec := doReq(t, mux, "GET", "/api/v1/admin/register", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRegister_400_Duplicate(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"admin:manage"})
	mock.ExpectQuery(`INSERT INTO admin_users`).
		WillReturnError(fmt.Errorf("duplicate key"))

	rec := doReq(t, mux, "POST", "/api/v1/admin/register",
		map[string]interface{}{"username": "dup", "password": "p", "level": 1},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRegister_403_Forbidden(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{})

	rec := doReq(t, mux, "POST", "/api/v1/admin/register",
		map[string]interface{}{"username": "x", "password": "y", "level": 1},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleListAdmins ====================
// handleListAdmins has internal requirePermission + requirePerm wrapper = 2 checks

func TestHandleListAdmins_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"admin:manage"})
	expectPermQuery(mock, 1, []string{"admin:manage"})
	mock.ExpectQuery(`SELECT id, username, password, COALESCE`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "username", "password", "admin_level", "created_at"}).
			AddRow(1, "admin1", "h", 1, "2025-01-01"))
	mock.ExpectQuery(`SELECT r.id, r.name`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "create_time"}))

	rec := doReq(t, mux, "GET", "/api/v1/admin/list", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"admin1"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleListAdmins_403_Forbidden(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{})

	rec := doReq(t, mux, "GET", "/api/v1/admin/list", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleRoles ====================
// handleRoles has internal requirePermission + requirePerm wrapper = 2 checks

func TestHandleRoles_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"role:manage"})
	expectPermQuery(mock, 1, []string{"role:manage"})
	mock.ExpectQuery(`SELECT id, name`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "description", "create_time"}).
			AddRow(1, "admin", "管理员", "2025-01-01"))

	rec := doReq(t, mux, "GET", "/api/v1/admin/roles", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"admin"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRoles_POST_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"role:manage"})
	expectPermQuery(mock, 1, []string{"role:manage"})
	mock.ExpectQuery(`INSERT INTO roles`).WithArgs("editor", "编辑").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))

	rec := doReq(t, mux, "POST", "/api/v1/admin/roles",
		map[string]string{"name": "editor", "description": "编辑"},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRoles_403_Forbidden(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{})

	rec := doReq(t, mux, "GET", "/api/v1/admin/roles", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleRolesByID ====================

func TestHandleRolesByID_PUT_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"role:manage"})
	mock.ExpectExec(`UPDATE roles SET name`).WithArgs("editor2", "编辑2", int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO audit_logs`).WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doReq(t, mux, "PUT", "/api/v1/admin/roles/5",
		map[string]string{"name": "editor2", "description": "编辑2"},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRolesByID_DELETE_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"role:manage"})
	mock.ExpectExec(`DELETE FROM roles WHERE id`).WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO audit_logs`).WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doReq(t, mux, "DELETE", "/api/v1/admin/roles/5", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRolesByID_Permissions_GET(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"role:manage"})
	mock.ExpectQuery(`SELECT permission_id FROM role_permissions WHERE role_id`).WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"permission_id"}).AddRow(1).AddRow(2))

	rec := doReq(t, mux, "GET", "/api/v1/admin/roles/5/permissions", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"data"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRolesByID_Permissions_PUT(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"role:manage"})
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id`).WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO role_permissions`).WithArgs(int64(5), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO role_permissions`).WithArgs(int64(5), int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectExec(`INSERT INTO audit_logs`).WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doReq(t, mux, "PUT", "/api/v1/admin/roles/5/permissions",
		map[string]interface{}{"permission_ids": []int64{1, 2}},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRolesByID_Templates_GET(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"role:manage"})

	rec := doReq(t, mux, "GET", "/api/v1/admin/roles/templates", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"platform-operation"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRolesByID_Template_PUT(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"role:manage"})
	mock.ExpectQuery(`SELECT id, code FROM permissions WHERE code = ANY`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).
			AddRow(1, "video:manage").AddRow(2, "category:manage").
			AddRow(3, "banner:manage").AddRow(4, "live:manage").
			AddRow(5, "subtitle:manage").AddRow(6, "search:manage").
			AddRow(7, "operation:manage").AddRow(8, "scheduled:manage"))
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id`).WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	for i := int64(1); i <= 8; i++ {
		mock.ExpectExec(`INSERT INTO role_permissions`).WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()
	mock.ExpectExec(`INSERT INTO audit_logs`).WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doReq(t, mux, "PUT", "/api/v1/admin/roles/5/template/platform-operation", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRolesByID_Template_PUT_NotFound(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"role:manage"})

	rec := doReq(t, mux, "PUT", "/api/v1/admin/roles/5/template/nonexistent", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"岗位模板不存在"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handlePermissions ====================

func TestHandlePermissions_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"role:manage"})
	mock.ExpectQuery(`SELECT id, name, code`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "code", "url", "method", "parent_id", "description"}).
			AddRow(1, "用户管理", "user:manage", "", "", 0, "desc"))

	rec := doReq(t, mux, "GET", "/api/v1/admin/permissions", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"user:manage"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleAuditLogs ====================

func TestHandleAuditLogs_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"audit:manage"})
	mock.ExpectQuery(`SELECT id, operator_id, operator_name`).WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "operator_id", "operator_name", "operator_role", "module", "action",
			"target_type", "target_id", "request_method", "request_uri", "client_ip",
			"user_agent", "result", "message", "detail", "created_at",
		}).AddRow(1, 1, "admin", "", "admin", "CREATE", "admin_users", "1",
			"POST", "/api", "127.0.0.1", "UA", 1, "ok", "", "2025-01-01"))
	mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rec := doReq(t, mux, "GET", "/api/v1/admin/audit-logs", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"list"`)
	assert.Contains(t, rec.Body.String(), `"total"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAuditLogs_EmptyList(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"audit:manage"})
	mock.ExpectQuery(`SELECT id, operator_id, operator_name`).WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "operator_id", "operator_name", "operator_role", "module", "action",
			"target_type", "target_id", "request_method", "request_uri", "client_ip",
			"user_agent", "result", "message", "detail", "created_at",
		}))
	mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	rec := doReq(t, mux, "GET", "/api/v1/admin/audit-logs", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"list"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAuditLogByID_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"audit:manage"})
	mock.ExpectQuery(`SELECT id, operator_id, operator_name`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "operator_id", "operator_name", "operator_role", "module", "action",
			"target_type", "target_id", "request_method", "request_uri", "client_ip",
			"user_agent", "result", "message", "detail", "created_at",
		}).AddRow(1, 1, "admin", "", "admin", "CREATE", "", "", "", "", "", "", 1, "ok", "", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)))

	rec := doReq(t, mux, "GET", "/api/v1/admin/audit-logs/1", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"id":1`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAuditLogByID_404(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"audit:manage"})
	mock.ExpectQuery(`SELECT id, operator_id, operator_name`).WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	rec := doReq(t, mux, "GET", "/api/v1/admin/audit-logs/999", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "audit log not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAuditLogByID_405_Method(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"audit:manage"})

	rec := doReq(t, mux, "DELETE", "/api/v1/admin/audit-logs/1", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleLoginLogs ====================

func TestHandleLoginLogs_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"security:manage"})
	mock.ExpectQuery(`SELECT id, user_id, ip, user_agent, status, login_time FROM login_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "ip", "user_agent", "status", "login_time"}).
			AddRow(1, 10, "127.0.0.1", "Mozilla/5.0", 1, "2025-01-01"))
	mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rec := doReq(t, mux, "GET", "/api/v1/admin/login-logs", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"list"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUserLoginLogs_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"security:manage"})
	mock.ExpectQuery(`SELECT id, user_id, ip, user_agent, status, login_time FROM login_logs WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "ip", "user_agent", "status", "login_time"}).
			AddRow(1, 10, "127.0.0.1", "Mozilla/5.0", 1, "2025-01-01"))

	rec := doReq(t, mux, "GET", "/api/v1/admin/login-logs/user/10", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"list"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUserLoginLogs_405_Method(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"security:manage"})

	rec := doReq(t, mux, "POST", "/api/v1/admin/login-logs/user/10", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleSecuritySettings ====================

func TestHandleSecuritySettings_GET_Default(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"security:manage"})
	mock.ExpectQuery(`SELECT config_value FROM system_configs WHERE config_key='security_settings'`).
		WillReturnError(sql.ErrNoRows)

	rec := doReq(t, mux, "GET", "/api/v1/admin/security-settings", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"password_policy"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSecuritySettings_GET_FromDB(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"security:manage"})
	settings := map[string]any{"password_policy": map[string]any{"min_length": 10}}
	raw, _ := json.Marshal(settings)
	mock.ExpectQuery(`SELECT config_value FROM system_configs WHERE config_key='security_settings'`).
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow(string(raw)))

	rec := doReq(t, mux, "GET", "/api/v1/admin/security-settings", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"password_policy"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSecuritySettings_GET_InvalidJSON(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"security:manage"})
	mock.ExpectQuery(`SELECT config_value FROM system_configs WHERE config_key='security_settings'`).
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow("invalid-json"))

	rec := doReq(t, mux, "GET", "/api/v1/admin/security-settings", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"password_policy"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSecuritySettings_PUT_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"security:manage"})
	mock.ExpectExec(`INSERT INTO system_configs`).WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doReq(t, mux, "PUT", "/api/v1/admin/security-settings",
		map[string]any{"password_policy": map[string]any{"min_length": 12}},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSecuritySettings_PUT_DBError(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"security:manage"})
	mock.ExpectExec(`INSERT INTO system_configs`).WillReturnError(fmt.Errorf("db error"))

	rec := doReq(t, mux, "PUT", "/api/v1/admin/security-settings",
		map[string]any{"key": "val"}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSecuritySettings_405_Method(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"security:manage"})

	rec := doReq(t, mux, "DELETE", "/api/v1/admin/security-settings", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleStorageMigrate ====================

func TestHandleStorageMigrate_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"storage:manage"})
	mock.ExpectExec(`UPDATE videos SET source_video_url = replace`).WillReturnResult(sqlmock.NewResult(0, 5))

	rec := doReq(t, mux, "POST", "/api/v1/admin/storage/migrate",
		map[string]string{"from": "oss", "to": "s3"},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.Contains(t, rec.Body.String(), `"from":"oss"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleStorageMigrate_DefaultTo(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"storage:manage"})
	mock.ExpectExec(`UPDATE videos SET source_video_url = replace`).WillReturnResult(sqlmock.NewResult(0, 0))

	rec := doReq(t, mux, "POST", "/api/v1/admin/storage/migrate",
		map[string]string{"from": "oss"},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"to":"s3"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleStorageMigrate_405_Method(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"storage:manage"})

	rec := doReq(t, mux, "GET", "/api/v1/admin/storage/migrate", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleStorageMigrate_DBError(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"storage:manage"})
	mock.ExpectExec(`UPDATE videos SET source_video_url = replace`).WillReturnError(fmt.Errorf("db error"))

	rec := doReq(t, mux, "POST", "/api/v1/admin/storage/migrate",
		map[string]string{"from": "oss", "to": "s3"},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleAdminByID ====================
// handleAdminByID has internal requirePermission (no wrapper requirePerm)

func TestHandleAdminByID_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"admin:manage"})
	mock.ExpectQuery(`SELECT id, username, COALESCE`).WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "admin_level"}).
			AddRow(5, "admin5", 1))
	mock.ExpectQuery(`SELECT r.id, r.name`).WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "create_time"}))

	rec := doReq(t, mux, "GET", "/api/v1/admin/5", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"admin5"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAdminByID_GET_404(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"admin:manage"})
	mock.ExpectQuery(`SELECT id, username, COALESCE`).WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	rec := doReq(t, mux, "GET", "/api/v1/admin/999", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "admin not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAdminByID_PUT_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"admin:manage"})
	mock.ExpectExec(`UPDATE admin_users SET updated_at`).WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doReq(t, mux, "PUT", "/api/v1/admin/5",
		map[string]string{"nickname": "new"},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAdminByID_InvalidID(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"admin:manage"})

	rec := doReq(t, mux, "GET", "/api/v1/admin/abc", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid admin id")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAdminByID_403_Forbidden(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{})

	rec := doReq(t, mux, "GET", "/api/v1/admin/5", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAdminByID_Roles_GET(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"admin:manage"})
	mock.ExpectQuery(`SELECT role_id FROM admin_user_roles WHERE admin_user_id`).WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"role_id"}).AddRow(1).AddRow(2))

	rec := doReq(t, mux, "GET", "/api/v1/admin/5/roles", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"data"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAdminByID_Roles_PUT(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"admin:manage"})
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM admin_user_roles WHERE admin_user_id`).WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO admin_user_roles`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectExec(`INSERT INTO audit_logs`).WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doReq(t, mux, "PUT", "/api/v1/admin/5/roles",
		map[string]interface{}{"role_ids": []int64{1}},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAdminByID_UnknownSubPath(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"admin:manage"})

	rec := doReq(t, mux, "GET", "/api/v1/admin/5/unknown", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAdminByID_EmptyPath(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"admin:manage"})

	rec := doReq(t, mux, "GET", "/api/v1/admin/", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAdminByID_405_Method(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"admin:manage"})

	rec := doReq(t, mux, "DELETE", "/api/v1/admin/5", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAdminByID_Roles_405_Method(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"admin:manage"})

	rec := doReq(t, mux, "DELETE", "/api/v1/admin/5/roles", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== requirePermission / CheckPermission ====================

func TestCheckPermission_OK(t *testing.T) {
	h, mock := newTestHandler(t)
	expectPermQuery(mock, 1, []string{"admin:manage"})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Admin-Id", "1")
	id, ok := h.CheckPermission(req, "admin:manage")
	assert.Equal(t, int64(1), id)
	assert.True(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckPermission_Denied(t *testing.T) {
	h, mock := newTestHandler(t)
	expectPermQuery(mock, 1, []string{})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Admin-Id", "1")
	id, ok := h.CheckPermission(req, "admin:manage")
	assert.Equal(t, int64(1), id)
	assert.False(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckPermission_NoAdminID(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest("GET", "/test", nil)
	id, ok := h.CheckPermission(req, "admin:manage")
	assert.Equal(t, int64(0), id)
	assert.False(t, ok)
}

func TestCheckPermission_JWTFallback(t *testing.T) {
	h, mock := newTestHandler(t)
	jwtTool := auth.NewJWT(testJWTSecret)
	token, err := jwtTool.GenerateAdmin(42)
	require.NoError(t, err)

	expectPermQuery(mock, 42, []string{"admin:manage"})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	id, ok := h.CheckPermission(req, "admin:manage")
	assert.Equal(t, int64(42), id)
	assert.True(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckPermission_JWTInvalid(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	id, ok := h.CheckPermission(req, "admin:manage")
	assert.Equal(t, int64(0), id)
	assert.False(t, ok)
}

// 回归：普通用户 token 绝不能被当成管理员身份。
// users.id 与 admin_users.id 是两套独立自增 ID（实测 users.id=4 的 string
// 撞上 admin_users.id=4 的 system_admin 即提权），因此 fallback 必须校验 IsAdmin。
func TestCheckPermission_RegularUserTokenRejected(t *testing.T) {
	h, _ := newTestHandler(t)
	jwtTool := auth.NewJWT(testJWTSecret)
	token, err := jwtTool.Generate(4) // 普通用户 id=4
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	id, ok := h.CheckPermission(req, "admin:manage")
	assert.Equal(t, int64(0), id, "不得把普通用户 id 当 admin_id 查询")
	assert.False(t, ok)
}

func TestCheckPermission_DBError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT DISTINCT p.code`).WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("db error"))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Admin-Id", "1")
	id, ok := h.CheckPermission(req, "admin:manage")
	assert.Equal(t, int64(1), id)
	assert.False(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== requirePerm middleware ====================

func TestRequirePerm_Forbidden(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{})

	rec := doReq(t, mux, "GET", "/api/v1/admin/list", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":403`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleOperationTasks ====================

func TestHandleOperationTasks_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"operation:manage"})
	mock.ExpectQuery(`SELECT id, task_key, task_type, task_name, target_type, COALESCE`).WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "task_key", "task_type", "task_name", "target_type", "target_id",
			"status", "progress", "error_message", "created_at", "updated_at",
		}).AddRow(1, "key1", "type1", "name1", "video", 1, "done", 100, "", "2025-01-01", "2025-01-01"))
	mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rec := doReq(t, mux, "GET", "/api/v1/admin/operation-tasks", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"list"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleOperationTasks_405_Method(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"operation:manage"})

	rec := doReq(t, mux, "POST", "/api/v1/admin/operation-tasks", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleOperationTaskByID_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"operation:manage"})
	mock.ExpectQuery(`SELECT id, task_key, task_type, task_name, target_type, COALESCE`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "task_key", "task_type", "task_name", "target_type", "target_id",
			"status", "progress", "error_message", "created_at", "updated_at",
		}).AddRow(1, "key1", "type1", "name1", "video", 1, "done", 100, "", "2025-01-01", "2025-01-01"))

	rec := doReq(t, mux, "GET", "/api/v1/admin/operation-tasks/1", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"id":1`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleOperationTaskByID_404(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"operation:manage"})
	mock.ExpectQuery(`SELECT id, task_key, task_type, task_name, target_type, COALESCE`).WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	rec := doReq(t, mux, "GET", "/api/v1/admin/operation-tasks/999", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleOperationTaskByID_405_Method(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"operation:manage"})

	rec := doReq(t, mux, "DELETE", "/api/v1/admin/operation-tasks/1", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleTranscodeConfig ====================

func TestHandleTranscodeConfig_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"transcode:manage"})
	mock.ExpectQuery(`SELECT config_value FROM system_configs WHERE config_key='transcode_encoder'`).
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow("auto"))

	rec := doReq(t, mux, "GET", "/api/v1/admin/transcode-config", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"encoder"`)
	assert.Contains(t, rec.Body.String(), `"auto"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleTranscodeConfig_GET_NoConfig(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"transcode:manage"})
	mock.ExpectQuery(`SELECT config_value FROM system_configs WHERE config_key='transcode_encoder'`).
		WillReturnError(sql.ErrNoRows)

	rec := doReq(t, mux, "GET", "/api/v1/admin/transcode-config", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"encoder"`)
	assert.Contains(t, rec.Body.String(), `"auto"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleTranscodeConfig_PUT_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"transcode:manage"})
	mock.ExpectExec(`INSERT INTO system_configs`).WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doReq(t, mux, "PUT", "/api/v1/admin/transcode-config",
		map[string]string{"encoder": "vaapi"},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleTranscodeConfig_PUT_InvalidEncoder(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"transcode:manage"})

	rec := doReq(t, mux, "PUT", "/api/v1/admin/transcode-config",
		map[string]string{"encoder": "nvenc"},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid encoder")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleTranscodeConfig_PUT_DBError(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"transcode:manage"})
	mock.ExpectExec(`INSERT INTO system_configs`).WillReturnError(fmt.Errorf("db error"))

	rec := doReq(t, mux, "PUT", "/api/v1/admin/transcode-config",
		map[string]string{"encoder": "x264"},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleTranscodeConfig_405_Method(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"transcode:manage"})

	rec := doReq(t, mux, "DELETE", "/api/v1/admin/transcode-config", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== SetScheduler ====================

func TestSetScheduler(t *testing.T) {
	h, _ := newTestHandler(t)
	assert.Nil(t, h.scheduler)
	h.SetScheduler(&Scheduler{})
	assert.NotNil(t, h.scheduler)
}

// ==================== handleScheduledTasks ====================

func TestHandleScheduledTasks_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})
	mock.ExpectQuery(`SELECT id, task_key, task_name, description, cron_expr, task_type, task_config`).WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "task_key", "task_name", "description", "cron_expr", "task_type", "task_config",
			"enabled", "last_run_at", "last_run_result", "last_run_message",
			"next_run_at", "run_count", "max_retries", "retry_count", "timeout_seconds",
			"created_at", "updated_at",
		}).AddRow(1, "key1", "task1", "desc", "0 * * * *", "cron", "{}", 1, nil, "", "", nil, 0, 3, 0, 60, "2025-01-01", "2025-01-01"))

	rec := doReq(t, mux, "GET", "/api/v1/admin/scheduled-tasks", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"list"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleScheduledTasks_POST_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})
	mock.ExpectQuery(`INSERT INTO scheduled_tasks`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	rec := doReq(t, mux, "POST", "/api/v1/admin/scheduled-tasks",
		map[string]interface{}{
			"task_key": "test_task", "task_name": "Test", "cron_expr": "0 * * * *",
			"task_type": "cron", "task_config": "{}", "enabled": 1, "max_retries": 3, "timeout_seconds": 60,
		}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"id":1`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleScheduledTasks_PUT_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})
	mock.ExpectExec(`UPDATE scheduled_tasks SET task_name`).WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doReq(t, mux, "PUT", "/api/v1/admin/scheduled-tasks",
		map[string]interface{}{
			"id": 1, "task_name": "Updated", "description": "desc", "cron_expr": "0 * * * *",
			"task_type": "cron", "task_config": "{}", "enabled": 1, "max_retries": 3, "timeout_seconds": 60,
		}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleScheduledTasks_DELETE_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})
	mock.ExpectExec(`DELETE FROM scheduled_tasks WHERE id`).WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doReq(t, mux, "DELETE", "/api/v1/admin/scheduled-tasks",
		map[string]interface{}{"id": 1}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleScheduledTasks_DELETE_BadID(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})

	rec := doReq(t, mux, "DELETE", "/api/v1/admin/scheduled-tasks",
		map[string]interface{}{"id": 0}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "id required")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleScheduledTasks_405_Method(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})

	rec := doReq(t, mux, "PATCH", "/api/v1/admin/scheduled-tasks", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleScheduledTaskToggle ====================

func TestHandleScheduledTaskToggle_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})
	mock.ExpectExec(`UPDATE scheduled_tasks SET enabled`).WillReturnResult(sqlmock.NewResult(0, 1))

	rec := doReq(t, mux, "POST", "/api/v1/admin/scheduled-tasks/toggle",
		map[string]interface{}{"id": 1, "enabled": 0},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleScheduledTaskToggle_405_Method(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})

	rec := doReq(t, mux, "GET", "/api/v1/admin/scheduled-tasks/toggle", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleScheduledTaskTrigger ====================

func TestHandleScheduledTaskTrigger_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})

	rec := doReq(t, mux, "POST", "/api/v1/admin/scheduled-tasks/trigger",
		map[string]string{"task_key": "backup"},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleScheduledTaskTrigger_EmptyKey(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})

	rec := doReq(t, mux, "POST", "/api/v1/admin/scheduled-tasks/trigger",
		map[string]string{"task_key": ""},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "task_key required")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleScheduledTaskTrigger_405_Method(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})

	rec := doReq(t, mux, "GET", "/api/v1/admin/scheduled-tasks/trigger", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleScheduledTaskTrigger_WithScheduler(t *testing.T) {
	h, mock := newTestHandler(t)
	scheduler := NewScheduler(h.svc)
	h.SetScheduler(scheduler)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})

	rec := doReq(t, mux, "POST", "/api/v1/admin/scheduled-tasks/trigger",
		map[string]string{"task_key": "nonexistent"},
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "task not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Helpers ====================

func TestDefaultSecuritySettings(t *testing.T) {
	s := defaultSecuritySettings()
	assert.NotNil(t, s["password_policy"])
	assert.NotNil(t, s["login_policy"])
	pp := s["password_policy"].(map[string]any)
	assert.Equal(t, 8, pp["min_length"])
	lp := s["login_policy"].(map[string]any)
	assert.Equal(t, 5, lp["max_attempts"])
}

func TestRoleTemplates(t *testing.T) {
	templates := roleTemplates()
	assert.Contains(t, templates, "super-admin")
	assert.Contains(t, templates, "platform-operation")
	assert.Contains(t, templates, "content-review")
	assert.Contains(t, templates, "customer-support")
	assert.Contains(t, templates, "data-analyst")
	assert.Contains(t, templates, "system-manager")
	sa := templates["super-admin"]
	assert.Greater(t, len(sa.PermissionCodes), 10)
}

func TestDetectVAAPI(t *testing.T) {
	h, _ := newTestHandler(t)
	assert.False(t, h.detectVAAPI())
}

func TestHandleScheduledTaskTrigger_NilScheduler(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})

	rec := doReq(t, mux, "POST", "/api/v1/admin/scheduled-tasks/trigger",
		map[string]string{"task_key": "test"}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
}
