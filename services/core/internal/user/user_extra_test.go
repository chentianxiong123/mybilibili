package user

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/pkg/auth"
	pb "mybilibili/pkg/pb"
)

func muxForUserExtend(h *UserExtendHandler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestHandler_Register_Login_GetUser(t *testing.T) {
	svc, mock, _ := newTestService(t)
	h := NewHandler(svc)
	ctx := context.Background()

	_, err := h.Register(ctx, &pb.RegisterRequest{})
	assert.Error(t, err)

	_, err = h.Login(ctx, &pb.LoginRequest{})
	assert.Error(t, err)

	mock.ExpectQuery(`FROM users WHERE id = \$1`).WithArgs(int64(99)).WillReturnError(sql.ErrNoRows)
	_, err = h.GetUser(ctx, &pb.GetUserRequest{UserId: 99})
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAwardExperience_SuccessNoLevelUp(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT level, experience FROM users`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"level", "experience"}).AddRow(1, 10))
	mock.ExpectExec(`UPDATE users SET experience = \$1, level = \$2, updated_at = NOW\(\) WHERE id = \$3`).
		WithArgs(int64(15), int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	leveled, level := AwardExperience(context.Background(), db, 1, 5)
	assert.False(t, leveled)
	assert.Equal(t, int32(1), level)
}

func TestAwardExperience_LevelUp(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	// LevelThreshold(1)=100，exp=95 + 10 = 105 -> 升级到 2，结转 5
	mock.ExpectQuery(`SELECT level, experience FROM users`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"level", "experience"}).AddRow(1, 95))
	mock.ExpectExec(`UPDATE users SET experience = \$1, level = \$2, updated_at = NOW\(\) WHERE id = \$3`).
		WithArgs(int64(5), int64(2), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	leveled, level := AwardExperience(context.Background(), db, 1, 10)
	assert.True(t, leveled)
	assert.Equal(t, int32(2), level)
}

func TestAwardExperience_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT level, experience FROM users`).WithArgs(int64(1)).WillReturnError(sql.ErrNoRows)
	leveled, level := AwardExperience(context.Background(), db, 1, 5)
	assert.False(t, leveled)
	assert.Equal(t, int32(0), level)
}

func TestAwardExperience_UpdateError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT level, experience FROM users`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"level", "experience"}).AddRow(1, 10))
	mock.ExpectExec(`UPDATE users SET experience`).WillReturnError(errors.New("boom"))
	leveled, level := AwardExperience(context.Background(), db, 1, 5)
	assert.False(t, leveled)
	assert.Equal(t, int32(1), level)
}

// ---- AdminHandler ----

type fakePerm struct {
	allowed map[string]bool
	adminID int64
}

func (f *fakePerm) CheckPermission(r *http.Request, permission string) (int64, bool) {
	if f.allowed[permission] {
		return f.adminID, true
	}
	return 0, false
}

type fakeAuditor struct {
	calls int
}

func (f *fakeAuditor) RecordAudit(ctx context.Context, operatorID int64, operatorName, module, action, targetType, targetID string, result int32, message, detail string) error {
	f.calls++
	return nil
}

func newUserAdminHandler(t *testing.T, perm PermChecker, auditor AuditRecorder) (*UserAdminHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewUserAdminHandler(db, auditor, perm, auth.NewJWT("secret")), mock
}

func muxForUserAdmin(h *UserAdminHandler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestUserAdminHandler_List_Unauthorized(t *testing.T) {
	h, _ := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{}}, nil)
	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/list", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/list", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUserAdminHandler_List_NoPermChecker(t *testing.T) {
	h, _ := newUserAdminHandler(t, nil, nil)
	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/list", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUserAdminHandler_List_Success(t *testing.T) {
	h, mock := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{"user:manage": true}, adminID: 1}, nil)
	now := time.Now()
	mock.ExpectQuery(`FROM users u ORDER BY id DESC`).
		WithArgs(int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "email", "avatar", "level", "status", "created_at", "phone", "follower_count", "following_count", "manuscript_count"}).
			AddRow(1, "u", "nick", "e", "a", 1, 1, now.Format("2006-01-02T15:04:05Z"), "138", 1, 2, 3))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/list", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserAdminHandler_List_QueryError(t *testing.T) {
	h, mock := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{"user:manage": true}}, nil)
	mock.ExpectQuery(`FROM users u ORDER BY id DESC`).WillReturnError(errors.New("boom"))
	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/list", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserAdminHandler_handleRoute_InvalidID(t *testing.T) {
	h, _ := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{"user:manage": true}}, nil)
	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/abc", nil, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserAdminHandler_handleGet(t *testing.T) {
	h, mock := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{"user:manage": true}, adminID: 1}, nil)
	now := time.Now().Format("2006-01-02T15:04:05Z")
	mock.ExpectQuery(`FROM users u WHERE id=\$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "email", "avatar", "level", "status", "created_at", "phone", "gender", "birthdate", "bio", "signature", "announcement", "follower_count", "following_count", "manuscript_count"}).
			AddRow(1, "u", "nick", "e", "a", 1, 1, now, "138", 0, "", "bio", "sig", "ann", 1, 2, 3))
	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/1", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserAdminHandler_handleGet_NotFound(t *testing.T) {
	h, mock := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{"user:manage": true}}, nil)
	mock.ExpectQuery(`FROM users u WHERE id=\$1`).WithArgs(int64(9)).WillReturnError(sql.ErrNoRows)
	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/9", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusNotFound, w.Code)

	mock.ExpectQuery(`FROM users u WHERE id=\$1`).WithArgs(int64(9)).WillReturnError(errors.New("boom"))
	w = doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/9", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserAdminHandler_handleUpdate(t *testing.T) {
	h, mock := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{"user:manage": true}, adminID: 1}, nil)

	w := doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1", `bad`, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE nickname = \$1 AND id != \$2`).
		WithArgs("dup", int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w = doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1", map[string]interface{}{"nickname": "dup"}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusConflict, w.Code)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE nickname = \$1 AND id != \$2`).
		WithArgs("newnick", int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`UPDATE users SET nickname=\$1, updated_at=NOW\(\) WHERE id=\$2`).WithArgs("newnick", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1", map[string]interface{}{"nickname": "newnick"}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectExec(`UPDATE users SET email=\$1, updated_at=NOW\(\) WHERE id=\$2`).WithArgs("e@x.com", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1", map[string]interface{}{"email": "e@x.com"}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectExec(`UPDATE users SET level=\$1, updated_at=NOW\(\) WHERE id=\$2`).WithArgs(int32(5), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1", map[string]interface{}{"level": 5}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserAdminHandler_handleStatus(t *testing.T) {
	auditor := &fakeAuditor{}
	h, mock := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{"user:manage": true}, adminID: 1}, auditor)

	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/1/status", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	w = doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1/status", `bad`, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mock.ExpectExec(`UPDATE users SET status=\$1, updated_at=NOW\(\) WHERE id=\$2`).WithArgs(int32(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1/status", map[string]interface{}{"status": 1}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, auditor.calls)

	mock.ExpectExec(`UPDATE users SET status=\$1`).WithArgs(int32(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 0))
	w = doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1/status", map[string]interface{}{"status": 1}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusNotFound, w.Code)

	mock.ExpectExec(`UPDATE users SET status=\$1`).WillReturnError(errors.New("boom"))
	w = doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1/status", map[string]interface{}{"status": 1}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserAdminHandler_handlePassword(t *testing.T) {
	auditor := &fakeAuditor{}
	h, mock := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{"user:manage": true}, adminID: 1}, auditor)

	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/1/password", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	w = doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1/password", `bad`, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1/password", map[string]interface{}{"newPassword": ""}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mock.ExpectExec(`UPDATE users SET password=\$1, updated_at=NOW\(\) WHERE id=\$2`).WithArgs(sha256hex("newpass"), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1/password", map[string]interface{}{"newPassword": "newpass"}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, auditor.calls)

	mock.ExpectExec(`UPDATE users SET password=\$1`).WithArgs(sha256hex("x"), int64(1)).WillReturnResult(sqlmock.NewResult(0, 0))
	w = doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1/password", map[string]interface{}{"newPassword": "x"}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusNotFound, w.Code)

	mock.ExpectExec(`UPDATE users SET password=\$1`).WillReturnError(errors.New("boom"))
	w = doRequest(t, muxForUserAdmin(h), "PUT", "/api/v1/user/admin/1/password", map[string]interface{}{"newPassword": "x"}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserAdminHandler_handleRoute_NotFoundAndMethod(t *testing.T) {
	h, _ := newUserAdminHandler(t, &fakePerm{allowed: map[string]bool{"user:manage": true}}, nil)
	w := doRequest(t, muxForUserAdmin(h), "GET", "/api/v1/user/admin/1/unknown", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = doRequest(t, muxForUserAdmin(h), "PATCH", "/api/v1/user/admin/1", nil, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestSha256hex(t *testing.T) {
	assert.Equal(t, "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8", sha256hex("password"))
}

func TestHandleAvatarUpload(t *testing.T) {
	h, mock := newTestHandler(t)

	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/1/avatar", nil, nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	w = doRequest(t, muxForUserExtend(h), "POST", "/api/v1/user/abc/avatar", nil, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", "a.png")
	require.NoError(t, err)
	fw.Write([]byte("fake-image"))
	mw.Close()

	mock.ExpectExec(`UPDATE users SET avatar = \$1, updated_at = NOW\(\) WHERE id = \$2`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	req := httptest.NewRequest("POST", "/api/v1/user/1/avatar", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	muxForUserExtend(h).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
