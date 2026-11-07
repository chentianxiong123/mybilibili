package creator

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRepo(t *testing.T) (*Repo, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewRepo(db), mock
}

func TestGrantCoupon(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`INSERT INTO traffic_coupons`).
		WithArgs(int64(7), 100, "测试", int64(0), sql.NullTime{}).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	id, err := repo.GrantCoupon(ctx, &Coupon{UserID: 7, Amount: 100, Reason: "测试"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListCoupons(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "amount", "reason", "task_id", "status", "granted_at", "used_at", "expires_at"}).
		AddRow(int64(1), int64(7), 100, "任务奖励", int64(0), 0, now, nil, nil)
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupons WHERE user_id`).WithArgs(int64(7)).WillReturnRows(rows)

	list, err := repo.ListCoupons(ctx, 7)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, 100, list[0].Amount)
}

func TestUsableCouponAmount(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\), 0\)`).WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(250))

	total, err := repo.UsableCouponAmount(ctx, 7)
	require.NoError(t, err)
	assert.Equal(t, 250, total)
}

func TestRedeemCoupon_OK(t *testing.T) {
	repo, mock := newRepo(t)
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

	amount, err := repo.RedeemCoupon(ctx, 1, 7, 50)
	require.NoError(t, err)
	assert.Equal(t, 100, amount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedeemCoupon_NotOwner(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).AddRow(100, int64(99), 0, nil))
	mock.ExpectRollback()

	_, err := repo.RedeemCoupon(ctx, 1, 7, 50)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestRedeemCoupon_AlreadyUsed(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).AddRow(100, int64(7), 1, nil))
	mock.ExpectRollback()

	_, err := repo.RedeemCoupon(ctx, 1, 7, 50)
	assert.ErrorIs(t, err, errCouponUsed)
}

func TestListCouponUses(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "coupon_id", "user_id", "manuscript_id", "amount", "created_at"}).
		AddRow(int64(1), int64(1), int64(7), int64(50), 100, now)
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupon_uses WHERE user_id`).WithArgs(int64(7)).WillReturnRows(rows)

	list, err := repo.ListCouponUses(ctx, 7)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(50), list[0].ManuscriptID)
}

func TestListTasks(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
		AddRow(int64(1), "first_upload", "首次投稿", "发一个视频", 100, 1, "once", true, now)
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE enabled = true`).WillReturnRows(rows)

	list, err := repo.ListTasks(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "first_upload", list[0].Code)
}

func TestGetTaskProgress(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WithArgs(int64(7), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "task_id", "progress", "done", "reward_claimed", "period_start", "updated_at"}).
			AddRow(int64(1), int64(7), int64(1), 3, false, false, now, now))

	p, err := repo.GetTaskProgress(ctx, 7, 1)
	require.NoError(t, err)
	assert.Equal(t, 3, p.Progress)
	assert.False(t, p.Done)
}

func TestGetTaskByID(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", 100, 1, "once", true, now))

	task, err := repo.GetTaskByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "first_upload", task.Code)
	assert.True(t, task.Enabled)
}