package manuscript

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRepo(t *testing.T) (*ManuscriptRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewManuscriptRepository(db), mock
}

var manuscriptCols = []string{
	"id", "title", "description", "cover_url", "user_id", "category_id",
	"view_count", "like_count", "coin_count", "collect_count", "share_count",
	"comment_count", "danmaku_count", "status", "review_status", "review_reason",
	"review_time", "reviewer_id", "upload_time", "updated_at", "duration", "duration_seconds", "source_type",
}

func scanManuscriptRow(row *sqlmock.Rows) {
	now := time.Now()
	row.AddRow(1, "title1", "desc1", "cover.jpg", int64(10), int64(1),
		int64(100), int64(20), int64(5), int64(3), int64(1),
		int64(0), int64(0), int32(3), int32(1), "ok",
		now, int64(99), now, now, "03:00", int32(180), "local")
}

func TestFindByID(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows(manuscriptCols)
	scanManuscriptRow(rows)
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(1)).WillReturnRows(rows)

	m, err := repo.FindByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), m.ID)
	assert.Equal(t, "title1", m.Title)
	assert.Equal(t, "desc1", m.Description)
	assert.Equal(t, "cover.jpg", m.CoverURL)
	assert.Equal(t, int64(10), m.UserID)
	assert.Equal(t, int64(1), m.CategoryID)
	assert.Equal(t, int64(100), m.ViewCount)
	assert.Equal(t, "local", m.SourceType)
	assert.Equal(t, "03:00", m.Duration)
	assert.Equal(t, int32(180), m.DurationSeconds)
}

func TestFindByID_NotFound(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols))

	m, err := repo.FindByID(ctx, 999)
	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestIncrementViewCount(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE manuscripts SET view_count`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.IncrementViewCount(ctx, 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFirstVideoIsVertical(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT is_vertical FROM videos`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"is_vertical"}).AddRow(1))

	v, err := repo.FirstVideoIsVertical(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int32(1), v)
}

func TestFirstVideoIsVertical_NoRows(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT is_vertical FROM videos`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"is_vertical"}))

	v, err := repo.FirstVideoIsVertical(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int32(0), v)
}

var videoCols = []string{
	"id", "manuscript_id", "video_order", "title",
	"play_url_hd", "play_url_sd", "play_url_ld",
	"upload_time", "updated_at", "process_progress", "process_stage",
	"has_subtitle", "has_summary", "process_status", "process_error",
	"source_video_url", "duration_seconds", "is_vertical",
}

func TestFindVideosByManuscriptID(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows(videoCols).
		AddRow(100, int64(1), int32(0), "P1", "hd.mp4", "sd.mp4", "ld.mp4",
			now, now, int32(100), "DONE", int32(1), int32(1), int32(5), "",
			"src.mp4", int32(180), int32(0))
	mock.ExpectQuery(`SELECT .+ FROM videos WHERE manuscript_id`).WithArgs(int64(1)).WillReturnRows(rows)

	videos, err := repo.FindVideosByManuscriptID(ctx, 1)
	require.NoError(t, err)
	require.Len(t, videos, 1)
	assert.Equal(t, int64(100), videos[0].ID)
	assert.Equal(t, "P1", videos[0].Title)
	assert.Equal(t, "hd.mp4", videos[0].PlayURLHd)
	assert.Equal(t, "DONE", videos[0].ProcessStage)
	assert.Equal(t, int32(5), videos[0].ProcessStatus)
}

func TestFindVideosByManuscriptID_Error(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM videos WHERE manuscript_id`).WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	videos, err := repo.FindVideosByManuscriptID(ctx, 1)
	assert.Nil(t, videos)
	assert.Error(t, err)
}

func TestFindTagsByVideoID(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"name"}).AddRow("go").AddRow("test")
	mock.ExpectQuery(`SELECT t.name FROM tags t`).WithArgs(int64(100)).WillReturnRows(rows)

	tags, err := repo.FindTagsByVideoID(ctx, 100)
	require.NoError(t, err)
	assert.Equal(t, []string{"go", "test"}, tags)
}

func TestFindTagsByVideoID_Error(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT t.name FROM tags t`).WithArgs(int64(100)).
		WillReturnError(sql.ErrConnDone)

	tags, err := repo.FindTagsByVideoID(ctx, 100)
	assert.Nil(t, tags)
	assert.Error(t, err)
}

func TestFindCategoryByID(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(5), "技术"))

	cat, err := repo.FindCategoryByID(ctx, 5)
	require.NoError(t, err)
	assert.Equal(t, "技术", cat.Name)
}

func TestFindCategoryByID_NotFound(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	cat, err := repo.FindCategoryByID(ctx, 999)
	assert.Nil(t, cat)
	assert.Error(t, err)
}

func TestListByUser_WithStatus(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = \$2`).
		WithArgs(int64(10), int32(3)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	rows := sqlmock.NewRows(manuscriptCols)
	scanManuscriptRow(rows)
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id = .+ ORDER BY upload_time`).
		WithArgs(int64(10), int32(3), int32(20), int32(0)).
		WillReturnRows(rows)

	list, total, err := repo.ListByUser(ctx, 10, 3, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
}

func TestListByUser_AllStatus(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1`).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id = \$1 ORDER BY upload_time`).
		WithArgs(int64(10), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols))

	list, total, err := repo.ListByUser(ctx, 10, -100, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, list)
}

func TestListByUser_CountError(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = \$2`).
		WithArgs(int64(10), int32(3)).
		WillReturnError(sql.ErrConnDone)

	_, _, err := repo.ListByUser(ctx, 10, 3, 1, 20)
	assert.Error(t, err)
}

func TestListRecommended(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows(manuscriptCols)
	scanManuscriptRow(rows)
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE status = 3 ORDER BY upload_time`).WillReturnRows(rows)

	list, err := repo.ListRecommended(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestListRecommended_Error(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE status = 3 ORDER BY upload_time`).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.ListRecommended(ctx)
	assert.Error(t, err)
}

func TestListHot(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows(manuscriptCols)
	scanManuscriptRow(rows)
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE status = 3 ORDER BY view_count`).WillReturnRows(rows)

	list, err := repo.ListHot(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestListHot_Error(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE status = 3 ORDER BY view_count`).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.ListHot(ctx)
	assert.Error(t, err)
}

func TestListByCategory(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE category_id = \$1 AND status = 3`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	rows := sqlmock.NewRows(manuscriptCols)
	scanManuscriptRow(rows)
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE category_id = .+ ORDER BY upload_time`).
		WithArgs(int64(5), int32(20), int32(0)).
		WillReturnRows(rows)

	list, total, err := repo.ListByCategory(ctx, 5, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
}

func TestListByCategory_CountError(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE category_id = \$1 AND status = 3`).
		WithArgs(int64(5)).
		WillReturnError(sql.ErrConnDone)

	_, _, err := repo.ListByCategory(ctx, 5, 1, 20)
	assert.Error(t, err)
}

func TestListCategories(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(int64(1), "生活").AddRow(int64(2), "技术")
	mock.ExpectQuery(`SELECT id, name FROM categories`).WillReturnRows(rows)

	cats, err := repo.ListCategories(ctx)
	require.NoError(t, err)
	assert.Len(t, cats, 2)
	assert.Equal(t, "生活", cats[0].Name)
}

func TestListCategories_Error(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, name FROM categories`).WillReturnError(sql.ErrConnDone)

	_, err := repo.ListCategories(ctx)
	assert.Error(t, err)
}

func TestUpdateStatus_Success(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE manuscripts SET status`).WithArgs(int32(3), int64(1), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateStatus(ctx, 1, 10, 3)
	require.NoError(t, err)
}

func TestUpdateStatus_NotFound(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE manuscripts SET status`).WithArgs(int32(3), int64(999), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.UpdateStatus(ctx, 999, 10, 3)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestDelete_Success(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM manuscripts WHERE id`).WithArgs(int64(1), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(ctx, 1, 10)
	require.NoError(t, err)
}

func TestSearchUser(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows(manuscriptCols)
	scanManuscriptRow(rows)
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id = .+ AND title ILIKE .+ ORDER BY upload_time`).
		WithArgs(int64(10), "%test%").
		WillReturnRows(rows)

	list, err := repo.SearchUser(ctx, 10, "test", "")
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestSearchUser_SortByView(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id = .+ AND title ILIKE .+ ORDER BY view_count`).
		WithArgs(int64(10), "%test%").
		WillReturnRows(sqlmock.NewRows(manuscriptCols))

	_, err := repo.SearchUser(ctx, 10, "test", "view")
	require.NoError(t, err)
}

func TestSearchUser_SortByLike(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id = .+ AND title ILIKE .+ ORDER BY like_count`).
		WithArgs(int64(10), "%test%").
		WillReturnRows(sqlmock.NewRows(manuscriptCols))

	_, err := repo.SearchUser(ctx, 10, "test", "like")
	require.NoError(t, err)
}

func TestSearchUser_Error(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id = .+ AND title ILIKE .+`).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.SearchUser(ctx, 10, "test", "")
	assert.Error(t, err)
}

func TestToPB(t *testing.T) {
	repo, _ := newRepo(t)
	now := time.Now()
	m := &Manuscript{
		ID: 1, Title: "t", Description: "d", CoverURL: "c.jpg",
		UserID: 10, CategoryID: 5, ViewCount: 100, LikeCount: 20,
		CoinCount: 5, CollectCount: 3, ShareCount: 1,
		CommentCount: 10, DanmakuCount: 0, Status: 3, ReviewStatus: 1,
		Duration: "03:00", DurationSeconds: 180, SourceType: "local",
		UploadTime: sql.NullTime{Time: now, Valid: true},
		UpdatedAt:  sql.NullTime{Time: now, Valid: true},
	}

	pb := repo.ToPB(m, "技术", nil, nil, nil, 0, "")
	assert.Equal(t, int64(1), pb.Id)
	assert.Equal(t, "t", pb.Title)
	assert.Equal(t, "d", pb.Description)
	assert.Equal(t, "技术", pb.CategoryName)
	assert.Equal(t, int32(180), pb.DurationSeconds)
	assert.Equal(t, "local", pb.SourceType)
}

func TestVideoToPB(t *testing.T) {
	v := &Video{
		ID: 100, Title: "v1", PlayURLHd: "hd.mp4", PlayURLSd: "sd.mp4",
		PlayURLld: "ld.mp4", DurationSeconds: 60, VideoOrder: 0,
		ProcessStatus: 5, ProcessProgress: 100, ProcessStage: "DONE",
		ProcessError: "", IsVertical: 0,
	}

	pb := videoToPB(v)
	assert.Equal(t, int64(100), pb.Id)
	assert.Equal(t, "v1", pb.Title)
	assert.Equal(t, "hd.mp4", pb.PlayUrlHd)
	assert.Equal(t, int32(60), pb.DurationSeconds)
	assert.Equal(t, int32(5), pb.ProcessStatus)
}

func TestItoa(t *testing.T) {
	assert.Equal(t, "0", itoa(0))
	assert.Equal(t, "1", itoa(1))
	assert.Equal(t, "42", itoa(42))
}
