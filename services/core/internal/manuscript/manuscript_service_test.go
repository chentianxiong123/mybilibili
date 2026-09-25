package manuscript

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mybilibili/core/internal/user"
	pb "mybilibili/pkg/pb"
)

func newService(t *testing.T) (*ManuscriptService, sqlmock.Sqlmock, sqlmock.Sqlmock) {
	t.Helper()
	msDB, msMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { msDB.Close() })

	userDB, userMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { userDB.Close() })

	svc := NewManuscriptService(NewManuscriptRepository(msDB), user.NewRepository(userDB))
	return svc, msMock, userMock
}

func TestGetManuscript_Success(t *testing.T) {
	svc, msMock, _ := newService(t)
	ctx := context.Background()
	now := time.Now()

	msMock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "title1", "desc1", "cover.jpg", int64(10), int64(1),
				int64(100), int64(20), int64(5), int64(3), int64(1),
				int64(0), int64(0), int32(3), int32(1), "ok",
				now, int64(99), now, now, "03:00", int32(180), "local"))

	msMock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "技术"))

	resp, err := svc.GetManuscript(ctx, &pb.GetManuscriptRequest{Id: 1})
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Manuscript.Id)
	assert.Equal(t, "技术", resp.Manuscript.CategoryName)
}

func TestGetManuscript_NotFound(t *testing.T) {
	svc, msMock, _ := newService(t)
	ctx := context.Background()

	msMock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols))

	_, err := svc.GetManuscript(ctx, &pb.GetManuscriptRequest{Id: 999})
	assert.Error(t, err)
}

func TestGetManuscriptWithVideos(t *testing.T) {
	svc, msMock, _ := newService(t)
	ctx := context.Background()
	now := time.Now()

	msMock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				now, int64(0), now, now, "00:00", int32(0), "local"))

	msMock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "cat"))

	msMock.ExpectQuery(`SELECT .+ FROM videos WHERE manuscript_id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "video_order", "title",
			"play_url_hd", "play_url_sd", "play_url_ld",
			"upload_time", "updated_at", "process_progress", "process_stage",
			"has_subtitle", "has_summary", "process_status", "process_error",
			"source_video_url", "duration_seconds", "is_vertical",
		}).AddRow(100, int64(1), int32(0), "P1", "hd.mp4", "sd.mp4", "ld.mp4",
			now, now, int32(100), "DONE", int32(1), int32(1), int32(5), "",
			"src.mp4", int32(60), int32(0)))

	msMock.ExpectQuery(`SELECT t.name FROM tags t`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("go"))

	resp, err := svc.GetManuscriptWithVideos(ctx, &pb.GetManuscriptWithVideosRequest{Id: 1, CurrentUserId: 0})
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Manuscript.Id)
	assert.Len(t, resp.Manuscript.Videos, 1)
}

func TestSvcListUserManuscripts(t *testing.T) {
	svc, msMock, _ := newService(t)
	ctx := context.Background()
	now := time.Now()

	msMock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1`).WithArgs(int64(10), int32(3)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	msMock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id`).WithArgs(int64(10), int32(3), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				now, int64(0), now, now, "", int32(0), "local"))

	msMock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))

	msMock.ExpectQuery(`SELECT is_vertical FROM videos`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"is_vertical"}).AddRow(0))

	resp, err := svc.ListUserManuscripts(ctx, &pb.ListUserManuscriptsRequest{
		UserId: 10, Status: 3, Page: 1, PageSize: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Total)
}

func TestSvcListRecommended(t *testing.T) {
	svc, msMock, userMock := newService(t)
	ctx := context.Background()
	now := time.Now()

	msMock.ExpectQuery(`WITH _seed AS .+ SELECT .+ FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				now, int64(0), now, now, "", int32(0), "local"))

	msMock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))

	userMock.ExpectQuery(`SELECT .+ FROM users WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nickname", "avatar", "level", "bio", "signature", "follower_count", "following_count", "liked_count"}).
			AddRow(int64(1), "u", "", 1, "", "", 0, 0, 0))

	msMock.ExpectQuery(`SELECT is_vertical FROM videos`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"is_vertical"}).AddRow(0))

	resp, err := svc.ListRecommended(ctx, &pb.ListRecommendedRequest{UserId: 1}, 0)
	require.NoError(t, err)
	assert.Len(t, resp.Manuscripts, 1)
}

func TestSvcListHot(t *testing.T) {
	svc, msMock, userMock := newService(t)
	ctx := context.Background()
	now := time.Now()

	msMock.ExpectQuery(`WITH _seed AS .+ SELECT .+ FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				now, int64(0), now, now, "", int32(0), "local"))

	msMock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))

	userMock.ExpectQuery(`SELECT .+ FROM users WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nickname", "avatar", "level", "bio", "signature", "follower_count", "following_count", "liked_count"}).
			AddRow(int64(1), "u", "", 1, "", "", 0, 0, 0))

	msMock.ExpectQuery(`SELECT is_vertical FROM videos`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"is_vertical"}).AddRow(0))

	resp, err := svc.ListHot(ctx, &pb.ListHotRequest{UserId: 1}, 0, 0)
	require.NoError(t, err)
	assert.Len(t, resp.Manuscripts, 1)
}

func TestSvcListByCategory(t *testing.T) {
	svc, msMock, _ := newService(t)
	ctx := context.Background()
	now := time.Now()

	msMock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE category_id = \$1 AND status = 3`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	msMock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE category_id = .+ AND status = 3`).
		WithArgs(int64(5), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(5),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				now, int64(0), now, now, "", int32(0), "local"))

	msMock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(5), "tech"))

	msMock.ExpectQuery(`SELECT is_vertical FROM videos`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"is_vertical"}).AddRow(0))

	resp, err := svc.ListByCategory(ctx, &pb.ListByCategoryRequest{CategoryId: 5, Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Total)
}

func TestSvcListCategories(t *testing.T) {
	svc, msMock, _ := newService(t)
	ctx := context.Background()

	msMock.ExpectQuery(`SELECT id, name FROM categories`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(int64(1), "生活").AddRow(int64(2), "技术"))

	resp, err := svc.ListCategories(ctx, &pb.ListCategoriesRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Categories, 2)
}

func TestSvcDeleteManuscript(t *testing.T) {
	svc, msMock, _ := newService(t)
	ctx := context.Background()

	msMock.ExpectExec(`DELETE FROM manuscripts WHERE id`).WithArgs(int64(1), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := svc.DeleteManuscript(ctx, &pb.DeleteManuscriptRequest{Id: 1, UserId: 10})
	require.NoError(t, err)
}

func TestSvcPublishManuscript(t *testing.T) {
	svc, msMock, _ := newService(t)
	ctx := context.Background()

	msMock.ExpectExec(`UPDATE manuscripts SET status`).WithArgs(int32(3), int64(1), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := svc.PublishManuscript(ctx, &pb.PublishManuscriptRequest{Id: 1, UserId: 10})
	require.NoError(t, err)
}

func TestSvcPublishManuscript_NotFound(t *testing.T) {
	svc, msMock, _ := newService(t)
	ctx := context.Background()

	msMock.ExpectExec(`UPDATE manuscripts SET status`).WithArgs(int32(3), int64(999), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	_, err := svc.PublishManuscript(ctx, &pb.PublishManuscriptRequest{Id: 999, UserId: 10})
	assert.Error(t, err)
}

func TestSvcUnpublishManuscript(t *testing.T) {
	svc, msMock, _ := newService(t)
	ctx := context.Background()

	msMock.ExpectExec(`UPDATE manuscripts SET status`).WithArgs(int32(-1), int64(1), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := svc.UnpublishManuscript(ctx, &pb.UnpublishManuscriptRequest{Id: 1, UserId: 10})
	require.NoError(t, err)
}

func TestSvcUnpublishManuscript_NotFound(t *testing.T) {
	svc, msMock, _ := newService(t)
	ctx := context.Background()

	msMock.ExpectExec(`UPDATE manuscripts SET status`).WithArgs(int32(-1), int64(999), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	_, err := svc.UnpublishManuscript(ctx, &pb.UnpublishManuscriptRequest{Id: 999, UserId: 10})
	assert.Error(t, err)
}

func TestSvcSearchUserManuscripts(t *testing.T) {
	svc, msMock, _ := newService(t)
	ctx := context.Background()
	now := time.Now()

	msMock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id = .+ AND title ILIKE .+`).
		WithArgs(int64(10), "%test%").
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "test", "d", "c.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				now, int64(0), now, now, "", int32(0), "local"))

	msMock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))

	msMock.ExpectQuery(`SELECT is_vertical FROM videos`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"is_vertical"}).AddRow(0))

	resp, err := svc.SearchUserManuscripts(ctx, &pb.SearchUserManuscriptsRequest{UserId: 10, Keyword: "test"})
	require.NoError(t, err)
	assert.Len(t, resp.Manuscripts, 1)
}

func TestSvcUnique(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, unique([]string{"a", "b", "a", "c", "b"}))
	assert.Nil(t, unique(nil))
	assert.Equal(t, []string{"x"}, unique([]string{"x"}))
}

func TestGetManuscriptWithVideos_CacheStore(t *testing.T) {
	svc, msMock, _ := newService(t)
	ctx := context.Background()
	now := time.Now()

	svc.SetCacheStore(&fakeCacheStore{})

	msMock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				now, int64(0), now, now, "", int32(0), "local"))

	msMock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))

	msMock.ExpectExec(`UPDATE manuscripts SET view_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	msMock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := svc.GetManuscriptWithVideos(ctx, &pb.GetManuscriptWithVideosRequest{Id: 1, CurrentUserId: 42})
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

type fakeCacheStore struct{}

func (f *fakeCacheStore) Get(_ context.Context, _ string) ([]byte, error)  { return nil, nil }
func (f *fakeCacheStore) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error {
	return nil
}
func (f *fakeCacheStore) Delete(_ context.Context, _ string) error              { return nil }
func (f *fakeCacheStore) Exists(_ context.Context, _ string) (bool, error)      { return false, nil }
func (f *fakeCacheStore) Lock(_ context.Context, _ string, _ time.Duration) (bool, error) {
	return true, nil
}
func (f *fakeCacheStore) Unlock(_ context.Context, _ string) error              { return nil }
func (f *fakeCacheStore) Incr(_ context.Context, _ string, _ time.Duration) (int64, error) {
	return 1, nil
}
func (f *fakeCacheStore) Close() error { return nil }
