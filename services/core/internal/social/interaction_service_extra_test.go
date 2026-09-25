package social

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/core/internal/events"
	"mybilibili/pkg/abstraction"
	pb "mybilibili/pkg/pb"
)

type fakeQueue struct{}

func (fakeQueue) Publish(ctx context.Context, topic string, msg abstraction.Message) error { return nil }
func (fakeQueue) Subscribe(ctx context.Context, topic, group string) (<-chan abstraction.Message, error) {
	return nil, nil
}
func (fakeQueue) Ack(ctx context.Context, topic string, msg abstraction.Message) error    { return nil }
func (fakeQueue) Nack(ctx context.Context, topic string, msg abstraction.Message) error   { return nil }
func (fakeQueue) Enqueue(ctx context.Context, queue string, msg abstraction.Message, delay time.Duration) error {
	return nil
}
func (fakeQueue) Close() error { return nil }

func TestInteractionService_SettersAndRepo(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	svc := NewInteractionService(NewInteractionRepository(db))
	fs := NewFollowService(NewFollowRepository(db))
	svc.SetFollowService(fs)
	svc.SetEventPublisher(events.NewEventPublisher(fakeQueue{}))
	svc.SetDB(db)

	assert.NotNil(t, svc.Repo())
	assert.NotNil(t, svc.followSvc)
	assert.NotNil(t, svc.publisher)

	_ = mock
}

func TestInteractionService_FollowUser_WithFollowService(t *testing.T) {
	_, _ = newInteractionSvc(t)
	fs := NewFollowService(NewFollowRepository(nil))
	db, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()
	fs.repo = NewFollowRepository(db)

	svc := NewInteractionService(NewInteractionRepository(nil))
	svc.SetFollowService(fs)

	m.ExpectBegin()
	m.ExpectExec(`INSERT INTO follows`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec(`UPDATE users SET following_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec(`UPDATE users SET follower_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()

	resp, err := svc.FollowUser(context.Background(), &pb.FollowUserRequest{UserId: 1, TargetUserId: 2})
	require.NoError(t, err)
	assert.True(t, resp.Following)
	require.NoError(t, m.ExpectationsWereMet())
}

func TestInteractionService_UnfollowUser_WithFollowService(t *testing.T) {
	fs := NewFollowService(NewFollowRepository(nil))
	db, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()
	fs.repo = NewFollowRepository(db)

	svc := NewInteractionService(NewInteractionRepository(nil))
	svc.SetFollowService(fs)

	m.ExpectBegin()
	m.ExpectExec(`DELETE FROM follows`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec(`UPDATE users SET following_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec(`UPDATE users SET follower_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()

	resp, err := svc.UnfollowUser(context.Background(), &pb.UnfollowUserRequest{UserId: 1, TargetUserId: 2})
	require.NoError(t, err)
	assert.False(t, resp.Following)
	require.NoError(t, m.ExpectationsWereMet())
}

func TestInteractionService_CheckFollow_WithFollowService(t *testing.T) {
	fs := NewFollowService(NewFollowRepository(nil))
	db, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()
	fs.repo = NewFollowRepository(db)

	svc := NewInteractionService(NewInteractionRepository(nil))
	svc.SetFollowService(fs)

	m.ExpectQuery(`SELECT EXISTS`).WithArgs(int64(1), int64(2)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	resp, err := svc.CheckFollow(context.Background(), &pb.CheckFollowRequest{UserId: 1, TargetUserId: 2})
	require.NoError(t, err)
	assert.True(t, resp.Following)
	require.NoError(t, m.ExpectationsWereMet())
}

func TestInteractionService_GetFollowCount_WithFollowService(t *testing.T) {
	fs := NewFollowService(NewFollowRepository(nil))
	db, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()
	fs.repo = NewFollowRepository(db)

	svc := NewInteractionService(NewInteractionRepository(nil))
	svc.SetFollowService(fs)

	m.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE follower_id`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	m.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(9))

	resp, err := svc.GetFollowCount(context.Background(), &pb.GetFollowCountRequest{UserId: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(5), resp.FollowingCount)
	assert.Equal(t, int32(9), resp.FollowerCount)
	require.NoError(t, m.ExpectationsWereMet())
}

func TestInteractionService_Like_WithPublisher(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	svc.SetEventPublisher(events.NewEventPublisher(fakeQueue{}))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).WithArgs(int64(1), "MANUSCRIPT", int64(9), "LIKE").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO user_interactions`).WithArgs(int64(1), "MANUSCRIPT", int64(9), "LIKE").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET like_count`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WithArgs(int64(9), int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT like_count FROM manuscripts`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"like_count"}).AddRow(4))

	resp, err := svc.LikeManuscript(context.Background(), &pb.LikeManuscriptRequest{UserId: 1, ManuscriptId: 9})
	require.NoError(t, err)
	assert.True(t, resp.Liked)
	require.NoError(t, mock.ExpectationsWereMet())
}
