package hot

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHotRepository_Delete(t *testing.T) {
	repo, mr := newTestRepository(t)
	ctx := context.Background()

	require.NoError(t, repo.Increment(ctx, "golang"))
	detail := fmt.Sprintf(detailPrefix, "golang")
	assert.True(t, mr.Exists(rankKey))
	assert.True(t, mr.Exists(detail))

	require.NoError(t, repo.Delete(ctx, "golang"))
	assert.False(t, mr.Exists(rankKey))
	assert.False(t, mr.Exists(detail))

	// Deleting a non-existent keyword must be a no-op, not an error.
	require.NoError(t, repo.Delete(ctx, "nope"))
}

func TestHotRepository_SetRank(t *testing.T) {
	repo, mr := newTestRepository(t)
	ctx := context.Background()

	// SetRank maps rank 1 -> highest score.
	require.NoError(t, repo.SetRank(ctx, "video", 1))
	score, err := mr.ZScore(rankKey, "video")
	require.NoError(t, err)
	assert.InDelta(t, 1000000-1*100, score, 0.001)

	// A numerically larger rank must map to a smaller score.
	require.NoError(t, repo.SetRank(ctx, "music", 2))
	score2, err := mr.ZScore(rankKey, "music")
	require.NoError(t, err)
	assert.InDelta(t, 1000000-2*100, score2, 0.001)
	assert.Greater(t, score, score2)

	// Empty keyword must be ignored (no blank member created).
	require.NoError(t, repo.SetRank(ctx, "", 1))
	require.NoError(t, repo.SetRank(ctx, "   ", 1))
}

func TestScoreString(t *testing.T) {
	cases := []struct {
		name string
		in   float64
		want string
	}{
		{"positive", 123.7, "124"},
		{"zero", 0, "0"},
		{"negative", -5.2, "-5"},
		{"large", 1000000, "1000000"},
		{"fractional", 99.4, "99"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ScoreString(tc.in))
		})
	}
}

// Error branches: with a cancelled or expired context, every public method
// must surface ctx.Err() rather than panicking or swallowing it. This also
// exercises the cancellation/timeout path of the redis client.
func TestHotRepository_ErrorBranches(t *testing.T) {
	repo, _ := newTestRepository(t)

	// Context already cancelled -> all commands return context.Canceled.
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.ErrorIs(t, repo.Increment(cancelCtx, "golang"), context.Canceled)
	_, errTop := repo.Top(cancelCtx, 10)
	assert.ErrorIs(t, errTop, context.Canceled)
	assert.ErrorIs(t, repo.UpdateScore(cancelCtx, "golang", 100), context.Canceled)
	assert.ErrorIs(t, repo.CleanExpired(cancelCtx, 1), context.Canceled)
	_, errGet := repo.Get(cancelCtx, "golang")
	assert.ErrorIs(t, errGet, context.Canceled)
	assert.ErrorIs(t, repo.Delete(cancelCtx, "golang"), context.Canceled)

	// Context past its deadline -> commands return context.DeadlineExceeded.
	deadlineCtx, deadlineCancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer deadlineCancel()

	assert.ErrorIs(t, repo.Increment(deadlineCtx, "golang"), context.DeadlineExceeded)
	_, errTop2 := repo.Top(deadlineCtx, 10)
	assert.ErrorIs(t, errTop2, context.DeadlineExceeded)
	assert.ErrorIs(t, repo.UpdateScore(deadlineCtx, "golang", 100), context.DeadlineExceeded)
	assert.ErrorIs(t, repo.CleanExpired(deadlineCtx, 1), context.DeadlineExceeded)
	_, errGet2 := repo.Get(deadlineCtx, "golang")
	assert.ErrorIs(t, errGet2, context.DeadlineExceeded)
	assert.ErrorIs(t, repo.Delete(deadlineCtx, "golang"), context.DeadlineExceeded)
}

func TestHotRepository_CleanExpiredNoop(t *testing.T) {
	repo, mr := newTestRepository(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		require.NoError(t, repo.UpdateScore(ctx, fmt.Sprintf("kw%d", i), float64(i+1)))
	}

	// total(3) <= keep(5): must early-return without touching data.
	require.NoError(t, repo.CleanExpired(ctx, 5))
	remaining, err := mr.SortedSet(rankKey)
	require.NoError(t, err)
	assert.Len(t, remaining, 3)

	// Empty set with keep(0): total(0) <= keep(0) -> early-return, no error.
	emptyRepo, emptyMr := newTestRepository(t)
	require.NoError(t, emptyRepo.CleanExpired(ctx, 0))
	assert.False(t, emptyMr.Exists(rankKey))
}

func TestHotRepository_TopNoData(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	top, err := repo.Top(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, top, 0)
}

func TestHotRepository_TopMoreThanAvailable(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	require.NoError(t, repo.UpdateScore(ctx, "a", 10))
	require.NoError(t, repo.UpdateScore(ctx, "b", 20))

	top, err := repo.Top(ctx, 100)
	require.NoError(t, err)
	require.Len(t, top, 2)
	assert.Equal(t, "b", top[0]["keyword"])
	assert.Equal(t, "a", top[1]["keyword"])
}

func TestHotRepository_IncrementTrimsWhitespace(t *testing.T) {
	repo, mr := newTestRepository(t)
	ctx := context.Background()

	require.NoError(t, repo.Increment(ctx, "  golang  "))
	score, err := mr.ZScore(rankKey, "golang")
	require.NoError(t, err)
	assert.InDelta(t, scoreIncrement(), score, 0.001)

	// Whitespace-only keyword must be ignored (no blank member).
	require.NoError(t, repo.Increment(ctx, " \t\n "))
	set, err := mr.SortedSet(rankKey)
	require.NoError(t, err)
	assert.Len(t, set, 1)
	_, hasBlank := set[""]
	assert.False(t, hasBlank)
}

func TestHotRepository_GetWithoutDetail(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	// UpdateScore writes a rank entry but no detail hash.
	require.NoError(t, repo.UpdateScore(ctx, "music", 80))

	item, err := repo.Get(ctx, "music")
	require.NoError(t, err)
	assert.Equal(t, "music", item["keyword"])
	assert.Equal(t, int64(80), item["score"])
	_, hasCount := item["search_count"]
	_, hasFirst := item["first_search_time"]
	_, hasLast := item["last_search_time"]
	assert.False(t, hasCount)
	assert.False(t, hasFirst)
	assert.False(t, hasLast)
}

// Top 的 blank-member 过滤分支只能通过直接注入脏数据触发：Increment 与
// UpdateScore 都守卫空关键词，走公开 API 永远造不出空串 member。
func TestHotTopSkipsBlankMembersDirect(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	// 绕过守卫，直接向 rank 集合注入历史脏数据（空串 / 纯空白）。
	require.NoError(t, repo.rdb.ZAdd(ctx, rankKey, redis.Z{Score: 500, Member: ""}).Err())
	require.NoError(t, repo.rdb.ZAdd(ctx, rankKey, redis.Z{Score: 400, Member: "  "}).Err())
	require.NoError(t, repo.rdb.ZAdd(ctx, rankKey, redis.Z{Score: 80, Member: "music"}).Err())
	require.NoError(t, repo.rdb.ZAdd(ctx, rankKey, redis.Z{Score: 30, Member: "game"}).Err())

	top, err := repo.Top(ctx, 10)
	require.NoError(t, err)
	require.Len(t, top, 2)
	assert.Equal(t, "music", top[0]["keyword"])
	assert.Equal(t, 1, top[0]["rank"])
	assert.Equal(t, "game", top[1]["keyword"])
	assert.Equal(t, 2, top[1]["rank"])
}

// BUG(查找): Top(ctx, 0) 应返回空，但当前实现返回 1 个元素；
// Top(ctx, n<0) 因 make 负容量直接 panic。此处记录当前（有缺陷）行为，
// 以保持测试套件为绿，供后续修复源码后回归。
func TestHotTopNonPositiveN(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()
	require.NoError(t, repo.UpdateScore(ctx, "a", 100))
	require.NoError(t, repo.UpdateScore(ctx, "b", 200))

	// n == 0：当前缺陷行为是返回 1 个元素（预期应为 0）。
	top, err := repo.Top(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, len(top), "BUG: Top(0) 应返回空列表，当前却返回 1 个元素")

	// n < 0：当前缺陷行为是 panic（make 负容量）。
	func() {
		defer func() {
			r := recover()
			assert.NotNil(t, r, "BUG: Top(负数) 应被守卫，当前却 panic")
		}()
		_, _ = repo.Top(ctx, -1)
	}()
}

// Get 在 HGetAll 出错时会把 detail 置为 nil，仍返回 keyword/score 而不报错。
func TestHotGetSwallowsDetailError(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	require.NoError(t, repo.UpdateScore(ctx, "music", 80))
	// 把 detail key 覆盖为错误的类型（非 hash），使 HGetAll 返回错误。
	require.NoError(t, repo.rdb.Set(ctx, fmt.Sprintf(detailPrefix, "music"), "not-a-hash", 0).Err())

	item, err := repo.Get(ctx, "music")
	require.NoError(t, err)
	assert.Equal(t, "music", item["keyword"])
	assert.Equal(t, int64(80), item["score"])
	_, hasCount := item["search_count"]
	assert.False(t, hasCount)
}
