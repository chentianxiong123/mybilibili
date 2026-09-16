package social

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockFollowSvc(t *testing.T) (*FollowService, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewFollowService(NewFollowRepository(db)), mock
}

func TestFollowService_FollowSelf(t *testing.T) {
	svc, _ := newMockFollowSvc(t)
	err := svc.Follow(context.Background(), 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot follow yourself")
}

func TestFollowService_Follow_Success(t *testing.T) {
	svc, mock := newMockFollowSvc(t)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO follows`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET following_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET follower_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, svc.Follow(context.Background(), 1, 2))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowService_Follow_DuplicateDoesNotIncrCounts(t *testing.T) {
	svc, mock := newMockFollowSvc(t)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO follows`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	require.NoError(t, svc.Follow(context.Background(), 1, 2))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowService_Unfollow(t *testing.T) {
	svc, mock := newMockFollowSvc(t)
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM follows`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET following_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET follower_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, svc.Unfollow(context.Background(), 1, 2))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFollowService_IsFollowing(t *testing.T) {
	svc, mock := newMockFollowSvc(t)
	mock.ExpectQuery(`SELECT EXISTS`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	ok, err := svc.IsFollowing(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestFollowService_FollowingCount(t *testing.T) {
	svc, mock := newMockFollowSvc(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE follower_id`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	n, err := svc.FollowingCount(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(7), n)
}

func TestFollowService_FollowerCount(t *testing.T) {
	svc, mock := newMockFollowSvc(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	n, err := svc.FollowerCount(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(3), n)
}
