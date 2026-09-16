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