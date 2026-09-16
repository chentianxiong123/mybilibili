package social

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockInteractionRepo(t *testing.T) (*InteractionRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewInteractionRepository(db), mock
}

func TestInteractionRepository_HasInteraction(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).
		WithArgs(int64(1), "manuscript", int64(9), "like").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	ok, err := repo.HasInteraction(context.Background(), 1, "manuscript", "like", 9)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestInteractionRepository_HasInteraction_None(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	ok, err := repo.HasInteraction(context.Background(), 1, "manuscript", "like", 9)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestInteractionRepository_HasInteraction_Error(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).WillReturnError(errors.New("db error"))
	_, err := repo.HasInteraction(context.Background(), 1, "manuscript", "like", 9)
	assert.Error(t, err)
}

func TestInteractionRepository_AddAndRemove(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectExec(`INSERT INTO user_interactions`).WithArgs(int64(1), "manuscript", int64(9), "like").WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.AddInteraction(context.Background(), 1, "manuscript", "like", 9))

	mock.ExpectExec(`DELETE FROM user_interactions`).WithArgs(int64(1), "manuscript", int64(9), "like").WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.RemoveInteraction(context.Background(), 1, "manuscript", "like", 9))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionRepository_Add_Error(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectExec(`INSERT INTO user_interactions`).WillReturnError(errors.New("insert failed"))
	assert.Error(t, repo.AddInteraction(context.Background(), 1, "manuscript", "like", 9))
}

func TestInteractionRepository_CountInteraction(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE target_type`).
		WithArgs("manuscript", int64(9), "like").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	n, err := repo.CountInteraction(context.Background(), "manuscript", "like", 9)
	require.NoError(t, err)
	assert.Equal(t, int32(5), n)
}

func TestInteractionRepository_GetInteractionIDs(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectQuery(`SELECT target_id FROM user_interactions`).
		WithArgs(int64(1), "manuscript", "like").
		WillReturnRows(sqlmock.NewRows([]string{"target_id"}).AddRow(3).AddRow(7))

	ids, err := repo.GetInteractionIDs(context.Background(), 1, "manuscript", "like")
	require.NoError(t, err)
	assert.Equal(t, []int64{3, 7}, ids)
}

func TestInteractionRepository_GetInteractionIDs_QueryError(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectQuery(`SELECT target_id FROM user_interactions`).WillReturnError(errors.New("boom"))
	_, err := repo.GetInteractionIDs(context.Background(), 1, "manuscript", "like")
	assert.Error(t, err)
}

func TestInteractionRepository_UpsertWatchHistory(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectExec(`INSERT INTO watch_history`).
		WithArgs(int64(1), int64(9), int32(30)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpsertWatchHistory(context.Background(), 1, 9, 30))
}

func TestInteractionRepository_GetWatchHistory(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectQuery(`SELECT m.id, m.title, m.cover_url`).
		WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "cover_url", "progress_seconds", "duration_seconds", "watched_at"}).
			AddRow(9, "标题", "http://cover", 30, 100, time.Now()))

	items, err := repo.GetWatchHistory(context.Background(), 1, 1, 20)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, int64(9), items[0].ManuscriptId)
	assert.Equal(t, "标题", items[0].Title)
	assert.Equal(t, int32(30), items[0].ProgressSeconds)
}

func TestInteractionRepository_GetWatchHistory_Empty(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectQuery(`SELECT m.id, m.title, m.cover_url`).WillReturnRows(sqlmock.NewRows([]string{"id", "title", "cover_url", "progress_seconds", "duration_seconds", "watched_at"}))
	items, err := repo.GetWatchHistory(context.Background(), 1, 1, 20)
	require.NoError(t, err)
	assert.Len(t, items, 0)
}

func TestInteractionRepository_ClearWatchHistory(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectExec(`DELETE FROM watch_history`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.ClearWatchHistory(context.Background(), 1))
}

func TestInteractionRepository_GetManuscriptTitle(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectQuery(`SELECT title FROM manuscripts`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("我的视频"))
	title, err := repo.GetManuscriptTitle(context.Background(), 9)
	require.NoError(t, err)
	assert.Equal(t, "我的视频", title)
}

func TestInteractionRepository_IncrementDecrementCount(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectExec(`UPDATE manuscripts SET like_count = like_count \+ 1`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.IncrementManuscriptCount(context.Background(), "like_count", 9))

	mock.ExpectExec(`UPDATE manuscripts SET coin_count = GREATEST\(coin_count - 1, 0\)`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.DecrementManuscriptCount(context.Background(), "coin_count", 9))
}

func TestInteractionRepository_GetUserCoinCountAndDeduct(t *testing.T) {
	repo, mock := newMockInteractionRepo(t)
	mock.ExpectQuery(`SELECT coin_count FROM users`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"coin_count"}).AddRow(3))
	n, err := repo.GetUserCoinCount(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(3), n)

	mock.ExpectExec(`UPDATE users SET coin_count = coin_count - 1`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.DeductCoin(context.Background(), 1))
}
