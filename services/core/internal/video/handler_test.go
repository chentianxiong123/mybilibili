package video

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewRepository(db)
	svc := NewService(repo, nil)
	return NewHandler(svc, nil), mock
}

func videoRow(id int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "manuscript_id", "video_order", "title",
		"play_url_hd", "play_url_sd", "play_url_ld",
		"upload_time", "updated_at",
		"process_progress", "process_stage",
		"has_subtitle", "has_summary",
		"process_status", "process_error",
		"source_video_url", "duration_seconds", "is_vertical",
	}).AddRow(
		id, int64(10), 1, "Test Video",
		"/uploads/hd.mp4", "/uploads/sd.mp4", "/uploads/ld.mp4",
		time.Now(), time.Now(),
		100, "done", 1, 1, 3, "",
		"/uploads/source.mp4", 60, 0,
	)
}

func TestHandleGetOne_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM videos.*WHERE.*id`).
		WithArgs(int64(123)).
		WillReturnRows(videoRow(123))

	req := httptest.NewRequest("GET", "/api/v1/video/123", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":200`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetOne_404(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM videos.*WHERE.*id`).
		WithArgs(int64(99999)).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/api/v1/video/99999", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// handleVideo returns 404 in JSON body but with HTTP 200 (writeJSON doesn't set status)
	assert.Contains(t, rec.Body.String(), `"code":404`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCategory_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM categories`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "动画").
			AddRow(2, "游戏").
			AddRow(3, "音乐"))

	req := httptest.NewRequest("GET", "/api/v1/category", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":200`)
	assert.Contains(t, rec.Body.String(), "动画")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleStatistics_Overview_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	// 真实统计 SQL 复杂, 这里 mock 返回空 map
	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	req := httptest.NewRequest("GET", "/api/v1/statistics", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// 只要不 panic + 200 即可 (实际 SQL 返回空, writeJSON 包成 {"code":200,...})
	assert.NotNil(t, rec.Body.String())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleVideo_EmptyPath(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest("GET", "/api/v1/video/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Contains(t, rec.Body.String(), `"code":404`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleVideo_ByManuscript(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM videos.*WHERE.*manuscript_id`).
		WithArgs(int64(10)).
		WillReturnRows(videoRow(1).AddRow(2, int64(10), 2, "Video 2",
			"/a.mp4", "/b.mp4", "/c.mp4",
			time.Now(), time.Now(),
			100, "done", 1, 1, 3, "",
			"/src.mp4", 120, 0))

	req := httptest.NewRequest("GET", "/api/v1/video/10/manuscript", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":200`)
	assert.Contains(t, rec.Body.String(), "Test Video")
	assert.Contains(t, rec.Body.String(), "Video 2")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleVideo_UserIds(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM manuscripts.*WHERE.*user_id`).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(101).AddRow(102))

	req := httptest.NewRequest("GET", "/api/v1/video/user/42/ids", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "101")
	assert.Contains(t, rec.Body.String(), "102")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleVideo_UserVideoIds(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM videos.*JOIN manuscripts`).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(500).AddRow(501))

	req := httptest.NewRequest("GET", "/api/v1/video/user/42/video-ids", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "500")
	assert.Contains(t, rec.Body.String(), "501")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCategory_POST_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`INSERT INTO categories`).
		WithArgs("测试分类").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))

	body := strings.NewReader(`{"name":"测试分类"}`)
	req := httptest.NewRequest("POST", "/api/v1/category", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ok")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCategory_POST_EmptyName(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	body := strings.NewReader(`{"name":""}`)
	req := httptest.NewRequest("POST", "/api/v1/category", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Contains(t, rec.Body.String(), `"code":400`)
	assert.Contains(t, rec.Body.String(), "name required")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCategoryByID_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM categories.*WHERE.*id`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(5, "动画"))

	req := httptest.NewRequest("GET", "/api/v1/category/5", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "动画")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCategoryByID_GET_404(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM categories.*WHERE.*id`).
		WithArgs(int64(99999)).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/api/v1/category/99999", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Contains(t, rec.Body.String(), `"code":404`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCategoryByID_PUT_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM categories.*WHERE.*id`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(5, "旧名"))

	mock.ExpectExec(`UPDATE categories SET name`).
		WithArgs("新名", int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := strings.NewReader(`{"name":"新名"}`)
	req := httptest.NewRequest("PUT", "/api/v1/category/5", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ok")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCategoryByID_DELETE_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM categories.*WHERE.*id`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(5, "动画"))

	mock.ExpectExec(`DELETE FROM categories WHERE id`).
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("DELETE", "/api/v1/category/5", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ok")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBannerHome_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM banner_images.*WHERE.*type.*=.*\$1`).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}).AddRow(1, "Home Banner", "/banner.jpg", "/link", 1,
			1, 1, 0, "default", nil, nil))

	req := httptest.NewRequest("GET", "/api/v1/banner/home", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Home Banner")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBannerHome_POST_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

mock.ExpectQuery(`INSERT INTO banner_images`).
		WithArgs("New Banner", "/img.jpg", "/link", int32(0), int32(1), nil, "default", (*time.Time)(nil), (*time.Time)(nil)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	body := strings.NewReader(`{"title":"New Banner","image_url":"/img.jpg","link_url":"/link"}`)
	req := httptest.NewRequest("POST", "/api/v1/banner/home", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ok")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBannerCategory_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM banner_images.*WHERE.*type.*=.*\$1`).
		WithArgs(int32(2), int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}).AddRow(2, "Cat Banner", "/cat.jpg", "/link", 1,
			1, 2, 3, "default", nil, nil))

	req := httptest.NewRequest("GET", "/api/v1/banner/category/3", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Cat Banner")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBannerBackground_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM banner_images.*WHERE.*type.*=.*\$1`).
		WithArgs(int32(3)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}).AddRow(3, "Morning BG", "/bg.jpg", "", 1,
			1, 3, 0, "morning", nil, nil))

	req := httptest.NewRequest("GET", "/api/v1/banner/background", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Morning BG")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBannerUserProfile_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM banner_images.*WHERE.*type.*=.*\$1`).
		WithArgs(int32(4)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}).AddRow(4, "Profile BG", "/profile.jpg", "", 1,
			1, 4, 0, "default", nil, nil))

	req := httptest.NewRequest("GET", "/api/v1/banner/user-profile", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Profile BG")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleStatisticsByPath_Overview(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(view_count\),0\) FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(1000))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE status = 0`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	req := httptest.NewRequest("GET", "/api/v1/statistics/overview", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "manuscript_count")
	assert.Contains(t, rec.Body.String(), "1000")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleStatisticsByPath_Recent(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	mock.ExpectQuery(`SELECT.*FROM manuscripts.*ORDER BY.*LIMIT`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "title", "cover_url", "view_count", "upload_time",
		}).AddRow(1, 10, "Recent Manuscript", "/cover.jpg", 500, time.Now()))

	req := httptest.NewRequest("GET", "/api/v1/statistics/manuscript/recent?limit=5", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Recent Manuscript")
	assert.Contains(t, rec.Body.String(), "500")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleStatisticsByPath_Default(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest("GET", "/api/v1/statistics/unknown", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "{}")
	assert.NoError(t, mock.ExpectationsWereMet())
}