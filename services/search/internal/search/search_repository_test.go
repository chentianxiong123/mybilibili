package search

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockRepo(t *testing.T) (*Repository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewRepository(db), mock
}

func TestSearchManuscripts_EmptyKeyword(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"id", "title", "description", "cover_url", "user_id", "category_id",
		"view_count", "like_count", "comment_count", "danmaku_count",
		"duration", "status", "upload_time",
		"id", "username", "nickname", "avatar", "level", "is_vertical"}
	rows := sqlmock.NewRows(columns).
		AddRow(1, "Test Title", "desc", "cover.jpg", 10, 2, 100, 10, 3, 5, "5:30", 3, "2025-01-01", 10, "alice", "Alice", "avatar.jpg", 5, 0)

	mock.ExpectQuery(`SELECT.*FROM manuscripts m`).WillReturnRows(rows)

	result, err := repo.SearchManuscripts(ctx, "", 0, 1, 10)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, int64(1), result[0]["id"])
	assert.Equal(t, "Test Title", result[0]["title"])
	assert.Equal(t, int64(100), result[0]["view_count"])

	u := result[0]["uploader"].(map[string]interface{})
	assert.Equal(t, "Alice", u["nickname"])
	assert.Equal(t, int64(5), u["level"])
}

func TestSearchManuscripts_WithKeyword(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"id", "title", "description", "cover_url", "user_id", "category_id",
		"view_count", "like_count", "comment_count", "danmaku_count",
		"duration", "status", "upload_time",
		"id", "username", "nickname", "avatar", "level", "is_vertical"}
	rows := sqlmock.NewRows(columns).
		AddRow(2, "Go Programming", "learn go", "go.jpg", 20, 1, 500, 50, 10, 20, "10:00", 3, "2025-06-01", 20, "bob", "Bob", "bob.jpg", 6, 1)

	mock.ExpectQuery(`SELECT.*plainto_tsquery`).WillReturnRows(rows)

	result, err := repo.SearchManuscripts(ctx, "test", 0, 1, 10)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, int64(2), result[0]["id"])
	assert.Equal(t, int64(1), result[0]["is_vertical"])

	u := result[0]["uploader"].(map[string]interface{})
	assert.Equal(t, "bob", u["username"])
}

func TestSearchManuscripts_EmptyResult(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"id", "title", "description", "cover_url", "user_id", "category_id",
		"view_count", "like_count", "comment_count", "danmaku_count",
		"duration", "status", "upload_time",
		"id", "username", "nickname", "avatar", "level", "is_vertical"}
	rows := sqlmock.NewRows(columns)

	mock.ExpectQuery(`SELECT.*FROM manuscripts m`).WillReturnRows(rows)

	result, err := repo.SearchManuscripts(ctx, "", 0, 1, 10)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestSearchUsers_EmptyKeyword(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"id", "username", "nickname", "avatar", "signature", "level", "follower_count", "manuscript_count"}
	rows := sqlmock.NewRows(columns).
		AddRow(1, "alice", "Alice", "avatar.jpg", "hello", 5, 100, 10)

	mock.ExpectQuery(`SELECT.*FROM users u`).WillReturnRows(rows)

	result, err := repo.SearchUsers(ctx, "", 1, 10)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, int64(1), result[0]["mid"])
	assert.Equal(t, "alice", result[0]["username"])
	assert.Equal(t, int64(100), result[0]["fans"])
	assert.Equal(t, int64(10), result[0]["videos"])
}

func TestSearchUsers_WithKeyword(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"id", "username", "nickname", "avatar", "signature", "level", "follower_count", "manuscript_count"}
	rows := sqlmock.NewRows(columns).
		AddRow(1, "alice", "Alice", "avatar.jpg", "hello", 5, 100, 10).
		AddRow(3, "alice_wonder", "AliceWonderland", "a.jpg", "wonder", 3, 50, 5)

	mock.ExpectQuery(`SELECT.*ILIKE`).WillReturnRows(rows)

	result, err := repo.SearchUsers(ctx, "alice", 1, 10)
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "alice", result[0]["username"])
	assert.Equal(t, "alice_wonder", result[1]["username"])
}

func TestSearchUsers_EmptyResult(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"id", "username", "nickname", "avatar", "signature", "level", "follower_count", "manuscript_count"}
	rows := sqlmock.NewRows(columns)

	mock.ExpectQuery(`SELECT.*FROM users u`).WillReturnRows(rows)

	result, err := repo.SearchUsers(ctx, "zzz", 1, 10)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetRecommendConfig_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"config_json"}).
		AddRow(`{"hot":true,"banners":[]}`)
	mock.ExpectQuery(`SELECT config_json FROM recommend_configs`).WillReturnRows(rows)

	config, err := repo.GetRecommendConfig(ctx)
	require.NoError(t, err)
	assert.Equal(t, `{"hot":true,"banners":[]}`, config)
}

func TestGetRecommendConfig_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT config_json FROM recommend_configs`).WillReturnError(sql.ErrNoRows)

	config, err := repo.GetRecommendConfig(ctx)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.Empty(t, config)
}

func TestUpdateRecommendConfig_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	mock.ExpectExec(`INSERT INTO recommend_configs`).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateRecommendConfig(ctx, `{"hot":false}`, "admin")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCountIndexed_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(42)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE search_vector IS NOT NULL`).WillReturnRows(rows)

	svc := NewService(NewRepository(db), nil)
	count, err := svc.CountIndexed(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(42), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCountIndexed_Zero(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE search_vector IS NOT NULL`).WillReturnRows(rows)

	svc := NewService(NewRepository(db), nil)
	count, err := svc.CountIndexed(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
