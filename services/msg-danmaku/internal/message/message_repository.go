package message

import (
	"context"
	"database/sql"
	"time"
)

type Message struct {
	ID             int64     `json:"id"`
	SenderID       int64     `json:"sender_id"`
	ReceiverID     int64     `json:"receiver_id"`
	ConversationID int64     `json:"conversation_id"`
	Content        string    `json:"content"`
	MessageType    int32     `json:"message_type"`
	IsRead         int32     `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}

type Conversation struct {
	ID                 int64  `json:"id"`
	UserID             int64  `json:"user_id"`
	TargetUserID       int64  `json:"target_user_id"`
	TargetUserName     string `json:"target_user_name"`
	TargetUserAvatar   string `json:"target_user_avatar"`
	LastMessageContent string `json:"last_message_content"`
	LastMessageTime    string `json:"last_message_time"`
	UnreadCount        int32  `json:"unread_count"`
}

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) DB() *sql.DB {
	return r.db
}

func (r *MessageRepository) SendMessage(ctx context.Context, senderID, receiverID int64, content string, msgType int32) (*Message, error) {
	// 会话是按人镜像的两行（sender→receiver、receiver→sender）。
	// 消息必须写双份，否则接收方读自己的会话行永远看不到对方发的
	// （曾导致私信"能发不能收"：历史与 SSE 都只对发送方可见）。
	senderConvID, receiverConvID, err := r.getOrCreateConversation(ctx, senderID, receiverID)
	if err != nil {
		return nil, err
	}

	insert := func(convID int64) (*Message, error) {
		msg := &Message{}
		err := r.db.QueryRowContext(ctx,
			`INSERT INTO messages (sender_id, receiver_id, conversation_id, content, message_type)
			 VALUES ($1, $2, $3, $4, $5) RETURNING id, sender_id, receiver_id, conversation_id, content, message_type, is_read, created_at`,
			senderID, receiverID, convID, content, msgType,
		).Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.ConversationID, &msg.Content, &msg.MessageType, &msg.IsRead, &msg.CreatedAt)
		return msg, err
	}

	msg, err := insert(senderConvID)
	if err != nil {
		return nil, err
	}
	if _, err := insert(receiverConvID); err != nil {
		return nil, err
	}

	// 发送方行只更新最后一条（未读不动：自己发的没有未读）
	r.db.ExecContext(ctx,
		`UPDATE conversations SET last_message_content = $1, last_message_time = NOW()
		 WHERE id = $2`, content, senderConvID)
	// 接收方行更新最后一条且未读+1
	r.db.ExecContext(ctx,
		`UPDATE conversations SET last_message_content = $1, last_message_time = NOW(), unread_count = unread_count + 1
		 WHERE id = $2`, content, receiverConvID)

	return msg, nil
}

// 返回发送方与接收方各自的会话行 id（不存在则建出镜像两行）。
// 只有一个调用方（SendMessage），签名可直接改。
func (r *MessageRepository) getOrCreateConversation(ctx context.Context, userID1, userID2 int64) (senderConvID, receiverConvID int64, err error) {
	err = r.db.QueryRowContext(ctx,
		`SELECT id FROM conversations WHERE user_id = $1 AND target_user_id = $2`, userID1, userID2).Scan(&senderConvID)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	if err == nil {
		// 发送方行存在时接收方行必然已建（建行时双写），直接查 id
		err = r.db.QueryRowContext(ctx,
			`SELECT id FROM conversations WHERE user_id = $1 AND target_user_id = $2`, userID2, userID1).Scan(&receiverConvID)
		if err != nil {
			return 0, 0, err
		}
		return senderConvID, receiverConvID, nil
	}

	err = r.db.QueryRowContext(ctx,
		`INSERT INTO conversations (user_id, target_user_id) VALUES ($1, $2) RETURNING id`, userID1, userID2).Scan(&senderConvID)
	if err != nil {
		return 0, 0, err
	}

	err = r.db.QueryRowContext(ctx,
		`INSERT INTO conversations (user_id, target_user_id) VALUES ($1, $2) ON CONFLICT (user_id, target_user_id) DO UPDATE SET target_user_id = EXCLUDED.target_user_id RETURNING id`, userID2, userID1).Scan(&receiverConvID)
	if err != nil {
		return 0, 0, err
	}
	return senderConvID, receiverConvID, nil
}

func (r *MessageRepository) GetConversations(ctx context.Context, userID int64) ([]*Conversation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT c.id, c.user_id, c.target_user_id,
		        COALESCE(u.nickname, u.username, ''), COALESCE(u.avatar, ''),
		        c.last_message_content, COALESCE(c.last_message_time::text,''),
		        (SELECT COUNT(*) FROM messages m WHERE m.conversation_id = c.id AND m.receiver_id = c.user_id AND m.is_read = 0) AS real_unread
		 FROM conversations c
		 LEFT JOIN users u ON u.id = c.target_user_id
		 WHERE c.user_id = $1 ORDER BY c.last_message_time DESC NULLS LAST`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Conversation
	for rows.Next() {
		c := &Conversation{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.TargetUserID, &c.TargetUserName, &c.TargetUserAvatar, &c.LastMessageContent, &c.LastMessageTime, &c.UnreadCount); err != nil {
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

	list := make([]*Message, 0)
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
		`SELECT COALESCE(COUNT(*),0) FROM messages WHERE receiver_id = $1 AND is_read = 0`, userID).Scan(&count)
	return count, err
}

// GetUnreadCountsByType 从 DB 重算六类未读数（源真相），供 Redis 缓存回源。
// message_type: 1=私信 2=回复 3=@ 4=点赞稿件 5=系统 6=点赞评论
func (r *MessageRepository) GetUnreadCountsByType(ctx context.Context, userID int64) map[string]int32 {
	counts := map[string]int32{
		"private": 0, "reply": 0, "at": 0, "like": 0, "system": 0, "dynamic": 0,
	}
	// 全部按「真实未读消息条数」统计，避免依赖可能漂移的 conversations.unread_count 列。
	// message_type: 1=私信 2=回复 3=@ 4=点赞稿件 5=系统 6=点赞评论
	var private, reply, at, like, system int32
	_ = r.db.QueryRowContext(ctx,
		`SELECT
		   COALESCE(COUNT(*) FILTER (WHERE message_type = 1),0),
		   COALESCE(COUNT(*) FILTER (WHERE message_type = 2),0),
		   COALESCE(COUNT(*) FILTER (WHERE message_type = 3),0),
		   COALESCE(COUNT(*) FILTER (WHERE message_type IN (4,6)),0),
		   COALESCE(COUNT(*) FILTER (WHERE message_type = 5),0)
		 FROM messages WHERE receiver_id = $1 AND is_read = 0`, userID).
		Scan(&private, &reply, &at, &like, &system)
	counts["private"] = private
	counts["reply"] = reply
	counts["at"] = at
	counts["like"] = like
	counts["system"] = system
	counts["dynamic"] = reply + at + like + system
	return counts
}

type NotificationBroadcaster struct {
	channels map[int64]chan *NotificationEvent
}

type NotificationEvent struct {
	Type      string         `json:"type"`
	Content   string         `json:"content,omitempty"`
	FromUID   int64          `json:"from_uid,omitempty"`
	FromName  string         `json:"from_name,omitempty"`
	CreatedAt string         `json:"created_at,omitempty"`
	Data      map[string]int32 `json:"data,omitempty"`
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

func (b *NotificationBroadcaster) Unsubscribe(userID int64, ch <-chan *NotificationEvent) {
	if b.channels[userID] == ch {
		delete(b.channels, userID)
	}
}

func (b *NotificationBroadcaster) Send(userID int64, event *NotificationEvent) {
	if b.channels[userID] != nil {
		select {
		case b.channels[userID] <- event:
		default:
		}
	}
}
