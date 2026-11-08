package creator

// repository_errors_test.go — error / boundary paths for Repo.
//
// Strategy notes (empirically verified against go-sqlmock v1.5.2):
//   - sqlmock.New() returns a *sql.DB wired to the mock driver, so
//     BeginTx / ExecContext / QueryContext work through the standard
//     database/sql context plumbing.
//   - A CANCELLED context makes database/sql return the context error
//     ("context canceled") from BeginTx, ExecContext, QueryContext and
//     QueryRowContext before any expectation is matched.
//   - An UNEXPECTED call (no matching expectation) returns a generic
//     "call ... was not expected" error, NOT sql.ErrNoRows.
//   - ExpectQuery(...).WillReturnError(e) is returned verbatim by
//     QueryRowContext().Scan, so we can make GetTaskProgress/GetTaskByID
//     return sql.ErrNoRows (used in service tests) or any sentinel.
//   - ExpectBegin().WillReturnError(e) makes BeginTx return e.

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRepo_NewRepo_DB verifies the DB() accessor (repository.go:68).
func TestRepo_NewRepo_DB(t *testing.T) {
	repo, _ := newRepo(t)
	assert.NotNil(t, repo.DB())
}

// TestRepo_UsableCouponAmount_Error covers repository.go:107-112 error branch.
func TestRepo_UsableCouponAmount_Error(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\), 0\)`).
		WillReturnError(errors.New("db down"))

	_, err := repo.UsableCouponAmount(context.Background(), 7)
	assert.Error(t, err)
}

// TestRepo_UsableCouponAmount_CancelledCtx verifies cancellation propagates.
func TestRepo_UsableCouponAmount_CancelledCtx(t *testing.T) {
	repo, _ := newRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.UsableCouponAmount(ctx, 7)
	assert.ErrorIs(t, err, context.Canceled)
}

// TestRepo_ListCoupons_QueryError covers repository.go:88-90.
func TestRepo_ListCoupons_QueryError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupons WHERE user_id`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.ListCoupons(context.Background(), 7)
	assert.Error(t, err)
}

// TestRepo_ListCoupons_Empty verifies the empty-result path (returns nil, no error).
func TestRepo_ListCoupons_Empty(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupons WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "amount", "reason", "task_id", "status", "granted_at", "used_at", "expires_at"}))

	list, err := repo.ListCoupons(context.Background(), 7)
	require.NoError(t, err)
	assert.Empty(t, list)
}

// TestRepo_ListCouponUses_QueryError covers repository.go:170-172.
func TestRepo_ListCouponUses_QueryError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupon_uses WHERE user_id`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.ListCouponUses(context.Background(), 7)
	assert.Error(t, err)
}

// TestRepo_ListCouponUses_Empty verifies the empty path.
func TestRepo_ListCouponUses_Empty(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupon_uses WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "coupon_id", "user_id", "manuscript_id", "amount", "created_at"}))

	list, err := repo.ListCouponUses(context.Background(), 7)
	require.NoError(t, err)
	assert.Empty(t, list)
}

// TestRepo_ListTasks_QueryError covers repository.go:192-194.
func TestRepo_ListTasks_QueryError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE enabled = true`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.ListTasks(context.Background())
	assert.Error(t, err)
}

// TestRepo_ListTasks_Empty verifies the empty path.
func TestRepo_ListTasks_Empty(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE enabled = true`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}))

	list, err := repo.ListTasks(context.Background())
	require.NoError(t, err)
	assert.Empty(t, list)
}

// TestRepo_GetTaskProgress_NoRows covers repository.go:216-218 (sql.ErrNoRows path).
func TestRepo_GetTaskProgress_NoRows(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`SELECT .+ FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnError(sql.ErrNoRows)

	p, err := repo.GetTaskProgress(context.Background(), 7, 1)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, p)
}

// TestRepo_GetTaskProgress_GenError covers a non-NoRows error.
func TestRepo_GetTaskProgress_GenError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`SELECT .+ FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnError(errors.New("db error"))

	_, err := repo.GetTaskProgress(context.Background(), 7, 1)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, sql.ErrNoRows)
}

// TestRepo_GetTaskByID_NoRows covers repository.go:298-300 (sql.ErrNoRows path).
func TestRepo_GetTaskByID_NoRows(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).
		WillReturnError(sql.ErrNoRows)

	task, err := repo.GetTaskByID(context.Background(), 999)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, task)
}

// TestRepo_GetTaskByID_GenError covers a generic error.
func TestRepo_GetTaskByID_GenError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).
		WillReturnError(errors.New("db error"))

	_, err := repo.GetTaskByID(context.Background(), 1)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, sql.ErrNoRows)
}

// TestRepo_GrantCoupon_QueryError covers repository.go:78-80.
func TestRepo_GrantCoupon_QueryError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`INSERT INTO traffic_coupons`).
		WillReturnError(errors.New("insert failed"))

	_, err := repo.GrantCoupon(context.Background(), &Coupon{UserID: 7, Amount: 100})
	assert.Error(t, err)
}

// ---- RedeemCoupon error branches ----

// TestRepo_RedeemCoupon_BeginError covers repository.go:119-121.
func TestRepo_RedeemCoupon_BeginError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin().WillReturnError(errors.New("cannot begin tx"))

	_, err := repo.RedeemCoupon(context.Background(), 1, 7, 50)
	assert.Error(t, err)
}

// TestRepo_RedeemCoupon_SelectError covers repository.go:131-133 (first SELECT inside tx).
func TestRepo_RedeemCoupon_SelectError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WillReturnError(errors.New("select failed"))
	mock.ExpectRollback()

	_, err := repo.RedeemCoupon(context.Background(), 1, 7, 50)
	assert.Error(t, err)
}

// TestRepo_RedeemCoupon_Expired covers repository.go:140-142.
func TestRepo_RedeemCoupon_Expired(t *testing.T) {
	repo, mock := newRepo(t)
	past := time.Now().Add(-time.Hour)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).
			AddRow(100, int64(7), 0, past))
	mock.ExpectRollback()

	_, err := repo.RedeemCoupon(context.Background(), 1, 7, 50)
	assert.ErrorIs(t, err, errCouponExpired)
}

// TestRepo_RedeemCoupon_UpdateCouponError covers repository.go:145-147.
func TestRepo_RedeemCoupon_UpdateCouponError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).
			AddRow(100, int64(7), 0, nil))
	mock.ExpectExec(`UPDATE traffic_coupons SET status = 1`).
		WillReturnError(errors.New("update failed"))
	mock.ExpectRollback()

	_, err := repo.RedeemCoupon(context.Background(), 1, 7, 50)
	assert.Error(t, err)
}

// TestRepo_RedeemCoupon_InsertUseError covers repository.go:150-152.
func TestRepo_RedeemCoupon_InsertUseError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).
			AddRow(100, int64(7), 0, nil))
	mock.ExpectExec(`UPDATE traffic_coupons SET status = 1`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO traffic_coupon_uses`).
		WillReturnError(errors.New("insert use failed"))
	mock.ExpectRollback()

	_, err := repo.RedeemCoupon(context.Background(), 1, 7, 50)
	assert.Error(t, err)
}

// TestRepo_RedeemCoupon_UpdateManuscriptError covers repository.go:155-157.
func TestRepo_RedeemCoupon_UpdateManuscriptError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).
			AddRow(100, int64(7), 0, nil))
	mock.ExpectExec(`UPDATE traffic_coupons SET status = 1`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO traffic_coupon_uses`).WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec(`UPDATE manuscripts SET exposure_weight`).
		WillReturnError(errors.New("update manuscript failed"))
	mock.ExpectRollback()

	_, err := repo.RedeemCoupon(context.Background(), 1, 7, 50)
	assert.Error(t, err)
}

// TestRepo_RedeemCoupon_CommitError covers repository.go:158-160.
func TestRepo_RedeemCoupon_CommitError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).
			AddRow(100, int64(7), 0, nil))
	mock.ExpectExec(`UPDATE traffic_coupons SET status = 1`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO traffic_coupon_uses`).WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec(`UPDATE manuscripts SET exposure_weight`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	_, err := repo.RedeemCoupon(context.Background(), 1, 7, 50)
	assert.Error(t, err)
}

// TestRepo_RedeemCoupon_CancelledCtx verifies ctx cancellation aborts the tx.
func TestRepo_RedeemCoupon_CancelledCtx(t *testing.T) {
	repo, _ := newRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.RedeemCoupon(ctx, 1, 7, 50)
	assert.ErrorIs(t, err, context.Canceled)
}

// ---- UpsertTaskProgress branches ----

// TestRepo_UpsertTaskProgress_BeginError covers repository.go:225-227.
func TestRepo_UpsertTaskProgress_BeginError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin().WillReturnError(errors.New("cannot begin tx"))

	_, done, err := repo.UpsertTaskProgress(context.Background(), 7, 1, 1, 5, nil)
	assert.Error(t, err)
	assert.False(t, done)
}

// TestRepo_UpsertTaskProgress_QueryError covers repository.go:243-245 (else-if err).
func TestRepo_UpsertTaskProgress_QueryError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnError(errors.New("progress query failed"))
	mock.ExpectRollback()

	_, done, err := repo.UpsertTaskProgress(context.Background(), 7, 1, 1, 5, nil)
	assert.Error(t, err)
	assert.False(t, done)
}

// TestRepo_UpsertTaskProgress_InsertError covers repository.go:238-240.
func TestRepo_UpsertTaskProgress_InsertError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO creator_task_progress`).
		WillReturnError(errors.New("insert progress failed"))
	mock.ExpectRollback()

	_, done, err := repo.UpsertTaskProgress(context.Background(), 7, 1, 1, 5, nil)
	assert.Error(t, err)
	assert.False(t, done)
}

// TestRepo_UpsertTaskProgress_AlreadyClaimed covers repository.go:251-253
// (done && claimed -> commit, no reward).
func TestRepo_UpsertTaskProgress_AlreadyClaimed(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin()
	// existing record: done=true
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnRows(sqlmock.NewRows([]string{"progress", "done"}).AddRow(5, true))
	// claimed=true -> early return (0, false, tx.Commit())
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(true))
	mock.ExpectCommit()

	reward, done, err := repo.UpsertTaskProgress(context.Background(), 7, 1, 1, 5, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, reward)
	assert.False(t, done)
}

// TestRepo_UpsertTaskProgress_NotDone covers repository.go:254-288
// (existing, not done -> update progress, no reward, commit).
func TestRepo_UpsertTaskProgress_NotDone(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnRows(sqlmock.NewRows([]string{"progress", "done"}).AddRow(2, false))
	// claimed check (inside existing branch): claimed=false -> continue
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	// progress update: 2+1=3 < target 5 -> done=false
	mock.ExpectExec(`UPDATE creator_task_progress SET progress`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// final claimed check (done=false, so this is skipped; commit directly)
	mock.ExpectCommit()

	reward, done, err := repo.UpsertTaskProgress(context.Background(), 7, 1, 1, 5, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, reward)
	assert.False(t, done)
}

// TestRepo_UpsertTaskProgress_ClaimedAfterNotDone covers the tail
// "if done { ... if !claimed { mark claimed; commit; return reward } }".
func TestRepo_UpsertTaskProgress_ClaimedAfterNotDone(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin()
	// existing: progress=4, target=5, delta=2 -> current=5 (capped), done=true
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnRows(sqlmock.NewRows([]string{"progress", "done"}).AddRow(4, false))
	// existing-branch claimed check: claimed=false
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	// update progress: current=5 >= target=5 -> done
	mock.ExpectExec(`UPDATE creator_task_progress SET progress`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// tail: done=true, re-check claimed
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	// mark claimed
	mock.ExpectExec(`UPDATE creator_task_progress SET reward_claimed`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	reward, done, err := repo.UpsertTaskProgress(context.Background(), 7, 1, 2, 5, nil)
	require.NoError(t, err)
	assert.True(t, done)
	assert.Equal(t, 5, reward) // returns target
}

// TestRepo_UpsertTaskProgress_UpdateError covers repository.go:261-263.
func TestRepo_UpsertTaskProgress_UpdateError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnRows(sqlmock.NewRows([]string{"progress", "done"}).AddRow(2, false))
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	mock.ExpectExec(`UPDATE creator_task_progress SET progress`).
		WillReturnError(errors.New("update progress failed"))
	mock.ExpectRollback()

	_, done, err := repo.UpsertTaskProgress(context.Background(), 7, 1, 1, 5, nil)
	assert.Error(t, err)
	assert.False(t, done)
}

// TestRepo_UpsertTaskProgress_MarkClaimedError covers repository.go:276-278.
func TestRepo_UpsertTaskProgress_MarkClaimedError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnRows(sqlmock.NewRows([]string{"progress", "done"}).AddRow(4, false))
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	mock.ExpectExec(`UPDATE creator_task_progress SET progress`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	mock.ExpectExec(`UPDATE creator_task_progress SET reward_claimed`).
		WillReturnError(errors.New("mark claimed failed"))
	mock.ExpectRollback()

	_, done, err := repo.UpsertTaskProgress(context.Background(), 7, 1, 2, 5, nil)
	assert.Error(t, err)
	assert.False(t, done)
}

// TestRepo_UpsertTaskProgress_FinalCommitError covers repository.go:285-287.
func TestRepo_UpsertTaskProgress_FinalCommitError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnRows(sqlmock.NewRows([]string{"progress", "done"}).AddRow(2, false))
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	mock.ExpectExec(`UPDATE creator_task_progress SET progress`).WillReturnResult(sqlmock.NewResult(0, 1))
	// final commit fails
	mock.ExpectCommit().WillReturnError(errors.New("final commit failed"))

	_, done, err := repo.UpsertTaskProgress(context.Background(), 7, 1, 1, 5, nil)
	assert.Error(t, err)
	assert.False(t, done)
}

// TestRepo_UpsertTaskProgress_TailCommitError covers repository.go:279-281.
func TestRepo_UpsertTaskProgress_TailCommitError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WillReturnRows(sqlmock.NewRows([]string{"progress", "done"}).AddRow(4, false))
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	mock.ExpectExec(`UPDATE creator_task_progress SET progress`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	mock.ExpectExec(`UPDATE creator_task_progress SET reward_claimed`).WillReturnResult(sqlmock.NewResult(0, 1))
	// tail commit fails
	mock.ExpectCommit().WillReturnError(errors.New("tail commit failed"))

	_, done, err := repo.UpsertTaskProgress(context.Background(), 7, 1, 2, 5, nil)
	assert.Error(t, err)
	assert.False(t, done)
}

// TestRepo_UpsertTaskProgress_CancelledCtx verifies ctx cancellation.
func TestRepo_UpsertTaskProgress_CancelledCtx(t *testing.T) {
	repo, _ := newRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, done, err := repo.UpsertTaskProgress(ctx, 7, 1, 1, 5, nil)
	assert.ErrorIs(t, err, context.Canceled)
	assert.False(t, done)
}

// TestRepo_GrantCoupon_CancelledCtx verifies cancellation.
func TestRepo_GrantCoupon_CancelledCtx(t *testing.T) {
	repo, _ := newRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.GrantCoupon(ctx, &Coupon{UserID: 7, Amount: 100})
	assert.ErrorIs(t, err, context.Canceled)
}
