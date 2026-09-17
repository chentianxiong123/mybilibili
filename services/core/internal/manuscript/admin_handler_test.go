package manuscript

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAdminHandler(t *testing.T) (*ManuscriptAdminHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	h := NewManuscriptAdminHandler(db)
	h.SetEventWriter(&fakeEventWriter{})
	h.SetEventPublisher(&fakeEventPublisher{})
	h.SetPermChecker(&fakePermChecker{ok: true})
	return h, mock
}

type fakeEventWriter struct{}

func (f *fakeEventWriter) RecordStatusEvent(_ context.Context, _, _ int64, _, _ int32, _, _ string, _ int64, _ string) error {
	return nil
}
func (f *fakeEventWriter) RecordVideoProcessEvent(_ context.Context, _, _ int64, _, _ int32, _ string, _ int) error {
	return nil
}
func (f *fakeEventWriter) RecordEditVersion(_ context.Context, _, _ int64, _, _ string, _ string) error {
	return nil
}

type fakeEventPublisher struct{}

func (f *fakeEventPublisher) PublishManuscriptIndex(_ context.Context, _ int64, _, _ string) error {
	return nil
}
func (f *fakeEventPublisher) PublishAnalytics(_ context.Context, _, _ int64, _, _ string, _ int64) error {
	return nil
}
func (f *fakeEventPublisher) PublishVideoProcess(_ context.Context, _, _ int64, _, _ string, _ int64) error {
	return nil
}

type fakePermChecker struct {
	ok bool
}

func (f *fakePermChecker) CheckPermission(_ *http.Request, _ string) (int64, bool) {
	return 1, f.ok
}

func TestAdminHandler_HandlePending_MethodNotAllowed(t *testing.T) {
	h, _ := newAdminHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/pending", nil)
	w := httptest.NewRecorder()
	h.handlePending(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestAdminHandler_HandlePending_Success(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts m WHERE m.review_status = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "user_id", "category_id", "status", "review_status",
			"view_count", "like_count", "comment_count", "upload_time", "updated_at",
		}).AddRow(1, "title1", int64(10), int64(1), int32(0), int32(0),
			int64(0), int64(0), int64(0), timeNow(), timeNow()))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE review_status = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/pending?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	h.handlePending(w, req)
	assert.Equal(t, 200, w.Code)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	assert.Equal(t, float64(1), data["total"])
}

func TestAdminHandler_HandleProcessing_MethodNotAllowed(t *testing.T) {
	h, _ := newAdminHandler(t)
	req := httptest.NewRequest("DELETE", "/api/v1/manuscript/admin/processing", nil)
	w := httptest.NewRecorder()
	h.handleProcessing(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestAdminHandler_HandleProcessing_Success(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts m WHERE m.review_status = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "user_id", "category_id", "status", "review_status",
			"view_count", "like_count", "comment_count", "upload_time", "updated_at",
		}))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE review_status = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/processing", nil)
	w := httptest.NewRecorder()
	h.handleProcessing(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_HandleAll_MethodNotAllowed(t *testing.T) {
	h, _ := newAdminHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/all", nil)
	w := httptest.NewRecorder()
	h.handleAll(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestAdminHandler_HandleAll_WithKeyword(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts m WHERE m.title ILIKE`).WillReturnRows(sqlmock.NewRows([]string{
		"id", "title", "user_id", "category_id", "status", "review_status",
		"view_count", "like_count", "comment_count", "upload_time", "updated_at",
	}))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/all?keyword=test", nil)
	w := httptest.NewRecorder()
	h.handleAll(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_HandleAll_WithStatus(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts m WHERE m.status = \$1`).WillReturnRows(sqlmock.NewRows([]string{
		"id", "title", "user_id", "category_id", "status", "review_status",
		"view_count", "like_count", "comment_count", "upload_time", "updated_at",
	}))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/all?status=3", nil)
	w := httptest.NewRecorder()
	h.handleAll(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_HandleAll_WithRows(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts m WHERE m.title ILIKE`).WillReturnRows(sqlmock.NewRows([]string{
		"id", "title", "user_id", "category_id", "status", "review_status",
		"view_count", "like_count", "comment_count", "upload_time", "updated_at",
	}).AddRow(1, "test", int64(10), int64(1), int32(3), int32(1),
		int64(100), int64(20), int64(5), timeNow(), timeNow()))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/all?keyword=test", nil)
	w := httptest.NewRecorder()
	h.handleAll(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_HandleAll_DBError(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts m WHERE m.title ILIKE`).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/all?keyword=test", nil)
	w := httptest.NewRecorder()
	h.handleAll(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_HandleStatistics_MethodNotAllowed(t *testing.T) {
	h, _ := newAdminHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/statistics", nil)
	w := httptest.NewRecorder()
	h.handleStatistics(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestAdminHandler_HandleStatistics_Success(t *testing.T) {
	h, mock := newAdminHandler(t)

	for _, q := range []string{
		`SELECT COUNT\(\*\) FROM manuscripts`,
		`SELECT COUNT\(\*\) FROM manuscripts WHERE review_status = 0`,
		`SELECT COUNT\(\*\) FROM manuscripts WHERE review_status = 1`,
		`SELECT COUNT\(\*\) FROM manuscripts WHERE review_status = 2`,
		`SELECT COUNT\(\*\) FROM manuscripts WHERE status = -1`,
		`SELECT COALESCE\(SUM\(view_count\),0\) FROM manuscripts`,
		`SELECT COALESCE\(SUM\(like_count\),0\) FROM manuscripts`,
		`SELECT COUNT\(\*\) FROM videos`,
	} {
		mock.ExpectQuery(q).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(10)))
	}

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/statistics", nil)
	w := httptest.NewRecorder()
	h.handleStatistics(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_HandleByID_EmptyPath(t *testing.T) {
	h, _ := newAdminHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestAdminHandler_HandleByID_GetDetail(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts m WHERE m.id = \$1`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "cover_url", "user_id", "category_id",
			"status", "review_status", "review_reason",
			"view_count", "like_count", "coin_count",
			"collect_count", "share_count", "comment_count", "danmaku_count",
			"duration", "upload_time", "updated_at", "source_type",
		}).AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
			int32(3), int32(1), "ok",
			int64(100), int64(20), int64(5),
			int64(3), int64(1), int64(0), int64(0),
			"03:00", timeNow(), timeNow(), "local"))

	mock.ExpectQuery(`SELECT id, video_order, title, play_url_hd, upload_time FROM videos WHERE manuscript_id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "video_order", "title", "play_url_hd", "upload_time",
		}).AddRow(100, 0, "P1", "hd.mp4", timeNow()))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/1", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_HandleByID_NotFound(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts m WHERE m.id = \$1`).WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "cover_url", "user_id", "category_id",
			"status", "review_status", "review_reason",
			"view_count", "like_count", "coin_count",
			"collect_count", "share_count", "comment_count", "danmaku_count",
			"duration", "upload_time", "updated_at", "source_type",
		}))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/999", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestAdminHandler_HandleByID_InvalidID(t *testing.T) {
	h, _ := newAdminHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/abc", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestAdminHandler_HandleByID_MethodNotAllowed(t *testing.T) {
	h, _ := newAdminHandler(t)
	req := httptest.NewRequest("DELETE", "/api/v1/manuscript/admin/1", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestAdminHandler_ReviewManuscript_Approve(t *testing.T) {
	h, mock := newAdminHandler(t)

	// exists check
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	// from status + uid
	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(0), int64(10)))
	// update
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))
	// triggerAllVideoProcess
	mock.ExpectQuery(`SELECT id FROM videos WHERE manuscript_id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	body := `{"reviewerId":"99","reason":"good"}`
	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/approve/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_ReviewManuscript_Reject(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(3), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"reason":"bad"}`
	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/reject/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_ReviewManuscript_NotFound(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/approve/999", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestAdminHandler_ReviewManuscript_InvalidID(t *testing.T) {
	h, _ := newAdminHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/approve/abc", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestAdminHandler_ReviewManuscript_MethodNotAllowed(t *testing.T) {
	h, _ := newAdminHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/approve/abc", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestAdminHandler_PublishManuscript(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT title FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("my ms"))
	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(1), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))
	// notification insert
	mock.ExpectExec(`INSERT INTO notifications`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/publish/1", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_UnpublishManuscript(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT title FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("my ms"))
	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(3), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/unpublish/1", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_RetryManuscript(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT title FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("my ms"))
	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(5), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/retry/1", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_Videos(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT id, video_order, title, play_url_hd, play_url_sd, play_url_ld, upload_time`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "video_order", "title", "play_url_hd", "play_url_sd", "play_url_ld", "upload_time",
		}).AddRow(100, 0, "P1", "hd.mp4", "sd.mp4", "ld.mp4", timeNow()))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/1/videos", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_Videos_Error(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT id, video_order, title, play_url_hd, play_url_sd, play_url_ld, upload_time`).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/1/videos", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_Transcode(t *testing.T) {
	h, mock := newAdminHandler(t)

	// exists
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	// manuscript_id, title
	mock.ExpectQuery(`SELECT manuscript_id, title FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "title"}).AddRow(int64(1), "P1"))
	// source URL
	mock.ExpectQuery(`SELECT COALESCE\(source_video_url`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"source_video_url"}).AddRow("src.mp4"))
	// uploader ID
	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	// from process
	mock.ExpectQuery(`SELECT process_status FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"process_status"}).AddRow(int32(0)))
	// update video
	mock.ExpectExec(`UPDATE videos SET process_status`).WillReturnResult(sqlmock.NewResult(0, 1))
	// update manuscript
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/transcode/100", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_Transcode_MethodNotAllowed(t *testing.T) {
	h, _ := newAdminHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/transcode/100", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestAdminHandler_Transcode_NotFound(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM videos WHERE id`).WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/transcode/999", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestAdminHandler_Transcode_InvalidVideoID(t *testing.T) {
	h, _ := newAdminHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/transcode/abc", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestAdminHandler_Transcode_MissingVideoID(t *testing.T) {
	h, _ := newAdminHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/transcode/", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestAdminHandler_ExtractAudio(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT manuscript_id, title FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "title"}).AddRow(int64(1), "P1"))
	mock.ExpectQuery(`SELECT COALESCE\(source_video_url`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"source_video_url"}).AddRow(""))
	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT process_status FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"process_status"}).AddRow(int32(0)))
	mock.ExpectExec(`UPDATE videos SET process_status`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/extract-audio/100", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_GenerateSubtitle(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT manuscript_id, title FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "title"}).AddRow(int64(1), "P1"))
	mock.ExpectQuery(`SELECT COALESCE\(source_video_url`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"source_video_url"}).AddRow(""))
	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT process_status FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"process_status"}).AddRow(int32(0)))
	mock.ExpectExec(`UPDATE videos SET process_status`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/generate-subtitle/100", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_AiSummary(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT manuscript_id, title FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "title"}).AddRow(int64(1), "P1"))
	mock.ExpectQuery(`SELECT COALESCE\(source_video_url`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"source_video_url"}).AddRow(""))
	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT process_status FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"process_status"}).AddRow(int32(0)))
	mock.ExpectExec(`UPDATE videos SET process_status`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/ai-summary/100", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_ProcessAll(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT manuscript_id, title FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "title"}).AddRow(int64(1), "P1"))
	mock.ExpectQuery(`SELECT COALESCE\(source_video_url`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"source_video_url"}).AddRow(""))
	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT process_status FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"process_status"}).AddRow(int32(0)))
	mock.ExpectExec(`UPDATE videos SET process_status`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/process-all/100", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_VideoSource(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COALESCE\(source_video_url`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"source_video_url", "title", "duration_seconds"}).
			AddRow("src.mp4", "P1", int64(60)))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/video-source/100", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_VideoSource_NotFound(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COALESCE\(source_video_url`).WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/video-source/999", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_ResetVideo(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectExec(`UPDATE videos SET process_status`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/reset/100", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_ResetVideo_NotFound(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM videos WHERE id`).WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/reset/999", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestAdminHandler_ApproveWithProcess(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(0), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET review_status`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id FROM videos WHERE manuscript_id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(100)))
	mock.ExpectExec(`UPDATE videos SET updated_at`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/1/approve-with-process", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_ApproveWithProcess_VideosError(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(0), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET review_status`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id FROM videos WHERE manuscript_id`).WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/1/approve-with-process", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_PermissionNotConfigured(t *testing.T) {
	db, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	defer db.Close()

	mux := http.NewServeMux()
	h := NewManuscriptAdminHandler(db)
	h.SetEventWriter(&fakeEventWriter{})
	// No perm checker set
	h.Register(mux)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/pending", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)
}

func TestAdminHandler_PermDenied(t *testing.T) {
	db, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	defer db.Close()

	mux := http.NewServeMux()
	h := NewManuscriptAdminHandler(db)
	h.SetEventWriter(&fakeEventWriter{})
	h.SetPermChecker(&fakePermChecker{ok: false})
	h.Register(mux)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/pending", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestAdminHandler_ReviewManuscript_DBError(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(0), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnError(sql.ErrConnDone)

	body := `{"reviewerId":"99","reason":"good"}`
	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/approve/1", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestAdminHandler_SetManuscriptStatus_NotFound(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/publish/999", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestAdminHandler_SetManuscriptStatus_DBError(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT title FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("title"))
	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(0), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/publish/1", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestAdminHandler_Transcode_DBError(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT manuscript_id, title FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "title"}).AddRow(int64(1), "P1"))
	mock.ExpectQuery(`SELECT COALESCE\(source_video_url`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"source_video_url"}).AddRow(""))
	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT process_status FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"process_status"}).AddRow(int32(0)))
	mock.ExpectExec(`UPDATE videos SET process_status`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/transcode/100", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestAdminHandler_ResetVideo_DBError(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectExec(`UPDATE videos SET process_status`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/reset/100", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestAdminHandler_PublishManuscript_DBError(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT title FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("title"))
	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(0), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/publish/1", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestAdminHandler_ApproveWithProcess_DBError(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(0), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET review_status`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/1/approve-with-process", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestParsePageAdmin(t *testing.T) {
	r := httptest.NewRequest("GET", "/?page=2&page_size=50", nil)
	page, size := parsePageAdmin(r)
	assert.Equal(t, int32(2), page)
	assert.Equal(t, int32(50), size)

	r = httptest.NewRequest("GET", "/?page=0&page_size=0", nil)
	page, size = parsePageAdmin(r)
	assert.Equal(t, int32(1), page)
	assert.Equal(t, int32(20), size)

	r = httptest.NewRequest("GET", "/?page=0&page_size=200", nil)
	page, size = parsePageAdmin(r)
	assert.Equal(t, int32(1), page)
	assert.Equal(t, int32(20), size)
}

func timeNow() sql.NullTime {
	return sql.NullTime{Time: time.Now(), Valid: true}
}

func timeNowWithTime() interface{} { return time.Now() }
