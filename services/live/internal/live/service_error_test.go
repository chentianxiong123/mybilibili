package live

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- CreateRoom 错误分支 ----------

func TestService_CreateRoom_RepoError(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`INSERT INTO live_rooms`).
		WillReturnError(sql.ErrConnDone)

	room, err := svc.CreateRoom(t.Context(), "直播间", "", "游戏", 5001, 0)
	assert.Error(t, err)
	assert.Nil(t, room)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- UpdateRoom 边界分支 ----------

func TestService_UpdateRoom_WithCoverAndName(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "旧名", 5001, "skey", "old-cover", "游戏", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET room_name=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// cover != "" 分支（L78-L80）应被执行
	err := svc.UpdateRoom(t.Context(), 1, 5001, "新名", "new-cover", "音乐")
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_UpdateRoom_OnlyCategory(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "旧名", 5001, "skey", "", "游戏", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET room_name=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// roomName=="" 与 cover=="" 分支（跳过赋值）应被执行
	err := svc.UpdateRoom(t.Context(), 1, 5001, "", "", "音乐")
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- ApplySeat 错误分支 ----------

func TestService_ApplySeat_GetSeatsError(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	err := svc.ApplySeat(t.Context(), 1, 100)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_ApplySeat_UpdateSeatError(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 0, SeatStatusEmpty, false, sql.NullTime{}))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnError(sql.ErrConnDone)

	err := svc.ApplySeat(t.Context(), 1, 100)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- AcceptSeat 错误分支 ----------

func TestService_AcceptSeat_RoomNotFound(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	err := svc.AcceptSeat(t.Context(), 1, 5001, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "room not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_AcceptSeat_GetSeatsError(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	err := svc.AcceptSeat(t.Context(), 1, 5001, 100)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_AcceptSeat_UpdateSeatError(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusPending, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnError(sql.ErrConnDone)

	err := svc.AcceptSeat(t.Context(), 1, 5001, 100)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- RejectSeat 错误分支 ----------

func TestService_RejectSeat_RoomNotFound(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	err := svc.RejectSeat(t.Context(), 1, 5001, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "room not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_RejectSeat_GetSeatsError(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	err := svc.RejectSeat(t.Context(), 1, 5001, 100)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_RejectSeat_UpdateSeatError(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusPending, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnError(sql.ErrConnDone)

	err := svc.RejectSeat(t.Context(), 1, 5001, 100)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- LeaveSeat 错误分支 ----------

func TestService_LeaveSeat_GetSeatByUserError(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(100)).
		WillReturnError(sql.ErrConnDone)

	err := svc.LeaveSeat(t.Context(), 1, 100)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_LeaveSeat_UpdateSeatError(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(100)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusOccupied, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnError(sql.ErrConnDone)

	err := svc.LeaveSeat(t.Context(), 1, 100)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- MuteSeat 错误分支 ----------

func TestService_MuteSeat_RoomNotFound(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	err := svc.MuteSeat(t.Context(), 1, 5001, 0, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "room not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_MuteSeat_GetSeatsError(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	err := svc.MuteSeat(t.Context(), 1, 5001, 0, true)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_MuteSeat_UpdateSeatError(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusOccupied, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnError(sql.ErrConnDone)

	err := svc.MuteSeat(t.Context(), 1, 5001, 0, true)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- LockSeat 错误分支 ----------

func TestService_LockSeat_RoomNotFound(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	err := svc.LockSeat(t.Context(), 1, 5001, 0, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "room not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_LockSeat_GetSeatsError(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	err := svc.LockSeat(t.Context(), 1, 5001, 0, true)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_LockSeat_UpdateSeatError(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusOccupied, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnError(sql.ErrConnDone)

	err := svc.LockSeat(t.Context(), 1, 5001, 0, true)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- KickSeat 错误分支 ----------

func TestService_KickSeat_RoomNotFound(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	err := svc.KickSeat(t.Context(), 1, 5001, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "room not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_KickSeat_UpdateSeatError(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(100)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusOccupied, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnError(sql.ErrConnDone)

	err := svc.KickSeat(t.Context(), 1, 5001, 100)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- GetRoom 错误分支 ----------

func TestService_GetRoom_NotFound(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	room, err := svc.GetRoom(t.Context(), 999)
	assert.Error(t, err)
	assert.Nil(t, room)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_ListLiveRooms_Error(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE status = .live.`).
		WillReturnError(sql.ErrConnDone)

	rooms, err := svc.ListLiveRooms(t.Context(), 1, 20)
	assert.Error(t, err)
	assert.Nil(t, rooms)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- HandleSRSCallback 补充 ----------

func TestService_HandleSRSCallback_EmptyAction(t *testing.T) {
	svc, _ := newTestService(t)
	err := svc.HandleSRSCallback(t.Context(), "", "skey123")
	assert.NoError(t, err)
}
