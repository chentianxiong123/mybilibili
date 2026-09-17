package message

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 覆盖 handleConversationByID messages 分支中 GetMessages 报错时 msgs==nil 的兜底。
func TestHandleConversationByID_Messages_RepoError(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT id, sender_id, receiver_id, conversation_id, content, message_type, is_read, created_at`).
		WithArgs(int64(50), int32(20), int32(0)).
		WillReturnError(errDB)

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/conversations/50/messages?page=1&size=20", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Data []any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
	require.NoError(t, mock.ExpectationsWereMet())
}

// 覆盖 handleConversationByID GET 命中已存在会话（循环内 c.ID == id）的分支。
func TestHandleConversationByID_GET_Found(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT c.id, c.user_id, c.target_user_id`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "target_user_id", "target_user_name", "target_user_avatar",
			"last_message_content", "last_message_time", "unread_count",
		}).AddRow(50, 1001, 1002, "对方", "http://avatar", "last", "2024-01-01 00:00:00", 3))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/conversations/50", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var c Conversation
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &c))
	assert.Equal(t, int64(50), c.ID)
	assert.Equal(t, int64(1002), c.TargetUserID)
	require.NoError(t, mock.ExpectationsWereMet())
}

// 覆盖 handleReplies 空结果时 list==nil 兜底为 []。
func TestHandleReplies_Empty(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`FROM messages m`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "username", "user_avatar", "content", "is_read", "created_at", "comment_id", "manuscript_id",
		}))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/replies", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Data []any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
	require.NoError(t, mock.ExpectationsWereMet())
}

// 覆盖 handleMessagesByUser GET 的 page/size 兜底分支。
func TestHandleMessagesByUser_GET_PaginationDefaults(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT id, sender_id, receiver_id, content, message_type, is_read, created_at`).
		WithArgs(int64(1001), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "receiver_id", "content", "message_type", "is_read", "created_at",
		}))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/user/1001?page=0&page_size=100", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Page int `json:"page"`
		Size int `json:"size"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 20, resp.Size)
	require.NoError(t, mock.ExpectationsWereMet())
}

// 覆盖 NotificationBroadcaster.Send buffer 满时的 default 丢弃分支。
func TestNotificationBroadcaster_Send_BufferFullDrops(t *testing.T) {
	b := NewNotificationBroadcaster()
	ch := b.Subscribe(4)

	// Subscribe 容量为 50：先塞满
	for i := 0; i < 50; i++ {
		b.Send(4, &NotificationEvent{Content: "x"})
	}
	// 第 51 条触发 default 丢弃
	b.Send(4, &NotificationEvent{Content: "dropped"})

	// 还能读到首条，且缓冲外的消息没有被偷偷写入
	got := make([]string, 0, 50)
drain:
	for {
		select {
		case ev := <-ch:
			got = append(got, ev.Content)
		default:
			break drain
		}
	}
	assert.Len(t, got, 50)
	assert.NotContains(t, got, "dropped")
	assert.Nil(t, b.channels[4+1]) // 未订阅用户 no-op
}