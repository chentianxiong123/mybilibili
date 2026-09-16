package comment

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "mybilibili/pkg/pb"
)

// argConverter 允许 sqlmock 处理 []int64 参数（生产环境由 lib/pq 转换为数组字面量）。
type argConverter struct{}

func (argConverter) ConvertValue(v interface{}) (driver.Value, error) {
	if arr, ok := v.([]int64); ok {
		s := "{"
		for i, x := range arr {
			if i > 0 {
				s += ","
			}
			s += strconv.FormatInt(x, 10)
		}
		return s + "}", nil
	}
	return driver.DefaultParameterConverter.ConvertValue(v)
}

func newMockCommentRepoConv(t *testing.T) (*CommentRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
		sqlmock.ValueConverterOption(argConverter{}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewCommentRepository(db), mock
}

func replyRow() []interface{} {
	return []interface{}{1, 100, 200, nil, "nice", 3, "NORMAL", time.Now(), time.Now()}
}

func TestCommentRepository_CreateReply(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	now := time.Now()
	mock.ExpectQuery(`INSERT INTO replies`).
		WithArgs(int64(100), int64(200), nil, "nice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))

	rep := &Reply{CommentID: 100, UserID: 200, Content: "nice"}
	id, err := repo.CreateReply(context.Background(), rep)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.Equal(t, now, rep.CreatedAt)
}

func TestCommentRepository_CreateReply_WithReplyTo(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	now := time.Now()
	mock.ExpectQuery(`INSERT INTO replies`).
		WithArgs(int64(100), int64(200), int64(50), "nice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(2, now))

	rep := &Reply{CommentID: 100, UserID: 200, Content: "nice", ReplyToUserID: sql.NullInt64{Int64: 50, Valid: true}}
	id, err := repo.CreateReply(context.Background(), rep)
	require.NoError(t, err)
	assert.Equal(t, int64(2), id)
}

func TestCommentRepository_CreateReply_Error(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`INSERT INTO replies`).WillReturnError(errors.New("boom"))
	_, err := repo.CreateReply(context.Background(), &Reply{CommentID: 1, UserID: 2, Content: "x"})
	assert.Error(t, err)
}

func TestCommentRepository_FindReplyByID(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT id, comment_id, user_id, reply_to_user_id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "user_id", "reply_to_user_id", "content", "like_count", "status", "created_at", "updated_at"}).
			AddRow(1, 100, 200, 50, "nice", 3, "NORMAL", now, now))
	rep, err := repo.FindReplyByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(100), rep.CommentID)
	assert.Equal(t, "nice", rep.Content)
}

func TestCommentRepository_FindReplyByID_Error(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT id, comment_id, user_id, reply_to_user_id`).WithArgs(int64(9)).WillReturnError(sql.ErrNoRows)
	_, err := repo.FindReplyByID(context.Background(), 9)
	assert.Error(t, err)
}

func TestCommentRepository_ListRepliesByComment(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`FROM replies WHERE comment_id = \$1 AND status = 'NORMAL' ORDER BY created_at ASC`).
		WithArgs(int64(100), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "user_id", "reply_to_user_id", "content", "like_count", "status", "created_at", "updated_at"}).
			AddRow(1, 100, 200, 50, "nice", 3, "NORMAL", time.Now(), time.Now()))
	list, err := repo.ListRepliesByComment(context.Background(), 100, 1, 20)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(200), list[0].UserID)
}

func TestCommentRepository_ListRepliesByComment_ScanError(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`FROM replies WHERE comment_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "user_id", "reply_to_user_id", "content", "like_count", "status", "created_at", "updated_at"}).
			AddRow("bad", 100, 200, nil, "x", 1, "NORMAL", time.Now(), time.Now()))
	_, err := repo.ListRepliesByComment(context.Background(), 100, 1, 20)
	assert.Error(t, err)
}

func TestCommentRepository_DeleteReply(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectExec(`UPDATE replies SET status = 'REMOVED'`).WithArgs(int64(1), int64(200)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.DeleteReply(context.Background(), 1, 200))

	mock.ExpectExec(`UPDATE replies SET status = 'REMOVED'`).WithArgs(int64(1), int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.DeleteReply(context.Background(), 1, 999)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestCommentRepository_IsReplyLiked(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).WithArgs(int64(1), int64(2)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	ok, err := repo.IsReplyLiked(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCommentRepository_LikeTarget_Comment(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO user_interactions`).WithArgs(int64(2), "COMMENT", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE comments SET like_count = like_count \+ 1`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.LikeTarget(context.Background(), "comment", 1, 2))
}

func TestCommentRepository_LikeTarget_CommentNotExists(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	err := repo.LikeTarget(context.Background(), "comment", 1, 2)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestCommentRepository_LikeTarget_CommentExistsError(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WithArgs(int64(1)).WillReturnError(errors.New("boom"))
	err := repo.LikeTarget(context.Background(), "comment", 1, 2)
	assert.Error(t, err)
}

func TestCommentRepository_LikeTarget_CommentDuplicate(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO user_interactions`).WithArgs(int64(2), "COMMENT", int64(1)).WillReturnResult(sqlmock.NewResult(0, 0))
	require.NoError(t, repo.LikeTarget(context.Background(), "comment", 1, 2))
}

func TestCommentRepository_LikeTarget_Reply(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM replies`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO user_interactions`).WithArgs(int64(2), "REPLY", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE replies SET like_count = like_count \+ 1`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.LikeTarget(context.Background(), "reply", 1, 2))
}

func TestCommentRepository_LikeTarget_ReplyNotExists(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM replies`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	err := repo.LikeTarget(context.Background(), "reply", 1, 2)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestCommentRepository_LikeTarget_InsertError(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO user_interactions`).WillReturnError(errors.New("boom"))
	err := repo.LikeTarget(context.Background(), "comment", 1, 2)
	assert.Error(t, err)
}

func TestCommentRepository_UnlikeTarget(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectExec(`DELETE FROM user_interactions`).WithArgs(int64(2), "COMMENT", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE comments SET like_count = GREATEST\(like_count - 1, 0\)`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UnlikeTarget(context.Background(), "comment", 1, 2))

	mock.ExpectExec(`DELETE FROM user_interactions`).WithArgs(int64(2), "COMMENT", int64(1)).WillReturnResult(sqlmock.NewResult(0, 0))
	require.NoError(t, repo.UnlikeTarget(context.Background(), "comment", 1, 2))

	mock.ExpectExec(`DELETE FROM user_interactions`).WithArgs(int64(2), "REPLY", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE replies SET like_count = GREATEST`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UnlikeTarget(context.Background(), "reply", 1, 2))

	mock.ExpectExec(`DELETE FROM user_interactions`).WillReturnError(errors.New("boom"))
	assert.Error(t, repo.UnlikeTarget(context.Background(), "comment", 1, 2))
}

func TestCommentRepository_BatchGetLikeCounts(t *testing.T) {
	repo, mock := newMockCommentRepo(t)

	out, err := repo.BatchGetLikeCounts(context.Background(), "COMMENT", nil)
	require.NoError(t, err)
	assert.Empty(t, out)

	mock.ExpectQuery(`SELECT id, like_count FROM comments WHERE id IN \(\$1,\$2\)`).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "like_count"}).AddRow(1, 5).AddRow(2, 3))
	out, err = repo.BatchGetLikeCounts(context.Background(), "COMMENT", []int64{1, 2})
	require.NoError(t, err)
	assert.Equal(t, map[int64]int64{1: 5, 2: 3}, out)

	mock.ExpectQuery(`SELECT id, like_count FROM replies WHERE id IN \(\$1\)`).
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "like_count"}).AddRow(3, 7))
	out, err = repo.BatchGetLikeCounts(context.Background(), "REPLY", []int64{3})
	require.NoError(t, err)
	assert.Equal(t, map[int64]int64{3: 7}, out)

	mock.ExpectQuery(`SELECT id, like_count FROM comments WHERE id IN`).WillReturnError(errors.New("boom"))
	_, err = repo.BatchGetLikeCounts(context.Background(), "COMMENT", []int64{1})
	assert.Error(t, err)
}

func TestCommentRepository_BatchIsLiked(t *testing.T) {
	repo, mock := newMockCommentRepo(t)

	out, err := repo.BatchIsLiked(context.Background(), 9, "COMMENT", nil)
	require.NoError(t, err)
	assert.Empty(t, out)

	mock.ExpectQuery(`SELECT target_id FROM user_interactions`).
		WithArgs(int64(9), "COMMENT", int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"target_id"}).AddRow(1))
	out, err = repo.BatchIsLiked(context.Background(), 9, "COMMENT", []int64{1, 2})
	require.NoError(t, err)
	assert.Equal(t, map[int64]bool{1: true}, out)

	mock.ExpectQuery(`SELECT target_id FROM user_interactions`).WillReturnError(errors.New("boom"))
	_, err = repo.BatchIsLiked(context.Background(), 9, "COMMENT", []int64{1})
	assert.Error(t, err)
}

func TestReplyToPB(t *testing.T) {
	rep := &Reply{
		ID: 1, CommentID: 2, UserID: 3, Content: "nice", LikeCount: 4,
		CreatedAt: time.Unix(0, 0).UTC(), Status: "NORMAL",
	}
	info := replyToPB(rep, "nick", "avatar", 5, "toname", 6, true)
	assert.Equal(t, int64(1), info.Id)
	assert.Equal(t, "nice", info.Content)
	assert.Equal(t, "toname", info.ReplyToUserName)
	assert.Equal(t, int64(6), info.ReplyToUserId)
	assert.Equal(t, "1970-01-01T00:00:00Z", info.CreatedAt)
}

func TestCommentRepository_ListCommentsByCreator(t *testing.T) {
	repo, mock := newMockCommentRepo(t)

	list, err := repo.ListCommentsByCreator(context.Background(), 100, 0, 1, 20, "latest", "reply")
	require.NoError(t, err)
	assert.Nil(t, list)

	now := time.Now()
	mock.ExpectQuery(`FROM comments JOIN manuscripts m ON m.id = comments.manuscript_id WHERE m.user_id = \$1 ORDER BY comments.created_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(int64(100), int32(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}).
			AddRow(1, 9, 200, "c", 3, 1, 0, now, now))
	list, err = repo.ListCommentsByCreator(context.Background(), 100, 0, 1, 20, "latest", "comment")
	require.NoError(t, err)
	require.Len(t, list, 1)

	mock.ExpectQuery(`AND comments.manuscript_id = \$2 ORDER BY comments.created_at ASC LIMIT \$3 OFFSET \$4`).
		WithArgs(int64(100), int64(9), int32(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}))
	list, err = repo.ListCommentsByCreator(context.Background(), 100, 9, 1, 20, "oldest", "comment")
	require.NoError(t, err)
	assert.Len(t, list, 0)

	mock.ExpectQuery(`ORDER BY comments.like_count DESC`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}))
	list, err = repo.ListCommentsByCreator(context.Background(), 100, 0, 1, 20, "likes", "comment")
	require.NoError(t, err)
	assert.Len(t, list, 0)

	mock.ExpectQuery(`FROM comments JOIN manuscripts`).WillReturnError(errors.New("boom"))
	_, err = repo.ListCommentsByCreator(context.Background(), 100, 0, 1, 20, "latest", "comment")
	assert.Error(t, err)
}

func TestCommentRepository_CountCommentsByCreator(t *testing.T) {
	repo, mock := newMockCommentRepo(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM replies r JOIN comments c ON r.comment_id = c.id`).
		WithArgs(int64(100), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	total, err := repo.CountCommentsByCreator(context.Background(), 100, 9, "reply")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM replies r JOIN comments c`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	total, err = repo.CountCommentsByCreator(context.Background(), 100, 0, "reply")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments JOIN manuscripts m ON m.id = comments.manuscript_id WHERE m.user_id = \$1`).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	total, err = repo.CountCommentsByCreator(context.Background(), 100, 0, "comment")
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)

	mock.ExpectQuery(`AND comments.manuscript_id = \$2`).WithArgs(int64(100), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	total, err = repo.CountCommentsByCreator(context.Background(), 100, 9, "comment")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
}

func TestCommentRepository_ListRepliesByCreator(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	now := time.Now()

	mock.ExpectQuery(`FROM replies r JOIN comments c ON r.comment_id = c.id JOIN manuscripts m`).
		WithArgs(int64(100), int32(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "user_id", "reply_to_user_id", "content", "like_count", "status", "created_at", "updated_at", "manuscript_id"}).
			AddRow(1, 2, 3, nil, "r", 1, "NORMAL", now, now, 9))
	list, err := repo.ListRepliesByCreator(context.Background(), 100, 0, 1, 20, "latest")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(9), list[0].ManuscriptID)

	mock.ExpectQuery(`AND c.manuscript_id = \$2 ORDER BY r.created_at ASC`).
		WithArgs(int64(100), int64(9), int32(10), int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "user_id", "reply_to_user_id", "content", "like_count", "status", "created_at", "updated_at", "manuscript_id"}))
	list, err = repo.ListRepliesByCreator(context.Background(), 100, 9, 2, 10, "oldest")
	require.NoError(t, err)
	assert.Len(t, list, 0)

	mock.ExpectQuery(`ORDER BY r.like_count DESC`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "user_id", "reply_to_user_id", "content", "like_count", "status", "created_at", "updated_at", "manuscript_id"}))
	_, err = repo.ListRepliesByCreator(context.Background(), 100, 0, 1, 20, "likes")
	require.NoError(t, err)

	mock.ExpectQuery(`FROM replies r JOIN comments`).WillReturnError(errors.New("boom"))
	_, err = repo.ListRepliesByCreator(context.Background(), 100, 0, 1, 20, "latest")
	assert.Error(t, err)
}

func TestCommentRepository_IsCommentOwnedByCreator(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments c JOIN manuscripts`).WithArgs(int64(1), int64(100)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	ok, err := repo.IsCommentOwnedByCreator(context.Background(), 1, 100)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCommentRepository_DeleteCommentByCreator(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectExec(`DELETE FROM comments`).WithArgs(int64(1), int64(100)).WillReturnResult(sqlmock.NewResult(0, 1))
	n, err := repo.DeleteCommentByCreator(context.Background(), 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	mock.ExpectExec(`DELETE FROM comments`).WillReturnError(errors.New("boom"))
	_, err = repo.DeleteCommentByCreator(context.Background(), 1, 100)
	assert.Error(t, err)
}

func TestCommentRepository_DeleteReplyByCreator(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectExec(`DELETE FROM replies`).WithArgs(int64(1), int64(100)).WillReturnResult(sqlmock.NewResult(0, 1))
	n, err := repo.DeleteReplyByCreator(context.Background(), 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	mock.ExpectExec(`DELETE FROM replies`).WillReturnError(errors.New("boom"))
	_, err = repo.DeleteReplyByCreator(context.Background(), 1, 100)
	assert.Error(t, err)
}

func TestCommentRepository_FindUsersByIDs_Success(t *testing.T) {
	repo, mock := newMockCommentRepoConv(t)
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users WHERE id = ANY\(\$1\)`).
		WithArgs("{9,10}").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).
			AddRow(9, "u9", "nick9", "a", 1).
			AddRow(10, "u10", "nick10", "b", 2))
	out := repo.FindUsersByIDs(context.Background(), []int64{9, 10})
	require.Len(t, out, 2)
	assert.Equal(t, "nick9", out[9].Nickname)
}

// ---- service ----

type fakeNotifier struct {
	sent int
}

func (f *fakeNotifier) SendMessage(ctx context.Context, senderID, receiverID int64, content string, msgType int32) {
	f.sent++
}

func TestCommentService_RepoAndSetters(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()
	svc := NewCommentService(NewCommentRepository(db))
	svc.SetDB(db)
	svc.SetNotifier(&fakeNotifier{})
	assert.NotNil(t, svc.Repo())
	_ = mock
}

func TestCommentService_AddReply_Success(t *testing.T) {
	svc, mock := newCommentSvc(t)
	svc.SetNotifier(&fakeNotifier{})
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO replies`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))
	mock.ExpectExec(`UPDATE comments SET reply_count = reply_count \+ 1`).WithArgs(int64(100)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id, manuscript_id, user_id, content`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}).
			AddRow(100, 9, 300, "parent", 1, 2, 0, now, now))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WithArgs(int64(9), int64(200), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET comment_count = comment_count \+ 1`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT user_id FROM comments`).WithArgs(int64(100)).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(300))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).WithArgs(int64(200)).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(200, "u", "nick", "a", 1))

	resp, err := svc.AddReply(ctx, &pb.AddReplyRequest{CommentId: 100, UserId: 200, Content: "nice", ReplyToUserId: 50})
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Reply.Id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentService_AddReply_EmptyContent(t *testing.T) {
	svc, _ := newCommentSvc(t)
	_, err := svc.AddReply(context.Background(), &pb.AddReplyRequest{CommentId: 1, UserId: 1})
	assert.Error(t, err)
}

func TestCommentService_AddReply_CreateError(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectQuery(`INSERT INTO replies`).WillReturnError(errors.New("boom"))
	_, err := svc.AddReply(context.Background(), &pb.AddReplyRequest{CommentId: 1, UserId: 1, Content: "x"})
	assert.Error(t, err)
}

func TestCommentService_GetReplies(t *testing.T) {
	svc, mock := newCommentSvc(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(`FROM replies WHERE comment_id = \$1 AND status = 'NORMAL'`).
		WithArgs(int64(100), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "user_id", "reply_to_user_id", "content", "like_count", "status", "created_at", "updated_at"}).
			AddRow(1, 100, 200, 50, "nice", 3, "NORMAL", now, now))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).WithArgs(int64(200)).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(200, "u", "nick", "a", 1))
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).WithArgs(int64(50)).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(50, "u5", "nick5", "a", 2))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).WithArgs(int64(1), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	resp, err := svc.GetReplies(ctx, &pb.GetRepliesRequest{CommentId: 100, UserId: 1, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, resp.Replies, 1)
	assert.Equal(t, "nice", resp.Replies[0].Content)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentService_DeleteReply(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectExec(`UPDATE replies SET status = 'REMOVED'`).WithArgs(int64(1), int64(200)).WillReturnResult(sqlmock.NewResult(0, 1))
	_, err := svc.DeleteReply(context.Background(), &pb.DeleteReplyRequest{Id: 1, UserId: 200})
	require.NoError(t, err)

	mock.ExpectExec(`UPDATE replies SET status = 'REMOVED'`).WithArgs(int64(1), int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	_, err = svc.DeleteReply(context.Background(), &pb.DeleteReplyRequest{Id: 1, UserId: 999})
	assert.Error(t, err)
}

func TestCommentService_UnlikeReply(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectExec(`DELETE FROM user_interactions`).WithArgs(int64(1), "REPLY", int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE replies SET like_count = GREATEST`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	_, err := svc.UnlikeReply(context.Background(), &pb.UnlikeReplyRequest{ReplyId: 9, UserId: 1})
	require.NoError(t, err)

	mock.ExpectExec(`DELETE FROM user_interactions`).WillReturnError(errors.New("boom"))
	_, err = svc.UnlikeReply(context.Background(), &pb.UnlikeReplyRequest{ReplyId: 9, UserId: 1})
	assert.Error(t, err)
}

func TestCommentService_LikeReply_NotFound(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM replies`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	_, err := svc.LikeReply(context.Background(), &pb.LikeReplyRequest{ReplyId: 9, UserId: 1})
	assert.Error(t, err)
}

func TestCommentService_LikeComment_Notification(t *testing.T) {
	svc, mock := newCommentSvc(t)
	notifier := &fakeNotifier{}
	svc.SetNotifier(notifier)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO user_interactions`).WithArgs(int64(1), "COMMENT", int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE comments SET like_count = like_count \+ 1`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT user_id FROM comments`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(50))

	_, err := svc.LikeComment(ctx, &pb.LikeCommentRequest{CommentId: 9, UserId: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, notifier.sent)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentService_LikeComment_OwnerIsSender(t *testing.T) {
	svc, mock := newCommentSvc(t)
	notifier := &fakeNotifier{}
	svc.SetNotifier(notifier)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO user_interactions`).WithArgs(int64(1), "COMMENT", int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE comments SET like_count = like_count \+ 1`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT user_id FROM comments`).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

	_, err := svc.LikeComment(ctx, &pb.LikeCommentRequest{CommentId: 9, UserId: 1})
	require.NoError(t, err)
	assert.Equal(t, 0, notifier.sent)
	require.NoError(t, mock.ExpectationsWereMet())
}
