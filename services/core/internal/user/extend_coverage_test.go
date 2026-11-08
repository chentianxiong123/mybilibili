package user

import (
	"bytes"
	"database/sql"
	"errors"
	"mime/multipart"
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

// ---- handleLogin / handleRegister / handleLogout: 方法与非 JSON 分支 ----

func TestHandleLogin_405_Method(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/login", nil, nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleLogin_400_BadJSON(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/user/login", strings.NewReader("not json"))
	rec := httptest.NewRecorder()
	muxForUserExtend(h).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleRegister_405_Method(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/register", nil, nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleRegister_400_BadJSON(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/user/register", strings.NewReader("not json"))
	rec := httptest.NewRecorder()
	muxForUserExtend(h).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleLogout_200(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doRequest(t, muxForUserExtend(h), "POST", "/api/v1/user/logout", nil, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
}

func TestHandleLogout_405_Method(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/logout", nil, nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// ---- handleRefresh: 405 / cookie 兜底 / 账号禁用 ----

func TestHandleRefresh_405_Method(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/token/refresh", nil, nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleRefresh_CookieFallback_200(t *testing.T) {
	h, mock := newTestHandler(t)
	svc := h.svc
	tok, err := svc.jwt.GenerateRefresh(7)
	require.NoError(t, err)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(7)).
		WillReturnRows(userRow(7, "u7", "h", "Nick7", 1))

	// body 不带 refreshToken，凭证从 HttpOnly cookie 读。
	req := httptest.NewRequest("POST", "/api/v1/user/token/refresh", strings.NewReader(`{}`))
	req.AddCookie(&http.Cookie{Name: auth.UserRefreshTokenCookie, Value: tok})
	rec := httptest.NewRecorder()
	muxForUserExtend(h).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRefresh_UserDisabled(t *testing.T) {
	h, mock := newTestHandler(t)
	svc := h.svc
	tok, err := svc.jwt.GenerateRefresh(8)
	require.NoError(t, err)

	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(8)).
		WillReturnRows(userRow(8, "u8", "h", "Nick8", 0 /* disabled */))

	w := doRequest(t, muxForUserExtend(h), "POST", "/api/v1/user/token/refresh", map[string]string{"refreshToken": tok}, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "disabled")
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- handleEmailCode / handleEmailVerify: 405 与空/无效分支 ----

func TestHandleEmailCode_405_Method(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/email/code", nil, nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleEmailCode_400_EmptyEmail(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doRequest(t, muxForUserExtend(h), "POST", "/api/v1/user/email/code", map[string]string{"email": ""}, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleEmailVerify_405_Method(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/email/verify", nil, nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleEmailVerify_InvalidCode(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id FROM verification_codes`).
		WithArgs("a@b.com", "000000").
		WillReturnError(sql.ErrNoRows)

	w := doRequest(t, muxForUserExtend(h), "POST", "/api/v1/user/email/verify", map[string]string{"email": "a@b.com", "code": "000000"}, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "failed")
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- handleBatch: 405 与查询失败 ----

func TestHandleBatch_405_Method(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/batch", nil, nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleBatch_QueryError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users WHERE id IN`).
		WillReturnError(errors.New("boom"))

	w := doRequest(t, muxForUserExtend(h), "POST", "/api/v1/user/batch", []int64{1, 2}, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- handleForgotPassword: 405 ----

func TestHandleForgotPassword_405_Method(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/password/forgot", nil, nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// ---- handlePinnedVideo: 未登录 POST 401 与 GET 非零视频 ----

func TestHandlePinnedVideo_NonGET_401(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doRequest(t, muxForUserExtend(h), "POST", "/api/v1/user/pinned-video", map[string]interface{}{"videoId": 1}, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandlePinnedVideo_GET_WithVideo(t *testing.T) {
	h, mock := newTestHandler(t)
	uploadTime := time.Now()
	mock.ExpectQuery(`SELECT COALESCE\(pinned_video_id,0\) FROM users WHERE id`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"pinned_video_id"}).AddRow(42))
	mock.ExpectQuery(`SELECT id, COALESCE\(title`).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "cover_url", "view_count", "comment_count", "duration", "upload_time"}).
			AddRow(42, "t", "c", 10, 2, "00:30", uploadTime))

	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/pinned-video?uid=5", nil, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":42`)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- handleLoginLogs / handleLoginLogCount: 401 与查询失败/条件分支 ----

func TestHandleLoginLogs_401_NoUser(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/login-logs", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleLoginLogs_QueryError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, user_id, ip, user_agent, status, login_time FROM login_logs`).
		WithArgs(int64(9), int32(20), int32(0)).
		WillReturnError(errors.New("boom"))

	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/login-logs", nil, map[string]string{"X-User-Id": "9"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLoginLogCount_401_NoUser(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/login-logs/count", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleLoginLogCount_Conditions(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM login_logs`).
		WithArgs(int64(5), "2025-01-01", "2025-01-02").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/login-logs/count?user_id=5&start_time=2025-01-01&end_time=2025-01-02", nil, map[string]string{"X-User-Id": "9"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total":3`)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- handlePrivacy GET: 查询失败走默认值 ----

func TestHandlePrivacy_GET_QueryError_Defaults(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT public_collection, public_birthday_tags`).
		WithArgs(int64(9)).
		WillReturnError(sql.ErrNoRows)

	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/privacy/test", nil, map[string]string{"X-User-Id": "9"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"publicCollection":1`)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- handleMe: 查询错误 / PUT 解码错 / 昵称冲突 / 头像更新 ----

func TestHandleMe_Get_NotFound(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(9)).
		WillReturnError(sql.ErrNoRows)

	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/me", nil, map[string]string{"X-User-Id": "9"})
	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMe_PUT_BadJSON(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest("PUT", "/api/v1/user/me", strings.NewReader("not json"))
	req.Header.Set("X-User-Id", "9")
	rec := httptest.NewRecorder()
	muxForUserExtend(h).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleMe_PUT_NicknameConflict(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE nickname`).
		WithArgs("dup", int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	w := doRequest(t, muxForUserExtend(h), "PUT", "/api/v1/user/me", map[string]interface{}{"nickname": "dup"}, map[string]string{"X-User-Id": "9"})
	assert.Equal(t, http.StatusConflict, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMe_PUT_Avatar(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE users SET avatar`).
		WithArgs("https://x/a.png", int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := doRequest(t, muxForUserExtend(h), "PUT", "/api/v1/user/me", map[string]interface{}{"avatar": "https://x/a.png"}, map[string]string{"X-User-Id": "9"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- handleMeAvatar: 401 与更新失败 ----

func TestHandleMeAvatar_401_NoUser(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doRequest(t, muxForUserExtend(h), "POST", "/api/v1/user/me/avatar", map[string]string{"avatar": "a"}, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleMeAvatar_UpdateError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE users SET avatar`).
		WithArgs("a.png", int64(9)).
		WillReturnError(errors.New("boom"))

	w := doRequest(t, muxForUserExtend(h), "POST", "/api/v1/user/me/avatar", map[string]string{"avatar": "a.png"}, map[string]string{"X-User-Id": "9"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- handleUserByID: PUT 更新失败 与 非 NoRows 数据库错误 ----

func TestHandleUserByID_PutUpdateError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE users SET signature`).
		WillReturnError(errors.New("boom"))

	w := doRequest(t, muxForUserExtend(h), "PUT", "/api/v1/user/1", map[string]interface{}{"signature": "s"}, nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUserByID_GetDBError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT.*FROM users.*WHERE id`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("boom"))

	w := doRequest(t, muxForUserExtend(h), "GET", "/api/v1/user/1", nil, nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- handleAvatarUpload: multipart 解析失败 / 缺文件 / 无扩展名 / DB 失败 ----

func TestHandleAvatarUpload_ParseFormError(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/user/1/avatar", strings.NewReader("plain text"))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	muxForUserExtend(h).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleAvatarUpload_FileRequired(t *testing.T) {
	h, _ := newTestHandler(t)
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	require.NoError(t, mw.Close()) // 无 file 字段

	req := httptest.NewRequest("POST", "/api/v1/user/1/avatar", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	muxForUserExtend(h).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleAvatarUpload_NoExt(t *testing.T) {
	h, mock := newTestHandler(t)
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", "noext")
	require.NoError(t, err)
	_, _ = fw.Write([]byte("fake-image"))
	require.NoError(t, mw.Close())

	mock.ExpectExec(`UPDATE users SET avatar`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/user/1/avatar", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	muxForUserExtend(h).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAvatarUpload_DBError(t *testing.T) {
	h, mock := newTestHandler(t)
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", "a.png")
	require.NoError(t, err)
	_, _ = fw.Write([]byte("fake-image"))
	require.NoError(t, mw.Close())

	mock.ExpectExec(`UPDATE users SET avatar`).
		WillReturnError(errors.New("boom"))

	req := httptest.NewRequest("POST", "/api/v1/user/1/avatar", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	muxForUserExtend(h).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- handleCaptcha verify 成功 ----

func TestHandleCaptcha_Verify_Success(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id FROM verification_codes`).
		WithArgs("abc", "42").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec(`UPDATE verification_codes SET used = 1`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := doRequest(t, muxForUserExtend(h), "POST", "/api/v1/captcha/verify", map[string]string{"captchaId": "abc", "answer": "42"}, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `true`)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- sendEmail: 已配置 SMTP（连接失败走错误日志分支） ----

func TestSendEmail_Configured(t *testing.T) {
	t.Setenv("SMTP_HOST", "127.0.0.1")
	t.Setenv("SMTP_PORT", "1") // 无监听端口 → 快速 connection refused
	t.Setenv("SMTP_USER", "u")
	t.Setenv("SMTP_PASS", "p")
	t.Setenv("SMTP_FROM", "from@x.com")
	// 不应 panic，连接失败只记日志。
	sendEmail("a@b.com", "123456")
}

func TestSendEmail_NotConfigured(t *testing.T) {
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_FROM", "")
	sendEmail("a@b.com", "123456")
}
