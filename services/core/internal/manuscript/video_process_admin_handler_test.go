package manuscript

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newVideoProcessHandler(t *testing.T) (*VideoProcessAdminHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewVideoProcessAdminHandler(db), mock
}

func TestVideoProcessAdminHandler_Hub(t *testing.T) {
	h, _ := newVideoProcessHandler(t)
	assert.NotNil(t, h.Hub())
}

func TestVideoProcessAdminHandler_HandleCurrent_NoRows(t *testing.T) {
	h, mock := newVideoProcessHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM videos v JOIN manuscripts m`).WillReturnRows(sqlmock.NewRows([]string{
		"id", "manuscript_id", "title", "process_stage", "stage_text", "process_status", "process_progress",
	}))

	req := httptest.NewRequest("GET", "/api/v1/video/process/admin/current", nil)
	w := httptest.NewRecorder()
	h.handleCurrent(w, req)
	assert.Equal(t, 200, w.Code)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	assert.Equal(t, false, data["processing"])
}

func TestVideoProcessAdminHandler_HandleCurrent_WithData(t *testing.T) {
	h, mock := newVideoProcessHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM videos v JOIN manuscripts m`).WillReturnRows(sqlmock.NewRows([]string{
		"id", "manuscript_id", "title", "process_stage", "stage_text", "process_status", "process_progress",
	}).AddRow(int64(100), int64(1), "P1", "TRANSCODING", "视频转码中", int32(1), int32(50)))

	req := httptest.NewRequest("GET", "/api/v1/video/process/admin/current", nil)
	w := httptest.NewRecorder()
	h.handleCurrent(w, req)
	assert.Equal(t, 200, w.Code)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	assert.Equal(t, true, data["processing"])
}

func TestVideoProcessAdminHandler_HandleCurrent_DBError(t *testing.T) {
	h, mock := newVideoProcessHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM videos v JOIN manuscripts m`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/video/process/admin/current", nil)
	w := httptest.NewRecorder()
	h.handleCurrent(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestVideoProcessAdminHandler_HandleQueue(t *testing.T) {
	h, mock := newVideoProcessHandler(t)

	mock.ExpectQuery(`SELECT process_status, COUNT\(\*\) FROM videos`).WillReturnRows(sqlmock.NewRows([]string{
		"process_status", "count",
	}).AddRow(0, int64(5)).AddRow(11, int64(3)).AddRow(21, int64(2)).AddRow(31, int64(1)))

	req := httptest.NewRequest("GET", "/api/v1/video/process/admin/queue", nil)
	w := httptest.NewRecorder()
	h.handleQueue(w, req)
	assert.Equal(t, 200, w.Code)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	assert.Equal(t, float64(5), data["waitingTranscode"])
	assert.Equal(t, float64(3), data["waitingAudio"])
	assert.Equal(t, float64(2), data["waitingSubtitle"])
	assert.Equal(t, float64(1), data["waitingAi"])
	assert.Equal(t, float64(11), data["queueSize"])
}

func TestVideoProcessAdminHandler_HandleQueue_DBError(t *testing.T) {
	h, mock := newVideoProcessHandler(t)

	mock.ExpectQuery(`SELECT process_status, COUNT\(\*\) FROM videos`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/video/process/admin/queue", nil)
	w := httptest.NewRecorder()
	h.handleQueue(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestVideoProcessAdminHandler_HandleStatistics(t *testing.T) {
	h, mock := newVideoProcessHandler(t)

	mock.ExpectQuery(`SELECT process_status, COUNT\(\*\) FROM videos`).WillReturnRows(sqlmock.NewRows([]string{
		"process_status", "count",
	}).AddRow(0, int64(10)).AddRow(11, int64(5)).AddRow(21, int64(3)).AddRow(31, int64(2)).AddRow(41, int64(1)).AddRow(5, int64(20)))

	req := httptest.NewRequest("GET", "/api/v1/video/process/admin/statistics", nil)
	w := httptest.NewRecorder()
	h.handleStatistics(w, req)
	assert.Equal(t, 200, w.Code)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	assert.Equal(t, float64(10), data["pending"])
	assert.Equal(t, float64(5), data["transcoding"])
	assert.Equal(t, float64(3), data["audioExtracting"])
	assert.Equal(t, float64(2), data["subtitleGenerating"])
	assert.Equal(t, float64(1), data["aiSummarizing"])
	assert.Equal(t, float64(20), data["completed"])
}

func TestVideoProcessAdminHandler_HandleStatistics_DBError(t *testing.T) {
	h, mock := newVideoProcessHandler(t)

	mock.ExpectQuery(`SELECT process_status, COUNT\(\*\) FROM videos`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/video/process/admin/statistics", nil)
	w := httptest.NewRecorder()
	h.handleStatistics(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestVideoProcessAdminHandler_Register(t *testing.T) {
	h, _ := newVideoProcessHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	assert.NotNil(t, mux)
}

func TestNewSSEHub(t *testing.T) {
	hub := newSSEHub()
	assert.NotNil(t, hub)
	assert.NotNil(t, hub.clients)

	ch := make(chan ProgressEvt, 16)
	hub.add(ch)
	assert.Len(t, hub.clients, 1)

	hub.remove(ch)
	assert.Len(t, hub.clients, 0)
}

func TestSSEHub_Broadcast(t *testing.T) {
	hub := newSSEHub()

	ch := make(chan ProgressEvt, 16)
	hub.add(ch)

	evt := ProgressEvt{VideoID: 1, Stage: "TRANSCODING", Progress: 50}
	hub.Broadcast(evt)

	received := <-ch
	assert.Equal(t, int64(1), received.VideoID)
	assert.Equal(t, "TRANSCODING", received.Stage)
	assert.Equal(t, int32(50), received.Progress)
}

func TestSSEHub_BroadcastSlow(t *testing.T) {
	hub := newSSEHub()

	ch := make(chan ProgressEvt) // unbuffered - will block
	hub.add(ch)

	// Should not block
	evt := ProgressEvt{VideoID: 1}
	hub.Broadcast(evt)

	hub.remove(ch)
}

func TestEventName(t *testing.T) {
	assert.Equal(t, "error", eventName(ProgressEvt{Error: "err"}))
	assert.Equal(t, "complete", eventName(ProgressEvt{Done: true}))
	assert.Equal(t, "progress", eventName(ProgressEvt{}))
}

func TestVideoProcessAdminHandler_BuildSnapshot(t *testing.T) {
	h, mock := newVideoProcessHandler(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM videos v JOIN manuscripts m`).WillReturnRows(sqlmock.NewRows([]string{
		"id", "manuscript_id", "title", "stage_text", "process_status", "process_progress",
	}))

	mock.ExpectQuery(`SELECT process_status, COUNT\(\*\) FROM videos`).WillReturnRows(sqlmock.NewRows([]string{
		"process_status", "count",
	}))

	payload := h.buildSnapshot(ctx)
	assert.NotNil(t, payload)

	var result map[string]any
	json.Unmarshal(payload, &result)
	assert.Contains(t, result, "current")
	assert.Contains(t, result, "statistics")
}

func TestVideoProcessAdminHandler_BuildSnapshot_WithData(t *testing.T) {
	h, mock := newVideoProcessHandler(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM videos v JOIN manuscripts m`).WillReturnRows(sqlmock.NewRows([]string{
		"id", "manuscript_id", "title", "stage_text", "process_status", "process_progress",
	}).AddRow(int64(100), int64(1), "P1", "视频转码中", int32(1), int32(50)))

	mock.ExpectQuery(`SELECT process_status, COUNT\(\*\) FROM videos`).WillReturnRows(sqlmock.NewRows([]string{
		"process_status", "count",
	}).AddRow(0, int64(5)))

	payload := h.buildSnapshot(ctx)
	var result map[string]any
	json.Unmarshal(payload, &result)

	current := result["current"].(map[string]any)
	assert.Equal(t, true, current["processing"])
	assert.Equal(t, float64(100), current["videoId"])
}

func TestVideoProcessAdminHandler_CollectStats(t *testing.T) {
	h, mock := newVideoProcessHandler(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT process_status, COUNT\(\*\) FROM videos`).WillReturnRows(sqlmock.NewRows([]string{
		"process_status", "count",
	}).AddRow(0, int64(5)).AddRow(5, int64(10)))

	stats := h.collectStats(ctx)
	assert.Equal(t, 5, stats["pending"])
	assert.Equal(t, 10, stats["completed"])
}

func TestVideoProcessAdminHandler_CollectStats_DBError(t *testing.T) {
	h, mock := newVideoProcessHandler(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT process_status, COUNT\(\*\) FROM videos`).WillReturnError(sql.ErrConnDone)

	stats := h.collectStats(ctx)
	assert.Equal(t, 0, stats["pending"])
}
