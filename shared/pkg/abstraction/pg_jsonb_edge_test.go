package abstraction

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 本文件覆盖 pg_jsonb.go 的错误/边界分支:
// - Insert / Update 的 json.Marshal 失败
// - Query 的 rows.Scan 失败(continue)
// - Query 返回的 doc 缺少 id 时用 rowID 回填

func TestPGJSONB_Insert_MarshalError(t *testing.T) {
	p, _ := newPGJSONBMock(t)
	// channel 无法 json.Marshal → 在触库前就返回错误
	_, err := p.Insert(context.Background(), "c", make(chan int))
	require.Error(t, err)
}

func TestPGJSONB_Update_MarshalError(t *testing.T) {
	p, _ := newPGJSONBMock(t)
	err := p.Update(context.Background(), "c", "id1", make(chan int))
	require.Error(t, err)
}

func TestPGJSONB_Query_ScanErrorContinues(t *testing.T) {
	p, mock := newPGJSONBMock(t)
	// 行只有 1 列, 而代码扫描 2 个目标 → 每次 rows.Scan 都报错 → continue 跳过
	rows := sqlmock.NewRows([]string{"doc_json"}).
		AddRow(`{"id":"1","name":"a"}`)
	mock.ExpectQuery(`SELECT id, doc_json FROM jsonb_documents`).
		WithArgs("users", 20, 0).
		WillReturnRows(rows)

	var out []map[string]any
	err := p.Query(context.Background(), "users", QueryFilter{}, &out)
	require.NoError(t, err)
	assert.Len(t, out, 0)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPGJSONB_Query_FillsMissingRowID(t *testing.T) {
	p, mock := newPGJSONBMock(t)
	// doc_json 里没有 id 字段 → 用行 id 回填到 map
	rows := sqlmock.NewRows([]string{"id", "doc_json"}).
		AddRow("row-42", `{"name":"alice"}`)
	mock.ExpectQuery(`SELECT id, doc_json FROM jsonb_documents`).
		WithArgs("users", 20, 0).
		WillReturnRows(rows)

	var out []map[string]any
	err := p.Query(context.Background(), "users", QueryFilter{}, &out)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "row-42", out[0]["id"])
	assert.Equal(t, "alice", out[0]["name"])

	require.NoError(t, mock.ExpectationsWereMet())
}