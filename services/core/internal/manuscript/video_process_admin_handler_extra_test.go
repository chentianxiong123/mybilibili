package manuscript

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- handleStream with real flusher writer: bridge events from hub ---

func TestVideoProcessAdminHandler_HandleStream_Bridge(t *testing.T) {
	h, mock := newVideoProcessHandler(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// buildSnapshot: current query returns no rows
	mock.ExpectQuery(`SELECT .+ FROM videos v JOIN manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "title", "stage_text", "process_status", "process_progress",
		}))
	// collectStats
	mock.ExpectQuery(`SELECT process_status, COUNT\(\*\) FROM videos`).
		WillReturnRows(sqlmock.NewRows([]string{"process_status", "count"}))

	done := make(chan string, 1)
	go func() {
		req := httptest.NewRequest("GET", "/api/v1/video/process/admin/stream", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		h.handleStream(w, req)
		done <- w.Body.String()
	}()

	// wait a little for the snapshot event to be written, then push events
	time.Sleep(100 * time.Millisecond)
	h.Hub().Broadcast(ProgressEvt{VideoID: 1, Stage: "TRANSCODING", Progress: 50})
	time.Sleep(100 * time.Millisecond)
	h.Hub().Broadcast(ProgressEvt{Done: true, VideoID: 1, Error: ""})
	time.Sleep(100 * time.Millisecond)
	h.Hub().Broadcast(ProgressEvt{Error: "boom", VideoID: 1})
	time.Sleep(100 * time.Millisecond)

	cancel()
	body := <-done
	assert.True(t, strings.Contains(body, "event: snapshot"), "missing snapshot event")
	assert.True(t, strings.Contains(body, "event: progress"), "missing progress event")
	assert.True(t, strings.Contains(body, "event: complete"), "missing complete event")
	assert.True(t, strings.Contains(body, "event: error"), "missing error event")
}

// --- collectStats mapping branches: statuses 11/21/31/41 ---

func TestVideoProcessAdminHandler_CollectStats_AllBranches(t *testing.T) {
	h, mock := newVideoProcessHandler(t)
	mock.ExpectQuery(`SELECT process_status, COUNT\(\*\) FROM videos`).
		WillReturnRows(sqlmock.NewRows([]string{"process_status", "count"}).
			AddRow(0, int64(1)).AddRow(11, int64(2)).AddRow(21, int64(3)).
			AddRow(31, int64(4)).AddRow(41, int64(5)).AddRow(5, int64(6)))

	stats := h.collectStats(context.Background())
	assert.Equal(t, 1, stats["pending"])
	assert.Equal(t, 2, stats["transcoding"])
	assert.Equal(t, 3, stats["audioExtracting"])
	assert.Equal(t, 4, stats["subtitleGenerating"])
	assert.Equal(t, 5, stats["aiSummarizing"])
	assert.Equal(t, 6, stats["completed"])
}

// --- buildSnapshot with stats error ---

func TestVideoProcessAdminHandler_BuildSnapshot_StatsError(t *testing.T) {
	h, mock := newVideoProcessHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM videos v JOIN manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "title", "stage_text", "process_status", "process_progress",
		}))
	mock.ExpectQuery(`SELECT process_status, COUNT\(\*\) FROM videos`).
		WillReturnError(sql.ErrConnDone)

	payload := h.buildSnapshot(context.Background())
	require.NotNil(t, payload)
	assert.True(t, strings.Contains(string(payload), `"processing":false`))
}

// --- buildSnapshot with current row present ---

func TestVideoProcessAdminHandler_BuildSnapshot_CurrentRow(t *testing.T) {
	h, mock := newVideoProcessHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM videos v JOIN manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "title", "stage_text", "process_status", "process_progress",
		}).AddRow(int64(7), int64(1), "t", "视频转码中", int32(1), int32(42)))
	mock.ExpectQuery(`SELECT process_status, COUNT\(\*\) FROM videos`).
		WillReturnRows(sqlmock.NewRows([]string{"process_status", "count"}).AddRow(5, int64(9)))

	payload := h.buildSnapshot(context.Background())
	require.NotNil(t, payload)
	assert.True(t, strings.Contains(string(payload), `"processing":true`))
	assert.True(t, strings.Contains(string(payload), `"videoId":7`))
}

// --- eventName branches ---

func TestEventName_Branches(t *testing.T) {
	assert.Equal(t, "error", eventName(ProgressEvt{Error: "boom"}))
	assert.Equal(t, "complete", eventName(ProgressEvt{Done: true}))
	assert.Equal(t, "progress", eventName(ProgressEvt{}))
}

// --- hub broadcast with multiple clients incl. a slow one ---

func TestSSEHub_Broadcast_MixedClients(t *testing.T) {
	hub := newSSEHub()

	fast := make(chan ProgressEvt, 16)
	slow := make(chan ProgressEvt) // unbuffered, never read -> dropped
	hub.add(fast)
	hub.add(slow)

	var wg sync.WaitGroup
	// concurrent broadcasts with slow client present must not deadlock
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			hub.Broadcast(ProgressEvt{VideoID: int64(i)})
		}(i)
	}
	wg.Wait()
	hub.remove(fast)
	hub.remove(slow)
	assert.Empty(t, hub.clients)
}
