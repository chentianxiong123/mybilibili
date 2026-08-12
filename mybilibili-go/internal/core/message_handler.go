package core

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

type MessageHTTPHandler struct {
	repo  *MessageRepository
	notif *NotificationBroadcaster
}

func NewMessageHTTPHandler(repo *MessageRepository, notif *NotificationBroadcaster) *MessageHTTPHandler {
	return &MessageHTTPHandler{repo: repo, notif: notif}
}

func (h *MessageHTTPHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/message/conversations", h.handleConversations)
	mux.HandleFunc("/api/v1/message/conversations/", h.handleConversationByID)
	mux.HandleFunc("/api/v1/message/send", h.handleSend)
	mux.HandleFunc("/api/v1/message/unread/", h.handleUnread)
	mux.HandleFunc("/api/v1/message/replies", h.handleReplies)
	mux.HandleFunc("/api/v1/message/at", h.handleAt)
	mux.HandleFunc("/api/v1/message/likes", h.handleLikes)
	mux.HandleFunc("/api/v1/message/system", h.handleSystem)
	mux.HandleFunc("/api/v1/message/settings", h.handleSettings)
	mux.HandleFunc("/api/v1/message/batch/read", h.handleBatchRead)
	mux.HandleFunc("/api/v1/message/", h.handleMessageByID)
	mux.HandleFunc("/api/v1/message/admin/", h.handleAdminBroadcast)
}

func (h *MessageHTTPHandler) handleConversations(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	list, _ := h.repo.GetConversations(r.Context(), userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *MessageHTTPHandler) handleConversationByID(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/message/conversations/"), "/")
	id, _ := strconv.ParseInt(parts[0], 10, 64)

	if len(parts) >= 2 && parts[1] == "messages" && r.Method == "GET" {
		page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
		size, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 50 {
			size = 20
		}
		msgs, _ := h.repo.GetMessages(r.Context(), id, int32(page), int32(size))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msgs)
		return
	}

	switch r.Method {
	case "GET":
		convs, _ := h.repo.GetConversations(r.Context(), userID)
		for _, c := range convs {
			if c.ID == id {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(c)
				return
			}
		}
		http.Error(w, "not found", 404)
	case "DELETE":
		_, err := h.repo.db.ExecContext(r.Context(), `DELETE FROM conversations WHERE id = $1 AND user_id = $2`, id, userID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *MessageHTTPHandler) handleSend(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	var req struct {
		ReceiverID int64  `json:"receiver_id"`
		Content    string `json:"content"`
		MsgType    int32  `json:"message_type"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Content == "" {
		http.Error(w, "content required", 400)
		return
	}
	if req.MsgType == 0 {
		req.MsgType = 1
	}
	msg, err := h.repo.SendMessage(r.Context(), userID, req.ReceiverID, req.Content, req.MsgType)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	h.notif.Send(req.ReceiverID, &NotificationEvent{
		Type: "message", Content: req.Content, FromUID: userID,
		CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func (h *MessageHTTPHandler) handleUnread(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	cnt, _ := h.repo.GetUnreadCount(r.Context(), userID)
	json.NewEncoder(w).Encode(map[string]int32{"unread_count": cnt})
}

func (h *MessageHTTPHandler) handleReplies(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	rows, _ := h.repo.db.QueryContext(r.Context(),
		`SELECT m.id, m.sender_id, m.content, m.created_at FROM messages m
		 WHERE m.receiver_id = $1 AND m.message_type = 2 ORDER BY m.created_at DESC LIMIT 20`, userID)
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, sID int64
		var content, t string
		rows.Scan(&id, &sID, &content, &t)
		list = append(list, map[string]interface{}{"id": id, "sender_id": sID, "content": content, "created_at": t})
	}
	json.NewEncoder(w).Encode(list)
}

func (h *MessageHTTPHandler) handleAt(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	rows, _ := h.repo.db.QueryContext(r.Context(),
		`SELECT m.id, m.sender_id, m.content, m.created_at FROM messages m
		 WHERE m.receiver_id = $1 AND m.message_type = 3 ORDER BY m.created_at DESC LIMIT 20`, userID)
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, sID int64
		var content, t string
		rows.Scan(&id, &sID, &content, &t)
		list = append(list, map[string]interface{}{"id": id, "sender_id": sID, "content": content, "created_at": t})
	}
	json.NewEncoder(w).Encode(list)
}

func (h *MessageHTTPHandler) handleLikes(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	rows, _ := h.repo.db.QueryContext(r.Context(),
		`SELECT m.id, m.sender_id, m.content, m.created_at FROM messages m
		 WHERE m.receiver_id = $1 AND m.message_type IN (4,6) ORDER BY m.created_at DESC LIMIT 20`, userID)
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, sID int64
		var content, t string
		rows.Scan(&id, &sID, &content, &t)
		list = append(list, map[string]interface{}{"id": id, "sender_id": sID, "content": content, "created_at": t})
	}
	json.NewEncoder(w).Encode(list)
}

func (h *MessageHTTPHandler) handleSystem(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	rows, _ := h.repo.db.QueryContext(r.Context(),
		`SELECT m.id, m.content, m.created_at FROM messages m
		 WHERE m.receiver_id = $1 AND m.message_type = 5 ORDER BY m.created_at DESC LIMIT 20`, userID)
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id int64
		var content, t string
		rows.Scan(&id, &content, &t)
		list = append(list, map[string]interface{}{"id": id, "content": content, "created_at": t})
	}
	json.NewEncoder(w).Encode(list)
}

func (h *MessageHTTPHandler) handleMessageByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/message/"), "/")
	id, _ := strconv.ParseInt(parts[0], 10, 64)

	switch r.Method {
	case "PUT": // mark read
		userID := getUserIDFromHeader(r)
		if len(parts) >= 2 && parts[1] == "read" {
			_, err := h.repo.db.ExecContext(r.Context(),
				`UPDATE messages SET is_read = 1 WHERE id = $1 AND receiver_id = $2`, id, userID)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Write([]byte(`{"status":"ok"}`))
		}
	case "DELETE":
		_, err := h.repo.db.ExecContext(r.Context(), `DELETE FROM messages WHERE id = $1`, id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *MessageHTTPHandler) handleSettings(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	switch r.Method {
	case "GET":
		row := h.repo.db.QueryRowContext(r.Context(),
			`SELECT COALESCE(private_message_notification,1), COALESCE(reply_notification,1), COALESCE(at_notification,1),
			        COALESCE(like_notification,1), COALESCE(system_notification,1)
			 FROM message_settings WHERE user_id = $1`, userID)
		var pm, reply, at, like, sys int32
		if err := row.Scan(&pm, &reply, &at, &like, &sys); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"private_message_notification": 1, "reply_notification": 1, "at_notification": 1,
				"like_notification": 1, "system_notification": 1,
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"private_message_notification": pm, "reply_notification": reply, "at_notification": at,
			"like_notification": like, "system_notification": sys,
		})
	case "PUT":
		var req struct {
			PrivateMessageNotification *int32 `json:"private_message_notification"`
			ReplyNotification          *int32 `json:"reply_notification"`
			AtNotification             *int32 `json:"at_notification"`
			LikeNotification           *int32 `json:"like_notification"`
			SystemNotification         *int32 `json:"system_notification"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		_, _ = h.repo.db.ExecContext(r.Context(),
			`INSERT INTO message_settings (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING`, userID)
		_, _ = h.repo.db.ExecContext(r.Context(),
			`UPDATE message_settings SET
			   private_message_notification = COALESCE($2, private_message_notification),
			   reply_notification          = COALESCE($3, reply_notification),
			   at_notification             = COALESCE($4, at_notification),
			   like_notification           = COALESCE($5, like_notification),
			   system_notification         = COALESCE($6, system_notification),
			   updated_at = NOW()
			 WHERE user_id = $1`,
			userID, req.PrivateMessageNotification, req.ReplyNotification, req.AtNotification,
			req.LikeNotification, req.SystemNotification)
		w.Write([]byte(`{"status":"ok"}`))
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *MessageHTTPHandler) handleBatchRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := getUserIDFromHeader(r)
	var req struct {
		IDs []int64 `json:"ids"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if len(req.IDs) > 0 {
		_, _ = h.repo.db.ExecContext(r.Context(),
			`UPDATE messages SET is_read = 1 WHERE id = ANY($1) AND receiver_id = $2`, pq.Array(req.IDs), userID)
	}
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *MessageHTTPHandler) handleAdminBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Content == "" {
		http.Error(w, "content required", 400)
		return
	}
	rows, _ := h.repo.db.QueryContext(r.Context(), `SELECT id FROM users`)
	defer rows.Close()
	for rows.Next() {
		var uid int64
		rows.Scan(&uid)
		msg, _ := h.repo.SendMessage(r.Context(), 0, uid, req.Content, 5)
		if msg != nil {
			h.notif.Send(uid, &NotificationEvent{Type: "system", Content: req.Content, CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z")})
		}
	}
	w.Write([]byte(`{"status":"broadcast_sent"}`))
}
