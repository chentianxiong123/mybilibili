package creator

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRedeemCoupon_AppliesToAnyManuscript pins a correctness bug in
// creator/repository.go:RedeemCoupon (lines ~144-157).
//
// ROOT CAUSE: RedeemCoupon validates that the COUPON belongs to userID
// (owner != userID → sql.ErrNoRows) but never validates that the MANUSCRIPT
// belongs to userID. The final
//   UPDATE manuscripts SET exposure_weight = exposure_weight + $1 WHERE id = $2
// runs against any manuscript id, so a creator can boost ANYONE ELSE's
// manuscript (cross-user exposure hijack) — and also their own.
//
// IMPACT: exposure_weight is a ranking/boosting signal in ListRecommended.
// A user can buy traffic exposure for a competitor's (or any) manuscript,
// corrupting recommendation ranking. No authorization/ownership gate.
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
	// Manuscript 50 is NOT owned by user 7, but the code applies the exposure anyway.
	// (A correct implementation would first SELECT user_id FROM manuscripts WHERE id=$1
	// and reject when it != userID; no such query is issued here.)
	mock.ExpectExec(`UPDATE manuscripts SET exposure_weight`).WithArgs(100, int64(50)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	amount, err := svc.Redeem(ctx, 7, 1, 50)
	require.NoError(t, err)
	assert.Equal(t, 100, amount)

	// BUG confirmed: no manuscript-ownership SELECT was ever issued.
	// A correct fix adds an ownership gate, which would make the mock fail here
	// with an unexpected query.
}

// TestRedeemCoupon_NonexistentManuscript_BurnsCoupon pins a second correctness
// bug in the same function: the manuscript UPDATE does not check RowsAffected.
//
// ROOT CAUSE: `UPDATE manuscripts SET exposure_weight = exposure_weight + $1
// WHERE id = $2` returns RowsAffected==0 when the manuscript id does not exist,
// but the result is discarded. The coupon is still marked used (status=1) and a
// traffic_coupon_uses row is written, then the tx commits.
//
// IMPACT: a coupon is silently burned (never recoverable) with zero effect —
// the creator loses the exposure points permanently without any error.
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
	// Manuscript 999999 does not exist → RowsAffected == 0, but the code ignores it.
	mock.ExpectExec(`UPDATE manuscripts SET exposure_weight`).WithArgs(100, int64(999999)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	amount, err := svc.Redeem(ctx, 7, 1, 999999)
	require.NoError(t, err)
	assert.Equal(t, 100, amount) // BUG: success despite zero manuscripts updated
}
