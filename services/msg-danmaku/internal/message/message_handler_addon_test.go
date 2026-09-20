package message

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/pkg/auth"
)

// doMessageNoAuth 与 doMessage 相同，但不注入 X-User-Id 头。
func doMessageNoAuth(t *testing.T, h *MessageHTTPHandler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestMessageHandlers_Unauthorized(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"conversations", http.MethodGet, "/api/v1/message/conversations", ""},
		{"send", http.MethodPost, "/api/v1/message/send", `{"receiverId":1,"content":"x"}`},
		{"unread", http.MethodGet, "/api/v1/message/unread", ""},
		{"unread_counts", http.MethodGet, "/api/v1/message/unread/counts", ""},
		{"replies", http.MethodGet, "/api/v1/message/replies", ""},
		{"at", http.MethodGet, "/api/v1/message/at", ""},
		{"likes", http.MethodGet, "/api/v1/message/likes", ""},
		{"system", http.MethodGet, "/api/v1/message/system", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, _ := newTestMessageHandler(t)
			rr := doMessageNoAuth(t, h, tc.method, tc.path, tc.body)
			assert.Equal(t, http.StatusUnauthorized, rr.Code)
		})
	}
}

func TestGetUserID_FromJWT(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)
	j := auth.NewJWT("test-secret")
	h := NewMessageHTTPHandler(repo, NewNotificationBroadcaster(), nil, j)

	token, err := j.Generate(1001)
	require.NoError(t, err)

	mock.ExpectQuery(`SELECT c.id, c.user_id, c.target_user_id`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "target_user_id", "target_user_name", "target_user_avatar",
			"last_message_content", "last_message_time", "unread_count",
		}))

	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/message/conversations", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserID_FromJWT_InvalidToken(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)
	j := auth.NewJWT("test-secret")
	h := NewMessageHTTPHandler(repo, NewNotificationBroadcaster(), nil, j)

	// jwt 解析失败会回退；此处无 header 且 token 非法 -> 401
	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/message/unread", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

type noFlusherWriter struct {
	w http.ResponseWriter
}

func (n noFlusherWriter) Header() http.Header       { return n.w.Header() }
func (n noFlusherWriter) Write(b []byte) (int, error) { return n.w.Write(b) }
func (n noFlusherWriter) WriteHeader(s int)         { n.w.WriteHeader(s) }

func TestHandleNotificationSSE_Unauthorized(t *testing.T) {
	h, _ := newTestMessageHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/sse/notification", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandleNotificationSSE_Unauthorized_InvalidToken(t *testing.T) {
	h, _ := newTestMessageHandler(t)
	h.jwt = auth.NewJWT("test-secret")
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/sse/notification?token=bogus", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandleNotificationSSE_NoFlusher(t *testing.T) {
	h, _ := newTestMessageHandler(t)
	h.jwt = auth.NewJWT("test-secret")
	token, _ := h.jwt.Generate(1001)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/sse/notification?token="+token, nil)
	rec := httptest.NewRecorder()
	w := noFlusherWriter{w: rec}
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandleNotificationSSE_Event(t *testing.T) {
	h, _ := newTestMessageHandler(t)
	h.jwt = auth.NewJWT("test-secret")
	token, _ := h.jwt.Generate(1001)
	mux := http.NewServeMux()
	h.Register(mux)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 预订阅：先写 map，避免与 handler 内部 Subscribe 并发写导致 data race
	h.notif.Subscribe(1001)

	req := httptest.NewRequest(http.MethodGet, "/sse/notification?token="+token, nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		mux.ServeHTTP(rr, req)
		close(done)
	}()

	// 等 handler 进入订阅循环后广播，再退出
	time.Sleep(100 * time.Millisecond)
	h.notif.Send(1001, &NotificationEvent{Type: "message", Content: "sse hello", FromUID: 101})
	time.Sleep(100 * time.Millisecond)
	cancel()
	<-done

	body := rr.Body.String()
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, body, "connected")
	assert.Contains(t, body, "sse hello")
}

func TestHandleConversationByID_GET_NotFound(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT c.id, c.user_id, c.target_user_id`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "target_user_id", "target_user_name", "target_user_avatar",
			"last_message_content", "last_message_time", "unread_count",
		}))

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/conversations/999", "")
	assert.Equal(t, http.StatusNotFound, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleConversationByID_PUT_NoUnreadField(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)
	h := NewMessageHTTPHandler(repo, NewNotificationBroadcaster(), nil)

	// body 无 unread_count 字段，不执行 SQL
	rr := doMessage(t, h, http.MethodPut, "/api/v1/message/conversations/50", `{"foo":1}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"status":"ok"}`, strings.TrimSpace(rr.Body.String()))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleConversationByID_DELETE_Error(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectExec(`DELETE FROM conversations WHERE id = \$1 AND user_id = \$2`).
		WithArgs(int64(50), int64(1001)).
		WillReturnError(errDB)

	rr := doMessage(t, h, http.MethodDelete, "/api/v1/message/conversations/50", "")
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessageByID_GET_NotFound(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectQuery(`SELECT id, sender_id, receiver_id, content, message_type, is_read, created_at`).
		WithArgs(int64(80), int64(1001)).
		WillReturnError(sql.ErrNoRows)

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/80", "")
	assert.Equal(t, http.StatusNotFound, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessageByID_PUT_ReadError(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectExec(`UPDATE messages SET is_read = 1 WHERE id = \$1 AND receiver_id = \$2`).
		WithArgs(int64(80), int64(1001)).
		WillReturnError(errDB)

	rr := doMessage(t, h, http.MethodPut, "/api/v1/message/80/read", "")
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMessageByID_DELETE_Error(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	mock.ExpectExec(`DELETE FROM messages WHERE id = \$1 AND receiver_id = \$2`).
		WithArgs(int64(80), int64(1001)).
		WillReturnError(errDB)

	rr := doMessage(t, h, http.MethodDelete, "/api/v1/message/80", "")
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSend_RepoError(t *testing.T) {
	h, mock := newTestMessageHandler(t)

	// getOrCreateConversation 查询报错 -> SendMessage 返回错误 -> 400
	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(1001), int64(1002)).
		WillReturnError(errDB)

	rr := doMessage(t, h, http.MethodPost, "/api/v1/message/send", `{"receiverId":1002,"content":"hi"}`)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSend_WithCacheInvalidate(t *testing.T) {
	mr, cache := newTestUnreadCache(t)
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)
	h := NewMessageHTTPHandler(repo, NewNotificationBroadcaster(), cache)

	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(1001), int64(1002)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(50))
	now := time.Now()
	mock.ExpectQuery(`INSERT INTO messages`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "receiver_id", "conversation_id", "content", "message_type", "is_read", "created_at",
		}).AddRow(100, 1001, 1002, 50, "hi", 1, 0, now))
	mock.ExpectExec(`UPDATE conversations SET last_message_content`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mr.HSet("unread:1002", "private", "3")
	rr := doMessage(t, h, http.MethodPost, "/api/v1/message/send", `{"receiverId":1002,"content":"hi"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.False(t, mr.Exists("unread:1002"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUnreadCounts_WithCacheHit(t *testing.T) {
	mr, cache := newTestUnreadCache(t)
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)
	h := NewMessageHTTPHandler(repo, NewNotificationBroadcaster(), cache)

	mock.ExpectQuery(`SELECT`).
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"private", "reply", "at", "like", "system"}).
			AddRow(0, 0, 0, 0, 0))
	mr.HSet("unread:1001", "private", "9", "reply", "1")

	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/unread/counts", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Private int32 `json:"private"`
			Reply   int32 `json:"reply"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, int32(9), resp.Data.Private)
	assert.Equal(t, int32(1), resp.Data.Reply)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBatchRead_405(t *testing.T) {
	h, _ := newTestMessageHandler(t)
	rr := doMessage(t, h, http.MethodGet, "/api/v1/message/batch/read", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleBatchRead_WithCacheInvalidate(t *testing.T) {
	mr, cache := newTestUnreadCache(t)
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)
	h := NewMessageHTTPHandler(repo, NewNotificationBroadcaster(), cache)

	mock.ExpectExec(`UPDATE messages SET is_read = 1 WHERE id = ANY`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE conversations SET unread_count = 0`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// pushUnread -> cache.Counts miss -> GetUnreadCountsByType
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"private", "reply", "at", "like", "system"}).AddRow(0, 0, 0, 0, 0))

	mr.HSet("unread:1001", "private", "2")
	rr := doMessage(t, h, http.MethodPut, "/api/v1/message/batch/read", `{"ids":[1]}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	// Invalidate deletes old cache; pushUnread repopulates with fresh counts (all 0)
	val := mr.HGet("unread:1001", "private")
	assert.Equal(t, "0", val)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExtractCommentText(t *testing.T) {
	assert.Equal(t, "", extractCommentText("没有开引号", "《"))
	assert.Equal(t, "标题", extractCommentText("赞了你的视频《标题》", "《"))
	assert.Equal(t, "好", extractCommentText(`赞了你的评论"好"`, `"`))
	assert.Equal(t, "中文右引号", extractCommentText("赞了你的评论\"中文右引号”", "\""))
	assert.Equal(t, "没闭合", extractCommentText("赞了你的视频《没闭合", "《"))
	assert.Equal(t, "尾部", extractCommentText(`赞了你的评论"尾部`, "\""))
}

func TestNotificationBroadcaster_SubscribeUnsubscribeSend(t *testing.T) {
	b := NewNotificationBroadcaster()

	ch := b.Subscribe(1)
	require.NotNil(t, ch)
	// 同用户返回同一 channel
	ch2 := b.Subscribe(1)
	assert.Equal(t, ch, ch2)

	// 已订阅的 receive
	b.Send(1, &NotificationEvent{Type: "a"})
	select {
	case ev := <-ch:
		assert.Equal(t, "a", ev.Type)
	default:
		t.Fatal("expected event")
	}

	// 未订阅用户 no-op，不 panic
	b.Send(999, &NotificationEvent{Type: "b"})

	// buffer 满时丢弃，不阻塞
	full := b.Subscribe(2)
	for i := 0; i < 50; i++ {
		b.Send(2, &NotificationEvent{Content: "x"})
	}
	select {
	case ev := <-full:
		assert.NotNil(t, ev)
	default:
		t.Fatal("expected buffered event")
	}

	// Unsubscribe 正确删除匹配 channel
	b.Unsubscribe(1, ch)
	assert.Nil(t, b.channels[1])

	// Unsubscribe 传入不匹配 channel：保留原 channel
	ch1 := b.Subscribe(3)
	other := make(chan *NotificationEvent, 1)
	b.Unsubscribe(3, other)
	assert.NotNil(t, b.channels[3])
	// 原 channel 仍可接收
	b.Send(3, &NotificationEvent{Type: "kept"})
	select {
	case ev := <-ch1:
		assert.Equal(t, "kept", ev.Type)
	default:
		t.Fatal("expected kept channel to still deliver")
	}
}

func TestHandleConversations_WithCacheNilAndJWT_NoToken(t *testing.T) {
	// jwt 非 nil 但无 Authorization 头，且无 X-User-Id -> 401
	db, _ := newMessageDB(t)
	repo := NewMessageRepository(db)
	j := auth.NewJWT("test-secret")
	h := NewMessageHTTPHandler(repo, NewNotificationBroadcaster(), nil, j)

	rr := doMessageNoAuth(t, h, http.MethodGet, "/api/v1/message/unread", "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}