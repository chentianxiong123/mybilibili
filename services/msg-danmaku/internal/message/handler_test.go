package message

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testUserID = "1001"

func newTestMessageHandler(t *testing.T) (*MessageHTTPHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewMessageRepository(db)
	h := NewMessageHTTPHandler(repo, NewNotificationBroadcaster(), nil)
	t.Cleanup(func() { db.Close() })
	return h, mock
}

func doMessage(t *testing.T, h *MessageHTTPHandler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	h.Register(mux)
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("X-User-Id", testUserID)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestHandleConversations_200(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT c.id, c.user_id, c.target_user_id`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "target_user_id", "target_user_name", "target_user_avatar",
			"last_message_content", "last_message_time", "unread_count",
		}).AddRow(1, 1001, 1002, "对方", "http://avatar", "你好", now.Format("2006-01-02 15:04:05"), 3))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/conversations", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int            `json:"code"`
		Data []*Conversation `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, int64(1002), resp.Data[0].TargetUserID)
	assert.Equal(t, "你好", resp.Data[0].LastMessageContent)
	assert.Equal(t, int32(3), resp.Data[0].UnreadCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSendMessage_200(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	// getOrCreateConversation：第一次 SELECT 无结果走 INSERT
	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(1001), int64(1002)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO conversations \(user_id, target_user_id\)`).
		WithArgs(int64(1001), int64(1002)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(50))
	mock.ExpectExec(`INSERT INTO conversations \(user_id, target_user_id\) VALUES \(\$1, \$2\) ON CONFLICT DO NOTHING`).
		WithArgs(int64(1002), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 写入消息并返回
	now := time.Now()
	mock.ExpectQuery(`INSERT INTO messages`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "receiver_id", "conversation_id", "content", "message_type", "is_read", "created_at",
		}).AddRow(100, 1001, 1002, 50, "你好呀", 1, 0, now))
	mock.ExpectExec(`UPDATE conversations SET last_message_content`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doMessage(t, h, http.MethodPost, "/api/v1/message/send", `{"receiverId":1002,"content":"你好呀"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			ID         int64  `json:"id"`
			ReceiverID int64  `json:"receiver_id"`
			Content    string `json:"content"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 200, resp.Code)
	assert.Equal(t, int64(100), resp.Data.ID)
	assert.Equal(t, "你好呀", resp.Data.Content)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUnread_200(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT COALESCE\(COUNT\(\*\),0\) FROM messages WHERE receiver_id = \$1 AND is_read = 0`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/unread", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Reply  int32 `json:"reply"`
			At     int32 `json:"at"`
			Like   int32 `json:"like"`
			System int32 `json:"system"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, int32(8), resp.Data.Reply)
	assert.Equal(t, int32(8), resp.Data.System)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleReplies_200(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	now := time.Now()
	mock.ExpectQuery(`FROM messages m`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "username", "user_avatar", "content", "is_read", "created_at", "comment_id", "manuscript_id",
		}).AddRow(300, 1002, "对方", "http://avatar", "回复内容", 0, now.Format("2006-01-02 15:04:05"), 55, 66))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/replies", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID        int64  `json:"id"`
			Content   string `json:"content"`
			ActionText string `json:"actionText"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "回复内容", resp.Data[0].Content)
	assert.Equal(t, "回复了你的评论", resp.Data[0].ActionText)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSendMessage_400(t *testing.T) {
	h, _ := newTestMessageHandler(t)

	rr := doMessage(t, h, http.MethodPost, "/api/v1/message/send", `{"receiverId":0,"content":""}`)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 400, resp.Code)
}

func TestHandleConversations_Empty(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT c.id, c.user_id, c.target_user_id`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "target_user_id", "target_user_name", "target_user_avatar",
			"last_message_content", "last_message_time", "unread_count",
		}))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/conversations", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int             `json:"code"`
		Data []*Conversation `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleConversationByID_Messages(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT id, sender_id, receiver_id, conversation_id, content, message_type, is_read, created_at`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "receiver_id", "conversation_id", "content", "message_type", "is_read", "created_at",
		}).AddRow(1, 1001, 1002, 50, "hello", 1, 0, now))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/conversations/50/messages?page=1&page_size=20", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID      int64  `json:"id"`
			Content string `json:"content"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "hello", resp.Data[0].Content)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleConversationByID_Messages_Empty(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT id, sender_id, receiver_id, conversation_id, content, message_type, is_read, created_at`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "receiver_id", "conversation_id", "content", "message_type", "is_read", "created_at",
		}))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/conversations/50/messages", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleConversationByID_PUT(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectExec(`UPDATE conversations SET unread_count`).
		WithArgs(int32(0), int64(50), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doMessage(t, h, http.MethodPut, "/api/v1/message/conversations/50", `{"unread_count":0}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleConversationByID_DELETE(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectExec(`DELETE FROM conversations WHERE id = \$1 AND user_id = \$2`).
		WithArgs(int64(50), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doMessage(t, h, http.MethodDelete, "/api/v1/message/conversations/50", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUnreadCounts_200(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"private", "reply", "at", "like", "system"}).
			AddRow(1, 2, 3, 4, 5))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/unread/counts", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Private int32 `json:"private"`
			Reply   int32 `json:"reply"`
			At      int32 `json:"at"`
			Like    int32 `json:"like"`
			System  int32 `json:"system"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, int32(1), resp.Data.Private)
	assert.Equal(t, int32(2), resp.Data.Reply)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleConversationUnread_200(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT unread_count FROM conversations`).
		WithArgs(int64(50), int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"unread_count"}).AddRow(int32(3)))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/conversation/unread/50", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		UnreadCount int32 `json:"unread_count"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, int32(3), resp.UnreadCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAt_200(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT m.id, m.sender_id`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "username", "user_avatar", "content", "is_read", "created_at", "manuscript_id",
		}).AddRow(500, 1002, "用户A", "http://a.png", "@你好", 0, now.Format("2006-01-02 15:04:05"), 10))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/at", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID         int64  `json:"id"`
			Content    string `json:"content"`
			ActionText string `json:"actionText"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "@你好", resp.Data[0].Content)
	assert.Equal(t, "在评论中@了我", resp.Data[0].ActionText)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAt_Empty(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT m.id, m.sender_id`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "username", "user_avatar", "content", "is_read", "created_at", "manuscript_id",
		}))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/at", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLikes_200(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT m.id, m.sender_id`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "username", "user_avatar", "content", "is_read", "created_at",
			"message_type", "manuscript_id", "comment_id", "video_title", "video_cover", "comment_content",
		}).AddRow(600, 1002, "用户B", "http://b.png", "赞了你的视频《测试》", 0, now.Format("2006-01-02 15:04:05"),
			4, 10, 0, "", "", ""))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/likes", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID         int64  `json:"id"`
			ActionText string `json:"actionText"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "赞了你的视频", resp.Data[0].ActionText)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLikes_CommentType(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT m.id, m.sender_id`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "username", "user_avatar", "content", "is_read", "created_at",
			"message_type", "manuscript_id", "comment_id", "video_title", "video_cover", "comment_content",
		}).AddRow(601, 1002, "用户C", "http://c.png", "赞了你的评论\"好\"", 0, now.Format("2006-01-02 15:04:05"),
			6, 10, 55, "", "", ""))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/likes", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ActionText string `json:"actionText"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "赞了你的评论", resp.Data[0].ActionText)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLikes_Empty(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT m.id, m.sender_id`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "username", "user_avatar", "content", "is_read", "created_at",
			"message_type", "manuscript_id", "comment_id", "video_title", "video_cover", "comment_content",
		}))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/likes", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSystem_200(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT m.id, m.content, m.is_read, COALESCE`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "content", "is_read", "created_at",
		}).AddRow(700, "系统通知内容", 0, now.Format("2006-01-02 15:04:05")))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/system", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID         int64  `json:"id"`
			Content    string `json:"content"`
			ActionText string `json:"actionText"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "系统通知内容", resp.Data[0].Content)
	assert.Equal(t, "系统通知", resp.Data[0].ActionText)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSystem_Empty(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT m.id, m.content, m.is_read, COALESCE`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "content", "is_read", "created_at",
		}))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/system", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessageByID_GET(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT id, sender_id, receiver_id, content, message_type, is_read, created_at`).
		WithArgs(int64(80), int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "receiver_id", "content", "message_type", "is_read", "created_at",
		}).AddRow(80, 1002, 1001, "你好", 1, 0, now))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/80", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		ID      int64  `json:"id"`
		Content string `json:"content"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, int64(80), resp.ID)
	assert.Equal(t, "你好", resp.Content)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessageByID_PUT_Read(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectExec(`UPDATE messages SET is_read = 1 WHERE id = \$1 AND receiver_id = \$2`).
		WithArgs(int64(80), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doMessage(t, h, http.MethodPut, "/api/v1/message/80/read", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessageByID_DELETE(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectExec(`DELETE FROM messages WHERE id = \$1 AND receiver_id = \$2`).
		WithArgs(int64(80), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doMessage(t, h, http.MethodDelete, "/api/v1/message/80", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSettings_GET(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT COALESCE\(private_message_notification`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"private_message_notification", "reply_notification", "at_notification", "like_notification", "system_notification",
		}).AddRow(int32(1), int32(1), int32(0), int32(1), int32(1)))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/settings", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(1), resp["private_message_notification"])
	assert.Equal(t, float64(0), resp["at_notification"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSettings_GET_Default(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT COALESCE\(private_message_notification`).
		WithArgs(int64(1001)).
		WillReturnError(sql.ErrNoRows)

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/settings", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(1), resp["private_message_notification"])
}

func TestHandleSettings_PUT(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectExec(`INSERT INTO message_settings`).
		WithArgs(int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE message_settings SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doMessage(t, h, http.MethodPut, "/api/v1/message/settings",
		`{"reply_notification":0}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSettings_405(t *testing.T) {
	h, _ := newTestMessageHandler(t)

	rr := doMessage(t, h, http.MethodPost, "/api/v1/message/settings", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleBatchRead_200(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectExec(`UPDATE messages SET is_read = 1 WHERE id = ANY`).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(`UPDATE conversations SET unread_count = 0`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doMessage(t, h, http.MethodPut, "/api/v1/message/batch/read", `{"ids":[1,2]}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBatchRead_EmptyIDs(t *testing.T) {
	h, _ := newTestMessageHandler(t)

	rr := doMessage(t, h, http.MethodPut, "/api/v1/message/batch/read", `{"ids":[]}`)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleMessagesByUser_GET(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT id, sender_id, receiver_id, content, message_type, is_read, created_at`).
		WithArgs(int64(1001), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "receiver_id", "content", "message_type", "is_read", "created_at",
		}).AddRow(90, 1002, 1001, "msg", 1, 0, now))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/user/1001?page=1&page_size=20", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		List []struct {
			ID int64 `json:"id"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.List, 1)
	assert.Equal(t, int64(90), resp.List[0].ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessagesByUser_DELETE(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectExec(`DELETE FROM messages WHERE receiver_id = \$1`).
		WithArgs(int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 5))

	rr := doMessage(t, h, http.MethodDelete, "/api/v1/message/user/1001", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessagesByUser_Forbidden(t *testing.T) {
	h, _ := newTestMessageHandler(t)

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/user/9999", "")
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandleMessagesByUser_InvalidID(t *testing.T) {
	h, _ := newTestMessageHandler(t)

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/user/0", "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleMessagesByUser_405(t *testing.T) {
	h, _ := newTestMessageHandler(t)

	rr := doMessage(t, h, http.MethodPost, "/api/v1/message/user/1001", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleAdminBroadcast_200(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT id FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1001).AddRow(1002))

	// For user 1001
	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(0), int64(1001)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO conversations \(user_id, target_user_id\)`).
		WithArgs(int64(0), int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100))
	mock.ExpectExec(`INSERT INTO conversations \(user_id, target_user_id\) VALUES \(\$1, \$2\) ON CONFLICT DO NOTHING`).
		WithArgs(int64(1001), int64(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	now := time.Now()
	mock.ExpectQuery(`INSERT INTO messages`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "receiver_id", "conversation_id", "content", "message_type", "is_read", "created_at",
		}).AddRow(200, 0, 1001, 100, "广播消息", 5, 0, now))
	mock.ExpectExec(`UPDATE conversations SET last_message_content`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// For user 1002
	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(0), int64(1002)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO conversations \(user_id, target_user_id\)`).
		WithArgs(int64(0), int64(1002)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(101))
	mock.ExpectExec(`INSERT INTO conversations \(user_id, target_user_id\) VALUES \(\$1, \$2\) ON CONFLICT DO NOTHING`).
		WithArgs(int64(1002), int64(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO messages`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "receiver_id", "conversation_id", "content", "message_type", "is_read", "created_at",
		}).AddRow(201, 0, 1002, 101, "广播消息", 5, 0, now))
	mock.ExpectExec(`UPDATE conversations SET last_message_content`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doMessage(t, h, http.MethodPost, "/api/v1/message/admin/broadcast", `{"content":"广播消息"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAdminBroadcast_EmptyContent(t *testing.T) {
	h, _ := newTestMessageHandler(t)

	rr := doMessage(t, h, http.MethodPost, "/api/v1/message/admin/broadcast", `{"content":""}`)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleAdminBroadcast_405(t *testing.T) {
	h, _ := newTestMessageHandler(t)

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/admin/broadcast", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleSendMessage_ReceiverID_Zero(t *testing.T) {
	h, _ := newTestMessageHandler(t)

	rr := doMessage(t, h, http.MethodPost, "/api/v1/message/send", `{"receiverId":0,"content":"hi"}`)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleSendMessage_InvalidJSON(t *testing.T) {
	h, _ := newTestMessageHandler(t)

	rr := doMessage(t, h, http.MethodPost, "/api/v1/message/send", "not json")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}