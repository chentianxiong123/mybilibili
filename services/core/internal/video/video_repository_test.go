package video

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockRepo(t *testing.T) (*Repository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewRepository(db), mock
}

func TestGetVideoByID(t *testing.T) {
	repo, mock := newMockRepo(t)
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
			1, 10, 1, "title",
			"http://hd", "http://sd", "http://ld",
			now, now, 100,
			"done", 1, 1,
			1, "", "http://src",
			120, 0,
		))

	v, err := repo.GetVideoByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), v.ID)
	assert.Equal(t, int64(10), v.ManuscriptID)
	assert.Equal(t, int32(1), v.VideoOrder)
	assert.Equal(t, "title", v.Title)
	assert.Equal(t, "done", v.ProcessStage)
	assert.Equal(t, int32(120), v.DurationSeconds)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetVideoByID_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT id, manuscript_id, video_order, title`).WillReturnError(sql.ErrNoRows)
	_, err := repo.GetVideoByID(context.Background(), 99)
	assert.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestListByManuscript(t *testing.T) {
	repo, mock := newMockRepo(t)
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
			AddRow(1, 10, 1, "v1", "h1", "s1", "l1", now, now, 100, "done", 1, 1, 1, "", "src1", 120, 0).
			AddRow(2, 10, 2, "v2", "h2", "s2", "l2", now, now, 100, "done", 1, 1, 1, "", "src2", 60, 1))

	list, err := repo.ListByManuscript(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "v1", list[0].Title)
	assert.Equal(t, int32(2), list[1].VideoOrder)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListByManuscript_Empty(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id, manuscript_id, video_order, title`).
		WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "video_order", "title",
			"play_url_hd", "play_url_sd", "play_url_ld",
			"upload_time", "updated_at", "process_progress",
			"process_stage", "has_subtitle", "has_summary",
			"process_status", "process_error", "source_video_url",
			"duration_seconds", "is_vertical",
		}))

	list, err := repo.ListByManuscript(context.Background(), 999)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestListUserManuscriptIDs(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id FROM manuscripts WHERE user_id`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10).AddRow(20))

	ids, err := repo.ListUserManuscriptIDs(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, []int64{10, 20}, ids)
}

func TestListUserManuscriptIDs_Empty(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id FROM manuscripts WHERE user_id`).
		WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	ids, err := repo.ListUserManuscriptIDs(context.Background(), 999)
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestListUserVideoIDs(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT v.id FROM videos v JOIN manuscripts m`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(3))

	ids, err := repo.ListUserVideoIDs(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 3}, ids)
}

func TestListUserVideoIDs_Empty(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT v.id FROM videos v JOIN manuscripts m`).
		WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	ids, err := repo.ListUserVideoIDs(context.Background(), 999)
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestBatchDeleteVideos(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectExec(`DELETE FROM videos WHERE id = ANY`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 2))

	require.NoError(t, repo.BatchDeleteVideos(context.Background(), []int64{1, 2}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBatchDeleteVideos_Empty(t *testing.T) {
	repo, _ := newMockRepo(t)

	require.NoError(t, repo.BatchDeleteVideos(context.Background(), []int64{}))
	require.NoError(t, repo.BatchDeleteVideos(context.Background(), nil))
}

func TestListBanners(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, title, image_url`).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}).
			AddRow(1, "banner1", "http://img1", "http://link1", 10, 1, 1, 0, "", nil, nil).
			AddRow(2, "banner2", "http://img2", "http://link2", 20, 1, 1, 5, "morning", now, nil))

	list, err := repo.ListBanners(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "banner1", list[0].Title)
	assert.Nil(t, list[0].StartTime)
	assert.NotNil(t, list[1].StartTime)
	assert.Equal(t, int64(5), list[1].CategoryID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListCategories(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id, name FROM categories`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "tech").
			AddRow(2, "music"))

	list, err := repo.ListCategories(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "tech", list[0].Name)
	assert.Equal(t, int64(2), list[1].ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCategoryByID(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "tech"))

	c, err := repo.GetCategoryByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "tech", c.Name)
}

func TestGetCategoryByID_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WillReturnError(sql.ErrNoRows)
	_, err := repo.GetCategoryByID(context.Background(), 99)
	assert.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestCreateCategory(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("new-cat").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))

	id, err := repo.CreateCategory(context.Background(), "new-cat")
	require.NoError(t, err)
	assert.Equal(t, int64(5), id)
}

func TestUpdateCategory(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectExec(`UPDATE categories SET name`).
		WithArgs("updated", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.UpdateCategory(context.Background(), 1, "updated"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCategory(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectExec(`DELETE FROM categories WHERE id`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.DeleteCategory(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateBanner(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`INSERT INTO banner_images`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	b := &BannerImage{Title: "b", ImageURL: "http://img", LinkURL: "http://link", SortOrder: 1, Type: 1}
	id, err := repo.CreateBanner(context.Background(), b)
	require.NoError(t, err)
	assert.Equal(t, int64(10), id)
}

func TestUpdateBanner(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectExec(`UPDATE banner_images SET title`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	b := &BannerImage{Title: "new", ImageURL: "u", LinkURL: "l", SortOrder: 1, Status: 1, Type: 1}
	require.NoError(t, repo.UpdateBanner(context.Background(), 1, b))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteBanner(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectExec(`DELETE FROM banner_images WHERE id`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.DeleteBanner(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetStatistics(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(view_count\),0\) FROM manuscripts`).WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(1000))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE status`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	stats, err := repo.GetStatistics(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(10), stats["manuscript_count"])
	assert.Equal(t, int64(5), stats["user_count"])
	assert.Equal(t, int64(1000), stats["view_count"])
	assert.Equal(t, int64(3), stats["pending_count"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListBanners_Error(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT id, title, image_url`).WillReturnError(errors.New("db down"))
	_, err := repo.ListBanners(context.Background(), 1)
	assert.Error(t, err)
}
