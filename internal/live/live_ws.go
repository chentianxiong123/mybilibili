package live

import (
	"encoding/json"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

type Client struct {
	userID int64
	roomID int64
	conn   *websocket.Conn
	send   chan []byte
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[int64]map[*Client]bool
	users map[int64]*Client
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[int64]map[*Client]bool),
		users: make(map[int64]*Client),
	}
}

func (h *Hub) register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[c.roomID] == nil {
		h.rooms[c.roomID] = make(map[*Client]bool)
	}
	h.rooms[c.roomID][c] = true
	if c.userID != 0 {
		h.users[c.userID] = c
	}
}

func (h *Hub) unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set := h.rooms[c.roomID]; set != nil {
		delete(set, c)
		close(c.send)
	}
	delete(h.users, c.userID)
}

func (h *Hub) BroadcastRoom(roomID int64, msg map[string]interface{}) {
	data, _ := json.Marshal(msg)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.rooms[roomID] {
		select {
		case c.send <- data:
		default:
		}
	}
}

func (h *Hub) SendToUser(userID int64, msg map[string]interface{}) {
	data, _ := json.Marshal(msg)
	h.mu.RLock()
	defer h.mu.RUnlock()
	if c := h.users[userID]; c != nil {
		select {
		case c.send <- data:
		default:
		}
	}
}

func (h *Hub) SendToRoomExcept(roomID int64, except *Client, msg map[string]interface{}) {
	data, _ := json.Marshal(msg)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.rooms[roomID] {
		if c == except {
			continue
		}
		select {
		case c.send <- data:
		default:
		}
	}
}

func (h *Hub) writePump(c *Client) {
	for data := range c.send {
		if err := websocket.Message.Send(c.conn, string(data)); err != nil {
			return
		}
	}
}

type wsMsg struct {
	Type       string `json:"type"`
	RoomID     int64  `json:"room_id"`
	SeatIndex  int32  `json:"seat_index,omitempty"`
	TargetID   int64  `json:"target_id,omitempty"`
	Muted      bool   `json:"muted,omitempty"`
	Locked     bool   `json:"locked,omitempty"`
	StreamerID int64  `json:"streamer_id,omitempty"`
	ViewerID   int64  `json:"viewer_id,omitempty"`
	SDP        string `json:"sdp,omitempty"`
	Candidate  string `json:"candidate,omitempty"`
}

func (h *Hub) ServeWS(ws *websocket.Conn, svc *Service) {
	userID := userIDFromQuery(ws.Request())
	roomID := roomIDFromQuery(ws.Request())
	if roomID == 0 {
		return
	}

	c := &Client{
		userID: userID,
		roomID: roomID,
		conn:   ws,
		send:   make(chan []byte, 256),
	}
	h.register(c)
	defer h.unregister(c)

	go h.writePump(c)

	for {
		var raw string
		if err := websocket.Message.Receive(ws, &raw); err != nil {
			return
		}

		var msg wsMsg
		if err := json.Unmarshal([]byte(raw), &msg); err != nil {
			continue
		}
		msg.RoomID = roomID
		h.handleMessage(svc, c, msg)
	}
}

func wsMsgToMap(msg wsMsg) map[string]interface{} {
	return map[string]interface{}{
		"type":        msg.Type,
		"room_id":     msg.RoomID,
		"seat_index":  msg.SeatIndex,
		"target_id":   msg.TargetID,
		"muted":       msg.Muted,
		"locked":      msg.Locked,
		"streamer_id": msg.StreamerID,
		"viewer_id":   msg.ViewerID,
		"sdp":         msg.SDP,
		"candidate":   msg.Candidate,
	}
}

func (h *Hub) handleMessage(svc *Service, c *Client, msg wsMsg) {
	ctx := c.conn.Request().Context()

	switch msg.Type {
	case "apply_seat":
		if err := svc.ApplySeat(ctx, msg.RoomID, c.userID); err != nil {
			h.SendToUser(c.userID, map[string]interface{}{"type": "error", "message": err.Error()})
		}
	case "leave_seat":
		svc.LeaveSeat(ctx, msg.RoomID, c.userID)
	case "accept_seat":
		svc.AcceptSeat(ctx, msg.RoomID, c.userID, msg.TargetID)
	case "reject_seat":
		svc.RejectSeat(ctx, msg.RoomID, c.userID, msg.TargetID)
	case "mute_self":
		seats, _ := svc.repo.GetSeats(ctx, msg.RoomID)
		for _, seat := range seats {
			if seat.UserID == c.userID {
				seat.Muted = !seat.Muted
				svc.repo.UpdateSeat(ctx, seat)
				h.BroadcastRoom(msg.RoomID, map[string]interface{}{
					"type": "seat_updated", "room_id": msg.RoomID, "seats": seatsToMap(seats),
				})
				break
			}
		}
	case "lock_seat":
		svc.LockSeat(ctx, msg.RoomID, c.userID, msg.SeatIndex, msg.Locked)
	case "kick":
		svc.KickSeat(ctx, msg.RoomID, c.userID, msg.TargetID)
	case "offer", "answer", "ice":
		h.SendToRoomExcept(msg.RoomID, c, wsMsgToMap(msg))
	default:
		h.SendToRoomExcept(msg.RoomID, c, wsMsgToMap(msg))
	}
}

func (h *Hub) HandleWS(svc *Service) http.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		h.ServeWS(ws, svc)
	})
}

func userIDFromQuery(r *http.Request) int64 {
	return parseInt64(r.URL.Query().Get("user_id"))
}

func roomIDFromQuery(r *http.Request) int64 {
	return parseInt64(r.URL.Query().Get("room_id"))
}

func parseInt64(s string) int64 {
	var id int64
	for _, b := range []byte(s) {
		if b >= '0' && b <= '9' {
			id = id*10 + int64(b-'0')
		}
	}
	return id
}
