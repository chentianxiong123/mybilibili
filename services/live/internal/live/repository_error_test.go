package live

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- ListLiveRooms 错误分支 ----------

func TestRepository_ListLiveRooms_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE status = 'live'`).
		WithArgs(int32(20), int32(0)).
		WillReturnError(sql.ErrConnDone)

	rooms, err := repo.ListLiveRooms(t.Context(), 1, 20)
	assert.Error(t, err)
	assert.Nil(t, rooms)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListLiveRooms_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	// 返回的列数少于 Scan 期望的列数 → Scan 报错
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE status = 'live'`).
		WithArgs(int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	rooms, err := repo.ListLiveRooms(t.Context(), 1, 20)
	assert.Error(t, err)
	assert.Nil(t, rooms)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- GetSeats 扫描错误分支 ----------

func TestRepository_GetSeats_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	seats, err := repo.GetSeats(t.Context(), 1)
	assert.Error(t, err)
	assert.Nil(t, seats)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- 其余 Exec 错误分支 ----------

func TestRepository_GetRoomByID_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	room, err := repo.GetRoomByID(t.Context(), 1)
	assert.Error(t, err)
	assert.Nil(t, room)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetRoomByStreamKey_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE stream_key = \$1`).
		WithArgs("skey").
		WillReturnError(sql.ErrConnDone)

	room, err := repo.GetRoomByStreamKey(t.Context(), "skey")
	assert.Error(t, err)
	assert.Nil(t, room)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetRoomByUserID_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE user_id = \$1`).
		WithArgs(int64(5001)).
		WillReturnError(sql.ErrConnDone)

	room, err := repo.GetRoomByUserID(t.Context(), 5001)
	assert.Error(t, err)
	assert.Nil(t, room)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetSeatByUser_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(100)).
		WillReturnError(sql.ErrConnDone)

	seat, err := repo.GetSeatByUser(t.Context(), 1, 100)
	assert.Error(t, err)
	assert.Nil(t, seat)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ScheduleRoom_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_rooms SET scheduled_at=\$1, updated_at=NOW\(\) WHERE id=\$2`).
		WillReturnError(sql.ErrConnDone)

	err = repo.ScheduleRoom(t.Context(), 1, sql.NullTime{Valid: false})
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateRoom_EmptyName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE live_rooms SET room_name=\$1`).
		WithArgs("", "", "", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	room := &Room{ID: 1, RoomName: "", CoverURL: "", Category: ""}
	err = repo.UpdateRoom(t.Context(), room)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
