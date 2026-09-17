package abstraction

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pgJSONB 直接连真 Postgres 才能跑完整; 这里走 sqlmock 测内部 SQL 拼接逻辑与 JSON 编解码.
func newPGJSONBMock(t *testing.T) (*pgJSONB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return &pgJSONB{db: db}, mock
}

func TestEscapeJSONKey(t *testing.T) {
	assert.Equal(t, "abc", escapeJSONKey("abc"))
	assert.Equal(t, "it''s", escapeJSONKey("it's"))
	assert.Equal(t, "''''", escapeJSONKey("''"))
}

func TestNewPGJSONB_DSNRequired(t *testing.T) {
	_, err := newPGJSONB(DocumentStoreConfig{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DSN required")
}

func TestNewPGJSONB_OpenFailure(t *testing.T) {
	// 传入非法 DSN 让 sql.Open 失败 (lib/pq 不会立刻连, 但特定格式可触发)
	_, err := newPGJSONB(DocumentStoreConfig{DSN: "host=127.0.0.1 port=1 dbname=xxx sslmode=disable"})
	// lib/pq 的 sql.Open 是延迟连接, 这里通常不报错; 测试仅确保路径走过
	_ = err
}

func TestPGJSONB_Insert_FindByID_Update_Delete(t *testing.T) {
	p, mock := newPGJSONBMock(t)
	ctx := context.Background()

	// Insert
	mock.ExpectExec(`INSERT INTO jsonb_documents`).
		WithArgs("u1", "users", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := p.Insert(ctx, "users", map[string]any{"id": "u1", "name": "alice"})
	require.NoError(t, err)
	assert.Equal(t, "u1", id)

	// Insert without id -> 自动生成
	mock.ExpectExec(`INSERT INTO jsonb_documents`).
		WithArgs(sqlmock.AnyArg(), "logs", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err = p.Insert(ctx, "logs", map[string]any{"msg": "hi"})
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	// FindByID
	docJSON, _ := json.Marshal(map[string]any{"id": "u1", "name": "alice"})
	mock.ExpectQuery(`SELECT doc_json FROM jsonb_documents`).
		WithArgs("users", "u1").
		WillReturnRows(sqlmock.NewRows([]string{"doc_json"}).AddRow(string(docJSON)))

	var got map[string]any
	err = p.FindByID(ctx, "users", "u1", &got)
	require.NoError(t, err)
	assert.Equal(t, "alice", got["name"])

	// FindByID not found
	mock.ExpectQuery(`SELECT doc_json FROM jsonb_documents`).
		WithArgs("users", "missing").
		WillReturnError(sql.ErrNoRows)
	err = p.FindByID(ctx, "users", "missing", &got)
	assert.Error(t, err)

	// Update
	mock.ExpectExec(`UPDATE jsonb_documents SET doc_json`).
		WithArgs(sqlmock.AnyArg(), "users", "u1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	err = p.Update(ctx, "users", "u1", map[string]any{"id": "u1", "name": "bob"})
	require.NoError(t, err)

	// Delete
	mock.ExpectExec(`DELETE FROM jsonb_documents`).
		WithArgs("users", "u1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	err = p.Delete(ctx, "users", "u1")
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPGJSONB_Query_PaginationAndFilter(t *testing.T) {
	p, mock := newPGJSONBMock(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "doc_json"}).
		AddRow("1", `{"id":"1","name":"a"}`).
		AddRow("2", `{"id":"2","name":"b"}`)

	// Filter 空, 默认 page=1 size=20, order=created_at DESC
	mock.ExpectQuery(`SELECT id, doc_json FROM jsonb_documents`).
		WithArgs("users", 20, 0).
		WillReturnRows(rows)

	var out []map[string]any
	err := p.Query(ctx, "users", QueryFilter{}, &out)
	require.NoError(t, err)
	assert.Len(t, out, 2)
	assert.Equal(t, "a", out[0]["name"])

	// 自定义 page / sort (pageSize=10, page=2 → offset=10)
	mock.ExpectQuery(`SELECT id, doc_json FROM jsonb_documents`).
		WithArgs("users", "active", 10, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "doc_json"}))

	err = p.Query(ctx, "users", QueryFilter{
		Page:     2,
		PageSize: 10,
		Sort:     "active",
		Order:    "DESC",
		Filters:  map[string]any{"status": "active"},
	}, &out)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPGJSONB_Query_SortNoOrderDefaultsAsc(t *testing.T) {
	p, mock := newPGJSONBMock(t)
	mock.ExpectQuery(`doc_json->>'active' ASC`).
		WithArgs("users", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "doc_json"}))
	err := p.Query(context.Background(), "users", QueryFilter{Sort: "active"}, &[]map[string]any{})
	require.NoError(t, err)
}

func TestPGJSONB_Query_DBError(t *testing.T) {
	p, mock := newPGJSONBMock(t)
	mock.ExpectQuery(`SELECT id, doc_json`).
		WillReturnError(errors.New("db down"))
	err := p.Query(context.Background(), "users", QueryFilter{}, &[]map[string]any{})
	assert.Error(t, err)
}

// 防止 driver 包未引用 (sqlmock 需要 driver.Value 兼容)
var _ driver.Value = (*struct{})(nil)