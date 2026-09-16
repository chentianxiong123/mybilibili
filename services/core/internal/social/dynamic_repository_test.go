package social

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

func newMockDynamicRepo(t *testing.T) (*DynamicRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewDynamicRepository(db), mock
}

func dynRow(cols ...interface{}) []interface{} {
	if len(cols) == 0 {
		return []interface{}{1, 100, "hello", 1, "http://img", 9, 3, 2, 1, 0, time.Now()}
	}
	return cols
}

func TestDynamicRepository_Create(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`INSERT INTO user_dynamics`).
		WithArgs(int64(100), "hello", int32(1), "http://img", int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	id, err := repo.Create(context.Background(), &Dynamic{UserID: 100, Content: "hello", DynamicType: 1, ImageURL: "http://img", RefManuscriptID: 9})
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
}

func TestDynamicRepository_Create_NilRef(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`INSERT INTO user_dynamics`).
		WithArgs(int64(100), "hello", int32(1), "", nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	id, err := repo.Create(context.Background(), &Dynamic{UserID: 100, Content: "hello", DynamicType: 1})
	require.NoError(t, err)
	assert.Equal(t, int64(2), id)
}

func TestDynamicRepository_Create_Error(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`INSERT INTO user_dynamics`).WillReturnError(errors.New("insert failed"))
	_, err := repo.Create(context.Background(), &Dynamic{UserID: 100, Content: "x"})
	assert.Error(t, err)
}

func TestDynamicRepository_GetByID(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT id, user_id, content, dynamic_type`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}).
			AddRow(1, 100, "hello", 1, "http://img", 9, 3, 2, 1, 0, now))

	d, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), d.ID)
	assert.Equal(t, int64(100), d.UserID)
	assert.Equal(t, "hello", d.Content)
	assert.Equal(t, int32(2), d.CommentCount)
	assert.Equal(t, now, d.CreatedAt)
}

func TestDynamicRepository_GetByID_Error(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`SELECT id, user_id, content, dynamic_type`).WithArgs(int64(9)).WillReturnError(sql.ErrNoRows)
	d, err := repo.GetByID(context.Background(), 9)
	assert.Error(t, err)
	assert.Nil(t, d)
}

func TestDynamicRepository_ListByUser(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`FROM user_dynamics WHERE user_id = \$1 AND status = 0`).
		WithArgs(int64(100), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}).
			AddRow(1, 100, "a", 1, "", 0, 1, 2, 3, 0, time.Now()).
			AddRow(2, 100, "b", 2, "", 0, 0, 0, 0, 0, time.Now()))

	list, err := repo.ListByUser(context.Background(), 100, 1, 20)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, int64(2), list[1].ID)
}

func TestDynamicRepository_ListByUser_Empty(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`FROM user_dynamics WHERE user_id = \$1 AND status = 0`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}))
	list, err := repo.ListByUser(context.Background(), 100, 1, 20)
	require.NoError(t, err)
	assert.Len(t, list, 0)
}

func TestDynamicRepository_ListByUser_QueryError(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`FROM user_dynamics WHERE user_id = \$1 AND status = 0`).WillReturnError(errors.New("boom"))
	_, err := repo.ListByUser(context.Background(), 100, 1, 20)
	assert.Error(t, err)
}

func TestDynamicRepository_ListByUser_ScanError(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`FROM user_dynamics WHERE user_id = \$1 AND status = 0`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}).
			AddRow("bad", 100, "a", 1, "", 0, 1, 2, 3, 0, time.Now()))
	_, err := repo.ListByUser(context.Background(), 100, 1, 20)
	assert.Error(t, err)
}

func TestDynamicRepository_ListFollowing(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`FROM user_dynamics d WHERE d.status = 0`).
		WithArgs(int64(100), int32(10), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}).
			AddRow(5, 200, "c", 1, "", 0, 0, 0, 0, 0, time.Now()))
	list, err := repo.ListFollowing(context.Background(), 100, 1, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(5), list[0].ID)
}

func TestDynamicRepository_ListFollowing_QueryError(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`FROM user_dynamics d WHERE d.status = 0`).WillReturnError(errors.New("boom"))
	_, err := repo.ListFollowing(context.Background(), 100, 1, 10)
	assert.Error(t, err)
}

func TestDynamicRepository_ListAll(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`FROM user_dynamics WHERE status = 0 ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(int32(10), int32(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}).
			AddRow(3, 300, "d", 1, "", 0, 0, 0, 0, 0, time.Now()))
	list, err := repo.ListAll(context.Background(), 2, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestDynamicRepository_Delete(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectExec(`UPDATE user_dynamics SET status = 1`).WithArgs(int64(1), int64(100)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Delete(context.Background(), 1, 100))
}

func TestDynamicRepository_Delete_Error(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectExec(`UPDATE user_dynamics SET status = 1`).WillReturnError(errors.New("boom"))
	assert.Error(t, repo.Delete(context.Background(), 1, 100))
}

func TestDynamicRepository_IncrCounts(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectExec(`like_count = GREATEST\(like_count \+ \$1, 0\)`).WithArgs(int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.IncrLikeCount(context.Background(), 1, 1))

	mock.ExpectExec(`comment_count = GREATEST\(comment_count \+ \$1, 0\)`).WithArgs(int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.IncrCommentCount(context.Background(), 1, 1))

	mock.ExpectExec(`share_count = GREATEST\(share_count \+ \$1, 0\)`).WithArgs(int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.IncrShareCount(context.Background(), 1, 1))
}

func TestDynamicRepository_IncrLikeCount_Error(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectExec(`like_count = GREATEST`).WillReturnError(errors.New("boom"))
	assert.Error(t, repo.IncrLikeCount(context.Background(), 1, 1))
}

func TestDynamicRepository_IsLiked(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM dynamic_likes`).WithArgs(int64(1), int64(100)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	ok, err := repo.IsLiked(context.Background(), 1, 100)
	require.NoError(t, err)
	assert.True(t, ok)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM dynamic_likes`).WithArgs(int64(1), int64(100)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	ok, err = repo.IsLiked(context.Background(), 1, 100)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestDynamicRepository_CreateComment(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`INSERT INTO dynamic_comments`).
		WithArgs(int64(1), int64(100), "nice", nil, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	id, err := repo.CreateComment(context.Background(), &DynamicComment{DynamicID: 1, UserID: 100, Content: "nice"})
	require.NoError(t, err)
	assert.Equal(t, int64(11), id)
}

func TestDynamicRepository_CreateComment_WithParent(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`INSERT INTO dynamic_comments`).
		WithArgs(int64(1), int64(100), "reply", int64(11), int64(200)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(12))
	id, err := repo.CreateComment(context.Background(), &DynamicComment{DynamicID: 1, UserID: 100, Content: "reply", ParentID: 11, ReplyUserID: 200})
	require.NoError(t, err)
	assert.Equal(t, int64(12), id)
}

func TestDynamicRepository_CreateComment_Error(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`INSERT INTO dynamic_comments`).WillReturnError(errors.New("boom"))
	_, err := repo.CreateComment(context.Background(), &DynamicComment{DynamicID: 1, UserID: 100, Content: "x"})
	assert.Error(t, err)
}

func TestDynamicRepository_ListComments_DefaultSort(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dynamic_id", "user_id", "content", "parent_id", "reply_user_id", "like_count", "status", "created_at", "updated_at"}).
			AddRow(1, 1, 100, "c1", 0, 0, 2, 0, time.Now(), time.Now()))
	list, err := repo.ListComments(context.Background(), 1, 1, 20, "new")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), list[0].ID)
}

func TestDynamicRepository_ListComments_HotSort(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`ORDER BY like_count DESC, created_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(int64(1), int32(10), int32(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dynamic_id", "user_id", "content", "parent_id", "reply_user_id", "like_count", "status", "created_at", "updated_at"}).
			AddRow(2, 1, 200, "c2", 0, 0, 5, 0, time.Now(), time.Now()))
	list, err := repo.ListComments(context.Background(), 1, 2, 10, "hot")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(2), list[0].ID)
}

func TestDynamicRepository_ListComments_QueryError(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`FROM dynamic_comments`).WillReturnError(errors.New("boom"))
	_, err := repo.ListComments(context.Background(), 1, 1, 20, "new")
	assert.Error(t, err)
}

func TestDynamicRepository_ListComments_ScanError(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`FROM dynamic_comments`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dynamic_id", "user_id", "content", "parent_id", "reply_user_id", "like_count", "status", "created_at", "updated_at"}).
			AddRow("bad", 1, 100, "c", 0, 0, 2, 0, time.Now(), time.Now()))
	_, err := repo.ListComments(context.Background(), 1, 1, 20, "new")
	assert.Error(t, err)
}

func TestDynamicRepository_ListReplies(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`FROM dynamic_comments WHERE parent_id = \$1 AND status = 0 ORDER BY created_at ASC`).
		WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dynamic_id", "user_id", "content", "parent_id", "reply_user_id", "like_count", "status", "created_at", "updated_at"}).
			AddRow(3, 1, 300, "reply", 1, 100, 1, 0, time.Now(), time.Now()))
	list, err := repo.ListReplies(context.Background(), 1, 1, 20)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(3), list[0].ID)
}

func TestDynamicRepository_ListReplies_QueryError(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectQuery(`FROM dynamic_comments WHERE parent_id`).WillReturnError(errors.New("boom"))
	_, err := repo.ListReplies(context.Background(), 1, 1, 20)
	assert.Error(t, err)
}

func TestDynamicRepository_DeleteComment(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectExec(`UPDATE dynamic_comments SET status = 1`).WithArgs(int64(1), int64(100)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.DeleteComment(context.Background(), 1, 100))
}

func TestDynamicRepository_DeleteComment_Error(t *testing.T) {
	repo, mock := newMockDynamicRepo(t)
	mock.ExpectExec(`UPDATE dynamic_comments SET status = 1`).WillReturnError(errors.New("boom"))
	assert.Error(t, repo.DeleteComment(context.Background(), 1, 100))
}
