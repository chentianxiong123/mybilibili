package social

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/pkg/pb"
)

func newInteractionSvc(t *testing.T) (*InteractionService, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	svc := NewInteractionService(NewInteractionRepository(db))
	svc.SetDB(db)
	return svc, mock
}

func TestInteractionService_LikeManuscript_FirstTime(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).
		WithArgs(int64(1), "MANUSCRIPT", int64(9), "LIKE").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET like_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE target_type`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	resp, err := svc.LikeManuscript(ctx, &pb.LikeManuscriptRequest{UserId: 1, ManuscriptId: 9})
	require.NoError(t, err)
	assert.True(t, resp.Liked)
	assert.Equal(t, int32(3), resp.LikeCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

type fakeProfileRecorder struct {
	likes    int
	collects int
	watches  int
	lastCat  int64
}

func (f *fakeProfileRecorder) RecordWatch(ctx context.Context, userID, categoryID int64, tags []string, duration int64) error {
	f.watches++
	f.lastCat = categoryID
	return nil
}
func (f *fakeProfileRecorder) RecordLike(ctx context.Context, userID, categoryID int64, tags []string) error {
	f.likes++
	f.lastCat = categoryID
	return nil
}
func (f *fakeProfileRecorder) RecordCollect(ctx context.Context, userID, categoryID int64, tags []string) error {
	f.collects++
	f.lastCat = categoryID
	return nil
}

type fakeNotifier struct {
	msgs []string
}

func (f *fakeNotifier) SendMessage(ctx context.Context, senderID, receiverID int64, content string, msgType int32) {
	f.msgs = append(f.msgs, content)
}

func TestInteractionService_LikeManuscript_WithRecorderAndNotifier(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	ctx := context.Background()

	rec := &fakeProfileRecorder{}
	not := &fakeNotifier{}
	svc.SetProfileRecorder(rec)
	svc.SetNotifier(not)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET like_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))
	// sendLikeNotification 路径：查稿件作者（先于 profileRecorder）
	mock.ExpectQuery(`SELECT user_id FROM manuscripts`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(100))
	// profileRecorder 路径：后查分类
	mock.ExpectQuery(`SELECT category_id FROM manuscripts`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"category_id"}).AddRow(2))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE target_type`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	resp, err := svc.LikeManuscript(ctx, &pb.LikeManuscriptRequest{UserId: 1, ManuscriptId: 9})
	require.NoError(t, err)
	assert.True(t, resp.Liked)
	assert.Equal(t, int32(3), resp.LikeCount)
	assert.Equal(t, 1, rec.likes)
	assert.Equal(t, int64(2), rec.lastCat)
	require.Len(t, not.msgs, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionService_LikeManuscript_NoOwnerSkipsNotify(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	ctx := context.Background()

	not := &fakeNotifier{}
	svc.SetNotifier(not)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET like_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))
	// 作者不存在 → 不通知
	mock.ExpectQuery(`SELECT user_id FROM manuscripts`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(0))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE target_type`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	_, err := svc.LikeManuscript(ctx, &pb.LikeManuscriptRequest{UserId: 1, ManuscriptId: 9})
	require.NoError(t, err)
	assert.Empty(t, not.msgs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionService_LikeManuscript_AlreadyLiked(t *testing.T) {	svc, mock := newInteractionSvc(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE target_type`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	resp, err := svc.LikeManuscript(ctx, &pb.LikeManuscriptRequest{UserId: 1, ManuscriptId: 9})
	require.NoError(t, err)
	assert.True(t, resp.Liked)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionService_UnlikeManuscript(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET like_count = GREATEST`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE target_type`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	resp, err := svc.UnlikeManuscript(ctx, &pb.UnlikeManuscriptRequest{UserId: 1, ManuscriptId: 9})
	require.NoError(t, err)
	assert.False(t, resp.Liked)
	assert.Equal(t, int32(0), resp.LikeCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionService_CoinManuscript_Success(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT coin_count FROM users`).WillReturnRows(sqlmock.NewRows([]string{"coin_count"}).AddRow(5))
	mock.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET coin_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE users SET coin_count = coin_count - 1`).WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := svc.CoinManuscript(ctx, &pb.CoinManuscriptRequest{UserId: 1, ManuscriptId: 9})
	require.NoError(t, err)
	assert.True(t, resp.Success)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionService_CoinManuscript_Insufficient(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	mock.ExpectQuery(`SELECT coin_count FROM users`).WillReturnRows(sqlmock.NewRows([]string{"coin_count"}).AddRow(0))

	_, err := svc.CoinManuscript(context.Background(), &pb.CoinManuscriptRequest{UserId: 1, ManuscriptId: 9})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient coins")
}

func TestInteractionService_CollectManuscript(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET collect_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := svc.CollectManuscript(ctx, &pb.CollectManuscriptRequest{UserId: 1, ManuscriptId: 9})
	require.NoError(t, err)
	assert.True(t, resp.Collected)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionService_UncollectManuscript(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	mock.ExpectExec(`DELETE FROM user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET collect_count = GREATEST`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := svc.UncollectManuscript(context.Background(), &pb.UncollectManuscriptRequest{UserId: 1, ManuscriptId: 9})
	require.NoError(t, err)
	assert.False(t, resp.Collected)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionService_ShareManuscript(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	mock.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET share_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := svc.ShareManuscript(context.Background(), &pb.ShareManuscriptRequest{UserId: 1, ManuscriptId: 9})
	require.NoError(t, err)
	assert.True(t, resp.Success)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionService_GetInteractionStatus(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	for i := 0; i < 3; i++ {
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	}
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE target_type`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	resp, err := svc.GetInteractionStatus(context.Background(), &pb.GetInteractionStatusRequest{UserId: 1, ManuscriptId: 9})
	require.NoError(t, err)
	assert.False(t, resp.Liked)
	assert.False(t, resp.Collected)
	assert.False(t, resp.Shared)
	assert.Equal(t, int32(2), resp.CoinCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionService_FollowUser_WithoutFollowService(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	resp, err := svc.FollowUser(context.Background(), &pb.FollowUserRequest{UserId: 1, TargetUserId: 2})
	require.NoError(t, err)
	assert.True(t, resp.Following)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionService_GetFollowCount_WithoutFollowService(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	resp, err := svc.GetFollowCount(context.Background(), &pb.GetFollowCountRequest{UserId: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.FollowingCount)
	assert.Equal(t, int32(0), resp.FollowerCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionService_GetLikedManuscripts(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	mock.ExpectQuery(`SELECT target_id FROM user_interactions`).
		WithArgs(int64(1), "MANUSCRIPT", "LIKE").
		WillReturnRows(sqlmock.NewRows([]string{"target_id"}).AddRow(5).AddRow(6))

	resp, err := svc.GetLikedManuscripts(context.Background(), &pb.GetLikedManuscriptsRequest{UserId: 1})
	require.NoError(t, err)
	assert.Equal(t, []int64{5, 6}, resp.ManuscriptIds)
}

func TestInteractionService_GetCollectedManuscripts(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	mock.ExpectQuery(`SELECT target_id FROM user_interactions`).
		WithArgs(int64(1), "MANUSCRIPT", "COLLECT").
		WillReturnRows(sqlmock.NewRows([]string{"target_id"}).AddRow(8))

	resp, err := svc.GetCollectedManuscripts(context.Background(), &pb.GetCollectedManuscriptsRequest{UserId: 1})
	require.NoError(t, err)
	assert.Equal(t, []int64{8}, resp.ManuscriptIds)
}

func TestInteractionService_WatchHistory(t *testing.T) {
	svc, mock := newInteractionSvc(t)
	ctx := context.Background()

	mock.ExpectExec(`INSERT INTO watch_history`).WithArgs(int64(1), int64(9), int32(30)).WillReturnResult(sqlmock.NewResult(0, 1))
	_, err := svc.AddWatchHistory(ctx, &pb.AddWatchHistoryRequest{UserId: 1, ManuscriptId: 9, ProgressSeconds: 30})
	require.NoError(t, err)

	mock.ExpectQuery(`SELECT m.id, m.title`).WillReturnRows(sqlmock.NewRows([]string{"id", "title", "cover_url", "progress_seconds", "duration_seconds", "watched_at"}))
	resp, err := svc.GetWatchHistory(ctx, &pb.GetWatchHistoryRequest{UserId: 1, Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Empty(t, resp.Items)

	mock.ExpectExec(`DELETE FROM watch_history`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	_, err = svc.ClearWatchHistory(ctx, &pb.ClearWatchHistoryRequest{UserId: 1})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
