package message

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "mybilibili/pkg/pb"
)

func TestNewGrpcServer(t *testing.T) {
	db, _ := newMessageDB(t)
	defer db.Close()
	repo := NewMessageRepository(db)
	notif := NewNotificationBroadcaster()
	cache := &UnreadCache{}
	s := NewGrpcServer(repo, notif, cache)
	require.NotNil(t, s)
	assert.Equal(t, repo, s.repo)
	assert.Equal(t, notif, s.notif)
	assert.Equal(t, cache, s.cache)
}

func TestGrpcServer_SendMessage_Success(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)
	notif := NewNotificationBroadcaster()
	s := NewGrpcServer(repo, notif, nil)

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

	ch := notif.Subscribe(1002)

	resp, err := s.SendMessage(context.Background(), &pb.SendMessageRequest{
		SenderId: 1001, ReceiverId: 1002, Content: "hi", MessageType: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(100), resp.MessageId)

	select {
	case ev := <-ch:
		assert.Equal(t, "message", ev.Type)
		assert.Equal(t, "hi", ev.Content)
		assert.Equal(t, int64(1001), ev.FromUID)
	default:
		t.Fatal("expected notification event")
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrpcServer_SendMessage_Error(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)
	s := NewGrpcServer(repo, nil, nil)

	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(1001), int64(1002)).
		WillReturnError(errDB)

	_, err := s.SendMessage(context.Background(), &pb.SendMessageRequest{
		SenderId: 1001, ReceiverId: 1002, Content: "hi", MessageType: 1,
	})
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrpcServer_SendMessage_WithCache(t *testing.T) {
	mr, cache := newTestUnreadCache(t)
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)
	s := NewGrpcServer(repo, nil, cache)

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

	mr.HSet("unread:1002", "private", "5")

	resp, err := s.SendMessage(context.Background(), &pb.SendMessageRequest{
		SenderId: 1001, ReceiverId: 1002, Content: "hi", MessageType: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(100), resp.MessageId)
	assert.False(t, mr.Exists("unread:1002"))
	require.NoError(t, mock.ExpectationsWereMet())
}