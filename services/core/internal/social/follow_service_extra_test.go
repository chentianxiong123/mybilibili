package social

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFollowService_ListFollowing(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()
	svc := NewFollowService(NewFollowRepository(db))

	mock.ExpectQuery(`SELECT following_id FROM follows`).
		WithArgs(int64(100), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"following_id"}).AddRow(2).AddRow(3))
	ids, err := svc.ListFollowing(context.Background(), 100, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, []int64{2, 3}, ids)
}

func TestFollowService_ListFollowers(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()
	svc := NewFollowService(NewFollowRepository(db))

	mock.ExpectQuery(`SELECT follower_id FROM follows`).
		WithArgs(int64(100), int32(10), int32(10)).
		WillReturnRows(sqlmock.NewRows([]string{"follower_id"}).AddRow(9))
	ids, err := svc.ListFollowers(context.Background(), 100, 2, 10)
	require.NoError(t, err)
	assert.Equal(t, []int64{9}, ids)
}
