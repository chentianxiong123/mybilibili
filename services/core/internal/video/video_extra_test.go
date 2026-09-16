package video

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func doReq(h http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func muxForVideo(h *Handler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestRepo_BatchDeleteVideos(t *testing.T) {
	repo := NewRepository(nil)
	require.NoError(t, repo.BatchDeleteVideos(context.Background(), nil))

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()
	repo = NewRepository(db)

	mock.ExpectExec(`DELETE FROM videos WHERE id = ANY\(\$1\)`).WithArgs("{1,2}").WillReturnResult(sqlmock.NewResult(0, 2))
	require.NoError(t, repo.BatchDeleteVideos(context.Background(), []int64{1, 2}))

	mock.ExpectExec(`DELETE FROM videos WHERE id = ANY\(\$1\)`).WillReturnError(errors.New("boom"))
	assert.Error(t, repo.BatchDeleteVideos(context.Background(), []int64{1}))
}

func TestRepo_UpdateBanner(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE banner_images SET title=\$1, image_url=\$2, link_url=\$3, sort_order=\$4, status=\$5, category_id=\$6, time_slot=\$7, start_time=\$8, end_time=\$9 WHERE id=\$10`).
		WithArgs("t", "u", "l", int32(1), int32(1), nil, "default", nil, nil, int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpdateBanner(context.Background(), 5, &BannerImage{Title: "t", ImageURL: "u", LinkURL: "l", SortOrder: 1, Status: 1, TimeSlot: "default"}))

	mock.ExpectExec(`DELETE FROM banner_images WHERE id = \$1`).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.DeleteBanner(context.Background(), 5))
}

func TestRepo_ListBannersByCategory(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`AND category_id = \$2`).
		WithArgs(int32(2), int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "image_url", "link_url", "sort_order", "status", "type", "category_id", "time_slot", "start_time", "end_time"}).
			AddRow(1, "t", "u", "l", 1, 1, 2, 5, "default", now, now))
	list, err := repo.ListBannersByCategory(context.Background(), 2, 5)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.NotNil(t, list[0].StartTime)
	assert.NotNil(t, list[0].EndTime)

	mock.ExpectQuery(`FROM banner_images WHERE type = \$1 AND status = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "image_url", "link_url", "sort_order", "status", "type", "category_id", "time_slot", "start_time", "end_time"}))
	list, err = repo.ListBannersByCategory(context.Background(), 2, 0)
	require.NoError(t, err)
	assert.Len(t, list, 0)

	mock.ExpectQuery(`FROM banner_images`).WillReturnError(errors.New("boom"))
	_, err = repo.ListBannersByCategory(context.Background(), 2, 5)
	assert.Error(t, err)
}

func TestRepo_GetStatistics(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts$`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(view_count\),0\) FROM manuscripts`).WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(100))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE status = 0`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	stats, err := repo.GetStatistics(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(5), stats["manuscript_count"])
	assert.Equal(t, int64(10), stats["user_count"])
	assert.Equal(t, int64(100), stats["view_count"])
	assert.Equal(t, int64(3), stats["pending_count"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_BannerAndBatchDelete(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()
	svc := NewService(NewRepository(db), nil)

	mock.ExpectExec(`UPDATE banner_images SET title=\$1`).WithArgs("t", "u", "l", int32(1), int32(1), nil, "default", nil, nil, int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.UpdateBanner(context.Background(), 5, &BannerImage{Title: "t", ImageURL: "u", LinkURL: "l", SortOrder: 1, Status: 1, TimeSlot: "default"}))

	mock.ExpectExec(`DELETE FROM banner_images WHERE id = \$1`).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.DeleteBanner(context.Background(), 5))

	mock.ExpectExec(`DELETE FROM videos WHERE id = ANY\(\$1\)`).WithArgs("{1}").WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.BatchDeleteVideos(context.Background(), []int64{1}))
}

func TestHandleBannerImages_Route(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM banner_images WHERE type = \$1 AND status = 1`).WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "image_url", "link_url", "sort_order", "status", "type", "category_id", "time_slot", "start_time", "end_time"}))
	w := doReq(muxForVideo(h), "GET", "/api/v1/banner-images/home", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBannerHome_Put(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE banner_images SET title=\$1`).WithArgs("t", "u", "l", int32(1), int32(1), nil, "default", nil, nil, int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForVideo(h), "PUT", "/api/v1/banner/home/5", `{"title":"t","image_url":"u","link_url":"l","sort_order":1,"status":1}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBannerHome_Delete(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`DELETE FROM banner_images WHERE id = \$1`).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForVideo(h), "DELETE", "/api/v1/banner/home/5", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBannerHome_MethodNotAllowed(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doReq(muxForVideo(h), "PATCH", "/api/v1/banner/home", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleBannerCategory_PostPutDelete(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`INSERT INTO banner_images`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	w := doReq(muxForVideo(h), "POST", "/api/v1/banner/category/5", `{"title":"t"}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectExec(`UPDATE banner_images SET title=\$1`).WithArgs("t", "u", "l", int32(1), int32(1), int64(5), "default", nil, nil, int64(3)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doReq(muxForVideo(h), "PUT", "/api/v1/banner/category/5/3", `{"title":"t","image_url":"u","link_url":"l","sort_order":1,"status":1}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectExec(`DELETE FROM banner_images WHERE id = \$1`).WithArgs(int64(3)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doReq(muxForVideo(h), "DELETE", "/api/v1/banner/category/5/3", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBannerSingle_Delete(t *testing.T) {
	h, mock := newTestHandler(t)
	now := time.Now()

	mock.ExpectQuery(`FROM banner_images WHERE type = \$1 AND status = 1`).WithArgs(int32(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "image_url", "link_url", "sort_order", "status", "type", "category_id", "time_slot", "start_time", "end_time"}).
			AddRow(1, "a", "u", "l", 1, 1, 3, 0, "morning", now, nil).
			AddRow(2, "b", "u", "l", 2, 1, 3, 0, "evening", now, nil))
	mock.ExpectExec(`DELETE FROM banner_images WHERE id = \$1`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM banner_images WHERE id = \$1`).WithArgs(int64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForVideo(h), "DELETE", "/api/v1/banner/background", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBannerSingle_Post(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`INSERT INTO banner_images`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	w := doReq(muxForVideo(h), "POST", "/api/v1/banner/background", `{"title":"t"}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBanner_NotFound(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doReq(muxForVideo(h), "GET", "/api/v1/banner/unknown", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	w = doReq(muxForVideo(h), "GET", "/api/v1/banner/", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleStatisticsByPath_ManuscriptStatus(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FILTER \(WHERE review_status=0\)`).
		WillReturnRows(sqlmock.NewRows([]string{"a", "b", "c", "d"}).AddRow(1, 2, 3, 4))
	w := doReq(muxForVideo(h), "GET", "/api/v1/statistics/manuscript/status", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleStatisticsByPath_ManuscriptRecent(t *testing.T) {
	h, mock := newTestHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT id, user_id, title, cover_url, view_count, upload_time`).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "cover_url", "view_count", "upload_time"}).
			AddRow(1, 2, "t", "c", 5, now))
	w := doReq(muxForVideo(h), "GET", "/api/v1/statistics/manuscript/recent?limit=10", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectQuery(`SELECT id, user_id, title, cover_url`).WillReturnError(errors.New("boom"))
	w = doReq(muxForVideo(h), "GET", "/api/v1/statistics/manuscript/recent", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectQuery(`SELECT id, user_id, title, cover_url, view_count, upload_time`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "cover_url", "view_count", "upload_time"}).
			AddRow("bad", 2, "t", "c", 5, now))
	w = doReq(muxForVideo(h), "GET", "/api/v1/statistics/manuscript/recent", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleStatisticsByPath_Others(t *testing.T) {
	h, _ := newTestHandler(t)
	for _, p := range []string{"video/play", "user/growth", "comment", "video/hot"} {
		w := doReq(muxForVideo(h), "GET", "/api/v1/statistics/"+p, "", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	}
	w := doReq(muxForVideo(h), "GET", "/api/v1/statistics/unknown", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleCategoryByID_Delete_NotFound(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, name FROM categories`).WithArgs(int64(99)).WillReturnError(errors.New("no rows"))
	w := doReq(muxForVideo(h), "DELETE", "/api/v1/category/99", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- AdminHandler ----

func muxForVideoAdmin(h *AdminHandler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func newAdminVideoHandler(t *testing.T) (*AdminHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewAdminHandler(db), mock
}

func TestAdminHandler_List(t *testing.T) {
	h, mock := newAdminVideoHandler(t)
	mock.ExpectQuery(`FROM videos v LEFT JOIN manuscripts m`).
		WithArgs("", "", int32(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "title", "process_status", "user_id", "manuscript_title"}).
			AddRow(1, 2, "t", 3, int64(4), "mt"))
	w := doReq(muxForVideoAdmin(h), "GET", "/api/v1/video/admin/list", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_List_QueryError(t *testing.T) {
	h, mock := newAdminVideoHandler(t)
	mock.ExpectQuery(`FROM videos v LEFT JOIN manuscripts m`).WillReturnError(errors.New("boom"))
	w := doReq(muxForVideoAdmin(h), "GET", "/api/v1/video/admin/list", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_BatchDelete(t *testing.T) {
	h, mock := newAdminVideoHandler(t)
	mock.ExpectExec(`DELETE FROM videos WHERE id=\$1`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM videos WHERE id=\$1`).WithArgs(int64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForVideoAdmin(h), "DELETE", "/api/v1/video/admin/batch", `[1,2]`, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_Detail(t *testing.T) {
	h, mock := newAdminVideoHandler(t)

	mock.ExpectQuery(`SELECT id, manuscript_id, video_order, title, play_url_hd`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "video_order", "title", "play_url_hd", "play_url_sd", "play_url_ld", "duration_seconds", "process_progress", "process_stage", "has_subtitle", "has_summary"}).
			AddRow(1, 2, 1, "t", "hd", "sd", "ld", 60, 100, "done", true, true))
	w := doReq(muxForVideoAdmin(h), "GET", "/api/v1/video/admin/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectQuery(`SELECT id, manuscript_id, video_order`).WithArgs(int64(99)).WillReturnError(errors.New("no rows"))
	w = doReq(muxForVideoAdmin(h), "GET", "/api/v1/video/admin/99", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_Delete(t *testing.T) {
	h, mock := newAdminVideoHandler(t)
	mock.ExpectExec(`DELETE FROM videos WHERE id=\$1`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForVideoAdmin(h), "DELETE", "/api/v1/video/admin/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_exec_Error(t *testing.T) {
	h, mock := newAdminVideoHandler(t)
	mock.ExpectExec(`DELETE FROM videos WHERE id=\$1`).WillReturnError(errors.New("boom"))
	w := doReq(muxForVideoAdmin(h), "DELETE", "/api/v1/video/admin/1", "", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
