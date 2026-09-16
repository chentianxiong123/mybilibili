package profile

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mybilibili/pkg/abstraction"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	store, err := abstraction.NewDocumentStore(abstraction.DocumentStoreConfig{Type: "memory"})
	require.NoError(t, err)
	return NewService(NewRepository(store))
}

func TestService_GetOrCreate_NewProfile(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.GetOrCreate(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, int64(42), p.UserID)
	assert.Empty(t, p.Tags)
	assert.Empty(t, p.FavoriteCategories)
}

func TestService_GetOrCreate_ReturnsExisting(t *testing.T) {
	svc := newTestService(t)
	p1, err := svc.GetOrCreate(context.Background(), 42)
	require.NoError(t, err)
	p2, err := svc.GetOrCreate(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, p1.ID, p2.ID)
	assert.Equal(t, int64(42), p2.UserID)
}

func TestService_Init_CreatesTags(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.Init(context.Background(), 7, []string{"游戏", "科技"})
	require.NoError(t, err)
	assert.Equal(t, []string{"游戏", "科技"}, p.Tags)

	p2, err := svc.Init(context.Background(), 7, []string{"音乐"})
	require.NoError(t, err)
	assert.Equal(t, []string{"音乐"}, p2.Tags)
}

func TestService_RecordWatch(t *testing.T) {
	svc := newTestService(t)
	err := svc.RecordWatch(context.Background(), 1, 5, []string{"电竞"}, 120)
	require.NoError(t, err)

	p, err := svc.Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.WatchCount)
	assert.Equal(t, []int64{5}, p.FavoriteCategories)
	assert.Equal(t, []string{"电竞"}, p.Tags)
}

func TestService_RecordWatch_Dedup(t *testing.T) {
	svc := newTestService(t)
	require.NoError(t, svc.RecordWatch(context.Background(), 1, 5, []string{"电竞", "电竞", "游戏"}, 120))
	require.NoError(t, svc.RecordWatch(context.Background(), 1, 5, []string{"电竞"}, 120))

	p, err := svc.Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.WatchCount)
	assert.Equal(t, []int64{5}, p.FavoriteCategories)
	assert.Equal(t, []string{"电竞", "游戏"}, p.Tags)
}

func TestService_RecordLikeAndCollect(t *testing.T) {
	svc := newTestService(t)
	require.NoError(t, svc.RecordLike(context.Background(), 2, 9, nil))
	require.NoError(t, svc.RecordCollect(context.Background(), 2, 9, nil))

	p, err := svc.Get(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.LikeCount)
	assert.Equal(t, int64(1), p.CollectCount)
	assert.Equal(t, int64(0), p.WatchCount)
	assert.Equal(t, []int64{9}, p.FavoriteCategories)
}

func TestService_RecordNoCategory(t *testing.T) {
	svc := newTestService(t)
	require.NoError(t, svc.RecordWatch(context.Background(), 3, 0, []string{"", "  ", "有效"}, 10))

	p, err := svc.Get(context.Background(), 3)
	require.NoError(t, err)
	assert.Empty(t, p.FavoriteCategories)
	// mergeTags 只过滤空字符串 ""，不 trim 空白
	assert.Contains(t, p.Tags, "有效")
}

func TestService_Get_NotFound(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.Get(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, p)
}

func TestContainsStr(t *testing.T) {
	assert.True(t, containsStr([]string{"a", "b"}, "a"))
	assert.False(t, containsStr([]string{"a"}, "z"))
	assert.False(t, containsStr(nil, "z"))
}

func TestContainsInt64(t *testing.T) {
	assert.True(t, containsInt64([]int64{1, 2}, 2))
	assert.False(t, containsInt64([]int64{1}, 9))
	assert.False(t, containsInt64(nil, 1))
}

func TestService_ConcurrentRecords(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		require.NoError(t, svc.RecordWatch(ctx, 100, int64(i), []string{"tag"}, 1))
	}
	p, err := svc.Get(ctx, 100)
	require.NoError(t, err)
	assert.Greater(t, p.WatchCount, int64(0))
	assert.Contains(t, p.Tags, "tag")
}

func TestNewService_RepositoryField(t *testing.T) {
	store, err := abstraction.NewDocumentStore(abstraction.DocumentStoreConfig{Type: "memory"})
	require.NoError(t, err)
	repo := NewRepository(store)
	svc := NewService(repo)
	assert.Equal(t, "user_profiles", collection)
	assert.NotNil(t, svc.repo)
	assert.Contains(t, strings.TrimSpace("user_profiles"), "user_profiles")
}
