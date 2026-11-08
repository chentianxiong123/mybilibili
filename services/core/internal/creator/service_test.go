package creator

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	myerr "mybilibili/pkg/errors"
)

func newService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewService(NewRepo(db)), mock
}

func TestService_Grant_InvalidAmount(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.Grant(context.Background(), 7, 0, "x")
	assert.ErrorIs(t, err, myerr.ErrInvalidArgument("coupon amount must be positive"))
}

func TestService_Balance(t *testing.T) {
	svc, mock := newService(t)
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\), 0\)`).WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(120))

	bal, err := svc.Balance(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, 120, bal)
}

func TestService_Coupons(t *testing.T) {
	svc, mock := newService(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM traffic_coupons WHERE user_id`).WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "amount", "reason", "task_id", "status", "granted_at", "used_at", "expires_at"}).
			AddRow(int64(1), int64(7), 100, "任务奖励", int64(0), 0, now, nil, nil))

	list, err := svc.Coupons(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, 100, list[0].Amount)
}

func TestService_Redeem_OK(t *testing.T) {
	svc, mock := newService(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).AddRow(100, int64(7), 0, nil))
	mock.ExpectExec(`UPDATE traffic_coupons SET status = 1`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO traffic_coupon_uses`).WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec(`UPDATE manuscripts SET exposure_weight`).WithArgs(100, int64(50)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	amount, err := svc.Redeem(ctx, 7, 1, 50)
	require.NoError(t, err)
	assert.Equal(t, 100, amount)
}

func TestService_Redeem_NotOwner(t *testing.T) {
	svc, mock := newService(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).AddRow(100, int64(99), 0, nil))
	mock.ExpectRollback()

	_, err := svc.Redeem(ctx, 7, 1, 50)
	assert.ErrorIs(t, err, myerr.ErrNotFound("coupon not found or not yours"))
}

func TestService_Redeem_Unauthenticated(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.Redeem(context.Background(), 0, 1, 50)
	assert.ErrorIs(t, err, myerr.ErrUnauthenticated("unauthorized"))
}

func TestService_Tasks(t *testing.T) {
	svc, mock := newService(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE enabled = true`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "发一个视频", 100, 1, "once", true, now))

	mock.ExpectQuery(`SELECT .+ FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WithArgs(int64(7), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "task_id", "progress", "done", "reward_claimed", "period_start", "updated_at"}).
			AddRow(int64(1), int64(7), int64(1), 1, true, true, now, now))

	list, err := svc.Tasks(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, 100, list[0].RewardCoupon)
	assert.True(t, list[0].Done)
}

func TestService_RecordTaskProgress_Completed(t *testing.T) {
	svc, mock := newService(t)
	ctx := context.Background()
	now := time.Now()

	// GetTaskByID
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", 100, 1, "once", true, now))

	// UpsertTaskProgress transaction -> returns reward claimed (1, true)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WithArgs(int64(7), int64(1)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO creator_task_progress`).WithArgs(int64(7), int64(1), 1, true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).WithArgs(int64(7), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	mock.ExpectExec(`UPDATE creator_task_progress SET reward_claimed`).WithArgs(int64(7), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// GrantCoupon (reward)
	mock.ExpectQuery(`INSERT INTO traffic_coupons`).WithArgs(int64(7), 100, "任务奖励", int64(1), sql.NullTime{}).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(2)))

	reward, err := svc.RecordTaskProgress(ctx, 7, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, 100, reward)
}

func TestService_RecordTaskProgress_Unauthenticated(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.RecordTaskProgress(context.Background(), 0, 1, 1)
	assert.ErrorIs(t, err, myerr.ErrUnauthenticated("unauthorized"))
}
