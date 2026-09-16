package live

import (
	"database/sql"
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

const liveHostID = "5001"

func newTestLiveHandler(t *testing.T) (*HTTPHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewRepository(db)
	svc := NewService(repo, NewHub())
	h := NewHTTPHandler(svc, NewHub())
	t.Cleanup(func() { db.Close() })
	return h, mock
}

func doLive(t *testing.T, h *HTTPHandler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	h.Register(mux)
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("X-User-Id", liveHostID)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestHandleCreateRoom_200(t *testing.T) {
	h, mock := newTestLiveHandler(t)

	now := time.Now()
	mock.ExpectQuery(`INSERT INTO live_rooms`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, now, now))

	rr := doLive(t, h, http.MethodPost, "/api/v1/live/room", `{"room_name":"测试直播间"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			ID       int64  `json:"id"`
			RoomName string `json:"roomName"`
			Status   string `json:"status"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, int64(1), resp.Data.ID)
	assert.Equal(t, "测试直播间", resp.Data.RoomName)
	assert.Equal(t, RoomStatusIdle, resp.Data.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetRoom_200(t *testing.T) {
	h, mock := newTestLiveHandler(t)

	now := time.Now()
	mock.ExpectQuery(`FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "测试直播间", 5001, "skey", "", "游戏", RoomStatusLive, 3, now, now))

	rr := doLive(t, h, http.MethodGet, "/api/v1/live/room/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			ID       int64  `json:"id"`
			RoomName string `json:"roomName"`
			Status   string `json:"status"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "测试直播间", resp.Data.RoomName)
	assert.Equal(t, RoomStatusLive, resp.Data.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetRoom_404(t *testing.T) {
	h, mock := newTestLiveHandler(t)

	mock.ExpectQuery(`FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	rr := doLive(t, h, http.MethodGet, "/api/v1/live/room/999", "")
	assert.Equal(t, http.StatusNotFound, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRoomList_200(t *testing.T) {
	h, mock := newTestLiveHandler(t)

	now := time.Now()
	mock.ExpectQuery(`FROM live_rooms WHERE status = 'live'`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "直播间A", 5001, "skey1", "", "游戏", RoomStatusLive, 10, now, now).
			AddRow(2, "直播间B", 5002, "skey2", "", "音乐", RoomStatusLive, 5, now, now))

	rr := doLive(t, h, http.MethodGet, "/api/v1/live/room/list", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID   int64  `json:"id"`
			Name string `json:"roomName"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 2)
	assert.Equal(t, "直播间A", resp.Data[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRoomStatus_200(t *testing.T) {
	h, mock := newTestLiveHandler(t)

	now := time.Now()
	// UpdateRoomStatus 先查房间校验 host，再更新状态
	mock.ExpectQuery(`FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "测试直播间", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET status=\$1`).
		WithArgs(RoomStatusLive, int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doLive(t, h, http.MethodPut, "/api/v1/live/room/1/status", `{"status":"live"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"status":"ok"`)
	require.NoError(t, mock.ExpectationsWereMet())
}