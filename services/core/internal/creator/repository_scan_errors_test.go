package creator

// repository_scan_errors_test.go — row-level scan error paths using
// sqlmock's RowError (covers the `rows.Scan` failure branches in
// ListCoupons / ListCouponUses / ListTasks, repository.go:96/177/200).

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRepo_ListCoupons_RowsErr covers repository.go:101 (rows.Err() after loop).
// A close error surfaced via rows.Err() (database/sql returns it after the
// loop completes; it is set by CloseError on the mock Rows).
func TestRepo_ListCoupons_RowsErr(t *testing.T) {
	repo, mock := newRepo(t)
	rows := sqlmock.NewRows([]string{"id", "user_id", "amount", "reason", "task_id", "status", "granted_at", "used_at", "expires_at"})
	rows.CloseError(errors.New("close failed"))
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupons WHERE user_id`).WillReturnRows(rows)

	// Close error is only surfaced via rows.Err() if Close is called while
	// iterating has not completed cleanly. database/sql surfaces a Close
	// error on the final Next (EOF) -> rows.Err() returns it.
	list, err := repo.ListCoupons(context.Background(), 7)
	_ = list
	assert.Error(t, err)
}

// TestRepo_ListCouponUses_RowsErr covers repository.go:182.
func TestRepo_ListCouponUses_RowsErr(t *testing.T) {
	repo, mock := newRepo(t)
	rows := sqlmock.NewRows([]string{"id", "coupon_id", "user_id", "manuscript_id", "amount", "created_at"})
	rows.CloseError(errors.New("close failed"))
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupon_uses WHERE user_id`).WillReturnRows(rows)

	_, err := repo.ListCouponUses(context.Background(), 7)
	assert.Error(t, err)
}

// TestRepo_ListTasks_RowsErr covers repository.go:205.
func TestRepo_ListTasks_RowsErr(t *testing.T) {
	repo, mock := newRepo(t)
	rows := sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"})
	rows.CloseError(errors.New("close failed"))
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE enabled = true`).WillReturnRows(rows)

	_, err := repo.ListTasks(context.Background())
	assert.Error(t, err)
}

// TestRepo_GetTaskProgress_CancelledCtx covers ctx-cancellation path.
func TestRepo_GetTaskProgress_CancelledCtx(t *testing.T) {
	repo, _ := newRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.GetTaskProgress(ctx, 7, 1)
	require.ErrorIs(t, err, context.Canceled)
}

// TestRepo_GetTaskByID_CancelledCtx covers ctx-cancellation path.
func TestRepo_GetTaskByID_CancelledCtx(t *testing.T) {
	repo, _ := newRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.GetTaskByID(ctx, 1)
	require.ErrorIs(t, err, context.Canceled)
}

// TestRepo_ListCouponUses_CancelledCtx covers ctx-cancellation path.
func TestRepo_ListCouponUses_CancelledCtx(t *testing.T) {
	repo, _ := newRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.ListCouponUses(ctx, 7)
	require.ErrorIs(t, err, context.Canceled)
}

// TestRepo_ListTasks_CancelledCtx covers ctx-cancellation path.
func TestRepo_ListTasks_CancelledCtx(t *testing.T) {
	repo, _ := newRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.ListTasks(ctx)
	require.ErrorIs(t, err, context.Canceled)
}
