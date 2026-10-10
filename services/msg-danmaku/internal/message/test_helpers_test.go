package message

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// errDB 用于模拟数据访问层错误。
var errDB = errors.New("db error")

// newMessageDB 构造一个 sqlmock 驱动的 *sql.DB 与 mock，并注册关闭清理。
func newMessageDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// expectSendMessageDualWrite 预置 SendMessage 双写的完整 sqlmock 期望序列：
// 查发送方行（1001→1002，得 50）、查接收方行（1002→1001，得 60）、
// 两份消息 INSERT（发送方份 id 100、接收方份 id 101）、两行会话 UPDATE。
func expectSendMessageDualWrite(mock sqlmock.Sqlmock, now time.Time) {
	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(1001), int64(1002)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(50))
	mock.ExpectQuery(`SELECT id FROM conversations`).
		WithArgs(int64(1002), int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(60))
	cols := []string{
		"id", "sender_id", "receiver_id", "conversation_id", "content", "message_type", "is_read", "created_at",
	}
	mock.ExpectQuery(`INSERT INTO messages`).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(100, 1001, 1002, 50, "hi", 1, 0, now))
	mock.ExpectQuery(`INSERT INTO messages`).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(101, 1001, 1002, 60, "hi", 1, 0, now))
	mock.ExpectExec(`UPDATE conversations SET last_message_content`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE conversations SET last_message_content`).
		WillReturnResult(sqlmock.NewResult(0, 1))
}