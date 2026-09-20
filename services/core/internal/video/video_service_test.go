package video

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewService(NewRepository(db), nil), mock
}

func TestService_GetVideo(t *testing.T) {
	svc, mock := newMockService(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, manuscript_id, video_order, title`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "video_order", "title",
			"play_url_hd", "play_url_sd", "play_url_ld",
			"upload_time", "updated_at", "process_progress",
			"process_stage", "has_subtitle", "has_summary",
			"process_status", "process_error", "source_video_url",
			"duration_seconds", "is_vertical",
		}).AddRow(
			1, 10, 1, "t", "h", "s", "l",
			now, now, 100, "done", 1, 1, 1, "", "src", 120, 0,
		))

	v, err := svc.GetVideo(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), v.ID)
	assert.Equal(t, "t", v.Title)
}

func TestService_ListByManuscript(t *testing.T) {
	svc, mock := newMockService(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, manuscript_id, video_order, title`).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "video_order", "title",
			"play_url_hd", "play_url_sd", "play_url_ld",
			"upload_time", "updated_at", "process_progress",
			"process_stage", "has_subtitle", "has_summary",
			"process_status", "process_error", "source_video_url",
			"duration_seconds", "is_vertical",
		}).
			AddRow(1, 10, 1, "v1", "h1", "s1", "l1", now, now, 100, "done", 1, 1, 1, "", "src1", 120, 0))

	list, err := svc.ListByManuscript(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "v1", list[0].Title)
}

func TestService_ListCategories(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery(`SELECT id, name FROM categories`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "tech"))

	list, err := svc.ListCategories(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "tech", list[0].Name)
}

func TestService_CreateCategory(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("new").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

	require.NoError(t, svc.CreateCategory(context.Background(), "new"))
}

func TestService_UpdateCategory(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectExec(`UPDATE categories SET name`).
		WithArgs("upd", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, svc.UpdateCategory(context.Background(), 1, "upd"))
}

func TestService_DeleteCategory(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectExec(`DELETE FROM categories WHERE id`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, svc.DeleteCategory(context.Background(), 1))
}

func TestService_ListBanners(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery(`SELECT id, title, image_url`).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}).AddRow(1, "b", "http://img", "http://link", 10, 1, 1, 0, "", nil, nil))

	list, err := svc.ListBanners(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "b", list[0].Title)
}

func TestService_ListBannersByCategory(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery(`SELECT id, title, image_url`).
		WithArgs(int32(1), int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}).AddRow(1, "b", "http://img", "http://link", 10, 1, 1, 5, "", nil, nil))

	list, err := svc.ListBannersByCategory(context.Background(), 1, 5)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(5), list[0].CategoryID)
}

func TestService_CreateBanner(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery(`INSERT INTO banner_images`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(20))

	b := &BannerImage{Title: "b", ImageURL: "u", LinkURL: "l", SortOrder: 1, Type: 1}
	require.NoError(t, svc.CreateBanner(context.Background(), b))
}

func TestService_UpdateBanner(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectExec(`UPDATE banner_images SET title`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	b := &BannerImage{Title: "new", ImageURL: "u", LinkURL: "l", SortOrder: 1, Status: 1, Type: 1}
	require.NoError(t, svc.UpdateBanner(context.Background(), 1, b))
}

func TestService_DeleteBanner(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectExec(`DELETE FROM banner_images WHERE id`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, svc.DeleteBanner(context.Background(), 1))
}

func TestService_Statistics(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(view_count\),0\) FROM manuscripts`).WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(500))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE status`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	stats, err := svc.Statistics(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(10), stats["manuscript_count"])
	assert.Equal(t, int64(5), stats["user_count"])
	assert.Equal(t, int64(500), stats["view_count"])
	assert.Equal(t, int64(2), stats["pending_count"])
}

func TestService_ListUserManuscriptIDs(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery(`SELECT id FROM manuscripts WHERE user_id`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10).AddRow(20))

	ids, err := svc.ListUserManuscriptIDs(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, []int64{10, 20}, ids)
}

func TestService_ListUserVideoIDs(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery(`SELECT v.id FROM videos v JOIN manuscripts m`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(3))

	ids, err := svc.ListUserVideoIDs(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 3}, ids)
}

func TestService_BatchDeleteVideos(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectExec(`DELETE FROM videos WHERE id = ANY`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 2))

	require.NoError(t, svc.BatchDeleteVideos(context.Background(), []int64{1, 2}))
}

func TestService_BatchDeleteVideos_Empty(t *testing.T) {
	svc, _ := newMockService(t)

	require.NoError(t, svc.BatchDeleteVideos(context.Background(), []int64{}))
	require.NoError(t, svc.BatchDeleteVideos(context.Background(), nil))
}
