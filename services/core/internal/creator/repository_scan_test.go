package creator

// repository_scan_test.go — type-mismatch techniques that force `rows.Scan`
// to fail, covering the 3 remaining uncovered scan-error branches:
//   - ListCoupons  repository.go:96-98
//   - ListCouponUses repository.go:177-179
//   - ListTasks    repository.go:200-202

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// TestRepo_ListCoupons_ScanTypeError covers repository.go:96-98.
// Scanning a string into an int64 destination fails inside rows.Scan.
func TestRepo_ListCoupons_ScanTypeError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupons WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "amount", "reason", "task_id", "status", "granted_at", "used_at", "expires_at"}).
			AddRow(int64(1), "not_an_int64", 100, "任务奖励", int64(0), 0, time.Now(), nil, nil))

	_, err := repo.ListCoupons(context.Background(), 7)
	assert.Error(t, err)
}

// TestRepo_ListCouponUses_ScanTypeError covers repository.go:177-179.
func TestRepo_ListCouponUses_ScanTypeError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`SELECT .+ FROM traffic_coupon_uses WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "coupon_id", "user_id", "manuscript_id", "amount", "created_at"}).
			AddRow(int64(1), int64(1), "bad", int64(50), 100, time.Now()))

	_, err := repo.ListCouponUses(context.Background(), 7)
	assert.Error(t, err)
}

// TestRepo_ListTasks_ScanTypeError covers repository.go:200-202.
func TestRepo_ListTasks_ScanTypeError(t *testing.T) {
	repo, mock := newRepo(t)
	mock.ExpectQuery(`SELECT .+ FROM creator_tasks WHERE enabled = true`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "description", "reward_coupon", "target", "period", "enabled", "created_at"}).
			AddRow(int64(1), "first_upload", "首次投稿", "d", "bad", 1, "once", true, time.Now()))

	_, err := repo.ListTasks(context.Background())
	assert.Error(t, err)
}
