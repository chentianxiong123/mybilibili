package creator

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePerm struct{ ok bool }

func (f *fakePerm) CheckPermission(r *http.Request, permission string) (int64, bool) {
	if f.ok {
		return 99, true
	}
	return 0, false
}

func newMockHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewHandler(NewService(NewRepo(db))), mock
}

func muxForCreator(h *Handler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func creatorReq(h http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestCoupons_Unauthorized(t *testing.T) {
	h, _ := newMockHandler(t)
	w := creatorReq(muxForCreator(h), "GET", "/api/v1/creator/coupons", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCoupons_OK(t *testing.T) {
	h, mock := newMockHandler(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM traffic_coupons WHERE user_id`).WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "amount", "reason", "task_id", "status", "granted_at", "used_at", "expires_at"}).
			AddRow(int64(1), int64(7), 100, "任务奖励", int64(0), 0, now, nil, nil))

	w := creatorReq(muxForCreator(h), "GET", "/api/v1/creator/coupons", "", map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBalance_OK(t *testing.T) {
	h, mock := newMockHandler(t)
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\), 0\)`).WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(120))

	w := creatorReq(muxForCreator(h), "GET", "/api/v1/creator/coupons/balance", "", map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "120")
}

func TestRedeem_Unauthorized(t *testing.T) {
	h, _ := newMockHandler(t)
	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/coupons/redeem", `{"coupon_id":1,"manuscript_id":50}`, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRedeem_OK(t *testing.T) {
	h, mock := newMockHandler(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).AddRow(100, int64(7), 0, nil))
	mock.ExpectExec(`UPDATE traffic_coupons SET status = 1`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO traffic_coupon_uses`).WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec(`UPDATE manuscripts SET exposure_weight`).WithArgs(100, int64(50), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/coupons/redeem",
		`{"coupon_id":1,"manuscript_id":50}`, map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "100")
}

func TestGrant_Forbidden_WhenNoPerm(t *testing.T) {
	h, _ := newMockHandler(t)
	h.perm = &fakePerm{ok: false}
	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/coupons/grant",
		`{"user_id":7,"amount":100}`, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGrant_NoPermChecker(t *testing.T) {
	h, _ := newMockHandler(t) // perm nil
	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/coupons/grant",
		`{"user_id":7,"amount":100}`, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGrant_OK(t *testing.T) {
	h, mock := newMockHandler(t)
	h.perm = &fakePerm{ok: true}

	mock.ExpectQuery(`INSERT INTO traffic_coupons`).WithArgs(int64(7), 100, "平台发放", int64(0), sql.NullTime{}).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/coupons/grant",
		`{"user_id":7,"amount":100}`, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "1")
}

func TestTasks_Anonymous_StillOK(t *testing.T) {
	h, mock := newMockHandler(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE enabled = true`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", 100, 1, "once", true, now))

	w := creatorReq(muxForCreator(h), "GET", "/api/v1/creator/tasks", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskProgress_OK(t *testing.T) {
	h, mock := newMockHandler(t)
	now := time.Now()

	// GetTaskByID
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", 0, 1, "once", true, now))
	// Upsert: already exists, no reward claimed -> return (0,false) after commit
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WithArgs(int64(7), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"progress", "done"}).AddRow(0, false))
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).WithArgs(int64(7), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	mock.ExpectExec(`UPDATE creator_task_progress SET progress`).WithArgs(int64(7), int64(1), 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).WithArgs(int64(7), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	mock.ExpectExec(`UPDATE creator_task_progress SET reward_claimed`).WithArgs(int64(7), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/tasks/progress",
		`{"task_id":1,"delta":1}`, map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusOK, w.Code)
}
