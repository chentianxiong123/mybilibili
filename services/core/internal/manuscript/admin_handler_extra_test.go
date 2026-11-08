package manuscript

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- eventWriter() lazy init ---

func TestAdminHandler_LazyEventWriter(t *testing.T) {
	db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	h := NewManuscriptAdminHandler(db)
	// no writer injected; eventWriter() must lazily create one
	ew := h.eventWriter()
	assert.NotNil(t, ew)
	assert.NotNil(t, h.events)
}

func TestAdminHandler_LazyEventWriter_Idempotent(t *testing.T) {
	db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	h := NewManuscriptAdminHandler(db)
	ew1 := h.eventWriter()
	ew2 := h.eventWriter()
	assert.Equal(t, ew1, ew2)
}

func TestNewManuscriptEventWriter_SatisfiesInterface(t *testing.T) {
	db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	var w ManuscriptEventWriter = NewManuscriptEventWriter(db)
	assert.NotNil(t, w)
}

// --- requirePerm branches via Register + ServeMux ---

func TestAdminHandler_RequirePerm_ForbiddenWithAdminID(t *testing.T) {
	db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	h := NewManuscriptAdminHandler(db)
	h.SetEventWriter(&fakeEventWriter{})
	h.SetPermChecker(&fakePermChecker{ok: false})

	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/pending", nil)
	req.Header.Set("X-Admin-Id", "4") // admin id present but perm check failed -> 403
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)
}

func TestAdminHandler_RequirePerm_UnauthorizedNoAdminID(t *testing.T) {
	db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	h := NewManuscriptAdminHandler(db)
	h.SetEventWriter(&fakeEventWriter{})
	h.SetPermChecker(&fakePermChecker{ok: false})

	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/pending", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestAdminHandler_RequirePerm_PermOK_CallsNext(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	h := NewManuscriptAdminHandler(db)
	h.SetEventWriter(&fakeEventWriter{})
	h.SetPermChecker(&fakePermChecker{ok: true})

	mock.ExpectQuery(`SELECT .+ FROM manuscripts m WHERE m.review_status = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "user_id", "category_id", "status", "review_status",
			"view_count", "like_count", "comment_count", "upload_time", "updated_at",
		}))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE review_status = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/pending", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

// --- listByStatus: db error path ---

func TestAdminHandler_ListByStatus_DBError(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts m WHERE m.review_status = \$1`).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/processing", nil)
	w := httptest.NewRecorder()
	h.handleProcessing(w, req)
	// on db error it writes "ok" with empty list
	assert.Equal(t, 200, w.Code)
}

// --- listByStatus: row scan error branch ---

func TestAdminHandler_ListByStatus_ScanError(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts m WHERE m.review_status = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "user_id", "category_id", "status", "review_status",
			"view_count", "like_count", "comment_count", "upload_time", "updated_at",
		}).AddRow("bad-id", "t", int64(10), int64(1), int32(0), int32(0),
			int64(0), int64(0), int64(0), timeNow(), timeNow()))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE review_status = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/pending", nil)
	w := httptest.NewRecorder()
	h.handlePending(w, req)
	// scan error -> row skipped, still 200
	assert.Equal(t, 200, w.Code)
}

// --- triggerAllVideoProcess: via reviewManuscript approve, with actual video rows ---

func TestAdminHandler_TriggerAllVideoProcess_WithData(t *testing.T) {
	h, mock := newAdminHandler(t)

	// reviewManuscript(approved=true, autoProcess=false)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(0), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	// triggerAllVideoProcess: videos exist
	mock.ExpectQuery(`SELECT id FROM videos WHERE manuscript_id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(100)).AddRow(int64(101)))
	for vid := int64(100); vid <= 101; vid++ {
		mock.ExpectQuery(`SELECT process_status FROM videos WHERE id`).WithArgs(vid).
			WillReturnRows(sqlmock.NewRows([]string{"process_status"}).AddRow(int32(0)))
		mock.ExpectExec(`UPDATE videos SET process_status`).WillReturnResult(sqlmock.NewResult(0, 1))
	}

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/approve/1", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_TriggerAllVideoProcess_QueryError(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(0), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))
	// videos query fails -> silent return
	mock.ExpectQuery(`SELECT id FROM videos WHERE manuscript_id`).WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/approve/1", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

// --- triggerVideoProcess: zero manuscript id branch (skips manuscript update) ---

func TestAdminHandler_TriggerVideoProcess_ZeroManuscriptID(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	// manuscript_id = 0 -> no manuscript status update
	mock.ExpectQuery(`SELECT manuscript_id, title FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "title"}).AddRow(int64(0), "P1"))
	mock.ExpectQuery(`SELECT COALESCE\(source_video_url`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"source"}).AddRow(""))
	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id`).WithArgs(int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"})) // no rows -> uid stays 0
	mock.ExpectQuery(`SELECT process_status FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"process_status"}).AddRow(int32(0)))
	mock.ExpectExec(`UPDATE videos SET process_status`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/transcode/100", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

// --- setManuscriptStatus: publisher unpublish branch ---

func TestAdminHandler_SetManuscriptStatus_Unpublish(t *testing.T) {
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

func TestAdminHandler_SetManuscriptStatus_Publish_NotificationError(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT title FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("my ms"))
	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(0), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))
	// notification insert fails (silently ignored)
	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectExec(`INSERT INTO notifications`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/publish/1", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

// --- getVideos: row scan error branch ---

func TestAdminHandler_GetVideos_ScanError(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT id, video_order, title, play_url_hd, play_url_sd, play_url_ld, upload_time`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "video_order", "title", "play_url_hd", "play_url_sd", "play_url_ld", "upload_time",
		}).AddRow(int64(7), 0, "P1", "hd.mp4", "sd.mp4", "ld.mp4", timeNow()).
			AddRow("not-an-int", 1, "P2", "x.mp4", "y.mp4", "z.mp4", timeNow()))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/1/videos", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	// scan error -> row skipped, still 200
	assert.Equal(t, 200, w.Code)
}

// --- handleByID: /{id}/publish & /{id}/unpublish sub-resource actions ---

func TestAdminHandler_ByID_PublishSubResource(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT title FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("my ms"))
	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(0), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectExec(`INSERT INTO notifications`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/1/publish", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_ByID_UnpublishSubResource(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT title FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("my ms"))
	mock.ExpectQuery(`SELECT status, user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "user_id"}).AddRow(int32(3), int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/1/unpublish", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 200, w.Code)
}

// --- handleAll: branches ---

func TestAdminHandler_HandleAll_NoFilter(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts m ORDER BY m.id DESC`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "user_id", "category_id", "status", "review_status",
			"view_count", "like_count", "comment_count", "upload_time", "updated_at",
		}).AddRow(1, "t", int64(10), int64(1), int32(3), int32(1),
			int64(0), int64(0), int64(0), timeNow(), timeNow()))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/all", nil)
	w := httptest.NewRecorder()
	h.handleAll(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_HandleAll_StatusParseFallback(t *testing.T) {
	h, mock := newAdminHandler(t)

	// status=abc -> Atoi fails -> 0
	mock.ExpectQuery(`SELECT .+ FROM manuscripts m WHERE m.status = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "user_id", "category_id", "status", "review_status",
			"view_count", "like_count", "comment_count", "upload_time", "updated_at",
		}))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/all?status=abc", nil)
	w := httptest.NewRecorder()
	h.handleAll(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAdminHandler_HandleAll_ScanError(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts m WHERE m.title ILIKE`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "user_id", "category_id", "status", "review_status",
			"view_count", "like_count", "comment_count", "upload_time", "updated_at",
		}).AddRow("bad-id", "t", int64(10), int64(1), int32(3), int32(1),
			int64(0), int64(0), int64(0), timeNow(), timeNow()))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/admin/all?keyword=foo", nil)
	w := httptest.NewRecorder()
	h.handleAll(w, req)
	assert.Equal(t, 200, w.Code)
}

// --- resetVideo: found but update succeeds after exists check is implicit ---

func TestAdminHandler_ResetVideo_FoundNoRows(t *testing.T) {
	h, mock := newAdminHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM videos WHERE id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/admin/reset/100", nil)
	w := httptest.NewRecorder()
	h.handleByID(w, req)
	assert.Equal(t, 404, w.Code)
}

// --- sseHub concurrency test ---

func TestSSEHub_Broadcast_Concurrent(t *testing.T) {
	hub := newSSEHub()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		ch := make(chan ProgressEvt, 16)
		hub.add(ch)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				hub.Broadcast(ProgressEvt{VideoID: int64(j)})
			}
		}()
	}
	time.Sleep(20 * time.Millisecond)
	wg.Wait()
	// concurrent add/remove + broadcast must not deadlock or race
	assert.GreaterOrEqual(t, len(hub.clients), 8)
}

// --- handleStream: non-flusher writer path (500) ---

func TestVideoProcessAdminHandler_HandleStream_NotFlusher(t *testing.T) {
	h, _ := newVideoProcessHandler(t)

	nonFlusher := &plainResponseWriter{}
	req := httptest.NewRequest("GET", "/api/v1/video/process/admin/stream", nil)
	h.handleStream(nonFlusher, req)
	assert.Equal(t, 500, nonFlusher.status)
}

type plainResponseWriter struct {
	status int
	buf    []byte
}

func (p *plainResponseWriter) Header() http.Header  { return http.Header{} }
func (p *plainResponseWriter) WriteHeader(code int) { p.status = code }
func (p *plainResponseWriter) Write(b []byte) (int, error) {
	p.buf = append(p.buf, b...)
	return len(b), nil
}
