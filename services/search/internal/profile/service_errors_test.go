package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServiceWithFake(t *testing.T) (*Service, *fakeDocumentStore) {
	t.Helper()
	store := newFakeDocumentStore()
	svc := NewService(NewRepository(store))
	return svc, store
}

func TestService_Init_CreateNew(t *testing.T) {
	svc, store := newTestServiceWithFake(t)
	p, err := svc.Init(context.Background(), 42, []string{"game"})
	require.NoError(t, err)
	assert.Equal(t, int64(42), p.UserID)
	assert.Equal(t, []string{"game"}, p.Tags)
	assert.NotEmpty(t, p.ID)
	assert.Equal(t, 1, store.insertCalls)
}

func TestService_Init_ExistingReplacesTags(t *testing.T) {
	svc, store := newTestServiceWithFake(t)
	store.profiles[7] = &Profile{ID: "existing", UserID: 7, Tags: []string{"old"}}

	p, err := svc.Init(context.Background(), 7, []string{"new"})
	require.NoError(t, err)
	assert.Equal(t, "existing", p.ID)
	assert.Equal(t, []string{"new"}, p.Tags)
	assert.Equal(t, 1, store.updateCalls)
	assert.Equal(t, 0, store.insertCalls)
}

func TestService_Init_UpstreamGetByUserErrorStillCreates(t *testing.T) {
	// Init swallows GetByUser errors; it proceeds as if no profile exists.
	svc, store := newTestServiceWithFake(t)
	store.queryFail = true
	store.queryErr = errors.New("query failed")

	p, err := svc.Init(context.Background(), 9, []string{"tag"})
	require.NoError(t, err)
	assert.Equal(t, int64(9), p.UserID)
	assert.Equal(t, []string{"tag"}, p.Tags)
	assert.NotEmpty(t, p.ID)
}

func TestService_Init_UpsertError(t *testing.T) {
	// Force Init's upsert path to fail: the memory/fake insert can fail.
	svc, store := newTestServiceWithFake(t)
	store.insertFail = true
	store.insertErr = errors.New("cannot insert")

	_, err := svc.Init(context.Background(), 42, []string{"a"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot insert")
}

func TestService_GetOrCreate_ReturnsExistingWithoutUpert(t *testing.T) {
	svc, store := newTestServiceWithFake(t)
	store.profiles[11] = &Profile{ID: "existing-11", UserID: 11, WatchCount: 5}

	p, err := svc.GetOrCreate(context.Background(), 11)
	require.NoError(t, err)
	assert.Equal(t, "existing-11", p.ID)
	assert.Equal(t, int64(5), p.WatchCount)
	// No upsert should happen when an existing profile is returned.
	assert.Equal(t, 0, store.updateCalls)
	assert.Equal(t, 0, store.insertCalls)
}

func TestService_GetOrCreate_GetByUserError(t *testing.T) {
	svc, store := newTestServiceWithFake(t)
	store.queryFail = true
	store.queryErr = errors.New("store down")

	p, err := svc.GetOrCreate(context.Background(), 1)
	require.Error(t, err)
	assert.Nil(t, p)
	assert.ErrorContains(t, err, "store down")
}

func TestService_GetOrCreate_CreateNew(t *testing.T) {
	svc, store := newTestServiceWithFake(t)
	p, err := svc.GetOrCreate(context.Background(), 21)
	require.NoError(t, err)
	assert.Equal(t, int64(21), p.UserID)
	assert.NotEmpty(t, p.ID)
	assert.Equal(t, 1, store.insertCalls)
}

func TestService_GetOrCreate_UpertErrorOnCreate(t *testing.T) {
	svc, store := newTestServiceWithFake(t)
	store.insertFail = true
	store.insertErr = errors.New("cannot create")

	p, err := svc.GetOrCreate(context.Background(), 22)
	require.Error(t, err)
	assert.Nil(t, p)
	assert.ErrorContains(t, err, "cannot create")
}

func TestService_RecordWatch_MergesCategoryAndTags(t *testing.T) {
	svc, _ := newTestServiceWithFake(t)
	err := svc.RecordWatch(context.Background(), 1, 5, []string{"e-sports", "game"}, 120)
	require.NoError(t, err)

	p, err := svc.Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.WatchCount)
	assert.Equal(t, []int64{5}, p.FavoriteCategories)
	assert.Equal(t, []string{"e-sports", "game"}, p.Tags)
}

func TestService_RecordWatch_DedupSkipsExisting(t *testing.T) {
	svc, _ := newTestServiceWithFake(t)
	require.NoError(t, svc.RecordWatch(context.Background(), 1, 5, []string{"tag"}, 10))
	require.NoError(t, svc.RecordWatch(context.Background(), 1, 5, []string{"tag", "new-tag"}, 10))

	p, err := svc.Get(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.WatchCount)
	// "tag" must appear only once.
	occurrences := 0
	for _, t := range p.Tags {
		if t == "tag" {
			occurrences++
		}
	}
	assert.Equal(t, 1, occurrences)
	assert.Contains(t, p.Tags, "new-tag")
	assert.Equal(t, []int64{5}, p.FavoriteCategories)
}

func TestService_RecordWatch_EmptyStringTagSkipped(t *testing.T) {
	svc, _ := newTestServiceWithFake(t)
	// "" should be filtered; whitespace "  " should NOT be filtered.
	require.NoError(t, svc.RecordWatch(context.Background(), 2, 0, []string{"", "  ", "ok"}, 5))

	p, err := svc.Get(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, []string{"  ", "ok"}, p.Tags)
	assert.Empty(t, p.FavoriteCategories)
}

func TestService_RecordWatch_GetOrCreateError(t *testing.T) {
	svc, store := newTestServiceWithFake(t)
	store.queryFail = true
	store.queryErr = errors.New("get or create failed")

	err := svc.RecordWatch(context.Background(), 3, 1, []string{"t"}, 1)
	require.Error(t, err)
	assert.ErrorContains(t, err, "get or create failed")
}

func TestService_RecordWatch_UpsertError(t *testing.T) {
	// Pre-create so GetOrCreate succeeds, then force the final upsert to fail.
	svc, store := newTestServiceWithFake(t)
	store.profiles[4] = &Profile{ID: "id-4", UserID: 4}
	store.updateFail = true
	store.updateErr = errors.New("upsert failed")

	err := svc.RecordWatch(context.Background(), 4, 1, []string{"t"}, 1)
	require.Error(t, err)
	assert.ErrorContains(t, err, "upsert failed")
}

func TestService_RecordLike_MergesAndCounts(t *testing.T) {
	svc, _ := newTestServiceWithFake(t)
	require.NoError(t, svc.RecordLike(context.Background(), 5, 9, []string{"music"}))
	require.NoError(t, svc.RecordLike(context.Background(), 5, 9, []string{"music", "new"}))

	p, err := svc.Get(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.LikeCount)
	assert.Equal(t, []int64{9}, p.FavoriteCategories)
	assert.Equal(t, []string{"music", "new"}, p.Tags)
	assert.Equal(t, int64(0), p.WatchCount)
}

func TestService_RecordLike_GetOrCreateError(t *testing.T) {
	svc, store := newTestServiceWithFake(t)
	store.queryFail = true
	store.queryErr = errors.New("boom")

	err := svc.RecordLike(context.Background(), 6, 1, nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "boom")
}

func TestService_RecordLike_UpsertError(t *testing.T) {
	svc, store := newTestServiceWithFake(t)
	store.profiles[7] = &Profile{ID: "id-7", UserID: 7}
	store.updateFail = true
	store.updateErr = errors.New("update boom")

	err := svc.RecordLike(context.Background(), 7, 1, []string{"x"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "update boom")
}

func TestService_RecordCollect_MergesAndCounts(t *testing.T) {
	svc, _ := newTestServiceWithFake(t)
	require.NoError(t, svc.RecordCollect(context.Background(), 8, 2, []string{"video"}))
	require.NoError(t, svc.RecordCollect(context.Background(), 8, 2, []string{"video", "extra"}))

	p, err := svc.Get(context.Background(), 8)
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.CollectCount)
	assert.Equal(t, []int64{2}, p.FavoriteCategories)
	assert.Equal(t, []string{"video", "extra"}, p.Tags)
}

func TestService_RecordCollect_GetOrCreateError(t *testing.T) {
	svc, store := newTestServiceWithFake(t)
	store.queryFail = true
	store.queryErr = errors.New("collect boom")

	err := svc.RecordCollect(context.Background(), 9, 1, nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "collect boom")
}

func TestService_RecordCollect_UpsertError(t *testing.T) {
	svc, store := newTestServiceWithFake(t)
	store.profiles[10] = &Profile{ID: "id-10", UserID: 10}
	store.updateFail = true
	store.updateErr = errors.New("collect upsert boom")

	err := svc.RecordCollect(context.Background(), 10, 1, []string{"y"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "collect upsert boom")
}

func TestService_MergeTags_ZeroCategoryNoOp(t *testing.T) {
	svc, _ := newTestServiceWithFake(t)
	// GetOrCreate with zero category and nil tags creates an empty profile.
	p, err := svc.GetOrCreate(context.Background(), 100)
	require.NoError(t, err)
	svc.mergeTags(p, 0, nil)
	assert.Empty(t, p.FavoriteCategories)
	assert.Empty(t, p.Tags)
}

func TestService_MergeTags_NegativeCategoryNoOp(t *testing.T) {
	svc, _ := newTestServiceWithFake(t)
	p, err := svc.GetOrCreate(context.Background(), 101)
	require.NoError(t, err)
	svc.mergeTags(p, -1, []string{"z"})
	assert.Empty(t, p.FavoriteCategories)
	assert.Equal(t, []string{"z"}, p.Tags)
}

func TestService_MergeTags_DuplicateTagSkipped(t *testing.T) {
	svc, _ := newTestServiceWithFake(t)
	p := &Profile{Tags: []string{"a", "b"}}
	svc.mergeTags(p, 0, []string{"a", "c"})
	assert.Equal(t, []string{"a", "b", "c"}, p.Tags)
}

func TestService_MergeTags_DuplicateCategorySkipped(t *testing.T) {
	svc, _ := newTestServiceWithFake(t)
	p := &Profile{FavoriteCategories: []int64{1, 2}}
	svc.mergeTags(p, 2, nil)
	assert.Equal(t, []int64{1, 2}, p.FavoriteCategories)
}
