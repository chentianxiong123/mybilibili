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

// ---------- helper: create service with sqlmock ----------

func newTestService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewRepository(db)
	svc := NewService(repo, NewHub())
	t.Cleanup(func() { db.Close() })
	return svc, mock
}

// ---------- Service 层测试 ----------

func TestService_CreateRoom_EmptyName(t *testing.T) {
	svc, _ := newTestService(t)
	room, err := svc.CreateRoom(t.Context(), "", "", "", 1001, 0)
	assert.Nil(t, room)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "room name required")
}

func TestService_ListLiveRooms_DefaultPagination(t *testing.T) {
	svc, mock := newTestService(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE status = .live.`).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}))

	// page=0 → page becomes 1, pageSize=0 → pageSize becomes 20
	rooms, err := svc.ListLiveRooms(t.Context(), 0, 0)
	require.NoError(t, err)
	assert.Empty(t, rooms)
	require.NoError(t, mock.ExpectationsWereMet())
	_ = now
}

func TestService_UpdateRoom_Success(t *testing.T) {
	svc, mock := newTestService(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "旧名字", 5001, "skey", "", "游戏", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET room_name=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.UpdateRoom(t.Context(), 1, 5001, "新名字", "", "音乐")
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_UpdateRoom_WrongUser(t *testing.T) {
	svc, mock := newTestService(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "测试", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))

	err := svc.UpdateRoom(t.Context(), 1, 9999, "新名字", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only host can update room")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_UpdateRoomStatus_WrongUser(t *testing.T) {
	svc, mock := newTestService(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "测试", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))

	err := svc.UpdateRoomStatus(t.Context(), 1, 9999, RoomStatusLive)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only host can update room status")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_ScheduleRoom_Success(t *testing.T) {
	svc, mock := newTestService(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "测试", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET scheduled_at=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	future := time.Now().Add(time.Hour)
	err := svc.ScheduleRoom(t.Context(), 1, 5001, &future)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_ScheduleRoom_NotFound(t *testing.T) {
	svc, mock := newTestService(t)

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	err := svc.ScheduleRoom(t.Context(), 999, 5001, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "room not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_SeatsToMap(t *testing.T) {
	seats := []*Seat{
		{SeatIndex: 0, UserID: 0, Status: SeatStatusEmpty, Muted: false},
		{SeatIndex: 1, UserID: 100, Status: SeatStatusOccupied, Muted: true},
		{SeatIndex: 2, UserID: 200, Status: SeatStatusPending, Muted: false},
	}
	result := seatsToMap(seats)
	require.Len(t, result, 3)

	assert.Equal(t, int32(0), result[0]["index"])
	assert.Equal(t, int32(SeatStatusEmpty), result[0]["status"])
	_, hasUID := result[0]["user_id"]
	assert.False(t, hasUID)

	assert.Equal(t, int32(1), result[1]["index"])
	assert.Equal(t, int64(100), result[1]["user_id"])
	assert.Equal(t, true, result[1]["muted"])

	assert.Equal(t, int32(2), result[2]["index"])
	assert.Equal(t, int64(200), result[2]["user_id"])
}

// ---------- HTTP Handler 层测试 ----------

func TestHandleCreateRoom_401(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/room", strings.NewReader(`{"room_name":"test"}`))
	req.Header.Del("X-User-Id")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandleCreateRoom_400_EmptyName(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/room", strings.NewReader(`{"room_name":""}`))
	req.Header.Set("X-User-Id", liveHostID)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleRoomByID_InvalidID(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/live/room/abc", nil)
	req.Header.Set("X-User-Id", liveHostID)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleRoomByID_My(t *testing.T) {
	h, mock := newTestLiveHandler(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE user_id = \$1`).
		WithArgs(int64(5001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "我的直播间", 5001, "skey", "", "游戏", RoomStatusLive, 0, now, now))

	rr := doLive(t, h, http.MethodGet, "/api/v1/live/room/my", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			RoomName string `json:"roomName"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "我的直播间", resp.Data.RoomName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRoomByID_My_401(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/live/room/my", nil)
	req.Header.Del("X-User-Id")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandleListRooms_405(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/room/list", nil)
	req.Header.Set("X-User-Id", liveHostID)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleSRSCallback_200(t *testing.T) {
	h, mock := newTestLiveHandler(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE stream_key = \$1`).
		WithArgs("test-stream-key").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "测试", 5001, "test-stream-key", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET status=\$1`).
		WithArgs(RoomStatusLive, int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	body := `{"action":"on_publish","stream":"test-stream-key","client_id":"c1","ip":"127.0.0.1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/srs/hook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"code":0`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSRSCallback_405(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/live/srs/hook", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleHealth_200(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/live/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "live ok")
}

// ---------- more HTTP handler edge cases ----------

func TestHandleRoom_405(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/live/room", nil)
	req.Header.Set("X-User-Id", liveHostID)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleCreateRoom_400_BadJSON(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/room", strings.NewReader(`{bad`))
	req.Header.Set("X-User-Id", liveHostID)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleRoomByID_Update_200(t *testing.T) {
	h, mock := newTestLiveHandler(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "旧名", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET room_name=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doLive(t, h, http.MethodPut, "/api/v1/live/room/1", `{"room_name":"新名","category":"音乐"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"status":"ok"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRoomByID_Update_401(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/live/room/1", strings.NewReader(`{"room_name":"x"}`))
	req.Header.Del("X-User-Id")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandleRoomByID_Schedule_200(t *testing.T) {
	h, mock := newTestLiveHandler(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "测试", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET scheduled_at=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doLive(t, h, http.MethodPut, "/api/v1/live/room/1/schedule", `{"scheduled_at":1700000000000}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"status":"ok"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRoomByID_Schedule_405_NotPUT(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/live/room/1/schedule", nil)
	req.Header.Set("X-User-Id", liveHostID)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleRoomByID_Status_405_NotPUT(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/live/room/1/status", nil)
	req.Header.Set("X-User-Id", liveHostID)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleRoomByID_My_404(t *testing.T) {
	h, mock := newTestLiveHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE user_id = \$1`).
		WithArgs(int64(5001)).
		WillReturnError(sql.ErrNoRows)

	rr := doLive(t, h, http.MethodGet, "/api/v1/live/room/my", "")
	assert.Equal(t, http.StatusNotFound, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRoomByID_DefaultMethod_405(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/live/room/1", nil)
	req.Header.Set("X-User-Id", liveHostID)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleRoomByID_Update_BadJSON(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/live/room/1", strings.NewReader(`{bad`))
	req.Header.Set("X-User-Id", liveHostID)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleSRSCallback_OtherAction(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	body := `{"action":"on_unpublish","stream":"key","client_id":"c1","ip":"127.0.0.1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/srs/hook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"code":0`)
}

func TestHandleSRSCallback_BadJSON(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/srs/hook", strings.NewReader(`{bad`))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleListRooms_WithParams(t *testing.T) {
	h, mock := newTestLiveHandler(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE status = .live.`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "A", 5001, "s", "", "", RoomStatusLive, 10, now, now))

	rr := doLive(t, h, http.MethodGet, "/api/v1/live/room/list?page=1&page_size=5", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Service 层: 座位相关测试 ----------

func newTestServiceWithSeats(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewRepository(db)
	svc := NewService(repo, NewHub())
	t.Cleanup(func() { db.Close() })
	return svc, mock
}

func roomRow() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
	})
}

func seatRow() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "room_id", "seat_index", "user_id", "status", "muted", "joined_at",
	})
}

func TestService_GetSeats(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusOccupied, false, now))

	seats, err := svc.GetSeats(t.Context(), 1)
	require.NoError(t, err)
	require.Len(t, seats, 1)
	assert.Equal(t, int64(100), seats[0].UserID)
}

func TestService_ApplySeat_Success(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().
			AddRow(1, 1, 0, 0, SeatStatusEmpty, false, sql.NullTime{}).
			AddRow(2, 1, 1, 0, SeatStatusEmpty, false, sql.NullTime{}))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	_ = now

	err := svc.ApplySeat(t.Context(), 1, 100)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_ApplySeat_AlreadyOccupied(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().
			AddRow(1, 1, 0, 100, SeatStatusOccupied, false, now))

	err := svc.ApplySeat(t.Context(), 1, 100)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_ApplySeat_NoAvailable(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().
			AddRow(1, 1, 0, 100, SeatStatusOccupied, false, now).
			AddRow(2, 1, 1, 200, SeatStatusPending, false, now))

	err := svc.ApplySeat(t.Context(), 1, 300)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no available seats")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_AcceptSeat_Success(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().
			AddRow(1, 1, 0, 100, SeatStatusPending, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.AcceptSeat(t.Context(), 1, 5001, 100)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_AcceptSeat_WrongUser(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))

	err := svc.AcceptSeat(t.Context(), 1, 9999, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only host can accept seat")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_AcceptSeat_NotFound(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow())

	err := svc.AcceptSeat(t.Context(), 1, 5001, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pending seat request not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_RejectSeat_Success(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().
			AddRow(1, 1, 0, 100, SeatStatusPending, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.RejectSeat(t.Context(), 1, 5001, 100)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_RejectSeat_WrongUser(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))

	err := svc.RejectSeat(t.Context(), 1, 9999, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only host can reject seat")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_RejectSeat_NotFound(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow())

	err := svc.RejectSeat(t.Context(), 1, 5001, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pending seat request not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_LeaveSeat_Success(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(100)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusOccupied, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 0, SeatStatusEmpty, false, sql.NullTime{}))

	err := svc.LeaveSeat(t.Context(), 1, 100)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_LeaveSeat_NotOnSeat(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(999)).
		WillReturnError(sql.ErrNoRows)

	err := svc.LeaveSeat(t.Context(), 1, 999)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_MuteSeat_Success(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusOccupied, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.MuteSeat(t.Context(), 1, 5001, 0, true)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_MuteSeat_Unmute(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusMuted, true, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.MuteSeat(t.Context(), 1, 5001, 0, false)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_MuteSeat_WrongUser(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))

	err := svc.MuteSeat(t.Context(), 1, 9999, 0, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only host can mute")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_MuteSeat_NotFound(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow())

	err := svc.MuteSeat(t.Context(), 1, 5001, 99, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "seat not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_LockSeat_Success(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 0, SeatStatusEmpty, false, sql.NullTime{}))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.LockSeat(t.Context(), 1, 5001, 0, true)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_LockSeat_Unlock(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 0, SeatStatusLocked, false, sql.NullTime{}))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.LockSeat(t.Context(), 1, 5001, 0, false)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_LockSeat_WrongUser(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))

	err := svc.LockSeat(t.Context(), 1, 9999, 0, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only host can lock seat")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_LockSeat_NotFound(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow())

	err := svc.LockSeat(t.Context(), 1, 5001, 99, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "seat not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_KickSeat_Success(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(100)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusOccupied, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 0, SeatStatusEmpty, false, sql.NullTime{}))

	err := svc.KickSeat(t.Context(), 1, 5001, 100)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_KickSeat_WrongUser(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))

	err := svc.KickSeat(t.Context(), 1, 9999, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only host can kick")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_KickSeat_UserNotOnSeat(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(100)).
		WillReturnError(sql.ErrNoRows)

	err := svc.KickSeat(t.Context(), 1, 5001, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not on seat")
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Service 层: Linkmic 测试 ----------

func TestService_CreateLinkmic_Success(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()
	mock.ExpectQuery(`INSERT INTO live_linkmic`).
		WithArgs(int64(1), int64(5001), int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))

	lm, err := svc.CreateLinkmic(t.Context(), 1, 5001, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), lm.ID)
	assert.Equal(t, int64(5001), lm.StreamerID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_EndLinkmic_Success(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	mock.ExpectExec(`UPDATE live_linkmic SET status=\$1`).
		WithArgs(int32(LinkmicStatusEnded), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.EndLinkmic(t.Context(), 1)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Hub 方法测试 ----------

func TestHub_SendToUser(t *testing.T) {
	hub := NewHub()
	hub.SendToUser(999, map[string]interface{}{"type": "test"})
}

func TestHub_SendToRoomExcept(t *testing.T) {
	hub := NewHub()
	hub.SendToRoomExcept(1, nil, map[string]interface{}{"type": "test"})
}

func TestHub_BroadcastRoom_Empty(t *testing.T) {
	hub := NewHub()
	hub.BroadcastRoom(1, map[string]interface{}{"type": "test"})
}

func TestWsMsgToMap(t *testing.T) {
	msg := wsMsg{Type: "test", RoomID: 1, SeatIndex: 2, TargetID: 3, Muted: true, Locked: false}
	m := wsMsgToMap(msg)
	assert.Equal(t, "test", m["type"])
	assert.Equal(t, int64(1), m["room_id"])
	assert.Equal(t, int32(2), m["seat_index"])
	assert.Equal(t, int64(3), m["target_id"])
	assert.Equal(t, true, m["muted"])
	assert.Equal(t, false, m["locked"])
}

func TestParseInt64(t *testing.T) {
	assert.Equal(t, int64(0), parseInt64(""))
	assert.Equal(t, int64(123), parseInt64("123"))
	assert.Equal(t, int64(42), parseInt64("abc42def"))
	assert.Equal(t, int64(0), parseInt64("abc"))
}

// ---------- Admin Handler 测试 ----------

func newTestAdminHandler(t *testing.T) (*AdminHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	h := NewAdminHandler(db)
	t.Cleanup(func() { db.Close() })
	return h, mock
}

func doAdmin(t *testing.T, h *AdminHandler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("X-Admin-Id", "1")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestAdminHandleRooms_200(t *testing.T) {
	h, mock := newTestAdminHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "status", "viewer_count", "created_at"}).
			AddRow(1, 5001, "test", "live", 10, now))

	rr := doAdmin(t, h, http.MethodGet, "/api/v1/live/admin/rooms")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandleStats_200(t *testing.T) {
	h, mock := newTestAdminHandler(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM live_rooms`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(`SELECT COALESCE\(SUM`).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(100))

	rr := doAdmin(t, h, http.MethodGet, "/api/v1/live/admin/stats")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandleRoomByID_200(t *testing.T) {
	h, mock := newTestAdminHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "status", "viewer_count", "created_at"}).
			AddRow(1, 5001, "test", "live", 10, now))

	rr := doAdmin(t, h, http.MethodGet, "/api/v1/live/admin/rooms/1")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandleRoomByID_404(t *testing.T) {
	h, mock := newTestAdminHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	rr := doAdmin(t, h, http.MethodGet, "/api/v1/live/admin/rooms/999")
	assert.Equal(t, http.StatusNotFound, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandleRoomStatus_200(t *testing.T) {
	h, mock := newTestAdminHandler(t)
	mock.ExpectExec(`UPDATE live_rooms SET status`).
		WithArgs("offline", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/live/admin/rooms/1/status", strings.NewReader(`{"status":"offline"}`))
	req.Header.Set("X-Admin-Id", "1")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandleRoute_InvalidID(t *testing.T) {
	h, _ := newTestAdminHandler(t)
	rr := doAdmin(t, h, http.MethodGet, "/api/v1/live/admin/rooms/abc/status")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminHandleRoute_NotFound(t *testing.T) {
	h, _ := newTestAdminHandler(t)
	rr := doAdmin(t, h, http.MethodGet, "/api/v1/live/admin/unknown")
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ---------- Linkmic Handler 测试 ----------

func newTestLinkmicHandler(t *testing.T) (*LinkmicHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewLinkmicRepository(db)
	hub := NewHub()
	rooms := NewRepository(db)
	svc := NewLinkmicService(repo, hub, rooms)
	h := NewLinkmicHandler(svc)
	t.Cleanup(func() { db.Close() })
	return h, mock
}

func doLinkmic(t *testing.T, h *LinkmicHandler, method, path, body string) *httptest.ResponseRecorder {
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

func TestLinkmicHandler_Apply_200(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`INSERT INTO live_linkmic`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "room_id", "streamer_id", "viewer_id", "status", "created_at"}).
			AddRow(1, 1, 5001, 100, LinkmicStatusApplying, now))

	rr := doLinkmic(t, h, http.MethodPost, "/api/v1/live/linkmic/apply/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicHandler_Accept_200(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	mock.ExpectExec(`UPDATE live_linkmic SET status`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doLinkmic(t, h, http.MethodPost, "/api/v1/live/linkmic/accept/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"status":"ok"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicHandler_Reject_200(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	mock.ExpectExec(`UPDATE live_linkmic SET status`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doLinkmic(t, h, http.MethodPost, "/api/v1/live/linkmic/reject/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"status":"ok"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicHandler_Disconnect_200(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	mock.ExpectExec(`UPDATE live_linkmic SET status`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doLinkmic(t, h, http.MethodPost, "/api/v1/live/linkmic/disconnect/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"status":"ok"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicHandler_Active_200(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_linkmic WHERE room_id = \$1 AND status IN`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "room_id", "streamer_id", "viewer_id", "status", "created_at"}).
			AddRow(1, 1, 5001, 100, LinkmicStatusConnected, now))

	rr := doLinkmic(t, h, http.MethodGet, "/api/v1/live/linkmic/active/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicHandler_Pending_200(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_linkmic WHERE room_id = \$1 AND status = 0`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "room_id", "streamer_id", "viewer_id", "status", "created_at"}).
			AddRow(1, 1, 5001, 100, LinkmicStatusApplying, now))

	rr := doLinkmic(t, h, http.MethodGet, "/api/v1/live/linkmic/pending/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicHandler_QueuePosition_200(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM live_linkmic`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	rr := doLinkmic(t, h, http.MethodGet, "/api/v1/live/linkmic/queue-position/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicHandler_ToggleAudio_200(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	mock.ExpectQuery(`UPDATE live_linkmic SET audio_enabled`).
		WillReturnRows(sqlmock.NewRows([]string{"audio_enabled"}).AddRow(1))

	rr := doLinkmic(t, h, http.MethodPost, "/api/v1/live/linkmic/toggle-audio/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicHandler_ToggleVideo_200(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	mock.ExpectQuery(`UPDATE live_linkmic SET video_enabled`).
		WillReturnRows(sqlmock.NewRows([]string{"video_enabled"}).AddRow(1))

	rr := doLinkmic(t, h, http.MethodPost, "/api/v1/live/linkmic/toggle-video/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Admin Handler: handleRoomStatus 边界测试 ----------

func TestAdminHandleRoomStatus_405_NotPUT(t *testing.T) {
	h, _ := newTestAdminHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/live/admin/rooms/1/status", nil)
	req.Header.Set("X-Admin-Id", "1")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAdminHandleRoomStatus_BadJSON(t *testing.T) {
	h, mock := newTestAdminHandler(t)
	mock.ExpectExec(`UPDATE live_rooms SET status`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/live/admin/rooms/1/status", strings.NewReader(`{bad`))
	req.Header.Set("X-Admin-Id", "1")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	// Bad JSON → req.Status is empty string, but Exec still succeeds
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandleRoomStatus_ExecError(t *testing.T) {
	h, mock := newTestAdminHandler(t)
	mock.ExpectExec(`UPDATE live_rooms SET status`).
		WillReturnError(sql.ErrConnDone)

	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/live/admin/rooms/1/status", strings.NewReader(`{"status":"live"}`))
	req.Header.Set("X-Admin-Id", "1")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Admin Handler: handleRooms 边界测试 ----------

func TestAdminHandleRooms_FilterStatus(t *testing.T) {
	h, mock := newTestAdminHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms`).
		WithArgs("live", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "status", "viewer_count", "created_at"}).
			AddRow(1, 5001, "直播中", "live", 50, now))

	rr := doAdmin(t, h, http.MethodGet, "/api/v1/live/admin/rooms?status=live")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandleRooms_EmptyResult(t *testing.T) {
	h, mock := newTestAdminHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "status", "viewer_count", "created_at"}))

	rr := doAdmin(t, h, http.MethodGet, "/api/v1/live/admin/rooms")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandleRooms_DBError(t *testing.T) {
	h, mock := newTestAdminHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms`).
		WillReturnError(sql.ErrConnDone)

	rr := doAdmin(t, h, http.MethodGet, "/api/v1/live/admin/rooms")
	assert.Equal(t, http.StatusOK, rr.Code)
	// DB error → returns empty list via WriteOK
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Admin Handler: handleStats 边界测试 ----------

func TestAdminHandleStats_QueryErrors(t *testing.T) {
	h, mock := newTestAdminHandler(t)
	// Both queries fail gracefully
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM live_rooms`).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectQuery(`SELECT COALESCE\(SUM`).
		WillReturnError(sql.ErrConnDone)

	rr := doAdmin(t, h, http.MethodGet, "/api/v1/live/admin/stats")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Admin Handler: handleRoute 额外测试 ----------

func TestAdminHandleRoute_StatsPOST_404(t *testing.T) {
	h, _ := newTestAdminHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/admin/stats", nil)
	req.Header.Set("X-Admin-Id", "1")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAdminHandleRoute_RoomsPOST_404(t *testing.T) {
	h, _ := newTestAdminHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/admin/rooms", nil)
	req.Header.Set("X-Admin-Id", "1")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAdminHandleRoute_RoomStatusNotPUT(t *testing.T) {
	h, _ := newTestAdminHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/live/admin/rooms/1/status", nil)
	req.Header.Set("X-Admin-Id", "1")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAdminHandleRoute_UnknownPath(t *testing.T) {
	h, _ := newTestAdminHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/live/admin/unknown/path", nil)
	req.Header.Set("X-Admin-Id", "1")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	// parts[1] = "path" fails ParseInt → 400
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ---------- Admin Handler: handleRoomByID 额外测试 ----------

func TestAdminHandleRoomByID_DBError(t *testing.T) {
	h, mock := newTestAdminHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	rr := doAdmin(t, h, http.MethodGet, "/api/v1/live/admin/rooms/1")
	assert.Equal(t, http.StatusNotFound, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Admin Handler: handleRoute 额外路径测试 ----------

func TestAdminHandleRoute_RoomByIDNotGET(t *testing.T) {
	h, _ := newTestAdminHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/admin/rooms/1", nil)
	req.Header.Set("X-Admin-Id", "1")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ---------- Cover Upload 测试 ----------

func TestHandleCoverUpload_405(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/live/room/cover", nil)
	req.Header.Set("X-User-Id", liveHostID)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleCoverUpload_NoFile(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/room/cover", strings.NewReader("not-a-form"))
	req.Header.Set("X-User-Id", liveHostID)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	// parse form fails or file missing
	assert.True(t, rr.Code == http.StatusBadRequest || rr.Code == http.StatusOK)
}

func TestHandleCoverUpload_WithFile(t *testing.T) {
	h, mock := newTestLiveHandler(t)
	// mock UpdateRoom for when roomID > 0
	mock.ExpectExec(`UPDATE live_rooms SET room_name=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)

	// Create a minimal multipart form with a file
	body := &strings.Builder{}
	boundary := "----TestBoundary"
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString(`Content-Disposition: form-data; name="file"; filename="test.jpg"` + "\r\n")
	body.WriteString("Content-Type: image/jpeg\r\n\r\n")
	body.WriteString("\xff\xd8\xff\xe0") // JPEG magic bytes
	body.WriteString("\r\n--" + boundary + "--\r\n")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/room/cover?roomId=1", strings.NewReader(body.String()))
	req.Header.Set("X-User-Id", liveHostID)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "uploaded")
}

func TestHandleCoverUpload_NoRoomID(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	body := &strings.Builder{}
	boundary := "----TestBoundary"
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString(`Content-Disposition: form-data; name="file"; filename="test.jpg"` + "\r\n")
	body.WriteString("Content-Type: image/jpeg\r\n\r\n")
	body.WriteString("\xff\xd8\xff\xe0")
	body.WriteString("\r\n--" + boundary + "--\r\n")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/room/cover", strings.NewReader(body.String()))
	req.Header.Set("X-User-Id", liveHostID)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "uploaded")
}

func TestHandleCoverUpload_BadMultipart(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/room/cover", strings.NewReader("not valid"))
	req.Header.Set("X-User-Id", liveHostID)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.True(t, rr.Code == http.StatusBadRequest || rr.Code == http.StatusOK)
}