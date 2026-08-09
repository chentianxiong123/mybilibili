package core

import (
	"context"
	"database/sql"
	"time"
)

type Message struct {
	ID             int64
	SenderID       int64
	ReceiverID     int64
	ConversationID int64
	Content        string
	MessageType    int32
	IsRead         int32
	CreatedAt      time.Time
}

type Conversation struct {
	ID                 int64
	UserID             int64
	TargetUserID       int64
	LastMessageContent string
	LastMessageAt      sql.NullTime
	UnreadCount        int32
}

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) SendMessage(ctx context.Context, senderID, receiverID int64, content string, msgType int32) (*Message, error) {
	convID, err := r.getOrCreateConversation(ctx, senderID, receiverID)
	if err != nil {
		return nil, err
	}

	msg := &Message{}
	err = r.db.QueryRowContext(ctx,
		`INSERT INTO messages (sender_id, receiver_id, conversation_id, content, message_type)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, sender_id, receiver_id, conversation_id, content, message_type, is_read, created_at`,
		senderID, receiverID, convID, content, msgType,
	).Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.ConversationID, &msg.Content, &msg.MessageType, &msg.IsRead, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}

	r.db.ExecContext(ctx,
		`UPDATE conversations SET last_message_content = $1, last_message_at = NOW(), unread_count = unread_count + 1
		 WHERE id = $2`, content, convID)

	return msg, nil
}

func (r *MessageRepository) getOrCreateConversation(ctx context.Context, userID1, userID2 int64) (int64, error) {
	var convID int64
	err := r.db.QueryRowContext(ctx,
		`SELECT id FROM conversations WHERE user_id = $1 AND target_user_id = $2`, userID1, userID2).Scan(&convID)
	if err == nil {
		return convID, nil
	}

	err = r.db.QueryRowContext(ctx,
		`INSERT INTO conversations (user_id, target_user_id) VALUES ($1, $2) RETURNING id`, userID1, userID2).Scan(&convID)
	if err != nil {
		return 0, err
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO conversations (user_id, target_user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID2, userID1)
	return convID, err
}

func (r *MessageRepository) GetConversations(ctx context.Context, userID int64) ([]*Conversation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT c.id, c.user_id, c.target_user_id, c.last_message_content, c.last_message_at, c.unread_count
		 FROM conversations c WHERE c.user_id = $1 ORDER BY c.last_message_at DESC NULLS LAST`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Conversation
	for rows.Next() {
		c := &Conversation{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.TargetUserID, &c.LastMessageContent, &c.LastMessageAt, &c.UnreadCount); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *MessageRepository) GetMessages(ctx context.Context, conversationID int64, page, pageSize int32) ([]*Message, error) {
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, sender_id, receiver_id, conversation_id, content, message_type, is_read, created_at
		 FROM messages WHERE conversation_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		conversationID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Message
	for rows.Next() {
		m := &Message{}
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.ConversationID, &m.Content, &m.MessageType, &m.IsRead, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

func (r *MessageRepository) MarkAsRead(ctx context.Context, conversationID, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE messages SET is_read = 1 WHERE conversation_id = $1 AND receiver_id = $2 AND is_read = 0`,
		conversationID, userID)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE conversations SET unread_count = 0 WHERE id = $1 AND user_id = $2`, conversationID, userID)
	return err
}

func (r *MessageRepository) GetUnreadCount(ctx context.Context, userID int64) (int32, error) {
	var count int32
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(unread_count), 0) FROM conversations WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

type NotificationBroadcaster struct {
	channels map[int64]chan *NotificationEvent
}

type NotificationEvent struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	FromUID   int64  `json:"from_uid"`
	FromName  string `json:"from_name"`
	CreatedAt string `json:"created_at"`
}

func NewNotificationBroadcaster() *NotificationBroadcaster {
	return &NotificationBroadcaster{channels: make(map[int64]chan *NotificationEvent)}
}

func (b *NotificationBroadcaster) Subscribe(userID int64) <-chan *NotificationEvent {
	if b.channels[userID] == nil {
		b.channels[userID] = make(chan *NotificationEvent, 50)
	}
	return b.channels[userID]
}

func (b *NotificationBroadcaster) Send(userID int64, event *NotificationEvent) {
	if b.channels[userID] != nil {
		select {
		case b.channels[userID] <- event:
		default:
		}
	}
}
