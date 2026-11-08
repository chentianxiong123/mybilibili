package live

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- handleRoom 错误分支 ----------

func TestHandleCreateRoom_400_CreateError(t *testing.T) {
	h, mock := newTestLiveHandler(t)
	mock.ExpectQuery(`INSERT INTO live_rooms`).
		WillReturnError(sql.ErrConnDone)

	rr := doLive(t, h, http.MethodPost, "/api/v1/live/room", `{"room_name":"测试直播间"}`)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- handleRoomByID 错误分支 ----------

func TestHandleRoomByID_EmptyPath_404(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	rr := doLive(t, h, http.MethodGet, "/api/v1/live/room/", "")
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandleRoomByID_My_NotGet_405(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	rr := doLive(t, h, http.MethodPost, "/api/v1/live/room/my", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleRoomByID_My_WithQuery_200(t *testing.T) {
	h, mock := newTestLiveHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE user_id = \$1`).
		WithArgs(int64(5001)).
		WillReturnRows(roomRow().AddRow(1, "我的直播间", 5001, "skey", "", "游戏", RoomStatusLive, 0, now, now))

	rr := doLive(t, h, http.MethodGet, "/api/v1/live/room/my?foo=bar", "")
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

func TestHandleRoomByID_Status_403(t *testing.T) {
	h, mock := newTestLiveHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "测试", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET status=\$1`).
		WillReturnError(sql.ErrConnDone)

	rr := doLive(t, h, http.MethodPut, "/api/v1/live/room/1/status", `{"status":"live"}`)
	assert.Equal(t, http.StatusForbidden, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRoomByID_Status_WrongUser_403(t *testing.T) {
	h, mock := newTestLiveHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "测试", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))

	// 请求的用户 9999 非房主 5001 → 403（无需触发 UPDATE）
	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/live/room/1/status", strings.NewReader(`{"status":"live"}`))
	req.Header.Set("X-User-Id", "9999")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRoomByID_Schedule_403(t *testing.T) {
	h, mock := newTestLiveHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "测试", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET scheduled_at=\$1`).
		WillReturnError(sql.ErrConnDone)

	rr := doLive(t, h, http.MethodPut, "/api/v1/live/room/1/schedule", `{"scheduled_at":1700000000000}`)
	assert.Equal(t, http.StatusForbidden, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRoomByID_Schedule_NilSchedule_200(t *testing.T) {
	h, mock := newTestLiveHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "测试", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET scheduled_at=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doLive(t, h, http.MethodPut, "/api/v1/live/room/1/schedule", `{"scheduled_at":null}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"status":"ok"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRoomByID_Update_403(t *testing.T) {
	h, mock := newTestLiveHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "旧名", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET room_name=\$1`).
		WillReturnError(sql.ErrConnDone)

	rr := doLive(t, h, http.MethodPut, "/api/v1/live/room/1", `{"room_name":"新名"}`)
	assert.Equal(t, http.StatusForbidden, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- handleListRooms 错误分支 ----------

func TestHandleListRooms_500(t *testing.T) {
	h, mock := newTestLiveHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE status = 'live'`).
		WillReturnError(sql.ErrConnDone)

	rr := doLive(t, h, http.MethodGet, "/api/v1/live/room/list", "")
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- handleSRSCallback 错误分支 ----------

func TestHandleSRSCallback_Failed_200(t *testing.T) {
	h, mock := newTestLiveHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE stream_key = \$1`).
		WithArgs("badkey").
		WillReturnError(sql.ErrNoRows)

	mux := http.NewServeMux()
	h.Register(mux)
	body := `{"action":"on_publish","stream":"badkey"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/srs/hook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"code":-1`)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- handleCoverUpload 错误分支 ----------

func TestHandleCoverUpload_ParseError_400(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/room/cover", strings.NewReader("plain text body"))
	req.Header.Set("X-User-Id", liveHostID)
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleCoverUpload_FileRequired_400(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	// 合法的 multipart 表单但没有 file 字段
	body := &strings.Builder{}
	boundary := "----TestBoundaryNoFileField"
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString(`Content-Disposition: form-data; name="other"` + "\r\n\r\n")
	body.WriteString("value\r\n")
	body.WriteString("--" + boundary + "--\r\n")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/room/cover", strings.NewReader(body.String()))
	req.Header.Set("X-User-Id", liveHostID)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleCoverUpload_NoExt(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	// 文件名无扩展名 → 默认 .jpg
	body := &strings.Builder{}
	boundary := "----TestBoundaryNoExt"
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString(`Content-Disposition: form-data; name="file"; filename="cover"` + "\r\n")
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

func TestHandleCoverUpload_CreateFileError_500(t *testing.T) {
	dir := "/tmp/mybilibili-uploads/live-covers"
	parent := filepath.Dir(dir)
	// 清理残留后，把 live-covers 变成一个普通文件，使 MkdirAll/Create 失败
	_ = os.RemoveAll(dir)
	require.NoError(t, os.MkdirAll(parent, 0o755))
	require.NoError(t, os.WriteFile(dir, []byte("block"), 0o644))
	t.Cleanup(func() { _ = os.RemoveAll(parent) })

	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	body := &strings.Builder{}
	boundary := "----TestBoundaryCreateErr"
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
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandleCoverUpload_NoRoomID_BadJSONPath(t *testing.T) {
	h, _ := newTestLiveHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	body := &strings.Builder{}
	boundary := "----TestBoundaryNoRoom"
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString(`Content-Disposition: form-data; name="file"; filename="a.png"` + "\r\n")
	body.WriteString("Content-Type: image/png\r\n\r\n")
	body.WriteString("\x89PNG\r\n")
	body.WriteString("\r\n--" + boundary + "--\r\n")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/room/cover", strings.NewReader(body.String()))
	req.Header.Set("X-User-Id", liveHostID)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "uploaded")
}
