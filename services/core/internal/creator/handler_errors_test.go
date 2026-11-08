package creator

// handler_errors_test.go — error / boundary paths for Handler.

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// TestSetPermChecker covers handler.go:26.
func TestSetPermChecker(t *testing.T) {
	h, _ := newMockHandler(t)
	perm := &fakePerm{ok: true}
	h.SetPermChecker(perm)
	assert.Same(t, perm, h.perm)
}

// ---- handleCoupons ----

// TestCoupons_MethodNotAllowed covers handler.go:42-45.
func TestCoupons_MethodNotAllowed(t *testing.T) {
	h, _ := newMockHandler(t)
	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/coupons", "", map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestCoupons_ServiceError covers handler.go:47-50.
func TestCoupons_ServiceError(t *testing.T) {
	h, mock := newMockHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupons WHERE user_id`).
		WillReturnError(assert.AnError)

	w := creatorReq(muxForCreator(h), "GET", "/api/v1/creator/coupons", "", map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---- handleBalance ----

// TestBalance_Unauthorized covers handler.go:56-58 (RequireUser false).
func TestBalance_Unauthorized(t *testing.T) {
	h, _ := newMockHandler(t)
	w := creatorReq(muxForCreator(h), "GET", "/api/v1/creator/coupons/balance", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestBalance_ServiceError covers handler.go:60-63.
func TestBalance_ServiceError(t *testing.T) {
	h, mock := newMockHandler(t)
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\), 0\)`).
		WillReturnError(assert.AnError)

	w := creatorReq(muxForCreator(h), "GET", "/api/v1/creator/coupons/balance", "", map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---- handleRedeem ----

// TestRedeem_MethodNotAllowed covers handler.go:73-76.
func TestRedeem_MethodNotAllowed(t *testing.T) {
	h, _ := newMockHandler(t)
	w := creatorReq(muxForCreator(h), "GET", "/api/v1/creator/coupons/redeem", "", map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestRedeem_InvalidBody covers handler.go:81-84.
func TestRedeem_InvalidBody(t *testing.T) {
	h, _ := newMockHandler(t)
	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/coupons/redeem", "{bad json", map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRedeem_ServiceError_NotFound covers handler.go:86-89 (WriteHTTPError -> 404).
func TestRedeem_ServiceError_NotFound(t *testing.T) {
	h, mock := newMockHandler(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).
			AddRow(100, int64(99), 0, nil))
	mock.ExpectRollback()

	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/coupons/redeem",
		`{"coupon_id":1,"manuscript_id":50}`, map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---- handleGrant ----

// TestGrant_MethodNotAllowed covers handler.go:102-105.
func TestGrant_MethodNotAllowed(t *testing.T) {
	h, _ := newMockHandler(t)
	h.perm = &fakePerm{ok: true}
	w := creatorReq(muxForCreator(h), "GET", "/api/v1/creator/coupons/grant", "", map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestGrant_InvalidBody covers handler.go:111-114.
func TestGrant_InvalidBody(t *testing.T) {
	h, _ := newMockHandler(t)
	h.perm = &fakePerm{ok: true}
	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/coupons/grant", "{bad json", map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGrant_ServiceError_InvalidAmount covers handler.go:116-119 (WriteHTTPError -> 400).
func TestGrant_ServiceError_InvalidAmount(t *testing.T) {
	h, _ := newMockHandler(t)
	h.perm = &fakePerm{ok: true}
	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/coupons/grant",
		`{"user_id":7,"amount":0}`, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGrant_ServiceError_RepoDown covers handler.go:116-119 -> 500.
func TestGrant_ServiceError_RepoDown(t *testing.T) {
	h, mock := newMockHandler(t)
	h.perm = &fakePerm{ok: true}
	mock.ExpectQuery(`INSERT INTO traffic_coupons`).WillReturnError(assert.AnError)

	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/coupons/grant",
		`{"user_id":7,"amount":100}`, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---- handleTasks ----

// TestTasks_ServiceError covers handler.go:126-129.
func TestTasks_ServiceError(t *testing.T) {
	h, mock := newMockHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE enabled = true`).
		WillReturnError(assert.AnError)

	w := creatorReq(muxForCreator(h), "GET", "/api/v1/creator/tasks", "", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestTasks_WithUser covers the authenticated path (progress looked up).
func TestTasks_WithUser(t *testing.T) {
	h, mock := newMockHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE enabled = true`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", 100, 1, "once", true, now))
	mock.ExpectQuery(`SELECT .+ FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "task_id", "progress", "done", "reward_claimed", "period_start", "updated_at"}).
			AddRow(int64(1), int64(7), int64(1), 1, true, true, now, now))

	w := creatorReq(muxForCreator(h), "GET", "/api/v1/creator/tasks", "", map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"done":true`)
}

// ---- handleTaskProgress ----

// TestTaskProgress_Unauthorized covers handler.go:135-137 (RequireUser false).
func TestTaskProgress_Unauthorized(t *testing.T) {
	h, _ := newMockHandler(t)
	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/tasks/progress",
		`{"task_id":1,"delta":1}`, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestTaskProgress_MethodNotAllowed covers handler.go:138-141.
func TestTaskProgress_MethodNotAllowed(t *testing.T) {
	h, _ := newMockHandler(t)
	w := creatorReq(muxForCreator(h), "GET", "/api/v1/creator/tasks/progress", "", map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestTaskProgress_InvalidBody covers handler.go:146-149.
func TestTaskProgress_InvalidBody(t *testing.T) {
	h, _ := newMockHandler(t)
	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/tasks/progress", "{bad json", map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTaskProgress_ServiceError_Internal covers handler.go:151-154
// (WriteHTTPError on a generic DB error -> 500).
func TestTaskProgress_ServiceError_Internal(t *testing.T) {
	h, mock := newMockHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).
		WillReturnError(assert.AnError)

	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/tasks/progress",
		`{"task_id":999,"delta":1}`, map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestTaskProgress_ServiceError_NotFound covers handler.go:151-154
// (WriteHTTPError on ErrNotFound -> 404).
func TestTaskProgress_ServiceError_NotFound(t *testing.T) {
	h, mock := newMockHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).
		WillReturnError(sql.ErrNoRows)

	w := creatorReq(muxForCreator(h), "POST", "/api/v1/creator/tasks/progress",
		`{"task_id":999,"delta":1}`, map[string]string{"X-User-Id": "7"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}
