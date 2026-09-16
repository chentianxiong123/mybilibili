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