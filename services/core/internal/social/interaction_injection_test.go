package social

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIncrementManuscriptCount_NoAllowlist pins a latent SQL-injection bug in
// interaction_repository.go:IncrementManuscriptCount / DecrementManuscriptCount
// / GetManuscriptCount (lines ~121-144).
//
// ROOT CAUSE: the `field` parameter is concatenated verbatim into the SQL text
// (`UPDATE manuscripts SET `+field+` = `+field+` + 1 ...`) with NO allowlist
// validation against a fixed set of counter columns. Today all call sites pass
// hardcoded literals ("like_count", "coin_count", "collect_count", "share_count"),
// so it is not exploitable through the public API — but any future caller that
// forwards a user-controlled string becomes a full SQL injection. The same
// unsafe pattern exists in shared/pkg/repository/metrics.go:UpsertDailyMetric.
//
// IMPACT: latent injection; also a correctness hazard (a typo'd column name is
// silently turned into a SQL error instead of being rejected early).
func TestIncrementManuscriptCount_NoAllowlist(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	ctx := context.Background()

	// A non-allowlisted identifier passes through verbatim with no validation.
	// A correct implementation whitelists {like_count, coin_count, collect_count,
	// share_count} and rejects anything else.
	mock.ExpectExec(`UPDATE manuscripts SET nonsense_col = nonsense_col \+ 1 WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.IncrementManuscriptCount(ctx, "nonsense_col", 1)
	require.NoError(t, err) // BUG: arbitrary column name accepted

	// GetManuscriptCount reads the same concatenated column.
	mock.ExpectQuery(`SELECT nonsense_col FROM manuscripts WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"nonsense_col"}).AddRow(int32(7)))
	count, err := repo.GetManuscriptCount(ctx, "nonsense_col", 1)
	require.NoError(t, err)
	assert.Equal(t, int32(7), count)
}
