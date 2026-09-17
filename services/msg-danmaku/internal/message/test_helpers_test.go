package message

import (
	"database/sql"
	"errors"
	"testing"

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