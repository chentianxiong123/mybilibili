package danmaku

import (
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