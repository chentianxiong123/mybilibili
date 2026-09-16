package moderation

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockModerationRepo(t *testing.T) (*Repository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewRepository(db), mock
}

func TestRepoListWords(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	createdAt := time.Now().Truncate(time.Second)
	mock.ExpectQuery(`SELECT id, word, match_type FROM prohibited_words`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "match_type", "category", "is_enabled", "created_at", "updated_at"}).
			AddRow(int64(1), "badword", "exact", "spam", int32(1), createdAt, createdAt))

	list, err := repo.ListWords(context.Background(), 1, 20)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "badword", list[0].Word)
	assert.Equal(t, int32(1), list[0].IsEnabled)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepoCreateWord(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	mock.ExpectExec(`INSERT INTO prohibited_words`).
		WithArgs("badword", "exact", "spam").
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, repo.CreateWord(context.Background(), "badword", "exact", "spam"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepoDeleteWord(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	mock.ExpectExec(`DELETE FROM prohibited_words WHERE id = \$1`).
		WithArgs(int64(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.DeleteWord(context.Background(), 3))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepoGetWord(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	createdAt := time.Now().Truncate(time.Second)
	mock.ExpectQuery(`SELECT id, word, match_type FROM prohibited_words WHERE id`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "match_type", "category", "is_enabled", "created_at", "updated_at"}).
			AddRow(int64(1), "word", "fuzzy", "ads", int32(0), createdAt, createdAt))

	p, err := repo.GetWord(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "word", p.Word)
	assert.Equal(t, "fuzzy", p.MatchType)
}
