package comment

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockCommentRepo(t *testing.T) (*CommentRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewCommentRepository(db), mock
}

func TestCreateComment(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO comments`).
		WithArgs(int64(9), int64(1), "hello").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(3, now))

	c := &Comment{ManuscriptID: 9, UserID: 1, Content: "hello"}
	id, err := repo.CreateComment(context.Background(), c)
	require.NoError(t, err)
	assert.Equal(t, int64(3), id)
	assert.Equal(t, now, c.CreatedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateComment_Error(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`INSERT INTO comments`).WillReturnError(errors.New("db down"))
	_, err := repo.CreateComment(context.Background(), &Comment{})
	assert.Error(t, err)
}

func TestFindByID(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, manuscript_id, user_id, content, like_count, reply_count, status, created_at, updated_at`).
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}).
			AddRow(3, 9, 1, "hi", 2, 1, 0, now, now))

	c, err := repo.FindByID(context.Background(), 3)
	require.NoError(t, err)
	assert.Equal(t, int64(3), c.ID)
	assert.Equal(t, "hi", c.Content)
	assert.Equal(t, int32(2), c.LikeCount)
}

func TestFindByID_NotFound(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT id, manuscript_id, user_id, content`).WillReturnError(sql.ErrNoRows)
	_, err := repo.FindByID(context.Background(), 99)
	assert.Error(t, err)
}

func TestListByManuscript_HotSort(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, manuscript_id, user_id, content`).
		WithArgs(int64(9), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}).
			AddRow(1, 9, 1, "a", 5, 0, 0, now, now).
			AddRow(2, 9, 2, "b", 1, 0, 0, now, now))

	list, err := repo.ListByManuscript(context.Background(), 9, 1, 20, "hot")
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, int64(1), list[0].ID)
	assert.Equal(t, "b", list[1].Content)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListByManuscript_Error(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT id, manuscript_id, user_id, content`).WillReturnError(errors.New("boom"))
	_, err := repo.ListByManuscript(context.Background(), 9, 1, 20, "")
	assert.Error(t, err)
}

func TestDelete_NotOwner(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectExec(`UPDATE comments SET status = 1`).WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.Delete(context.Background(), 1, 999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestDelete_Success(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectExec(`UPDATE comments SET status = 1`).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Delete(context.Background(), 1, 1))
}

func TestIncrementReplyCount(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectExec(`UPDATE comments SET reply_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.IncrementReplyCount(context.Background(), 1))
}

func TestFindUserByID(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(1, "u", "n", "a", 2))

	u, err := repo.FindUserByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "n", u.Nickname)
	assert.Equal(t, int32(2), u.Level)
}

func TestFindUserByID_NotFound(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).WillReturnError(sql.ErrNoRows)
	_, err := repo.FindUserByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestFindUsersByIDs_Empty(t *testing.T) {
	repo, _ := newMockCommentRepo(t)
	m := repo.FindUsersByIDs(context.Background(), nil)
	assert.Empty(t, m)
}

func TestFindUsersByIDs_QueryError(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users`).WillReturnError(errors.New("boom"))
	m := repo.FindUsersByIDs(context.Background(), []int64{1})
	assert.Empty(t, m)
}

func TestWriteContentReview(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectExec(`INSERT INTO content_reviews`).WithArgs("comment", int64(1), "c").WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.WriteContentReview(context.Background(), "comment", 1, "c"))
}

func TestUpdateCommentStatus(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	mock.ExpectExec(`UPDATE comments SET status = \$1`).WithArgs(int64(1), int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpdateCommentStatus(context.Background(), 5, 1))
}
