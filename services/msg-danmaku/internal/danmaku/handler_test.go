package danmaku

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDanmaku(t *testing.T) (*HTTPHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewDanmakuRepository(db)
	svc := NewDanmakuService(repo, NewDanmakuBroadcaster())
	h := NewHTTPHandler(svc, NewDanmakuBroadcaster())
	t.Cleanup(func() { db.Close() })
	return h, mock
}

func doDanmaku(t *testing.T, h *HTTPHandler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	h.Register(mux)
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("X-User-Id", "1001")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestHandleDanmakuSend_200(t *testing.T) {
	h, mock := newTestDanmaku(t)

	// Send -> Create (INSERT ... RETURNING id)
	mock.ExpectQuery(`INSERT INTO danmaku`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

	rr := doDanmaku(t, h, http.MethodPost, "/api/v1/danmaku/send",
		`{"video_id":99,"manuscript_id":0,"content":"测试弹幕","time":1.5,"color":"#fff","mode":1}`)
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			ID      int64  `json:"id"`
			VideoID int64  `json:"video_id"`
			Content string `json:"content"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, int64(7), resp.Data.ID)
	assert.Equal(t, int64(99), resp.Data.VideoID)
	assert.Equal(t, "测试弹幕", resp.Data.Content)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDanmakuVideo_200(t *testing.T) {
	h, mock := newTestDanmaku(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT id, video_id, manuscript_id, user_id, content, time, color, mode, created_at FROM danmaku WHERE video_id = \$1 ORDER BY time`).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at",
		}).AddRow(1, 99, 1, 1001, "第一条", 0.5, "#fff", 1, now).
			AddRow(2, 99, 1, 1002, "第二条", 1.0, "#ff0", 1, now))

	rr := doDanmaku(t, h, http.MethodGet, "/api/v1/danmaku/video/99", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID      int64  `json:"id"`
			Content string `json:"content"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 2)
	assert.Equal(t, "第一条", resp.Data[0].Content)
	assert.Equal(t, "第二条", resp.Data[1].Content)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDanmakuDelete_200(t *testing.T) {
	h, mock := newTestDanmaku(t)

	mock.ExpectExec(`DELETE FROM danmaku WHERE id = \$1 AND user_id = \$2`).
		WithArgs(int64(7), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doDanmaku(t, h, http.MethodDelete, "/api/v1/danmaku/7", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDanmakuBatchCount_200(t *testing.T) {
	h, mock := newTestDanmaku(t)

	mock.ExpectQuery(`SELECT manuscript_id, COUNT\(\*\) FROM danmaku`).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "count"}).
			AddRow(10, 5).AddRow(20, 9))

	rr := doDanmaku(t, h, http.MethodGet, "/api/v1/danmaku/batch-count?ids=10,20,30", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int            `json:"code"`
		Data map[string]int64 `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, int64(5), resp.Data["10"])
	assert.Equal(t, int64(9), resp.Data["20"])
	assert.Equal(t, int64(0), resp.Data["30"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDanmakuSend_400(t *testing.T) {
	h, _ := newTestDanmaku(t)

	rr := doDanmaku(t, h, http.MethodPost, "/api/v1/danmaku/send",
		`{"video_id":0,"manuscript_id":0,"content":"","time":0,"color":"#fff","mode":1}`)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 400, resp.Code)
}

func TestHandleDanmakuVideo_Empty(t *testing.T) {
	h, mock := newTestDanmaku(t)

	mock.ExpectQuery(`SELECT id, video_id, manuscript_id, user_id, content, time, color, mode, created_at FROM danmaku WHERE video_id = \$1 ORDER BY time`).
		WithArgs(int64(888)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at",
		}))

	rr := doDanmaku(t, h, http.MethodGet, "/api/v1/danmaku/video/888", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDanmakuVideo_TimeRange(t *testing.T) {
	h, mock := newTestDanmaku(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT id, video_id, manuscript_id, user_id, content, time, color, mode, created_at FROM danmaku WHERE video_id = \$1 AND time >= \$2 AND time <= \$3`).
		WithArgs(int64(99), 0.5, 2.0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at",
		}).AddRow(1, 99, 1, 1001, "弹幕", 1.0, "#fff", 1, now))

	rr := doDanmaku(t, h, http.MethodGet, "/api/v1/danmaku/video/99?start_time=0.5&end_time=2.0", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDanmakuTrend_200(t *testing.T) {
	h, mock := newTestDanmaku(t)

	mock.ExpectQuery(`SELECT TO_CHAR`).
		WillReturnRows(sqlmock.NewRows([]string{"day", "count"}).
			AddRow("2024-01-01", 5).AddRow("2024-01-02", 8))

	rr := doDanmaku(t, h, http.MethodGet, "/api/v1/danmaku/trend?ids=10,20&start_date=2024-01-01&end_date=2024-01-31", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int            `json:"code"`
		Data map[string]int `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 5, resp.Data["2024-01-01"])
	assert.Equal(t, 8, resp.Data["2024-01-02"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDanmakuTrend_400(t *testing.T) {
	h, _ := newTestDanmaku(t)

	rr := doDanmaku(t, h, http.MethodGet, "/api/v1/danmaku/trend", "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleDanmakuCreatorList_200(t *testing.T) {
	h, mock := newTestDanmaku(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT danmaku\.id, danmaku\.video_id`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at",
		}).AddRow(1, 10, 10, 1001, "弹幕", 1.0, "#fff", 1, now))

	rr := doDanmaku(t, h, http.MethodGet, "/api/v1/creator/danmaku/list?page=1&page_size=20", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			List  []map[string]interface{} `json:"list"`
			Total int64                    `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, int64(1), resp.Data.Total)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDanmakuCreatorList_WithVideoID(t *testing.T) {
	h, mock := newTestDanmaku(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT danmaku\.id, danmaku\.video_id`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at",
		}).AddRow(1, 50, 10, 1001, "弹幕", 1.0, "#fff", 1, now))

	rr := doDanmaku(t, h, http.MethodGet, "/api/v1/creator/danmaku/list?video_id=50&page=1&page_size=20", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDanmakuCreatorDelete_200(t *testing.T) {
	h, mock := newTestDanmaku(t)

	mock.ExpectExec(`DELETE FROM danmaku WHERE id = \$1 AND video_id IN`).
		WithArgs(int64(5), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doDanmaku(t, h, http.MethodDelete, "/api/v1/creator/danmaku/5", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDanmakuByPath_405(t *testing.T) {
	h, _ := newTestDanmaku(t)

	rr := doDanmaku(t, h, http.MethodPut, "/api/v1/danmaku/99", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleDanmakuByPath_InvalidID(t *testing.T) {
	h, _ := newTestDanmaku(t)

	rr := doDanmaku(t, h, http.MethodGet, "/api/v1/danmaku/abc", "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleDanmakuCreatorByPath_405(t *testing.T) {
	h, _ := newTestDanmaku(t)

	rr := doDanmaku(t, h, http.MethodPut, "/api/v1/creator/danmaku/99", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleDanmakuSend_MethodNotAllowed(t *testing.T) {
	h, _ := newTestDanmaku(t)

	rr := doDanmaku(t, h, http.MethodGet, "/api/v1/danmaku/send", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleDanmakuSend_InvalidJSON(t *testing.T) {
	h, _ := newTestDanmaku(t)

	rr := doDanmaku(t, h, http.MethodPost, "/api/v1/danmaku/send", "not json")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleDanmakuBatchCount_400(t *testing.T) {
	h, _ := newTestDanmaku(t)

	rr := doDanmaku(t, h, http.MethodGet, "/api/v1/danmaku/batch-count", "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDanmakuRepository_UpsertDailyMetric(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewDanmakuRepository(db)

	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).
		WithArgs(int64(10), int64(1001), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpsertDailyMetric(context.Background(), 10, 1001, "danmaku_count", 1)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDanmakuRepository_UpsertDailyMetric_DifferentField(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewDanmakuRepository(db)

	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).
		WithArgs(int64(20), int64(2002), 5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpsertDailyMetric(context.Background(), 20, 2002, "view_count", 5)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDanmakuRepository_CountByVideo(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewDanmakuRepository(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku WHERE video_id`).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))

	count, err := repo.CountByVideo(context.Background(), 99)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDanmakuRepository_CountByVideo_Zero(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewDanmakuRepository(db)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku WHERE video_id`).
		WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	count, err := repo.CountByVideo(context.Background(), 999)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDanmakuRepository_CountByManuscriptIDs_Empty(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewDanmakuRepository(db)

	result, err := repo.CountByManuscriptIDs(context.Background(), []int64{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestDanmakuRepository_CountByManuscriptIDs_Multiple(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewDanmakuRepository(db)

	mock.ExpectQuery(`SELECT manuscript_id, COUNT\(\*\) FROM danmaku`).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "count"}).
			AddRow(10, 5).AddRow(30, 12))

	result, err := repo.CountByManuscriptIDs(context.Background(), []int64{10, 20, 30})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), result[10])
	assert.Equal(t, int64(0), result[20])
	assert.Equal(t, int64(12), result[30])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDanmakuService_Broadcaster(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDanmakuRepository(db)
	broadcaster := NewDanmakuBroadcaster()
	svc := NewDanmakuService(repo, broadcaster)

	// Test Broadcaster() accessor
	assert.NotNil(t, svc.Broadcaster())

	// Test Subscribe
	ch := broadcaster.Subscribe(100)
	assert.NotNil(t, ch)

	// Test Broadcast - should send to subscriber
	event := &DanmakuEvent{
		ID: 1, VideoID: 100, UserID: 1001,
		Content: "hello", Time: 1.0, Color: "#fff", Mode: 1,
		CreatedAt: "2024-01-01T00:00:00Z",
	}
	broadcaster.Broadcast(100, event)

	received := <-ch
	assert.Equal(t, int64(1), received.ID)
	assert.Equal(t, "hello", received.Content)

	// Test Broadcast to non-subscribed video (no panic)
	broadcaster.Broadcast(999, event)

	// Test Subscribe same video returns same channel
	ch2 := broadcaster.Subscribe(100)
	assert.Equal(t, ch, ch2)

	// Test Unsubscribe (no-op, just ensure no panic)
	broadcaster.Unsubscribe(100, ch)
}

func TestDanmakuService_Broadcaster_BufferFull(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDanmakuRepository(db)
	broadcaster := NewDanmakuBroadcaster()
	_ = NewDanmakuService(repo, broadcaster)

	ch := broadcaster.Subscribe(200)

	// Fill the buffer (capacity 100)
	for i := 0; i < 100; i++ {
		broadcaster.Broadcast(200, &DanmakuEvent{ID: int64(i)})
	}
	// 101st should not block (non-blocking send)
	broadcaster.Broadcast(200, &DanmakuEvent{ID: 999})

	// Should still be able to read
	received := <-ch
	assert.Equal(t, int64(0), received.ID)
}

func TestDanmakuService_CountByVideo(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDanmakuRepository(db)
	svc := NewDanmakuService(repo, NewDanmakuBroadcaster())

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku WHERE video_id`).
		WithArgs(int64(55)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	count, err := svc.CountByVideo(context.Background(), 55)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSSE_InvalidVideoID(t *testing.T) {
	h, _ := newTestDanmaku(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/sse/danmaku?video_id=abc", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleSSE_Events(t *testing.T) {
	h, _ := newTestDanmaku(t)
	mux := http.NewServeMux()
	h.Register(mux)

	// Subscribe first to get the channel
	videoID := int64(777)
	_ = h.broadcaster.Subscribe(videoID)

	// Start SSE in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/sse/danmaku?video_id=777", nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	rr.Header().Set("Content-Type", "text/event-stream")

	done := make(chan struct{})
	go func() {
		mux.ServeHTTP(rr, req)
		close(done)
	}()

	// Send an event
	event := &DanmakuEvent{ID: 1, VideoID: 777, Content: "sse test"}
	h.broadcaster.Broadcast(videoID, event)

	// Wait a bit for the event to be sent, then cancel
	time.Sleep(100 * time.Millisecond)
	cancel()
	<-done

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleDanmakuByPath_Delete_Unauthorized(t *testing.T) {
	h, _ := newTestDanmaku(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/danmaku/99", nil)
	// No X-User-Id header
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandleDanmakuCreatorByPath_Delete_Unauthorized(t *testing.T) {
	h, _ := newTestDanmaku(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/creator/danmaku/99", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandleDanmakuSend_Unauthorized(t *testing.T) {
	h, _ := newTestDanmaku(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/danmaku/send",
		strings.NewReader(`{"video_id":1,"content":"test","time":0,"color":"#fff","mode":1}`))
	// No X-User-Id header
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandleDanmakuCreatorList_Unauthorized(t *testing.T) {
	h, _ := newTestDanmaku(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/danmaku/list?page=1&page_size=20", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}