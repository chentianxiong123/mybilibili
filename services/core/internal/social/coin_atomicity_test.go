package social

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeductCoin_IgnoresRowsAffected pins a coin-balance arithmetic bug in
// interaction_repository.go:DeductCoin (lines ~152-155).
//
// ROOT CAUSE: `UPDATE users SET coin_count = coin_count - 1 WHERE id = $1 AND
// coin_count > 0` guards against negatives, but the caller (InteractionService.
// CoinManuscript, interaction_service.go:96-107) never checks RowsAffected and
// the whole read-check-deduct sequence is NOT transactional. The balance is
// read via GetUserCoinCount, then (in separate statements) the interaction is
// recorded, manuscript coin_count is incremented, and DeductCoin runs. If two
// requests race with coin_count=1, both pass the check, both increment the
// manuscript coin_count, but only one DeductCoin actually affects a row.
//
// IMPACT: manuscript coin_count can be inflated beyond the coins actually
// spent (double-spend under concurrency, with no negative balance). Also
// CoinManuscript ignores req.CoinCount entirely and always deducts exactly 1
// coin, so a client requesting N coins is silently charged 1.
func TestDeductCoin_IgnoresRowsAffected(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	ctx := context.Background()

	// User already has 0 coins (or lost the race): the guarded UPDATE affects 0
	// rows, but the error is not returned.
	mock.ExpectExec(`UPDATE users SET coin_count = coin_count - 1 WHERE id = \$1 AND coin_count > 0`).
		WithArgs(int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 0)) // RowsAffected = 0

	err := repo.DeductCoin(ctx, 7)
	require.NoError(t, err) // BUG: no error despite zero coins deducted

	// BUG confirmed: DeductCoin reports success even when no coin was deducted.
	// A correct implementation returns an error on RowsAffected==0 (or the
	// service must wrap the check+deduct+increment in one transaction).
	assert.NoError(t, err)
}
