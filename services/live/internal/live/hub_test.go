package live

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(userID, roomID int64) *Client {
	return &Client{
		userID: userID,
		roomID: roomID,
		send:   make(chan []byte, 256),
	}
}

func TestHub_RegisterAndUnregister(t *testing.T) {
	hub := NewHub()
	c1 := newTestClient(100, 1)
	c2 := newTestClient(200, 1)
	c3 := newTestClient(300, 2)

	hub.register(c1)
	hub.register(c2)
	hub.register(c3)

	hub.mu.RLock()
	assert.Len(t, hub.rooms[1], 2)
	assert.Len(t, hub.rooms[2], 1)
	assert.Equal(t, c1, hub.users[100])
	assert.Equal(t, c2, hub.users[200])
	assert.Equal(t, c3, hub.users[300])
	hub.mu.RUnlock()

	hub.unregister(c1)
	hub.mu.RLock()
	assert.Len(t, hub.rooms[1], 1)
	_, exists := hub.users[100]
	assert.False(t, exists)
	_, stillThere := hub.users[200]
	assert.True(t, stillThere)
	hub.mu.RUnlock()

	hub.unregister(c2)
	hub.mu.RLock()
	assert.Empty(t, hub.rooms[1])
	hub.mu.RUnlock()

	hub.unregister(c3)
	hub.mu.RLock()
	assert.Empty(t, hub.rooms[2])
	assert.Empty(t, hub.users)
	hub.mu.RUnlock()
}

func TestHub_RegisterZeroUserID(t *testing.T) {
	hub := NewHub()
	c := &Client{userID: 0, roomID: 1, send: make(chan []byte, 256)}
	hub.register(c)

	hub.mu.RLock()
	assert.Len(t, hub.rooms[1], 1)
	_, exists := hub.users[0]
	assert.False(t, exists)
	hub.mu.RUnlock()

	hub.unregister(c)
}

func TestHub_RegisterCreatesRoom(t *testing.T) {
	hub := NewHub()
	c := newTestClient(100, 99)
	hub.register(c)

	hub.mu.RLock()
	assert.NotNil(t, hub.rooms[99])
	assert.Len(t, hub.rooms[99], 1)
	hub.mu.RUnlock()

	hub.unregister(c)
}

func TestHub_SendToRoomExcept_NormalBroadcast(t *testing.T) {
	hub := NewHub()
	c1 := newTestClient(100, 1)
	c2 := newTestClient(200, 1)
	c3 := newTestClient(300, 1)
	hub.register(c1)
	hub.register(c2)
	hub.register(c3)
	defer hub.unregister(c1)
	defer hub.unregister(c2)
	defer hub.unregister(c3)

	hub.SendToRoomExcept(1, c1, map[string]interface{}{"type": "test"})

	// c2 and c3 should receive, c1 (except) should not
	select {
	case msg := <-c2.send:
		assert.Contains(t, string(msg), "test")
	case <-time.After(time.Second):
		t.Fatal("c2 did not receive message")
	}
	select {
	case msg := <-c3.send:
		assert.Contains(t, string(msg), "test")
	case <-time.After(time.Second):
		t.Fatal("c3 did not receive message")
	}
	select {
	case <-c1.send:
		t.Fatal("c1 (except) should not receive message")
	default:
		// good, c1 didn't receive
	}
}

func TestHub_SendToRoomExcept_EmptyRoom(t *testing.T) {
	hub := NewHub()
	hub.SendToRoomExcept(999, nil, map[string]interface{}{"type": "test"})
}

func TestHub_SendToRoomExcept_FullChannel(t *testing.T) {
	hub := NewHub()
	c := &Client{
		userID: 100,
		roomID: 1,
		send:   make(chan []byte, 1), // buffer of 1
	}
	hub.register(c)
	defer hub.unregister(c)

	// fill the channel
	c.send <- []byte(`{"filled":true}`)

	// now SendToRoomExcept should skip (default branch in select)
	hub.SendToRoomExcept(1, nil, map[string]interface{}{"type": "test"})

	// channel should still only have the original message
	assert.Len(t, c.send, 1)
}

func TestHub_BroadcastRoom_WithClients(t *testing.T) {
	hub := NewHub()
	c1 := newTestClient(100, 1)
	c2 := newTestClient(200, 1)
	hub.register(c1)
	hub.register(c2)
	defer hub.unregister(c1)
	defer hub.unregister(c2)

	hub.BroadcastRoom(1, map[string]interface{}{"type": "broadcast"})

	select {
	case msg := <-c1.send:
		assert.Contains(t, string(msg), "broadcast")
	case <-time.After(time.Second):
		t.Fatal("c1 did not receive broadcast")
	}
	select {
	case msg := <-c2.send:
		assert.Contains(t, string(msg), "broadcast")
	case <-time.After(time.Second):
		t.Fatal("c2 did not receive broadcast")
	}
}

func TestHub_BroadcastRoom_FullChannel(t *testing.T) {
	hub := NewHub()
	c := &Client{
		userID: 100,
		roomID: 1,
		send:   make(chan []byte, 1),
	}
	hub.register(c)
	defer hub.unregister(c)

	c.send <- []byte(`{"filled":true}`)
	hub.BroadcastRoom(1, map[string]interface{}{"type": "test"})
	assert.Len(t, c.send, 1)
}

func TestHub_BroadcastRoom_NonExistentRoom(t *testing.T) {
	hub := NewHub()
	hub.BroadcastRoom(999, map[string]interface{}{"type": "test"})
}

func TestHub_SendToUser_Registered(t *testing.T) {
	hub := NewHub()
	c := newTestClient(100, 1)
	hub.register(c)
	defer hub.unregister(c)

	hub.SendToUser(100, map[string]interface{}{"type": "dm"})
	select {
	case msg := <-c.send:
		assert.Contains(t, string(msg), "dm")
	case <-time.After(time.Second):
		t.Fatal("client did not receive DM")
	}
}

func TestHub_SendToUser_NotRegistered(t *testing.T) {
	hub := NewHub()
	hub.SendToUser(999, map[string]interface{}{"type": "dm"})
}

func TestHub_SendToUser_FullChannel(t *testing.T) {
	hub := NewHub()
	c := &Client{
		userID: 100,
		roomID: 1,
		send:   make(chan []byte, 1),
	}
	hub.register(c)
	defer hub.unregister(c)

	c.send <- []byte(`{"filled":true}`)
	hub.SendToUser(100, map[string]interface{}{"type": "dm"})
	assert.Len(t, c.send, 1)
}

func TestHub_MultipleRooms(t *testing.T) {
	hub := NewHub()
	c1 := newTestClient(100, 1)
	c2 := newTestClient(200, 2)
	c3 := newTestClient(300, 1)
	hub.register(c1)
	hub.register(c2)
	hub.register(c3)
	defer hub.unregister(c1)
	defer hub.unregister(c2)
	defer hub.unregister(c3)

	// Send to room 1 only
	hub.SendToRoomExcept(1, nil, map[string]interface{}{"room": 1})

	select {
	case msg := <-c1.send:
		assert.Contains(t, string(msg), `"room":1`)
	case <-time.After(time.Second):
		t.Fatal("c1 in room 1 did not receive")
	}
	select {
	case msg := <-c3.send:
		assert.Contains(t, string(msg), `"room":1`)
	case <-time.After(time.Second):
		t.Fatal("c3 in room 1 did not receive")
	}
	select {
	case <-c2.send:
		t.Fatal("c2 in room 2 should not receive")
	default:
		// good
	}
}

func TestHub_UnregisterNonExistentUser(t *testing.T) {
	hub := NewHub()
	c := &Client{userID: 500, roomID: 10, send: make(chan []byte, 1)}
	// unregister without registering should not panic
	hub.unregister(c)
}

func TestWsMsgToMap_FullFields(t *testing.T) {
	msg := wsMsg{
		Type:       "offer",
		RoomID:     42,
		SeatIndex:  3,
		TargetID:   777,
		Muted:      true,
		Locked:     true,
		StreamerID: 100,
		ViewerID:   200,
		SDP:        "v=0",
		Candidate:  "candidate:udp",
	}
	m := wsMsgToMap(msg)
	require.NotNil(t, m)
	assert.Equal(t, "offer", m["type"])
	assert.Equal(t, int64(42), m["room_id"])
	assert.Equal(t, int32(3), m["seat_index"])
	assert.Equal(t, int64(777), m["target_id"])
	assert.Equal(t, true, m["muted"])
	assert.Equal(t, true, m["locked"])
	assert.Equal(t, int64(100), m["streamer_id"])
	assert.Equal(t, int64(200), m["viewer_id"])
	assert.Equal(t, "v=0", m["sdp"])
	assert.Equal(t, "candidate:udp", m["candidate"])
}

func TestParseInt64_EdgeCases(t *testing.T) {
	assert.Equal(t, int64(0), parseInt64(""))
	assert.Equal(t, int64(0), parseInt64("abc"))
	assert.Equal(t, int64(123), parseInt64("123"))
	assert.Equal(t, int64(42), parseInt64("abc42def"))
	assert.Equal(t, int64(0), parseInt64("0"))
	assert.Equal(t, int64(999999), parseInt64("999999"))
}

func TestUserIDFromQuery(t *testing.T) {
	tests := []struct {
		url  string
		want int64
	}{
		{"http://localhost/ws?user_id=123", 123},
		{"http://localhost/ws?user_id=abc", 0},
		{"http://localhost/ws", 0},
		{"http://localhost/ws?user_id=0", 0},
		{"http://localhost/ws?user_id=999999", 999999},
	}
	for _, tt := range tests {
		r, _ := http.NewRequest("GET", tt.url, nil)
		got := userIDFromQuery(r)
		assert.Equal(t, tt.want, got, "url=%s", tt.url)
	}
}

func TestRoomIDFromQuery(t *testing.T) {
	tests := []struct {
		url  string
		want int64
	}{
		{"http://localhost/ws?room_id=42", 42},
		{"http://localhost/ws?room_id=abc", 0},
		{"http://localhost/ws", 0},
		{"http://localhost/ws?room_id=0", 0},
		{"http://localhost/ws?room_id=12345", 12345},
	}
	for _, tt := range tests {
		r, _ := http.NewRequest("GET", tt.url, nil)
		got := roomIDFromQuery(r)
		assert.Equal(t, tt.want, got, "url=%s", tt.url)
	}
}

func TestGenRoomCode(t *testing.T) {
	code1 := genRoomCode()
	code2 := genRoomCode()
	assert.Len(t, code1, 8)
	assert.Len(t, code2, 8)
	// Very unlikely to be equal, but verify non-empty
	assert.NotEmpty(t, code1)
	assert.NotEmpty(t, code2)
}
