package live

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Finding A (HIGH): Hub.unregister corrupts the per-user client map when a
// user holds multiple concurrent WS connections.
//
// live_ws.go:50 `delete(h.users, c.userID)` removes the map entry UNCONDITIONALLY.
// If user 100 has two connections (A then B), register overwrites users[100]=B.
// When A disconnects, unregister(A) deletes users[100] even though B is still
// connected. SendToUser (used by KickSeat to send "kicked") then misses B.
// ---------------------------------------------------------------------------
func TestHub_Unregister_MultiConnSameUser_RemovesLiveEntry(t *testing.T) {
	hub := NewHub()
	a := newTestClient(100, 1)
	b := newTestClient(100, 1)
	hub.register(a)
	hub.register(b)

	// Both connections registered; users[100] points to the newest (b).
	hub.mu.RLock()
	assert.Equal(t, b, hub.users[100], "users[100] should point to the newest connection b")
	hub.mu.RUnlock()

	// Older connection A disconnects.
	hub.unregister(a)

	// FINDING-053 修复：A 断开后 users[100] 仍指向仍连接的 B。
	hub.mu.RLock()
	_, exists := hub.users[100]
	hub.mu.RUnlock()
	assert.True(t, exists,
		"FINDING-053 修复: users[100] 应保留（B 仍连接），SendToUser 应能达 B")

	// Impact: SendToUser(100) now reaches the surviving connection b (FINDING-053 修复).
	hub.SendToUser(100, map[string]interface{}{"type": "kicked", "room_id": int64(1)})
	select {
	case <-b.send:
		// FINDING-053 修复：b 仍能收到 kick 通知。
	case <-time.After(200 * time.Millisecond):
		t.Fatal("FINDING-053 修复失败: b 应收到 kick，但 users[100] 未保留")
	}

	// b is still correctly present in the room set (only the user map is corrupted).
	hub.mu.RLock()
	assert.True(t, hub.rooms[1][b], "b should remain in the room set")
	hub.mu.RUnlock()
	hub.unregister(b)
}

// ---------------------------------------------------------------------------
// Finding B (MEDIUM): CreateRoom ignores maxSeats and never initializes seats,
// so a freshly created room has zero live_seats rows and ApplySeat always
// returns "no available seats" for a brand-new room.
//
// live_service.go:34 CreateRoom signature `_ int32` discards maxSeats;
// InitSeats (live_repository.go:187) is never called anywhere.
// ---------------------------------------------------------------------------
func TestService_CreateRoom_NeverSeedsSeats(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`INSERT INTO live_rooms`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, now, now))
	// FINDING-054 修复：CreateRoom 现调用 InitSeats 播种 maxSeats=8 个座位。
	mock.ExpectExec(`INSERT INTO live_seats`).WillReturnResult(sqlmock.NewResult(0, 8))

	room, err := svc.CreateRoom(t.Context(), "新直播间", "", "游戏", 5001, 8)
	require.NoError(t, err)
	assert.Equal(t, int64(1), room.ID)

	// 座位已播种：第一个 ApplySeat 应能占座（此前 live_seats 0 行 → "no available seats"）。
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "room_id", "seat_index", "user_id", "status", "muted", "joined_at"}).
			AddRow(1, int64(1), 0, int64(0), int32(0), false, nil))
	mock.ExpectExec(`UPDATE live_seats`).WillReturnResult(sqlmock.NewResult(0, 1))
	err = svc.ApplySeat(t.Context(), room.ID, 100)
	require.NoError(t, err, "FINDING-054 修复：座位已播种，ApplySeat 应成功")
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// Finding C (MEDIUM): ApplySeat only short-circuits when the user is already
// SeatStatusOccupied. If the user already has a Pending seat and applies again,
// it claims a second empty seat -> the same user ends up Pending on two seats.
//
// live_service.go:137-141 checks `seat.Status == SeatStatusOccupied` only.
// ---------------------------------------------------------------------------
func TestService_ApplySeat_PendingUserClaimsSecondSeat(t *testing.T) {
	svc, mock := newTestServiceWithSeats(t)
	now := time.Now()

	// First apply: user 100 grabs empty seat 0 -> Pending.
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().
			AddRow(1, 1, 0, 0, SeatStatusEmpty, false, sql.NullTime{}).
			AddRow(2, 1, 1, 0, SeatStatusEmpty, false, sql.NullTime{}))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WithArgs(int64(100), int32(SeatStatusPending), false, sqlmock.AnyArg(), int64(1), int32(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.ApplySeat(t.Context(), 1, 100))

	// Second apply while user 100 is still PENDING on seat 0.
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().
			AddRow(1, 1, 0, 100, SeatStatusPending, false, now).
			AddRow(2, 1, 1, 0, SeatStatusEmpty, false, sql.NullTime{}))
	// Second apply while user 100 is still PENDING on seat 0 → 已 Pending 不再申请第二座。
	err := svc.ApplySeat(t.Context(), 1, 100)
	assert.NoError(t, err, "FINDING-055 修复: 已 Pending 用户不得占第二座（返回成功但不新增）")
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// Finding D (MEDIUM): admin handler functional bugs.
//
//	D1 handleStats compares varchar status to integer 1 (`WHERE status = 1`)
//	   instead of `status = 'live'` (live_admin_handler.go:116-117). In PG this
//	   errors and the error is swallowed (`_ =`), so live_count/total_viewers
//	   are always 0.
//	D2 handleRooms/handleRoomByID select a non-existent `title` column while the
//	   live_rooms schema column is `room_name` (live_admin_handler.go:67,90).
//	   The query errors and is swallowed, so the room list is always empty and
//	   by-id always 404.
//
// ---------------------------------------------------------------------------
func TestAdminHandler_StatsUsesIntegerCompare(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	h := NewAdminHandler(db)

	// The handler issues `WHERE status = 1`; pin it by matching that buggy SQL.
	// Correct SQL would be `WHERE status = 'live'`.
	mock.ExpectQuery(`status = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(`status = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(42))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/live/admin/stats", nil)
	h.handleStats(rr, req)

	assert.Contains(t, rr.Body.String(), `"live_count":5`,
		"FINDING-D1: handler queries WHERE status = 1 (integer) instead of status='live' (varchar)")
	assert.Contains(t, rr.Body.String(), `"total_viewers":42`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_RoomsUseNonExistentTitleColumn(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	h := NewAdminHandler(db)

	// Handler issues `SELECT id, user_id, title, ...`; pin by matching the
	// buggy `title` column. Schema column is `room_name` (sql/000:478).
	mock.ExpectQuery(`SELECT .+title.+ FROM live_rooms`).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "status", "viewer_count", "created_at"}).
			AddRow(1, "5001", "直播间", "live", 7, "2026-01-01 00:00:00"))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/live/admin/rooms", nil)
	h.handleRooms(rr, req)

	assert.Contains(t, rr.Body.String(), `"id":1`,
		"FINDING-D2: handler selects non-existent `title` column; should be `room_name`")
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// Finding F (HIGH, auth gap): POST /api/v1/live/linkmic/apply/{room} does not
// require authentication. handleLinkmic reads userID via GetUserIDFromHeader
// (linkmic_handler.go:27) but never rejects userID==0 (unlike handleRoom),
// and LinkmicService.Apply (linkmic.go:117-130) inserts the linkmic with
// viewer_id=0 for anonymous callers -> anonymous queue pollution / spam.
// ---------------------------------------------------------------------------
func TestLinkmicApply_NoAuth_InsertsViewerZero(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "r", 5001, "s", "", "", RoomStatusIdle, 0, now, now))
	// Anonymous apply: viewer_id is 0 (no X-User-Id header).
	mock.ExpectQuery(`INSERT INTO live_linkmic`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "room_id", "streamer_id", "viewer_id", "status", "created_at"}).
			AddRow(1, 1, 5001, 0, LinkmicStatusApplying, now))

	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/linkmic/apply/1", nil)
	// No X-User-Id header -> unauthenticated.
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code,
		"FINDING-026 修复: unauthenticated linkmic apply 应 401")
}

// ---------------------------------------------------------------------------
// Finding E (LOW): mute_self toggles seat.Muted but never updates seat.Status
// to SeatStatusMuted/SeatStatusOccupied, unlike the host MuteSeat path
// (live_service.go:259-263). The DB row is left with status=Occupied while
// muted=true -> client mute state is inconsistent.
//
// live_ws.go:175-186 mute_self branch.
// ---------------------------------------------------------------------------
func TestHandleMessage_MuteSelf_DoesNotUpdateStatus(t *testing.T) {
	hub, svc, mock, c := newHandleMessageEnv(t)
	// Seat 0 is Occupied(2), muted=false.
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusOccupied, false, sql.NullTime{}))
	// The UPDATE passes status=Occupied(2) (unchanged) with muted=true.
	// A correct self-mute would pass status=SeatStatusMuted(3).
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WithArgs(int64(100), int32(SeatStatusOccupied), true, sqlmock.AnyArg(), int64(1), int32(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	hub.handleMessage(svc, c, wsMsg{Type: "mute_self", RoomID: 1})

	select {
	case msg := <-c.send:
		assert.Contains(t, string(msg), "seat_updated")
	default:
		t.Fatal("expected broadcast message")
	}
	require.NoError(t, mock.ExpectationsWereMet(),
		"FINDING-E: mute_self should set status=SeatStatusMuted, not leave status=Occupied")
}
