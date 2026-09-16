package social

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockFollowRepo(t *testing.T) (*FollowRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewFollowRepository(db), mock
}

func TestFollowRepository_Follow_Success(t *testing.T) {
	repo, mock := newMockFollowRepo(t)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO follows`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET following_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET follower_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.Follow(context.Background(), 1, 2))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowRepository_Follow_Duplicate(t *testing.T) {
	repo, mock := newMockFollowRepo(t)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO follows`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	require.NoError(t, repo.Follow(context.Background(), 1, 2))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowRepository_Follow_ExecError(t *testing.T) {
	repo, mock := newMockFollowRepo(t)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO follows`).WillReturnError(errors.New("constraint violated"))

	err := repo.Follow(context.Background(), 1, 2)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowRepository_Follow_BeginError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()
	repo := NewFollowRepository(db)

	mock.ExpectBegin().WillReturnError(errors.New("conn lost"))
	err = repo.Follow(context.Background(), 1, 2)
	assert.Error(t, err)
}

func TestFollowRepository_Unfollow_Success(t *testing.T) {
	repo, mock := newMockFollowRepo(t)
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM follows`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET following_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET follower_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.Unfollow(context.Background(), 1, 2))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowRepository_Unfollow_NotFollowing(t *testing.T) {
	repo, mock := newMockFollowRepo(t)
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM follows`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	require.NoError(t, repo.Unfollow(context.Background(), 1, 2))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowRepository_Unfollow_CommitError(t *testing.T) {
	repo, mock := newMockFollowRepo(t)
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM follows`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET following_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET follower_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	err := repo.Unfollow(context.Background(), 1, 2)
	assert.Error(t, err)
}

func TestFollowRepository_IsFollowing(t *testing.T) {
	repo, mock := newMockFollowRepo(t)
	mock.ExpectQuery(`SELECT EXISTS`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	ok, err := repo.IsFollowing(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.True(t, ok)

	mock.ExpectQuery(`SELECT EXISTS`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	ok, err = repo.IsFollowing(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestFollowRepository_ListFollowing(t *testing.T) {
	repo, mock := newMockFollowRepo(t)
	mock.ExpectQuery(`SELECT following_id FROM follows`).
		WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"following_id"}).AddRow(2).AddRow(3))

	ids, err := repo.ListFollowing(context.Background(), 1, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, []int64{2, 3}, ids)
}

func TestFollowRepository_ListFollowers(t *testing.T) {
	repo, mock := newMockFollowRepo(t)
	mock.ExpectQuery(`SELECT follower_id FROM follows`).
		WithArgs(int64(5), int32(10), int32(20)).
		WillReturnRows(sqlmock.NewRows([]string{"follower_id"}).AddRow(9))

	ids, err := repo.ListFollowers(context.Background(), 5, 3, 10)
	require.NoError(t, err)
	assert.Equal(t, []int64{9}, ids)
}

func TestFollowRepository_List_QueryError(t *testing.T) {
	repo, mock := newMockFollowRepo(t)
	mock.ExpectQuery(`SELECT following_id FROM follows`).WillReturnError(errors.New("db down"))
	_, err := repo.ListFollowing(context.Background(), 1, 1, 10)
	assert.Error(t, err)
}

func TestFollowRepository_Counts(t *testing.T) {
	repo, mock := newMockFollowRepo(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE follower_id`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(11))
	n, err := repo.FollowingCount(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(11), n)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))
	n, err = repo.FollowerCount(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(4), n)
}

func TestParseInt64(t *testing.T) {
	assert.Equal(t, int64(42), parseInt64("42"))
	assert.Equal(t, int64(0), parseInt64("abc"))
	assert.Equal(t, int64(-1), parseInt64("-1"))
}
