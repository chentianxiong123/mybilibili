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

// seedProhibited 直接往违禁词内存缓存里塞词。
// 违禁词检查已从"每次查库"改为内存匹配，所以测试不再需要 mock 那条 SQL。
func seedProhibited(svc *CommentService, words ...ProhibitedWord) {
	store := newProhibitedWordStore(nil, time.Minute)
	store.words = words
	store.loaded = true
	svc.prohibited = store
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

	seedProhibited(svc, ProhibitedWord{Word: "脏话", MatchType: matchTypeContains})
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
		svc.commentLimiter.record(1, time.Now())
	}
	_, err := svc.AddComment(ctx, &pb.AddCommentRequest{UserId: 1, ManuscriptId: 9, Content: "x"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many comments")
}

// 评论与回复现在各用独立的桶，回复打满不影响评论。
func TestAddComment_ReplyLimitIsIndependent(t *testing.T) {
	svc, mock := newCommentSvc(t)
	ctx := context.Background()
	now := time.Now()

	for i := 0; i < 20; i++ {
		svc.replyLimiter.record(1, time.Now())
	}

	// 回复桶满了，但评论桶是空的，评论应该照发
	mock.ExpectQuery(`INSERT INTO comments`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(8, now))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET comment_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO content_reviews`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(1, "u", "u", "", 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	_, err := svc.AddComment(ctx, &pb.AddCommentRequest{UserId: 1, ManuscriptId: 9, Content: "x"})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// 回复超限要给出"回复"而不是"评论"的提示，便于用户知道该降频哪一类。
func TestAddReply_RateLimited(t *testing.T) {
	svc, _ := newCommentSvc(t)
	ctx := context.Background()

	for i := 0; i < 20; i++ {
		svc.replyLimiter.record(1, time.Now())
	}
	_, err := svc.AddReply(ctx, &pb.AddReplyRequest{UserId: 1, CommentId: 3, Content: "x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many replies")
}

// 回复命中违禁词：落库标 PENDING（公开列表只查 NORMAL），且不计数、不推送通知。
func TestAddReply_ProhibitedWord(t *testing.T) {
	svc, mock := newCommentSvc(t)
	ctx := context.Background()
	now := time.Now()

	seedProhibited(svc, ProhibitedWord{Word: "脏话", MatchType: matchTypeContains})

	// 断言真的把 PENDING 写进了 replies.status
	mock.ExpectQuery(`INSERT INTO replies`).
		WithArgs(int64(3), int64(1), nil, "你个脏话", statusPendingReview).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(11, now))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(1, "u", "u", "", 1))

	resp, err := svc.AddReply(ctx, &pb.AddReplyRequest{UserId: 1, CommentId: 3, Content: "你个脏话"})
	require.NoError(t, err)
	require.NotNil(t, resp.Reply)
	// 关键：没有 UPDATE replies SET reply_count / manuscripts.comment_count / 通知
	require.NoError(t, mock.ExpectationsWereMet())
}

// 回复正常时行为不变：计数 + 通知都照旧。
func TestAddReply_NormalStillCounts(t *testing.T) {
	svc, mock := newCommentSvc(t)
	ctx := context.Background()
	now := time.Now()

	seedProhibited(svc, ProhibitedWord{Word: "脏话", MatchType: matchTypeContains})

	mock.ExpectQuery(`INSERT INTO replies`).
		WithArgs(int64(3), int64(1), nil, "正常回复", replyStatusNormal).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(12, now))
	mock.ExpectExec(`UPDATE comments SET reply_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id, manuscript_id, user_id, content, like_count, reply_count, status, created_at, updated_at
		 FROM comments`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}).
			AddRow(3, 9, 1, "父评论", 0, 0, 0, now, now))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(1, "u", "u", "", 1))

	_, err := svc.AddReply(ctx, &pb.AddReplyRequest{UserId: 1, CommentId: 3, Content: "正常回复"})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAddComment_ReviewRejects(t *testing.T) {
	svc, mock := newCommentSvc(t)
	ctx := context.Background()
	now := time.Now()

	svc.SetReviewService(fakeReview{passed: false})

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
