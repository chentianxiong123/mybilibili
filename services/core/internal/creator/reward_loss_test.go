package creator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	myerr "mybilibili/pkg/errors"
)

// TestRecordTaskProgress_RewardLostOnGrantFailure pins a reward-logic bug
// distinct from FINDING-003 (duplicate award). Here the bug is the OPPOSITE:
// a reward is permanently LOST when the coupon grant fails.
//
// ROOT CAUSE: creator/service.go RecordTaskProgress (lines ~158-191) calls
// repo.UpsertTaskProgress FIRST, which commits `reward_claimed = true` inside
// its own transaction. ONLY THEN does it call repo.GrantCoupon in a separate
// transaction. If GrantCoupon fails (e.g. transient DB error), the error is
// returned to the caller but the progress row already has reward_claimed=true.
//
// IMPACT: the user is told the request failed, but on retry UpsertTaskProgress
// sees `done && claimed` and short-circuits (repository.go:251-253), so the
// coupon is never granted. The reward is permanently lost. The "reward claimed"
// flag is set BEFORE the reward is actually granted — the two operations are
// not atomic.
func TestRecordTaskProgress_RewardLostOnGrantFailure(t *testing.T) {
	svc, mock := newService(t)
	ctx := context.Background()
	now := time.Now()

	// GetTaskByID: enabled task, reward_coupon=100, target=1.
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE id = \$1`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", 100, 1, "once", true, now))

	// UpsertTaskProgress: no existing row → INSERT (delta=1 >= target=1 → done=true),
	// then mark reward_claimed=true and COMMIT.
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT progress, done FROM creator_task_progress WHERE user_id = \$1 AND task_id = \$2`).
		WithArgs(int64(7), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"progress", "done"})) // no rows → ErrNoRows
	mock.ExpectExec(`INSERT INTO creator_task_progress`).
		WithArgs(int64(7), int64(1), 1, true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`SELECT reward_claimed FROM creator_task_progress`).WithArgs(int64(7), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"reward_claimed"}).AddRow(false))
	mock.ExpectExec(`UPDATE creator_task_progress SET reward_claimed`).WithArgs(int64(7), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit() // reward_claimed=true is now durable

	// GrantCoupon then FAILS (separate transaction).
	mock.ExpectQuery(`INSERT INTO traffic_coupons`).WillReturnError(errors.New("db down"))

	reward, err := svc.RecordTaskProgress(ctx, 7, 1, 1)
	require.Error(t, err)
	assert.Equal(t, 0, reward)
	assert.ErrorIs(t, err, myerr.ErrInternal("database error"))

	// BUG confirmed: UpsertTaskProgress committed reward_claimed=true BEFORE
	// GrantCoupon. A correct implementation grants the coupon atomically (same tx)
	// or re-checks reward_claimed only after a successful grant.
}
