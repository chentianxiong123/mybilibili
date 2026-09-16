package comment

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "mybilibili/pkg/pb"
)

func newCommentHandler(t *testing.T) (*CommentHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	svc := NewCommentService(NewCommentRepository(db))
	svc.SetDB(db)
	return NewCommentHandler(svc), mock
}

func TestCommentHandler_AddComment(t *testing.T) {
	h, mock := newCommentHandler(t)
	now := time.Now()
	mock.ExpectQuery(`INSERT INTO comments`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WithArgs(int64(9), int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET comment_count = comment_count \+ 1`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO content_reviews`).WithArgs("comment", int64(1), "hello").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(1, "u", "nick", "a", 1))

	resp, err := h.AddComment(context.Background(), &pb.AddCommentRequest{ManuscriptId: 9, UserId: 1, Content: "hello"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Comment.Id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentHandler_ListComments(t *testing.T) {
	h, mock := newCommentHandler(t)
	now := time.Now()
	mock.ExpectQuery(`FROM comments WHERE manuscript_id = \$1 AND status = 0`).
		WithArgs(int64(9), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}).
			AddRow(1, 9, 200, "c", 1, 0, 0, now, now))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).WithArgs(int64(200)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(200, "u", "nick", "a", 1))

	resp, err := h.ListComments(context.Background(), &pb.ListCommentsRequest{ManuscriptId: 9, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, resp.Comments, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentHandler_DeleteComment(t *testing.T) {
	h, mock := newCommentHandler(t)
	mock.ExpectExec(`UPDATE comments SET status = 1`).WithArgs(int64(1), int64(200)).WillReturnResult(sqlmock.NewResult(0, 1))
	_, err := h.DeleteComment(context.Background(), &pb.DeleteCommentRequest{Id: 1, UserId: 200})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentHandler_AddReply(t *testing.T) {
	h, mock := newCommentHandler(t)
	now := time.Now()
	mock.ExpectQuery(`INSERT INTO replies`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))
	mock.ExpectExec(`UPDATE comments SET reply_count = reply_count \+ 1`).WithArgs(int64(100)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id, manuscript_id, user_id, content`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}).
			AddRow(100, 9, 300, "parent", 1, 1, 0, now, now))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WithArgs(int64(9), int64(200), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET comment_count = comment_count \+ 1`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).WithArgs(int64(200)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(200, "u", "nick", "a", 1))

	resp, err := h.AddReply(context.Background(), &pb.AddReplyRequest{CommentId: 100, UserId: 200, Content: "hi"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Reply.Id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentHandler_GetReplies(t *testing.T) {
	h, mock := newCommentHandler(t)
	now := time.Now()
	mock.ExpectQuery(`FROM replies WHERE comment_id = \$1 AND status = 'NORMAL'`).
		WithArgs(int64(100), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "user_id", "reply_to_user_id", "content", "like_count", "status", "created_at", "updated_at"}).
			AddRow(1, 100, 200, nil, "r", 1, "NORMAL", now, now))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).WithArgs(int64(200)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(200, "u", "nick", "a", 1))

	resp, err := h.GetReplies(context.Background(), &pb.GetRepliesRequest{CommentId: 100, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, resp.Replies, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentHandler_DeleteReply(t *testing.T) {
	h, mock := newCommentHandler(t)
	mock.ExpectExec(`UPDATE replies SET status = 'REMOVED'`).WithArgs(int64(1), int64(200)).WillReturnResult(sqlmock.NewResult(0, 1))
	_, err := h.DeleteReply(context.Background(), &pb.DeleteReplyRequest{Id: 1, UserId: 200})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentHandler_LikeComment(t *testing.T) {
	h, mock := newCommentHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO user_interactions`).WithArgs(int64(1), "COMMENT", int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE comments SET like_count = like_count \+ 1`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := h.LikeComment(context.Background(), &pb.LikeCommentRequest{CommentId: 9, UserId: 1})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentHandler_UnlikeComment(t *testing.T) {
	h, mock := newCommentHandler(t)
	mock.ExpectExec(`DELETE FROM user_interactions`).WithArgs(int64(1), "COMMENT", int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE comments SET like_count = GREATEST`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	_, err := h.UnlikeComment(context.Background(), &pb.UnlikeCommentRequest{CommentId: 9, UserId: 1})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentHandler_LikeReply(t *testing.T) {
	h, mock := newCommentHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM replies`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO user_interactions`).WithArgs(int64(1), "REPLY", int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE replies SET like_count = like_count \+ 1`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := h.LikeReply(context.Background(), &pb.LikeReplyRequest{ReplyId: 9, UserId: 1})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentHandler_UnlikeReply(t *testing.T) {
	h, mock := newCommentHandler(t)
	mock.ExpectExec(`DELETE FROM user_interactions`).WithArgs(int64(1), "REPLY", int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE replies SET like_count = GREATEST`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	_, err := h.UnlikeReply(context.Background(), &pb.UnlikeReplyRequest{ReplyId: 9, UserId: 1})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
