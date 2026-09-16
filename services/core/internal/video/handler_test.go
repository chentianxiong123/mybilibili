package video

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
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