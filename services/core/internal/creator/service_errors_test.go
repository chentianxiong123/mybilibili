package creator

// service_errors_test.go — error / boundary paths for Service.

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	myerr "mybilibili/pkg/errors"
)

// ---- Grant ----

func TestService_Grant_InvalidUser(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.Grant(context.Background(), 0, 100, "x")
	assert.ErrorIs(t, err, myerr.ErrInvalidArgument("invalid user"))
}

func TestService_Grant_NegativeUser(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.Grant(context.Background(), -1, 100, "x")
	assert.ErrorIs(t, err, myerr.ErrInvalidArgument("invalid user"))
}

func TestService_Grant_NegativeAmount(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.Grant(context.Background(), 7, -5, "x")
	assert.ErrorIs(t, err, myerr.ErrInvalidArgument("coupon amount must be positive"))
}

// TestService_Grant_OK covers service.go:48-55 (default reason + repo call).
func TestService_Grant_OK(t *testing.T) {
	svc, mock := newService(t)
	mock.ExpectQuery(`INSERT INTO traffic_coupons`).
		WithArgs(int64(7), 100, "平台发放", int64(0), sql.NullTime{}).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	id, err := svc.Grant(context.Background(), 7, 100, "")
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
}

// TestService_Grant_WithReason covers service.go:51-55 (custom reason).
func TestService_Grant_WithReason(t *testing.T) {
	svc, mock := newService(t)
	mock.ExpectQuery(`INSERT INTO traffic_coupons`).
		WithArgs(int64(7), 100, "自定义原因", int64(0), sql.NullTime{}).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(2)))

	id, err := svc.Grant(context.Background(), 7, 100, "自定义原因")
	require.NoError(t, err)
	assert.Equal(t, int64(2), id)
}

// TestService_Grant_RepoError covers service.go:51-55 error propagation.
func TestService_Grant_RepoError(t *testing.T) {
	svc, mock := newService(t)
	mock.ExpectQuery(`INSERT INTO traffic_coupons`).
		WillReturnError(errors.New("insert failed"))

	_, err := svc.Grant(context.Background(), 7, 100, "x")
	assert.Error(t, err)
}

// ---- Balance ----

func TestService_Balance_RepoError(t *testing.T) {
	svc, mock := newService(t)
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\), 0\)`).
		WillReturnError(errors.New("db error"))

	_, err := svc.Balance(context.Background(), 7)
	assert.Error(t, err)
}

// ---- Coupons ----

// TestService_Coupons_ListError covers service.go:66-68.
func TestService_Coupons_ListError(t *testing.T) {
	svc, mock := newService(t)
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupons WHERE user_id`).
		WillReturnError(errors.New("list failed"))

	_, err := svc.Coupons(context.Background(), 7)
	assert.Error(t, err)
}

// TestService_Coupons_ValidTimes covers service.go:75-82 (UsedAt/ExpiresAt
// pointers set when Valid).
func TestService_Coupons_ValidTimes(t *testing.T) {
	svc, mock := newService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupons WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "amount", "reason", "task_id", "status", "granted_at", "used_at", "expires_at"}).
			AddRow(int64(1), int64(7), 100, "任务奖励", int64(3), 1, now, now, now))

	list, err := svc.Coupons(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.NotNil(t, list[0].UsedAt)
	assert.NotNil(t, list[0].ExpiresAt)
	assert.True(t, list[0].UsedAt.Equal(now))
	assert.True(t, list[0].ExpiresAt.Equal(now))
}

// TestService_Coupons_Empty verifies empty slice, no error.
func TestService_Coupons_Empty(t *testing.T) {
	svc, mock := newService(t)
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupons WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "amount", "reason", "task_id", "status", "granted_at", "used_at", "expires_at"}))

	list, err := svc.Coupons(context.Background(), 7)
	require.NoError(t, err)
	assert.Empty(t, list)
}

// ---- Redeem ----

// TestService_Redeem_InvalidCouponID covers service.go:93-95.
func TestService_Redeem_InvalidCouponID(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.Redeem(context.Background(), 7, 0, 50)
	assert.ErrorIs(t, err, myerr.ErrInvalidArgument("invalid coupon or manuscript"))
}

// TestService_Redeem_InvalidManuscriptID covers service.go:93-95.
func TestService_Redeem_InvalidManuscriptID(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.Redeem(context.Background(), 7, 1, 0)
	assert.ErrorIs(t, err, myerr.ErrInvalidArgument("invalid coupon or manuscript"))
}

// TestService_Redeem_Used covers service.go:101-102 (errCouponUsed branch).
func TestService_Redeem_Used(t *testing.T) {
	svc, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).
			AddRow(100, int64(7), 1, nil))
	mock.ExpectRollback()

	_, err := svc.Redeem(context.Background(), 7, 1, 50)
	assert.ErrorIs(t, err, myerr.ErrInvalidArgument("coupon already used"))
}

// TestService_Redeem_Expired covers service.go:103-104 (errCouponExpired branch).
func TestService_Redeem_Expired(t *testing.T) {
	svc, mock := newService(t)
	past := time.Now().Add(-time.Hour)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).
			AddRow(100, int64(7), 0, past))
	mock.ExpectRollback()

	_, err := svc.Redeem(context.Background(), 7, 1, 50)
	assert.ErrorIs(t, err, myerr.ErrInvalidArgument("coupon expired"))
}

// TestService_Redeem_Internal covers service.go:105-106 (default branch).
func TestService_Redeem_Internal(t *testing.T) {
	svc, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WillReturnError(errors.New("db boom"))
	mock.ExpectRollback()

	_, err := svc.Redeem(context.Background(), 7, 1, 50)
	assert.ErrorIs(t, err, myerr.ErrInternal("database error"))
}

// ---- Uses ----

// TestService_Uses covers service.go:113-115.
func TestService_Uses(t *testing.T) {
	svc, mock := newService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupon_uses WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "coupon_id", "user_id", "manuscript_id", "amount", "created_at"}).
			AddRow(int64(1), int64(1), int64(7), int64(50), 100, now))

	list, err := svc.Uses(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(50), list[0].ManuscriptID)
}

// TestService_uses_RepoError covers service.go:114-115 error propagation.
func TestService_Uses_RepoError(t *testing.T) {
	svc, mock := newService(t)
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupon_uses WHERE user_id`).
		WillReturnError(errors.New("db error"))

	_, err := svc.Uses(context.Background(), 7)
	assert.Error(t, err)
}

// ---- Tasks ----

// TestService_Tasks_ListError covers service.go:136-138.
func TestService_Tasks_ListError(t *testing.T) {
	svc, mock := newService(t)
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE enabled = true`).
		WillReturnError(errors.New("list failed"))

	_, err := svc.Tasks(context.Background(), 7)
	assert.Error(t, err)
}

// TestService_Tasks_Anonymous covers service.go:145 (userID<=0 skips progress).
func TestService_Tasks_Anonymous(t *testing.T) {
	svc, mock := newService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE enabled = true`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", 100, 1, "once", true, now))

	list, err := svc.Tasks(context.Background(), 0)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.False(t, list[0].Done) // no progress looked up
}

// TestService_Tasks_EmptyList verifies empty task list.
func TestService_Tasks_EmptyList(t *testing.T) {
	svc, mock := newService(t)
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE enabled = true`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}))

	list, err := svc.Tasks(context.Background(), 7)
	require.NoError(t, err)
	assert.Empty(t, list)
}

// TestService_Tasks_ProgressError covers service.go:146 (perr != nil is
// silently ignored — no progress attached).
func TestService_Tasks_ProgressError(t *testing.T) {
	svc, mock := newService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE enabled = true`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", 100, 1, "once", true, now))
	// GetTaskProgress returns an error -> progress stays zero
	mock.ExpectQuery(`SELECT .+ FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnError(errors.New("progress db error"))

	list, err := svc.Tasks(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.False(t, list[0].Done)
}

// ---- RecordTaskProgress ----

// TestService_RecordTaskProgress_NegativeDelta covers service.go:162-164
// (delta<=0 coerced to 1).
func TestService_RecordTaskProgress_NegativeDelta(t *testing.T) {
	svc, mock := newService(t)
	now := time.Now()
	// GetTaskByID
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", 0, 1, "once", true, now))
	// Upsert: delta forced to 1, no existing record, target=1 -> done
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO creator_task_progress`).WithArgs(int64(7), int64(1), 1, true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	mock.ExpectExec(`UPDATE creator_task_progress SET reward_claimed`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	_, err := svc.RecordTaskProgress(context.Background(), 7, 1, -1)
	require.NoError(t, err)
}

// TestService_RecordTaskProgress_ZeroDelta covers service.go:162-164.
func TestService_RecordTaskProgress_ZeroDelta(t *testing.T) {
	svc, mock := newService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", 0, 1, "once", true, now))
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO creator_task_progress`).WithArgs(int64(7), int64(1), 1, true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	mock.ExpectExec(`UPDATE creator_task_progress SET reward_claimed`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	_, err := svc.RecordTaskProgress(context.Background(), 7, 1, 0)
	require.NoError(t, err)
}

// TestService_RecordTaskProgress_TaskNotFound covers service.go:166-168.
func TestService_RecordTaskProgress_TaskNotFound(t *testing.T) {
	svc, mock := newService(t)
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.RecordTaskProgress(context.Background(), 7, 999, 1)
	assert.ErrorIs(t, err, myerr.ErrNotFound("task not found"))
}

// TestService_RecordTaskProgress_TaskDBError covers service.go:170-174.
func TestService_RecordTaskProgress_TaskDBError(t *testing.T) {
	svc, mock := newService(t)
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).
		WillReturnError(errors.New("db boom"))

	_, err := svc.RecordTaskProgress(context.Background(), 7, 1, 1)
	assert.ErrorIs(t, err, myerr.ErrInternal("database error"))
}

// TestService_RecordTaskProgress_TaskDisabled covers service.go:172-174.
func TestService_RecordTaskProgress_TaskDisabled(t *testing.T) {
	svc, mock := newService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", 0, 1, "once", false, now))

	_, err := svc.RecordTaskProgress(context.Background(), 7, 1, 1)
	assert.ErrorIs(t, err, myerr.ErrInvalidArgument("task disabled"))
}

// TestService_RecordTaskProgress_UpsertError covers service.go:176-178.
func TestService_RecordTaskProgress_UpsertError(t *testing.T) {
	svc, mock := newService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", 100, 1, "once", true, now))
	mock.ExpectBegin().WillReturnError(errors.New("cannot begin"))

	_, err := svc.RecordTaskProgress(context.Background(), 7, 1, 1)
	assert.ErrorIs(t, err, myerr.ErrInternal("database error"))
}

// TestService_RecordTaskProgress_GrantRewardError covers service.go:185-187.
func TestService_RecordTaskProgress_GrantRewardError(t *testing.T) {
	svc, mock := newService(t)
	now := time.Now()
	// GetTaskByID
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", 100, 1, "once", true, now))
	// Upsert returns done=true
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO creator_task_progress`).WithArgs(int64(7), int64(1), 1, true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	mock.ExpectExec(`UPDATE creator_task_progress SET reward_claimed`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	// GrantCoupon (reward) fails
	mock.ExpectQuery(`INSERT INTO traffic_coupons`).
		WillReturnError(errors.New("grant reward failed"))

	_, err := svc.RecordTaskProgress(context.Background(), 7, 1, 1)
	assert.ErrorIs(t, err, myerr.ErrInternal("database error"))
}

// TestService_RecordTaskProgress_DoneNoReward covers service.go:189-191
// (done but RewardCoupon==0 -> returns 0, no grant).
func TestService_RecordTaskProgress_DoneNoReward(t *testing.T) {
	svc, mock := newService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", 0, 1, "once", true, now))
	// Upsert: existing record already done+claimed -> returns (0,false,commit)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnRows(sqlmock.NewRows([]string{"progress", "done"}).AddRow(1, true))
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(true))
	mock.ExpectCommit()

	reward, err := svc.RecordTaskProgress(context.Background(), 7, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, 0, reward)
}
