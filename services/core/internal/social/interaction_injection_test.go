package social

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIncrementManuscriptCount_NoAllowlist pins FINDING-062（已修复）：
// IncrementManuscriptCount/DecrementManuscriptCount/GetManuscriptCount 现在对
// field 做 allowlist 校验（{like_count,coin_count,collect_count,share_count,view_count}），
// 非白名单列名返回 error，不再原样拼进 SQL（防未来调用方转发用户输入 → 完整 SQLi）。
func TestIncrementManuscriptCount_NoAllowlist(t *testing.T) {
	repo, _ := newMockInteractionRepo(t)
	ctx := context.Background()

	err := repo.IncrementManuscriptCount(ctx, "nonsense_col", 1)
	assert.Error(t, err, "FINDING-062 修复：非白名单列名应拒绝")
}

// TestIncrementManuscriptCount_Allowlisted confirms allowlisted fields still work.
func TestIncrementManuscriptCount_Allowlisted(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	ctx := context.Background()
	mock.ExpectExec(`UPDATE manuscripts SET like_count = like_count \+ 1 WHERE id = \$1`).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.IncrementManuscriptCount(ctx, "like_count", 1))
}
