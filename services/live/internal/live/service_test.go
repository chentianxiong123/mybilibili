package live

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreateRoom_Success(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`INSERT INTO live_rooms`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, now, now))

	room, err := svc.CreateRoom(t.Context(), "新直播间", "", "游戏", 5001, 0)
	require.NoError(t, err)
	assert.NotNil(t, room)
	assert.Equal(t, "新直播间", room.RoomName)
	assert.Equal(t, RoomStatusIdle, room.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetRoom_Success(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "测试", 5001, "skey", "", "游戏", RoomStatusLive, 5, now, now))

	room, err := svc.GetRoom(t.Context(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), room.ID)
	assert.Equal(t, RoomStatusLive, room.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetRoomByHost_Success(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE user_id = \$1`).
		WithArgs(int64(5001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "我的直播间", 5001, "skey", "", "音乐", RoomStatusIdle, 0, now, now))

	room, err := svc.GetRoomByHost(t.Context(), 5001)
	require.NoError(t, err)
	assert.Equal(t, int64(1), room.ID)
	assert.Equal(t, int64(5001), room.UserID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetRoomByHost_NotFound(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE user_id = \$1`).
		WithArgs(int64(9999)).
		WillReturnError(sql.ErrNoRows)

	room, err := svc.GetRoomByHost(t.Context(), 9999)
	assert.Error(t, err)
	assert.Nil(t, room)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_UpdateRoomStatus_Success(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "测试", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET status=\$1`).
		WithArgs(RoomStatusLive, int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.UpdateRoomStatus(t.Context(), 1, 5001, RoomStatusLive)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_UpdateRoomStatus_NotFound(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	err := svc.UpdateRoomStatus(t.Context(), 999, 5001, RoomStatusLive)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "room not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_UpdateRoom_NotFound(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	err := svc.UpdateRoom(t.Context(), 999, 5001, "new", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "room not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_HandleSRSCallback_OnPublish_Success(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE stream_key = \$1`).
		WithArgs("skey123").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "测试", 5001, "skey123", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET status=\$1`).
		WithArgs(RoomStatusLive, int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.HandleSRSCallback(t.Context(), "on_publish", "skey123")
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_HandleSRSCallback_OtherAction(t *testing.T) {
	svc, _ := newTestService(t)
	err := svc.HandleSRSCallback(t.Context(), "on_unpublish", "skey123")
	assert.NoError(t, err)
}

func TestService_HandleSRSCallback_InvalidStreamKey(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE stream_key = \$1`).
		WithArgs("badkey").
		WillReturnError(sql.ErrNoRows)

	err := svc.HandleSRSCallback(t.Context(), "on_publish", "badkey")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid stream key")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_HandleSRSCallback_UpdateStatusError(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE stream_key = \$1`).
		WithArgs("skey123").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "测试", 5001, "skey123", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET status=\$1`).
		WillReturnError(sql.ErrConnDone)

	err := svc.HandleSRSCallback(t.Context(), "on_publish", "skey123")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_ScheduleRoom_WrongUser(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "测试", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))

	future := time.Now().Add(time.Hour)
	err := svc.ScheduleRoom(t.Context(), 1, 9999, &future)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only host can schedule room")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_ScheduleRoom_ClearSchedule(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "测试", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET scheduled_at=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.ScheduleRoom(t.Context(), 1, 5001, nil)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_ListLiveRooms_CustomPagination(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE status = .live.`).
		WithArgs(5, 10).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "A", 5001, "s", "", "", RoomStatusLive, 10, now, now))

	rooms, err := svc.ListLiveRooms(t.Context(), 3, 5)
	require.NoError(t, err)
	assert.Len(t, rooms, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_ListLiveRooms_PageSizeTooLarge(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	// pageSize > 50 → defaults to 20
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE status = .live.`).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}))

	_, err := svc.ListLiveRooms(t.Context(), 1, 100)
	assert.NoError(t, err)
	_ = now
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_ApplySeat_BroadcastsUpdate(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().
			AddRow(1, 1, 0, 0, SeatStatusEmpty, false, sql.NullTime{}))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.ApplySeat(t.Context(), 1, 100)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_LeaveSeat_BroadcastsUpdate(t *testing.T) {
	svc, mock := newTestService(t)
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

func TestService_KickSeat_BroadcastsAndNotifies(t *testing.T) {
	svc, mock := newTestService(t)
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

func TestService_AcceptSeat_Success_Broadcasts(t *testing.T) {
	svc, mock := newTestService(t)
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

func TestService_RejectSeat_Success_Broadcasts(t *testing.T) {
	svc, mock := newTestService(t)
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

func TestService_MuteSeat_BroadcastsUpdate(t *testing.T) {
	svc, mock := newTestService(t)
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

func TestService_LockSeat_BroadcastsUpdate(t *testing.T) {
	svc, mock := newTestService(t)
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

func TestService_SeatsToMap_Empty(t *testing.T) {
	result := seatsToMap([]*Seat{})
	assert.Empty(t, result)
}

func TestService_SeatsToMap_WithZeros(t *testing.T) {
	seats := []*Seat{
		{SeatIndex: 0, UserID: 0, Status: SeatStatusEmpty, Muted: false},
	}
	result := seatsToMap(seats)
	require.Len(t, result, 1)
	_, hasUID := result[0]["user_id"]
	assert.False(t, hasUID)
	assert.Equal(t, int32(0), result[0]["status"])
	assert.Equal(t, false, result[0]["muted"])
}

// ---------- Repository 层测试 ----------

func TestRepository_IncrementViewerCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_rooms SET viewer_count = viewer_count \+ 1`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.IncrementViewerCount(t.Context(), 1)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DecrementViewerCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_rooms SET viewer_count = GREATEST\(viewer_count - 1, 0\)`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DecrementViewerCount(t.Context(), 1)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_InitSeats(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`INSERT INTO live_seats`).
		WillReturnResult(sqlmock.NewResult(0, 3))

	err = repo.InitSeats(t.Context(), 1, 3)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetSeats_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	seats, err := repo.GetSeats(t.Context(), 1)
	assert.Error(t, err)
	assert.Nil(t, seats)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetSeats_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow())

	seats, err := repo.GetSeats(t.Context(), 1)
	assert.NoError(t, err)
	assert.Empty(t, seats)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateSeat(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	seat := &Seat{
		RoomID:    1,
		SeatIndex: 0,
		UserID:    100,
		Status:    SeatStatusOccupied,
		Muted:     false,
		JoinedAt:  sql.NullTime{Valid: false},
	}
	err = repo.UpdateSeat(t.Context(), seat)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateSeat_WithUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	seat := &Seat{
		RoomID:    1,
		SeatIndex: 0,
		UserID:    100,
		Status:    SeatStatusOccupied,
		Muted:     false,
		JoinedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	}
	err = repo.UpdateSeat(t.Context(), seat)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetSeatByUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(100)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusOccupied, false, now))

	seat, err := repo.GetSeatByUser(t.Context(), 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(100), seat.UserID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetSeatByUser_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(999)).
		WillReturnError(sql.ErrNoRows)

	seat, err := repo.GetSeatByUser(t.Context(), 1, 999)
	assert.Error(t, err)
	assert.Nil(t, seat)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateRoomStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_rooms SET status=\$1, updated_at=NOW\(\) WHERE id=\$2`).
		WithArgs("live", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateRoomStatus(t.Context(), 1, "live")
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ScheduleRoom(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_rooms SET scheduled_at=\$1, updated_at=NOW\(\) WHERE id=\$2`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.ScheduleRoom(t.Context(), 1, sql.NullTime{Valid: false})
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateLinkmicStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_linkmic SET status=\$1 WHERE id=\$2`).
		WithArgs(int32(1), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateLinkmicStatus(t.Context(), 1, 1)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListLiveRooms_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE status = 'live'`).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}))

	rooms, err := repo.ListLiveRooms(t.Context(), 1, 20)
	assert.NoError(t, err)
	assert.Empty(t, rooms)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListLiveRooms_WithRooms(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE status = 'live'`).
		WithArgs(5, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "A", 5001, "s", "", "", RoomStatusLive, 10, now, now).
			AddRow(2, "B", 5002, "s2", "", "", RoomStatusLive, 5, now, now))

	rooms, err := repo.ListLiveRooms(t.Context(), 1, 5)
	assert.NoError(t, err)
	assert.Len(t, rooms, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_CreateRoom_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO live_rooms`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, now, now))

	room := &Room{RoomName: "test", UserID: 5001, CoverURL: "", Category: "game"}
	err = repo.CreateRoom(t.Context(), room)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), room.ID)
	assert.NotEmpty(t, room.StreamKey)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateRoom_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_rooms SET room_name=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	room := &Room{ID: 1, RoomName: "new", CoverURL: "url", Category: "cat"}
	err = repo.UpdateRoom(t.Context(), room)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetRoomByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "test", 5001, "skey", "", "game", RoomStatusLive, 5, now, now))

	room, err := repo.GetRoomByID(t.Context(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), room.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetRoomByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	room, err := repo.GetRoomByID(t.Context(), 999)
	assert.Error(t, err)
	assert.Nil(t, room)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetRoomByStreamKey_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE stream_key = \$1`).
		WithArgs("skey123").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "test", 5001, "skey123", "", "", RoomStatusIdle, 0, now, now))

	room, err := repo.GetRoomByStreamKey(t.Context(), "skey123")
	require.NoError(t, err)
	assert.Equal(t, "skey123", room.StreamKey)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetRoomByStreamKey_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE stream_key = \$1`).
		WithArgs("bad").
		WillReturnError(sql.ErrNoRows)

	room, err := repo.GetRoomByStreamKey(t.Context(), "bad")
	assert.Error(t, err)
	assert.Nil(t, room)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetRoomByUserID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE user_id = \$1`).
		WithArgs(int64(5001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "my room", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))

	room, err := repo.GetRoomByUserID(t.Context(), 5001)
	require.NoError(t, err)
	assert.Equal(t, int64(5001), room.UserID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetRoomByUserID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE user_id = \$1`).
		WithArgs(int64(9999)).
		WillReturnError(sql.ErrNoRows)

	room, err := repo.GetRoomByUserID(t.Context(), 9999)
	assert.Error(t, err)
	assert.Nil(t, room)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_CreateLinkmic_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO live_linkmic`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(1, now))

	lm, err := repo.CreateLinkmic(t.Context(), 1, 5001, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), lm.ID)
	assert.Equal(t, int64(1), lm.RoomID)
	assert.Equal(t, int64(5001), lm.StreamerID)
	assert.Equal(t, int64(100), lm.ViewerID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_CreateLinkmic_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`INSERT INTO live_linkmic`).
		WillReturnError(sql.ErrConnDone)

	lm, err := repo.CreateLinkmic(t.Context(), 1, 5001, 100)
	assert.Error(t, err)
	assert.Nil(t, lm)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateLinkmicStatus_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_linkmic SET status=\$1 WHERE id=\$2`).
		WillReturnError(sql.ErrConnDone)

	err = repo.UpdateLinkmicStatus(t.Context(), 1, 1)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_InitSeats_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`INSERT INTO live_seats`).
		WillReturnError(sql.ErrConnDone)

	err = repo.InitSeats(t.Context(), 1, 2)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateRoomStatus_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_rooms SET status=\$1, updated_at=NOW\(\) WHERE id=\$2`).
		WillReturnError(sql.ErrConnDone)

	err = repo.UpdateRoomStatus(t.Context(), 1, "live")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ScheduleRoom_WithTime(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_rooms SET scheduled_at=\$1, updated_at=NOW\(\) WHERE id=\$2`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.ScheduleRoom(t.Context(), 1, sql.NullTime{Time: time.Now(), Valid: true})
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_IncrementViewerCount_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_rooms SET viewer_count = viewer_count \+ 1`).
		WillReturnError(sql.ErrConnDone)

	err = repo.IncrementViewerCount(t.Context(), 1)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DecrementViewerCount_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_rooms SET viewer_count = GREATEST\(viewer_count - 1, 0\)`).
		WillReturnError(sql.ErrConnDone)

	err = repo.DecrementViewerCount(t.Context(), 1)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateRoom_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_rooms SET room_name=\$1`).
		WillReturnError(sql.ErrConnDone)

	room := &Room{ID: 1, RoomName: "new", CoverURL: "", Category: ""}
	err = repo.UpdateRoom(t.Context(), room)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateSeat_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnError(sql.ErrConnDone)

	seat := &Seat{RoomID: 1, SeatIndex: 0, UserID: 100, Status: SeatStatusOccupied, Muted: false}
	err = repo.UpdateSeat(t.Context(), seat)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
