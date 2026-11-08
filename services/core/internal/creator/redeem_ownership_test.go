package creator

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRedeemCoupon_AppliesToAnyManuscript pins FINDING-059（已修复）：
// RedeemCoupon 现在对稿件 UPDATE 加 `AND user_id = $3`（归属校验）——稿件非本人
// 所有时 RowsAffected==0 → 报错回滚，coupon 不被消耗。此前可给任意稿件加 exposure。
func TestRedeemCoupon_AppliesToAnyManuscript(t *testing.T) {
	svc, mock := newService(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).
			AddRow(100, int64(7), 0, nil)) // coupon owner = 7
	mock.ExpectExec(`UPDATE traffic_coupons SET status = 1`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO traffic_coupon_uses`).WillReturnResult(sqlmock.NewResult(1, 1))
	// 稿件 50 非本人（user_id != 7）→ WHERE id AND user_id 不匹配 → RowsAffected 0。
	mock.ExpectExec(`UPDATE manuscripts SET exposure_weight`).WithArgs(100, int64(50), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	_, err := svc.Redeem(ctx, 7, 1, 50)
	assert.Error(t, err, "FINDING-059 修复：非本人稿件不应加权重")
}

// TestRedeemCoupon_NonexistentManuscript_BurnsCoupon pins FINDING-059（已修复）：
// 稿件不存在时 RowsAffected==0 → 报错回滚，coupon 不被消耗（status 不置 1）。
func TestRedeemCoupon_NonexistentManuscript_BurnsCoupon(t *testing.T) {
	svc, mock := newService(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"amount", "user_id", "status", "expires_at"}).
			AddRow(100, int64(7), 0, nil))
	mock.ExpectExec(`UPDATE traffic_coupons SET status = 1`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO traffic_coupon_uses`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE manuscripts SET exposure_weight`).WithArgs(100, int64(999999), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	_, err := svc.Redeem(ctx, 7, 1, 999999)
	assert.Error(t, err, "FINDING-059 修复：稿件不存在不应消耗 coupon")
	require.NoError(t, mock.ExpectationsWereMet())
}
