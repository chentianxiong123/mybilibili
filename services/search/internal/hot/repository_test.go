package hot

import (
	"context"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepository(t *testing.T) (*Repository, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := NewRepository(client)
	t.Cleanup(func() { client.Close() })
	return repo, mr
}

func TestHotIncrement(t *testing.T) {
	repo, mr := newTestRepository(t)

	err := repo.Increment(context.Background(), "golang")
	require.NoError(t, err)
	err = repo.Increment(context.Background(), "golang")
	require.NoError(t, err)

	score, err := mr.ZScore(rankKey, "golang")
	require.NoError(t, err)
	assert.InDelta(t, scoreIncrement()*2, score, 0.001)

	count := mr.HGet(fmt.Sprintf(detailPrefix, "golang"), "count")
	assert.Equal(t, "2", count)
}

func TestHotGetRank(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	require.NoError(t, repo.UpdateScore(ctx, "video", 50))
	require.NoError(t, repo.UpdateScore(ctx, "music", 80))
	require.NoError(t, repo.UpdateScore(ctx, "game", 30))

	top, err := repo.Top(ctx, 10)
	require.NoError(t, err)
	require.Len(t, top, 3)

	assert.Equal(t, 1, top[0]["rank"])
	assert.Equal(t, "music", top[0]["keyword"])
	assert.Equal(t, int64(80), top[0]["score"])

	assert.Equal(t, "video", top[1]["keyword"])
	assert.Equal(t, "game", top[2]["keyword"])

	item, err := repo.Get(ctx, "music")
	require.NoError(t, err)
	assert.Equal(t, "music", item["keyword"])
	assert.Equal(t, int64(80), item["score"])
}

func TestHotCleanExpired(t *testing.T) {
	repo, mr := newTestRepository(t)
	ctx := context.Background()

	// 写入 5 个关键词，分数 1..5
	for i := 0; i < 5; i++ {
		require.NoError(t, repo.UpdateScore(ctx, "kw"+string(rune('1'+i)), float64(i+1)))
	}

	// 只保留 2 个（分数最高的 kw5=5, kw4=4）
	require.NoError(t, repo.CleanExpired(ctx, 2))

	// 最高分关键词（kw5）应保留
	top5, err := mr.ZScore(rankKey, "kw5")
	require.NoError(t, err)
	assert.InDelta(t, 5.0, top5, 0.001)

	// kw4（第二高）也应保留
	top4, err := mr.ZScore(rankKey, "kw4")
	require.NoError(t, err)
	assert.InDelta(t, 4.0, top4, 0.001)

	// 只有 kw4/kw5 保留，其余（kw1/kw2/kw3）应被裁剪
	remaining, err := mr.SortedSet(rankKey)
	require.NoError(t, err)
	assert.Len(t, remaining, 2)
	_, hasHigh := remaining["kw4"]
	_, hasHighest := remaining["kw5"]
	assert.True(t, hasHigh && hasHighest)
	_, hasLow := remaining["kw1"]
	_, hasMid := remaining["kw3"]
	assert.False(t, hasLow || hasMid)
}

func TestScoreIncrementPositive(t *testing.T) {
	assert.Greater(t, scoreIncrement(), 0.0)
}

func TestHotRepository_Top(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	require.NoError(t, repo.UpdateScore(ctx, "alpha", 100))
	require.NoError(t, repo.UpdateScore(ctx, "beta", 200))
	require.NoError(t, repo.UpdateScore(ctx, "gamma", 50))

	top, err := repo.Top(ctx, 2)
	require.NoError(t, err)
	require.Len(t, top, 2)
	assert.Equal(t, "beta", top[0]["keyword"])
	assert.Equal(t, int64(200), top[0]["score"])
	assert.Equal(t, 1, top[0]["rank"])
	assert.Equal(t, "alpha", top[1]["keyword"])
	assert.Equal(t, int64(100), top[1]["score"])
	assert.Equal(t, 2, top[1]["rank"])
}

func TestHotRepository_Get(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	require.NoError(t, repo.Increment(ctx, "golang"))
	require.NoError(t, repo.Increment(ctx, "golang"))

	item, err := repo.Get(ctx, "golang")
	require.NoError(t, err)
	assert.Equal(t, "golang", item["keyword"])
	assert.NotNil(t, item["score"])
	assert.NotNil(t, item["search_count"])
	assert.NotNil(t, item["first_search_time"])
	assert.NotNil(t, item["last_search_time"])

	_, err = repo.Get(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestHotRepository_CleanExpired(t *testing.T) {
	repo, mr := newTestRepository(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		require.NoError(t, repo.UpdateScore(ctx, fmt.Sprintf("kw%d", i), float64(i+1)))
	}

	require.NoError(t, repo.CleanExpired(ctx, 3))

	remaining, err := mr.SortedSet(rankKey)
	require.NoError(t, err)
	assert.Len(t, remaining, 3)

	for _, kw := range []string{"kw2", "kw3", "kw4"} {
		_, exists := remaining[kw]
		assert.True(t, exists, "expected %s to be retained", kw)
	}
	for _, kw := range []string{"kw0", "kw1"} {
		_, exists := remaining[kw]
		assert.False(t, exists, "expected %s to be removed", kw)
	}
}

// 空关键词不得进热度榜：否则 ZIncrBy 会凭空造出一个空串 member，
// 顶到 rank 1，被前端当成关键词渲染出来。
func TestHotIncrementIgnoresEmptyKeyword(t *testing.T) {
	repo, mr := newTestRepository(t)
	ctx := context.Background()

	require.NoError(t, repo.Increment(ctx, ""))
	require.NoError(t, repo.Increment(ctx, "   "))

	assert.False(t, mr.Exists(rankKey))

	// 正常关键词仍可写入
	require.NoError(t, repo.Increment(ctx, "golang"))
	assert.True(t, mr.Exists(rankKey))
}

// Top 必须跳过历史脏数据（空串/纯空白 member），且 rank 从 1 连续编号。
func TestHotTopSkipsBlankMembers(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	require.NoError(t, repo.UpdateScore(ctx, "", 500)) // 脏数据：空关键词却排第一
	require.NoError(t, repo.UpdateScore(ctx, "  ", 400))
	require.NoError(t, repo.UpdateScore(ctx, "music", 80))
	require.NoError(t, repo.UpdateScore(ctx, "game", 30))

	top, err := repo.Top(ctx, 10)
	require.NoError(t, err)
	require.Len(t, top, 2)

	assert.Equal(t, "music", top[0]["keyword"])
	assert.Equal(t, 1, top[0]["rank"])
	assert.Equal(t, "game", top[1]["keyword"])
	assert.Equal(t, 2, top[1]["rank"])
}

// 过滤脏数据后仍要凑满 n 条（多扫 extraScan 条再截断）。
func TestHotTopReturnsUpToNAfterFiltering(t *testing.T) {
	repo, _ := newTestRepository(t)
	ctx := context.Background()

	// 3 条脏数据 + 5 条正常数据
	for i := 0; i < 3; i++ {
		require.NoError(t, repo.UpdateScore(ctx, "", float64(1000+i)))
	}
	for i := 0; i < 5; i++ {
		require.NoError(t, repo.UpdateScore(ctx, fmt.Sprintf("kw%d", i), float64(100+i)))
	}

	top, err := repo.Top(ctx, 4)
	require.NoError(t, err)
	require.Len(t, top, 4)
	for i, item := range top {
		assert.NotEmpty(t, item["keyword"])
		assert.Equal(t, i+1, item["rank"])
	}
}
