package comment

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mybilibili/pkg/abstraction"
	pb "mybilibili/pkg/pb"
)

func newCommentSvc(t *testing.T) (*CommentService, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	svc := NewCommentService(NewCommentRepository(db))
	svc.SetDB(db)
	return svc, mock
}

func TestAddComment_EmptyContent(t *testing.T) {
	svc, _ := newCommentSvc(t)
	_, err := svc.AddComment(context.Background(), &pb.AddCommentRequest{UserId: 1, ManuscriptId: 9})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "content required")
}

func TestAddComment_Success(t *testing.T) {
	svc, mock := newCommentSvc(t)
	ctx := context.Background()
	now := time.Now()

	// 违禁词检查 → 无
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM prohibited_words`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	// 创建评论
	mock.ExpectQuery(`INSERT INTO comments`).
		WithArgs(int64(9), int64(1), "好看").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(5, now))
	// daily metric
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))
	// 更新稿件评论数
	mock.ExpectExec(`UPDATE manuscripts SET comment_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	// 写内容审核记录
	mock.ExpectExec(`INSERT INTO content_reviews`).WillReturnResult(sqlmock.NewResult(0, 1))
	// buildComment: 查用户
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(1, "u", "昵称", "http://a", 3))
	// buildComment: 是否已赞
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	resp, err := svc.AddComment(ctx, &pb.AddCommentRequest{UserId: 1, ManuscriptId: 9, Content: "好看"})
	require.NoError(t, err)
	require.NotNil(t, resp.Comment)
	assert.Equal(t, int64(5), resp.Comment.Id)
	assert.Equal(t, "好看", resp.Comment.Content)
	assert.Equal(t, "昵称", resp.Comment.UserName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAddComment_ProhibitedWord(t *testing.T) {
	svc, mock := newCommentSvc(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM prohibited_words`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// 违禁词 → 直接建评论 status=1，无 daily metric / review
	mock.ExpectQuery(`INSERT INTO comments`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(6, now))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(1, "u", "u", "", 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	resp, err := svc.AddComment(ctx, &pb.AddCommentRequest{UserId: 1, ManuscriptId: 9, Content: "脏话"})
	require.NoError(t, err)
	assert.NotNil(t, resp.Comment)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAddComment_RateLimited(t *testing.T) {
	svc, _ := newCommentSvc(t)
	ctx := context.Background()

	for i := 0; i < 20; i++ {
		svc.limiter.record(1, time.Now())
	}
	_, err := svc.AddComment(ctx, &pb.AddCommentRequest{UserId: 1, ManuscriptId: 9, Content: "x"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many comments")
}

func TestAddComment_ReviewRejects(t *testing.T) {
	svc, mock := newCommentSvc(t)
	ctx := context.Background()
	now := time.Now()

	svc.SetReviewService(fakeReview{passed: false})

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM prohibited_words`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`INSERT INTO comments`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(7, now))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET comment_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO content_reviews`).WillReturnResult(sqlmock.NewResult(0, 1))
	// 审核不过 → 更新状态
	mock.ExpectExec(`UPDATE comments SET status`).WithArgs(int64(1), int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(1, "u", "u", "", 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	resp, err := svc.AddComment(ctx, &pb.AddCommentRequest{UserId: 1, ManuscriptId: 9, Content: "可疑内容"})
	require.NoError(t, err)
	assert.NotNil(t, resp.Comment)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAddComment_CreateError(t *testing.T) {
	svc, mock := newCommentSvc(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM prohibited_words`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`INSERT INTO comments`).WillReturnError(sql.ErrConnDone)

	_, err := svc.AddComment(ctx, &pb.AddCommentRequest{UserId: 1, ManuscriptId: 9, Content: "x"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create comment")
}

func TestDeleteComment_NotFound(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectExec(`UPDATE comments SET status = 1`).WillReturnResult(sqlmock.NewResult(0, 0))

	_, err := svc.DeleteComment(context.Background(), &pb.DeleteCommentRequest{Id: 1, UserId: 1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete")
}

func TestDeleteComment_Success(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectExec(`UPDATE comments SET status = 1`).WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := svc.DeleteComment(context.Background(), &pb.DeleteCommentRequest{Id: 1, UserId: 1})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeComment_NotFound(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	_, err := svc.LikeComment(context.Background(), &pb.LikeCommentRequest{CommentId: 1, UserId: 1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "comment not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeComment_Success(t *testing.T) {
	svc, mock := newCommentSvc(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE comments SET like_count = like_count \+ 1`).WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := svc.LikeComment(ctx, &pb.LikeCommentRequest{CommentId: 1, UserId: 1})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUnlikeComment_Success(t *testing.T) {
	svc, mock := newCommentSvc(t)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE comments SET like_count = GREATEST`).WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := svc.UnlikeComment(ctx, &pb.UnlikeCommentRequest{CommentId: 1, UserId: 1})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeComment_ReplySuccess(t *testing.T) {
	svc, mock := newCommentSvc(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM replies`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE replies SET like_count = like_count \+ 1`).WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := svc.LikeReply(ctx, &pb.LikeReplyRequest{ReplyId: 2, UserId: 1})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeComment_DuplicateNoIncrement(t *testing.T) {
	svc, mock := newCommentSvc(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// 已点赞 → INSERT 影响 0 行，不再累加
	mock.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 0))

	_, err := svc.LikeComment(ctx, &pb.LikeCommentRequest{CommentId: 1, UserId: 1})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListComments_Empty(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectQuery(`SELECT id, manuscript_id, user_id, content`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}))

	resp, err := svc.ListComments(context.Background(), &pb.ListCommentsRequest{ManuscriptId: 9, Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Empty(t, resp.Comments)
	require.NoError(t, mock.ExpectationsWereMet())
}

type fakeReview struct {
	passed bool
}

func (f fakeReview) ReviewComment(ctx context.Context, content string) (bool, error) {
	return f.passed, nil
}

func TestCommentService_SetCacheStore(t *testing.T) {
	svc, _ := newCommentSvc(t)
	cache, err := abstraction.NewCacheStore(abstraction.CacheStoreConfig{Type: "memory"})
	require.NoError(t, err)
	svc.SetCacheStore(cache)
	require.NotNil(t, svc.cacheStore)
}
