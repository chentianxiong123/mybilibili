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

func newManuscriptHandlerWithMocks(t *testing.T) (*ManuscriptHandler, sqlmock.Sqlmock) {
	t.Helper()
	msDB, msMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { msDB.Close() })

	userDB, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { userDB.Close() })

	svc := NewManuscriptService(NewManuscriptRepository(msDB), user.NewRepository(userDB))
	h := NewManuscriptHandler(svc)
	return h, msMock
}

func TestManuscriptHandler_GetManuscript(t *testing.T) {
	h, mock := newManuscriptHandlerWithMocks(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).AddRow(
			1, "t", "d", "c.jpg", int64(10), int64(1),
			int64(0), int64(0), int64(0), int64(0), int64(0),
			int64(0), int64(0), int32(3), int32(1), "",
			now, int64(0), now, now, "", int32(0), "local"))

	mock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))

	resp, err := h.GetManuscript(ctx, &pb.GetManuscriptRequest{Id: 1})
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Manuscript.Id)
}

func TestManuscriptHandler_ListUserManuscripts(t *testing.T) {
	h, mock := newManuscriptHandlerWithMocks(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id`).WillReturnRows(sqlmock.NewRows(manuscriptCols))

	resp, err := h.ListUserManuscripts(ctx, &pb.ListUserManuscriptsRequest{UserId: 10, Status: 3, Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(0), resp.Total)
}

func TestManuscriptHandler_ListRecommended(t *testing.T) {
	h, mock := newManuscriptHandlerWithMocks(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE status = 3 ORDER BY upload_time`).WillReturnRows(sqlmock.NewRows(manuscriptCols))

	resp, err := h.ListRecommended(ctx, &pb.ListRecommendedRequest{UserId: 1})
	require.NoError(t, err)
	assert.Empty(t, resp.Manuscripts)
}

func TestManuscriptHandler_ListHot(t *testing.T) {
	h, mock := newManuscriptHandlerWithMocks(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE status = 3 ORDER BY view_count`).WillReturnRows(sqlmock.NewRows(manuscriptCols))

	resp, err := h.ListHot(ctx, &pb.ListHotRequest{UserId: 1})
	require.NoError(t, err)
	assert.Empty(t, resp.Manuscripts)
}

func TestManuscriptHandler_ListByCategory(t *testing.T) {
	h, mock := newManuscriptHandlerWithMocks(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE category_id = \$1 AND status = 3`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE category_id = .+ AND status = 3`).WillReturnRows(sqlmock.NewRows(manuscriptCols))

	resp, err := h.ListByCategory(ctx, &pb.ListByCategoryRequest{CategoryId: 5, Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(0), resp.Total)
}

func TestManuscriptHandler_ListCategories(t *testing.T) {
	h, mock := newManuscriptHandlerWithMocks(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, name FROM categories`).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))

	resp, err := h.ListCategories(ctx, &pb.ListCategoriesRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Categories, 1)
}

func TestManuscriptHandler_DeleteManuscript(t *testing.T) {
	h, mock := newManuscriptHandlerWithMocks(t)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM manuscripts WHERE id`).WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := h.DeleteManuscript(ctx, &pb.DeleteManuscriptRequest{Id: 1, UserId: 10})
	require.NoError(t, err)
}

func TestManuscriptHandler_PublishManuscript(t *testing.T) {
	h, mock := newManuscriptHandlerWithMocks(t)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := h.PublishManuscript(ctx, &pb.PublishManuscriptRequest{Id: 1, UserId: 10})
	require.NoError(t, err)
}

func TestManuscriptHandler_UnpublishManuscript(t *testing.T) {
	h, mock := newManuscriptHandlerWithMocks(t)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := h.UnpublishManuscript(ctx, &pb.UnpublishManuscriptRequest{Id: 1, UserId: 10})
	require.NoError(t, err)
}

func TestManuscriptHandler_SearchUserManuscripts(t *testing.T) {
	h, mock := newManuscriptHandlerWithMocks(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id = .+ AND title ILIKE`).WillReturnRows(sqlmock.NewRows(manuscriptCols))

	resp, err := h.SearchUserManuscripts(ctx, &pb.SearchUserManuscriptsRequest{UserId: 10, Keyword: "test"})
	require.NoError(t, err)
	assert.Empty(t, resp.Manuscripts)
}

func TestManuscriptHandler_GetManuscriptWithVideos(t *testing.T) {
	h, mock := newManuscriptHandlerWithMocks(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).AddRow(
			1, "t", "d", "c.jpg", int64(10), int64(1),
			int64(0), int64(0), int64(0), int64(0), int64(0),
			int64(0), int64(0), int32(3), int32(1), "",
			now, int64(0), now, now, "", int32(0), "local"))

	mock.ExpectQuery(`SELECT id, name FROM categories WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "c"))

	mock.ExpectQuery(`SELECT .+ FROM videos WHERE manuscript_id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "video_order", "title",
			"play_url_hd", "play_url_sd", "play_url_ld",
			"upload_time", "updated_at", "process_progress", "process_stage",
			"has_subtitle", "has_summary", "process_status", "process_error",
			"source_video_url", "duration_seconds", "is_vertical",
		}).AddRow(100, int64(1), int32(0), "P1", "hd.mp4", "sd.mp4", "ld.mp4",
			now, now, int32(100), "DONE", int32(0), int32(0), int32(5), "", "src.mp4", int32(60), int32(0)))

	mock.ExpectQuery(`SELECT t.name FROM tags t`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"name"}))

	resp, err := h.GetManuscriptWithVideos(ctx, &pb.GetManuscriptWithVideosRequest{Id: 1, CurrentUserId: 0})
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Manuscript.Id)
}
