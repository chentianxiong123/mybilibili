package message

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageRepository_DB(t *testing.T) {
	db, _ := newMessageDB(t)
	defer db.Close()
	repo := NewMessageRepository(db)
	assert.Same(t, db, repo.DB())
}

func TestMessageRepository_MarkAsRead(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)

	mock.ExpectExec(`UPDATE messages SET is_read = 1`).
		WithArgs(int64(50), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec(`UPDATE conversations SET unread_count = 0`).
		WithArgs(int64(50), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.MarkAsRead(context.Background(), 50, 1001)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_MarkAsRead_FirstError(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)

	mock.ExpectExec(`UPDATE messages SET is_read = 1`).
		WithArgs(int64(50), int64(1001)).
		WillReturnError(errDB)

	err := repo.MarkAsRead(context.Background(), 50, 1001)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_MarkAsRead_SecondError(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)

	mock.ExpectExec(`UPDATE messages SET is_read = 1`).
		WithArgs(int64(50), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec(`UPDATE conversations SET unread_count = 0`).
		WithArgs(int64(50), int64(1001)).
		WillReturnError(errDB)

	err := repo.MarkAsRead(context.Background(), 50, 1001)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_GetConversations_QueryError(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)

	mock.ExpectQuery(`SELECT c.id, c.user_id, c.target_user_id`).
		WithArgs(int64(1001)).
		WillReturnError(errDB)

	list, err := repo.GetConversations(context.Background(), 1001)
	assert.Error(t, err)
	assert.Nil(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_GetConversations_ScanError(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)

	mock.ExpectQuery(`SELECT c.id, c.user_id, c.target_user_id`).
		WithArgs(int64(1001)).
		// 只有 1 列，Scan 8 个字段必然出错
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	list, err := repo.GetConversations(context.Background(), 1001)
	assert.Error(t, err)
	assert.Nil(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_GetMessages_Empty(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)

	mock.ExpectQuery(`SELECT id, sender_id, receiver_id, conversation_id, content, message_type, is_read, created_at`).
		WithArgs(int64(50), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "sender_id", "receiver_id", "conversation_id", "content", "message_type", "is_read", "created_at",
		}))

	list, err := repo.GetMessages(context.Background(), 50, 1, 20)
	require.NoError(t, err)
	assert.Empty(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_GetMessages_QueryError(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)

	mock.ExpectQuery(`SELECT id, sender_id, receiver_id, conversation_id, content, message_type, is_read, created_at`).
		WithArgs(int64(50), int32(20), int32(0)).
		WillReturnError(errDB)

	list, err := repo.GetMessages(context.Background(), 50, 1, 20)
	assert.Error(t, err)
	assert.Nil(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_GetMessages_ScanError(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)

	mock.ExpectQuery(`SELECT id, sender_id, receiver_id, conversation_id, content, message_type, is_read, created_at`).
		WithArgs(int64(50), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	list, err := repo.GetMessages(context.Background(), 50, 1, 20)
	assert.Error(t, err)
	assert.Nil(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_SendMessage_ExistingConversation(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)

	// 双写：发送方行与接收方行都已存在，各查各的 id
	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(1001), int64(1002)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(50))
	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(1002), int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(60))
	now := time.Now()
	msgCols := []string{
		"id", "sender_id", "receiver_id", "conversation_id", "content", "message_type", "is_read", "created_at",
	}
	// 发送方份（conversation 50）
	mock.ExpectQuery(`INSERT INTO messages`).
		WithArgs(int64(1001), int64(1002), int64(50), "hi", int32(1)).
		WillReturnRows(sqlmock.NewRows(msgCols).AddRow(100, 1001, 1002, 50, "hi", 1, 0, now))
	// 接收方份（conversation 60）
	mock.ExpectQuery(`INSERT INTO messages`).
		WithArgs(int64(1001), int64(1002), int64(60), "hi", int32(1)).
		WillReturnRows(sqlmock.NewRows(msgCols).AddRow(101, 1001, 1002, 60, "hi", 1, 0, now))
	// 发送方行只更新最后一条
	mock.ExpectExec(`UPDATE conversations SET last_message_content = \$1, last_message_time = NOW\(\)`).
		WithArgs("hi", int64(50)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 接收方行更新最后一条且未读+1
	mock.ExpectExec(`UPDATE conversations SET last_message_content = \$1, last_message_time = NOW\(\), unread_count = unread_count \+ 1`).
		WithArgs("hi", int64(60)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	msg, err := repo.SendMessage(context.Background(), 1001, 1002, "hi", 1)
	require.NoError(t, err)
	// 返回发送方的那份（调用方是发送方）
	assert.Equal(t, int64(100), msg.ID)
	assert.Equal(t, int64(50), msg.ConversationID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_SendMessage_CreateConversations(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)

	// 发送方行不存在：建发送方行，upsert 接收方行
	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(1001), int64(1002)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`INSERT INTO conversations \(user_id, target_user_id\) VALUES`).
		WithArgs(int64(1001), int64(1002)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(50))
	mock.ExpectQuery(`ON CONFLICT \(user_id, target_user_id\)`).
		WithArgs(int64(1002), int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(60))
	now := time.Now()
	msgCols := []string{
		"id", "sender_id", "receiver_id", "conversation_id", "content", "message_type", "is_read", "created_at",
	}
	mock.ExpectQuery(`INSERT INTO messages`).
		WillReturnRows(sqlmock.NewRows(msgCols).AddRow(100, 1001, 1002, 50, "hi", 1, 0, now))
	mock.ExpectQuery(`INSERT INTO messages`).
		WillReturnRows(sqlmock.NewRows(msgCols).AddRow(101, 1001, 1002, 60, "hi", 1, 0, now))
	mock.ExpectExec(`UPDATE conversations SET last_message_content`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE conversations SET last_message_content`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	msg, err := repo.SendMessage(context.Background(), 1001, 1002, "hi", 1)
	require.NoError(t, err)
	assert.Equal(t, int64(50), msg.ConversationID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_SendMessage_ConversationSelectError(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)

	// SELECT 出真错（非 ErrNoRows）直接失败，不再盲目 INSERT
	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(1001), int64(1002)).
		WillReturnError(errDB)

	msg, err := repo.SendMessage(context.Background(), 1001, 1002, "hi", 1)
	assert.Error(t, err)
	assert.Nil(t, msg)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_SendMessage_MessageInsertError(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)

	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(1001), int64(1002)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(50))
	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(1002), int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(60))
	mock.ExpectQuery(`INSERT INTO messages`).
		WillReturnError(errDB)

	msg, err := repo.SendMessage(context.Background(), 1001, 1002, "hi", 1)
	assert.Error(t, err)
	assert.Nil(t, msg)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_SendMessage_ReverseInsertError(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)

	// SELECT ErrNoRows -> INSERT 成功 -> 反向 upsert 报错
	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(1001), int64(1002)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO conversations \(user_id, target_user_id\)`).
		WithArgs(int64(1001), int64(1002)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(50))
	mock.ExpectQuery(`ON CONFLICT \(user_id, target_user_id\)`).
		WithArgs(int64(1002), int64(1001)).
		WillReturnError(errDB)

	msg, err := repo.SendMessage(context.Background(), 1001, 1002, "hi", 1)
	assert.Error(t, err)
	assert.Nil(t, msg)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_GetUnreadCount_Error(t *testing.T) {
	db, mock := newMessageDB(t)
	repo := NewMessageRepository(db)

	mock.ExpectQuery(`SELECT COALESCE\(COUNT\(\*\),0\) FROM messages`).
		WithArgs(int64(1001)).
		WillReturnError(errDB)

	count, err := repo.GetUnreadCount(context.Background(), 1001)
	assert.Error(t, err)
	assert.Equal(t, int32(0), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

