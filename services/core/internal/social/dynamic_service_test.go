package social

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockDynamicSvc(t *testing.T) (*DynamicService, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewDynamicService(NewDynamicRepository(db)), mock
}

func TestDynamicService_Publish(t *testing.T) {
	svc, mock := newMockDynamicSvc(t)
	now := time.Now()
	mock.ExpectQuery(`INSERT INTO user_dynamics`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`SELECT id, user_id, content, dynamic_type`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}).
			AddRow(1, 100, "hello", 1, "", 0, 0, 0, 0, 0, now))

	d, err := svc.Publish(context.Background(), 100, "hello", 1, "", 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), d.ID)
}

func TestDynamicService_Publish_CreateError(t *testing.T) {
	svc, mock := newMockDynamicSvc(t)
	mock.ExpectQuery(`INSERT INTO user_dynamics`).WillReturnError(errors.New("boom"))
	_, err := svc.Publish(context.Background(), 100, "hello", 1, "", 0)
	assert.Error(t, err)
}

func TestDynamicService_GetByID(t *testing.T) {
	svc, mock := newMockDynamicSvc(t)
	mock.ExpectQuery(`SELECT id, user_id, content, dynamic_type`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}).
			AddRow(1, 100, "hello", 1, "", 0, 0, 0, 0, 0, time.Now()))
	d, err := svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(100), d.UserID)
}

func TestDynamicService_ListByUser(t *testing.T) {
	svc, mock := newMockDynamicSvc(t)
	mock.ExpectQuery(`FROM user_dynamics WHERE user_id = \$1 AND status = 0`).
		WithArgs(int64(100), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}).
			AddRow(1, 100, "a", 1, "", 0, 0, 0, 0, 0, time.Now()))
	list, err := svc.ListByUser(context.Background(), 100, 1, 20)
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestDynamicService_ListFollowing(t *testing.T) {
	svc, mock := newMockDynamicSvc(t)
	mock.ExpectQuery(`FROM user_dynamics d WHERE d.status = 0`).
		WithArgs(int64(100), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}))
	list, err := svc.ListFollowing(context.Background(), 100, 1, 20)
	require.NoError(t, err)
	assert.Len(t, list, 0)
}

func TestDynamicService_ListAll(t *testing.T) {
	svc, mock := newMockDynamicSvc(t)
	mock.ExpectQuery(`FROM user_dynamics WHERE status = 0 ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(int32(20), int32(20)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "dynamic_type", "image_url", "ref_manuscript_id", "like_count", "comment_count", "share_count", "status", "created_at"}).
			AddRow(2, 200, "b", 1, "", 0, 0, 0, 0, 0, time.Now()))
	list, err := svc.ListAll(context.Background(), 2, 20)
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestDynamicService_Delete(t *testing.T) {
	svc, mock := newMockDynamicSvc(t)
	mock.ExpectExec(`UPDATE user_dynamics SET status = 1`).WithArgs(int64(1), int64(100)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.Delete(context.Background(), 1, 100))
}

func TestDynamicService_LikeUnlikeShare(t *testing.T) {
	svc, mock := newMockDynamicSvc(t)
	mock.ExpectExec(`like_count = GREATEST`).WithArgs(int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.Like(context.Background(), 1, 100))

	mock.ExpectExec(`like_count = GREATEST`).WithArgs(int64(-1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.Unlike(context.Background(), 1, 100))

	mock.ExpectExec(`share_count = GREATEST`).WithArgs(int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.ShareDynamic(context.Background(), 1, 100))

	mock.ExpectExec(`comment_count = GREATEST`).WithArgs(int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.IncrCommentCount(context.Background(), 1, 1))
}

func TestDynamicService_IsLiked(t *testing.T) {
	svc, mock := newMockDynamicSvc(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM dynamic_likes`).WithArgs(int64(1), int64(100)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	ok, err := svc.IsLiked(context.Background(), 1, 100)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestDynamicService_AddComment(t *testing.T) {
	svc, mock := newMockDynamicSvc(t)
	mock.ExpectQuery(`INSERT INTO dynamic_comments`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	mock.ExpectExec(`comment_count = GREATEST`).WithArgs(int64(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	dc, err := svc.AddComment(context.Background(), 1, 100, "nice", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, "nice", dc.Content)
	assert.Equal(t, int64(100), dc.UserID)
}

func TestDynamicService_AddComment_CreateError(t *testing.T) {
	svc, mock := newMockDynamicSvc(t)
	mock.ExpectQuery(`INSERT INTO dynamic_comments`).WillReturnError(errors.New("boom"))
	_, err := svc.AddComment(context.Background(), 1, 100, "nice", 0, 0)
	assert.Error(t, err)
}

func TestDynamicService_ListCommentsReplies(t *testing.T) {
	svc, mock := newMockDynamicSvc(t)
	mock.ExpectQuery(`ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dynamic_id", "user_id", "content", "parent_id", "reply_user_id", "like_count", "status", "created_at", "updated_at"}).
			AddRow(1, 1, 100, "c", 0, 0, 1, 0, time.Now(), time.Now()))
	list, err := svc.ListComments(context.Background(), 1, 1, 20, "new")
	require.NoError(t, err)
	require.Len(t, list, 1)

	mock.ExpectQuery(`FROM dynamic_comments WHERE parent_id`).WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dynamic_id", "user_id", "content", "parent_id", "reply_user_id", "like_count", "status", "created_at", "updated_at"}))
	replies, err := svc.ListReplies(context.Background(), 1, 1, 20)
	require.NoError(t, err)
	assert.Len(t, replies, 0)
}

func newMockCollectionRepo(t *testing.T) (*CollectionRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewCollectionRepository(db), mock
}

func TestCollectionRepository_Create(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectQuery(`INSERT INTO manuscript_collections`).
		WithArgs("title", "desc", "cover", int64(100), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	id, err := repo.Create(context.Background(), 100, "title", "desc", "cover", 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
}

func TestCollectionRepository_Create_Error(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectQuery(`INSERT INTO manuscript_collections`).WillReturnError(errors.New("boom"))
	_, err := repo.Create(context.Background(), 100, "title", "desc", "cover", 0)
	assert.Error(t, err)
}

func TestCollectionRepository_GetByID(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT id, title, description, cover_url`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "cover_url", "user_id", "manuscript_count", "view_count", "status", "created_at", "updated_at"}).
			AddRow(1, "title", "desc", "cover", 100, 3, 9, 0, now, now))
	m, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), m["id"])
	assert.Equal(t, "title", m["title"])
	assert.Equal(t, int64(3), m["manuscript_count"])
}

func TestCollectionRepository_GetByID_Nulls(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectQuery(`SELECT id, title, description, cover_url`).WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "cover_url", "user_id", "manuscript_count", "view_count", "status", "created_at", "updated_at"}).
			AddRow(2, "t", nil, nil, 100, nil, nil, nil, nil, nil))
	m, err := repo.GetByID(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), m["id"])
}

func TestCollectionRepository_GetByID_Error(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectQuery(`SELECT id, title, description, cover_url`).WithArgs(int64(9)).WillReturnError(errors.New("boom"))
	_, err := repo.GetByID(context.Background(), 9)
	assert.Error(t, err)
}

func TestCollectionRepository_ListByUser(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	now := time.Now()
	mock.ExpectQuery(`FROM manuscript_collections WHERE user_id = \$1`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "cover_url", "user_id", "manuscript_count", "view_count", "status", "created_at", "updated_at"}).
			AddRow(1, "t1", "d1", "c1", 100, 1, 2, 0, now, now).
			AddRow(2, "t2", "d2", "c2", 100, 0, 0, 1, now, now))
	list, err := repo.ListByUser(context.Background(), 100)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "t2", list[1]["title"])
}

func TestCollectionRepository_ListByUser_QueryError(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectQuery(`FROM manuscript_collections WHERE user_id = \$1`).WillReturnError(errors.New("boom"))
	_, err := repo.ListByUser(context.Background(), 100)
	assert.Error(t, err)
}

func TestCollectionRepository_ListByUser_ScanError(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectQuery(`FROM manuscript_collections WHERE user_id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "cover_url", "user_id", "manuscript_count", "view_count", "status", "created_at", "updated_at"}).
			AddRow("bad", "t", nil, nil, 100, nil, nil, nil, nil, nil))
	_, err := repo.ListByUser(context.Background(), 100)
	assert.Error(t, err)
}

func TestCollectionRepository_UpdateDelete(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectExec(`UPDATE manuscript_collections SET title=\$1, description=\$2, status=\$3, updated_at=NOW\(\)`).
		WithArgs("nt", "nd", int32(0), int64(1), int64(100)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Update(context.Background(), 1, 100, "nt", "nd", 0))

	mock.ExpectExec(`DELETE FROM manuscript_collections WHERE id=\$1 AND user_id=\$2`).
		WithArgs(int64(1), int64(100)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Delete(context.Background(), 1, 100))
}

func TestCollectionRepository_Update_Error(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectExec(`UPDATE manuscript_collections`).WillReturnError(errors.New("boom"))
	assert.Error(t, repo.Update(context.Background(), 1, 100, "nt", "nd", 0))
}

func TestCollectionRepository_AddManuscript(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectExec(`INSERT INTO manuscript_collection_relations`).
		WithArgs(int64(9), int64(1), int64(0)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscript_collections SET manuscript_count = \(SELECT COUNT`).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.AddManuscript(context.Background(), 1, 9, 0))
}

func TestCollectionRepository_AddManuscript_InsertError(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectExec(`INSERT INTO manuscript_collection_relations`).WillReturnError(errors.New("boom"))
	assert.Error(t, repo.AddManuscript(context.Background(), 1, 9, 0))
}

func TestCollectionRepository_AddManuscript_UpdateError(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectExec(`INSERT INTO manuscript_collection_relations`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscript_collections SET manuscript_count`).WillReturnError(errors.New("boom"))
	assert.Error(t, repo.AddManuscript(context.Background(), 1, 9, 0))
}

func TestCollectionRepository_RemoveManuscript(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectExec(`DELETE FROM manuscript_collection_relations WHERE collection_id=\$1 AND manuscript_id=\$2`).
		WithArgs(int64(1), int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscript_collections SET manuscript_count`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.RemoveManuscript(context.Background(), 1, 9))
}

func TestCollectionRepository_RemoveManuscript_Error(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectExec(`DELETE FROM manuscript_collection_relations`).WillReturnError(errors.New("boom"))
	assert.Error(t, repo.RemoveManuscript(context.Background(), 1, 9))
}

func TestCollectionRepository_ListManuscripts(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectQuery(`SELECT manuscript_id FROM manuscript_collection_relations`).
		WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id"}).AddRow(9).AddRow(10))
	ids, err := repo.ListManuscripts(context.Background(), 1, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, []int64{9, 10}, ids)
}

func TestCollectionRepository_ListManuscripts_QueryError(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectQuery(`SELECT manuscript_id FROM manuscript_collection_relations`).WillReturnError(errors.New("boom"))
	_, err := repo.ListManuscripts(context.Background(), 1, 1, 20)
	assert.Error(t, err)
}

func TestCollectionRepository_ListManuscriptsDetail(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	now := time.Now()
	mock.ExpectQuery(`JOIN manuscripts m ON m.id = cr.manuscript_id`).
		WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "cover_url", "view_count", "upload_time", "duration", "user_id", "status"}).
			AddRow(9, "title", "cover", 100, now, "00:10", 200, 0).
			AddRow(10, "title2", "cover2", 1, now, "00:05", 201, 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscript_collection_relations`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	list, total, err := repo.ListManuscriptsDetail(context.Background(), 1, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, list, 2)
	assert.Equal(t, "title", list[0]["title"])
}

func TestCollectionRepository_ListManuscriptsDetail_QueryError(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	mock.ExpectQuery(`JOIN manuscripts m ON m.id = cr.manuscript_id`).WillReturnError(errors.New("boom"))
	_, _, err := repo.ListManuscriptsDetail(context.Background(), 1, 1, 20)
	assert.Error(t, err)
}

func TestCollectionService_AllMethods(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	svc := NewCollectionService(repo)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO manuscript_collections`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`SELECT id, title, description, cover_url`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "cover_url", "user_id", "manuscript_count", "view_count", "status", "created_at", "updated_at"}).
			AddRow(1, "t", "d", "c", 100, 0, 0, 0, now, now))
	m, err := svc.Create(context.Background(), 100, "t", "d", "c", 0)
	require.NoError(t, err)
	assert.Equal(t, "t", m["title"])

	mock.ExpectQuery(`SELECT id, title, description, cover_url`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "cover_url", "user_id", "manuscript_count", "view_count", "status", "created_at", "updated_at"}).
			AddRow(1, "t", "d", "c", 100, 0, 0, 0, now, now))
	m, err = svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(100), m["user_id"])

	mock.ExpectQuery(`FROM manuscript_collections WHERE user_id`).WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "cover_url", "user_id", "manuscript_count", "view_count", "status", "created_at", "updated_at"}))
	list, err := svc.ListByUser(context.Background(), 100)
	require.NoError(t, err)
	assert.Len(t, list, 0)

	mock.ExpectExec(`UPDATE manuscript_collections SET title=\$1`).WithArgs("nt", "nd", int32(1), int64(1), int64(100)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.Update(context.Background(), 1, 100, "nt", "nd", 1))

	mock.ExpectExec(`DELETE FROM manuscript_collections`).WithArgs(int64(1), int64(100)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.Delete(context.Background(), 1, 100))

	mock.ExpectExec(`INSERT INTO manuscript_collection_relations`).WithArgs(int64(9), int64(1), int64(0)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscript_collections SET manuscript_count`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.AddManuscript(context.Background(), 1, 9, 100))

	mock.ExpectExec(`DELETE FROM manuscript_collection_relations`).WithArgs(int64(1), int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscript_collections SET manuscript_count`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.RemoveManuscript(context.Background(), 1, 9, 100))

	mock.ExpectQuery(`SELECT manuscript_id FROM manuscript_collection_relations`).WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id"}).AddRow(9))
	ids, err := svc.ListManuscripts(context.Background(), 1, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, []int64{9}, ids)

	mock.ExpectQuery(`JOIN manuscripts m`).WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "cover_url", "view_count", "upload_time", "duration", "user_id", "status"}).
			AddRow(9, "t", "c", 1, now, "00:10", 200, 0))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscript_collection_relations`).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	detail, total, err := svc.ListManuscriptsDetail(context.Background(), 1, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, detail, 1)
}

func TestCollectionService_Create_Error(t *testing.T) {
	repo, mock := newMockCollectionRepo(t)
	svc := NewCollectionService(repo)
	mock.ExpectQuery(`INSERT INTO manuscript_collections`).WillReturnError(errors.New("boom"))
	_, err := svc.Create(context.Background(), 100, "t", "d", "c", 0)
	assert.Error(t, err)
}

func newMockShareRepo(t *testing.T) (*ShareRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewShareRepository(db), mock
}

func TestShareRepository_Record(t *testing.T) {
	repo, mock := newMockShareRepo(t)
	mock.ExpectExec(`INSERT INTO shares`).WithArgs(int64(100), int64(9), "wechat", "1.2.3.4").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET share_count = share_count \+ 1`).WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Record(context.Background(), 100, 9, "wechat", "1.2.3.4"))
}

func TestShareRepository_Record_InsertError(t *testing.T) {
	repo, mock := newMockShareRepo(t)
	mock.ExpectExec(`INSERT INTO shares`).WillReturnError(errors.New("boom"))
	assert.Error(t, repo.Record(context.Background(), 100, 9, "wechat", "1.2.3.4"))
}

func TestShareRepository_Record_UpdateError(t *testing.T) {
	repo, mock := newMockShareRepo(t)
	mock.ExpectExec(`INSERT INTO shares`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET share_count`).WillReturnError(errors.New("boom"))
	assert.Error(t, repo.Record(context.Background(), 100, 9, "wechat", "1.2.3.4"))
}
