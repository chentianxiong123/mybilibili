package live

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- LinkmicService.Apply 错误分支 ----------

func newLinkmicServiceEnv(t *testing.T) (*LinkmicService, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewLinkmicRepository(db)
	rooms := NewRepository(db)
	svc := NewLinkmicService(repo, NewHub(), rooms)
	t.Cleanup(func() { db.Close() })
	return svc, mock
}

func TestLinkmicService_Apply_RoomError(t *testing.T) {
	svc, mock := newLinkmicServiceEnv(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	lm, err := svc.Apply(context.Background(), 1, 100)
	assert.Error(t, err)
	assert.Nil(t, lm)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicService_Apply_RepoError(t *testing.T) {
	svc, mock := newLinkmicServiceEnv(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`INSERT INTO live_linkmic`).
		WillReturnError(sql.ErrConnDone)

	lm, err := svc.Apply(context.Background(), 1, 100)
	assert.Error(t, err)
	assert.Nil(t, lm)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicService_Apply_Success(t *testing.T) {
	svc, mock := newLinkmicServiceEnv(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`INSERT INTO live_linkmic`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "room_id", "streamer_id", "viewer_id", "status", "created_at"}).
			AddRow(1, 1, 5001, 100, LinkmicStatusApplying, now))

	lm, err := svc.Apply(context.Background(), 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), lm.ID)
	assert.Equal(t, int64(5001), lm.StreamerID)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- LinkmicService 其余方法 ----------

func TestLinkmicService_Active_Delegation(t *testing.T) {
	svc, mock := newLinkmicServiceEnv(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_linkmic WHERE room_id = \$1 AND status IN`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "room_id", "streamer_id", "viewer_id", "status", "created_at"}).
			AddRow(1, 5, 5001, 100, LinkmicStatusConnected, now))

	list, err := svc.Active(context.Background(), 5)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, LinkmicStatusConnected, list[0].Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicService_Pending_Delegation(t *testing.T) {
	svc, mock := newLinkmicServiceEnv(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_linkmic WHERE room_id = \$1 AND status = 0`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "room_id", "streamer_id", "viewer_id", "status", "created_at"}).
			AddRow(1, 5, 5001, 100, LinkmicStatusApplying, now))

	list, err := svc.Pending(context.Background(), 5)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, LinkmicStatusApplying, list[0].Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicService_QueuePosition_Delegation(t *testing.T) {
	svc, mock := newLinkmicServiceEnv(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM live_linkmic`).
		WithArgs(int64(5), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))

	pos, err := svc.QueuePosition(context.Background(), 5, 7)
	require.NoError(t, err)
	assert.Equal(t, 5, pos)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicService_ToggleAudio_Delegation(t *testing.T) {
	svc, mock := newLinkmicServiceEnv(t)
	mock.ExpectQuery(`UPDATE live_linkmic SET audio_enabled`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"audio_enabled"}).AddRow(int32(1)))

	enabled, err := svc.ToggleAudio(context.Background(), 9)
	require.NoError(t, err)
	assert.Equal(t, int32(1), enabled)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicService_ToggleVideo_Delegation(t *testing.T) {
	svc, mock := newLinkmicServiceEnv(t)
	mock.ExpectQuery(`UPDATE live_linkmic SET video_enabled`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"video_enabled"}).AddRow(int32(0)))

	enabled, err := svc.ToggleVideo(context.Background(), 9)
	require.NoError(t, err)
	assert.Equal(t, int32(0), enabled)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- LinkmicHandler 错误分支 ----------

func TestLinkmicHandler_Apply_Error(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	rr := doLinkmic(t, h, http.MethodPost, "/api/v1/live/linkmic/apply/1", "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicHandler_ToggleAudio_Error(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	mock.ExpectQuery(`UPDATE live_linkmic SET audio_enabled`).
		WillReturnError(sql.ErrConnDone)

	rr := doLinkmic(t, h, http.MethodPost, "/api/v1/live/linkmic/toggle-audio/1", "")
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicHandler_ToggleVideo_Error(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	mock.ExpectQuery(`UPDATE live_linkmic SET video_enabled`).
		WillReturnError(sql.ErrConnDone)

	rr := doLinkmic(t, h, http.MethodPost, "/api/v1/live/linkmic/toggle-video/1", "")
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- LinkmicHandler 未知路径（不匹配任何 case，默认 200 空响应） ----------

func TestLinkmicHandler_UnknownPath(t *testing.T) {
	h, _ := newTestLinkmicHandler(t)
	rr := doLinkmic(t, h, http.MethodPost, "/api/v1/live/linkmic/unknown/1", "")
	// 无匹配 case → 不写任何响应，默认 200
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestLinkmicHandler_Apply_BadID(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	// roomID 解析失败被忽略（ParseInt 返回 0），随后 Apply 查询房间 → 报错 → 400
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(0)).
		WillReturnError(sql.ErrNoRows)

	rr := doLinkmic(t, h, http.MethodPost, "/api/v1/live/linkmic/apply/abc", "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
