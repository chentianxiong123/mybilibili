package live

import (
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
)

// newWSConnPair 建立一条真实的 websocket 服务端/客户端连接对。
// 服务端 handler 在捕获连接后阻塞，直到测试结束（done 关闭）。
// 返回服务端连接（其 Request() 非 nil，可用于 handleMessage 的 Client）与客户端连接。
func newWSConnPair(t *testing.T) (*websocket.Conn, *websocket.Conn, *httptest.Server) {
	t.Helper()
	serverConnCh := make(chan *websocket.Conn, 1)
	done := make(chan struct{})
	srv := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		serverConnCh <- ws
		<-done
	}))
	url := "ws://" + srv.Listener.Addr().String() + "/ws?room_id=1&user_id=100"
	client, err := websocket.Dial(url, "", srv.URL)
	require.NoError(t, err)
	var server *websocket.Conn
	select {
	case server = <-serverConnCh:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not accept websocket connection")
	}
	t.Cleanup(func() {
		close(done)
		_ = client.Close()
		srv.Close()
	})
	return server, client, srv
}

// newServeWSEnv 搭建基于 hub.HandleWS 的完整 websocket 服务端环境。
func newServeWSEnv(t *testing.T) (*Hub, *Service, sqlmock.Sqlmock, *httptest.Server) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	repo := NewRepository(db)
	hub := NewHub()
	svc := NewService(repo, hub)
	srv := httptest.NewServer(hub.HandleWS(svc))
	t.Cleanup(func() { srv.Close() })
	return hub, svc, mock, srv
}

// newHandleMessageEnv 返回 hub、svc、mock 与一个已注册的 Client（使用真实服务端连接）。
func newHandleMessageEnv(t *testing.T) (*Hub, *Service, sqlmock.Sqlmock, *Client) {
	t.Helper()
	serverConn, _, _ := newWSConnPair(t)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	repo := NewRepository(db)
	hub := NewHub()
	svc := NewService(repo, hub)
	c := &Client{userID: 100, roomID: 1, conn: serverConn, send: make(chan []byte, 256)}
	hub.register(c)
	t.Cleanup(func() { hub.unregister(c) })
	return hub, svc, mock, c
}

// ---------- writePump ----------

func TestWritePump_SendsMessage(t *testing.T) {
	server, client, _ := newWSConnPair(t)
	hub := NewHub()
	c := &Client{userID: 100, roomID: 1, conn: server, send: make(chan []byte, 256)}

	done := make(chan struct{})
	go func() {
		hub.writePump(c)
		close(done)
	}()

	c.send <- []byte(`{"type":"ping"}`)

	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	var got string
	require.NoError(t, websocket.Message.Receive(client, &got))
	assert.Equal(t, `{"type":"ping"}`, got)

	close(c.send)
	<-done
}

func TestWritePump_SendError_Returns(t *testing.T) {
	server, _, _ := newWSConnPair(t)
	hub := NewHub()
	c := &Client{userID: 100, roomID: 1, conn: server, send: make(chan []byte, 256)}

	done := make(chan struct{})
	go func() {
		hub.writePump(c)
		close(done)
	}()

	// 关闭服务端连接 → 后续 Send 失败 → writePump 返回
	_ = server.Close()
	c.send <- []byte(`{"type":"ping"}`)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("writePump did not exit after connection closed")
	}
}

// ---------- ServeWS ----------

func TestServeWS_NoRoomID_Returns(t *testing.T) {
	_, _, _, srv := newServeWSEnv(t)
	url := "ws://" + srv.Listener.Addr().String() + "/ws"
	client, err := websocket.Dial(url, "", srv.URL)
	if err != nil {
		// 服务端因 roomID==0 立即返回并关闭连接，Dial 可能已失败
		return
	}
	defer client.Close()
	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	var resp string
	assert.Error(t, websocket.Message.Receive(client, &resp))
}

func TestServeWS_ApplySeatError_RespondsToClient(t *testing.T) {
	_, _, mock, srv := newServeWSEnv(t)
	url := "ws://" + srv.Listener.Addr().String() + "/ws?room_id=1&user_id=100"
	client, err := websocket.Dial(url, "", srv.URL)
	require.NoError(t, err)
	defer client.Close()

	// 空座位表 → ApplySeat 返回 "no available seats" → 服务端向该用户发送 error 消息
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow())

	require.NoError(t, websocket.Message.Send(client, `{"type":"apply_seat"}`))

	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	var resp string
	require.NoError(t, websocket.Message.Receive(client, &resp))
	assert.Contains(t, resp, "no available seats")
	assert.Contains(t, resp, "error")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServeWS_BadJSON_Continues(t *testing.T) {
	_, _, mock, srv := newServeWSEnv(t)
	url := "ws://" + srv.Listener.Addr().String() + "/ws?room_id=1&user_id=100"
	client, err := websocket.Dial(url, "", srv.URL)
	require.NoError(t, err)
	defer client.Close()

	// 坏 JSON 被 continue 忽略；随后有效 apply_seat 触发 DB 并回发 error
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow())

	require.NoError(t, websocket.Message.Send(client, `{bad`))
	require.NoError(t, websocket.Message.Send(client, `{"type":"apply_seat"}`))

	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	var resp string
	require.NoError(t, websocket.Message.Receive(client, &resp))
	assert.Contains(t, resp, "no available seats")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServeWS_Unregistered_NoRoom(t *testing.T) {
	_, _, _, srv := newServeWSEnv(t)
	// 连接后立即关闭，服务端 Receive 失败 → ServeWS 返回 → unregister
	url := "ws://" + srv.Listener.Addr().String() + "/ws?room_id=1&user_id=100"
	client, err := websocket.Dial(url, "", srv.URL)
	require.NoError(t, err)
	_ = client.Close()
	// 无 DB 交互，连接关闭后服务端清理
}

// ---------- handleMessage 各分支 ----------

func TestHandleMessage_ApplySeat_Success(t *testing.T) {
	hub, svc, mock, c := newHandleMessageEnv(t)
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 0, SeatStatusEmpty, false, sql.NullTime{}))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	hub.handleMessage(svc, c, wsMsg{Type: "apply_seat", RoomID: 1})

	select {
	case msg := <-c.send:
		assert.Contains(t, string(msg), "seat_updated")
	default:
		t.Fatal("expected broadcast message")
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessage_ApplySeat_Error(t *testing.T) {
	hub, svc, mock, c := newHandleMessageEnv(t)
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow())

	hub.handleMessage(svc, c, wsMsg{Type: "apply_seat", RoomID: 1})

	select {
	case msg := <-c.send:
		assert.Contains(t, string(msg), "error")
		assert.Contains(t, string(msg), "no available seats")
	default:
		t.Fatal("expected error message to user")
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessage_LeaveSeat(t *testing.T) {
	hub, svc, mock, c := newHandleMessageEnv(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(100)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusOccupied, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 0, SeatStatusEmpty, false, sql.NullTime{}))

	hub.handleMessage(svc, c, wsMsg{Type: "leave_seat", RoomID: 1})

	select {
	case msg := <-c.send:
		assert.Contains(t, string(msg), "seat_updated")
	default:
		t.Fatal("expected broadcast message")
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessage_AcceptSeat(t *testing.T) {
	hub, svc, mock, c := newHandleMessageEnv(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 100, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 200, SeatStatusPending, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	hub.handleMessage(svc, c, wsMsg{Type: "accept_seat", RoomID: 1, TargetID: 200})

	select {
	case msg := <-c.send:
		assert.Contains(t, string(msg), "seat_updated")
	default:
		t.Fatal("expected broadcast message")
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessage_RejectSeat(t *testing.T) {
	hub, svc, mock, c := newHandleMessageEnv(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 100, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 200, SeatStatusPending, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	hub.handleMessage(svc, c, wsMsg{Type: "reject_seat", RoomID: 1, TargetID: 200})

	select {
	case msg := <-c.send:
		assert.Contains(t, string(msg), "seat_updated")
	default:
		t.Fatal("expected broadcast message")
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessage_MuteSelf_Match(t *testing.T) {
	hub, svc, mock, c := newHandleMessageEnv(t)
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusOccupied, false, sql.NullTime{}))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	hub.handleMessage(svc, c, wsMsg{Type: "mute_self", RoomID: 1})

	select {
	case msg := <-c.send:
		assert.Contains(t, string(msg), "seat_updated")
	default:
		t.Fatal("expected broadcast message")
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessage_MuteSelf_NoMatch(t *testing.T) {
	hub, svc, mock, c := newHandleMessageEnv(t)
	mock.ExpectQuery(`SELECT .+ FROM live_seats`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 200, SeatStatusOccupied, false, sql.NullTime{}))

	hub.handleMessage(svc, c, wsMsg{Type: "mute_self", RoomID: 1})

	select {
	case <-c.send:
		t.Fatal("should not broadcast when no matching seat")
	default:
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessage_LockSeat(t *testing.T) {
	hub, svc, mock, c := newHandleMessageEnv(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 100, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 0, SeatStatusEmpty, false, sql.NullTime{}))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	hub.handleMessage(svc, c, wsMsg{Type: "lock_seat", RoomID: 1, SeatIndex: 0, Locked: true})

	select {
	case msg := <-c.send:
		assert.Contains(t, string(msg), "seat_updated")
	default:
		t.Fatal("expected broadcast message")
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessage_Kick(t *testing.T) {
	hub, svc, mock, c := newHandleMessageEnv(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(roomRow().AddRow(1, "r", 100, "s", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(200)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 200, SeatStatusOccupied, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 0, SeatStatusEmpty, false, sql.NullTime{}))

	hub.handleMessage(svc, c, wsMsg{Type: "kick", RoomID: 1, TargetID: 200})

	select {
	case msg := <-c.send:
		assert.Contains(t, string(msg), "seat_updated")
	default:
		t.Fatal("expected broadcast message")
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessage_Offer_RelaysToOthers(t *testing.T) {
	hub, svc, mock, c := newHandleMessageEnv(t)
	c2 := newTestClient(200, 1)
	hub.register(c2)
	defer hub.unregister(c2)

	hub.handleMessage(svc, c, wsMsg{Type: "offer", RoomID: 1, SDP: "v=0"})

	select {
	case msg := <-c2.send:
		assert.Contains(t, string(msg), "offer")
	default:
		t.Fatal("expected relayed message to other client")
	}
	// c（发送者）不应收到
	select {
	case <-c.send:
		t.Fatal("sender should not receive its own offer")
	default:
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessage_Answer_RelaysToOthers(t *testing.T) {
	hub, svc, mock, c := newHandleMessageEnv(t)
	c2 := newTestClient(200, 1)
	hub.register(c2)
	defer hub.unregister(c2)

	hub.handleMessage(svc, c, wsMsg{Type: "answer", RoomID: 1, SDP: "a=1"})

	select {
	case msg := <-c2.send:
		assert.Contains(t, string(msg), "answer")
	default:
		t.Fatal("expected relayed message to other client")
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessage_ICE_RelaysToOthers(t *testing.T) {
	hub, svc, mock, c := newHandleMessageEnv(t)
	c2 := newTestClient(200, 1)
	hub.register(c2)
	defer hub.unregister(c2)

	hub.handleMessage(svc, c, wsMsg{Type: "ice", RoomID: 1, Candidate: "candidate:udp"})

	select {
	case msg := <-c2.send:
		assert.Contains(t, string(msg), "ice")
	default:
		t.Fatal("expected relayed message to other client")
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessage_Default_RelaysToOthers(t *testing.T) {
	hub, svc, mock, c := newHandleMessageEnv(t)
	c2 := newTestClient(200, 1)
	hub.register(c2)
	defer hub.unregister(c2)

	hub.handleMessage(svc, c, wsMsg{Type: "unknown_type", RoomID: 1})

	select {
	case msg := <-c2.send:
		assert.Contains(t, string(msg), "unknown_type")
	default:
		t.Fatal("expected relayed message to other client")
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- ServeWS 端到端（真实消息循环） ----------

func TestServeWS_LeaveSeat_EndToEnd(t *testing.T) {
	_, _, mock, srv := newServeWSEnv(t)
	url := "ws://" + srv.Listener.Addr().String() + "/ws?room_id=1&user_id=100"
	client, err := websocket.Dial(url, "", srv.URL)
	require.NoError(t, err)
	defer client.Close()

	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(100)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 100, SeatStatusOccupied, false, now))
	mock.ExpectExec(`UPDATE live_seats SET user_id=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT .+ FROM live_seats WHERE room_id = \$1 ORDER BY seat_index`).
		WithArgs(int64(1)).
		WillReturnRows(seatRow().AddRow(1, 1, 0, 0, SeatStatusEmpty, false, sql.NullTime{}))

	require.NoError(t, websocket.Message.Send(client, `{"type":"leave_seat"}`))

	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	var resp string
	require.NoError(t, websocket.Message.Receive(client, &resp))
	assert.Contains(t, resp, "seat_updated")

	require.NoError(t, mock.ExpectationsWereMet())
}
