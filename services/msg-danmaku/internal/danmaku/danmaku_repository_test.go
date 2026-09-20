package danmaku

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockRepo(t *testing.T) (*DanmakuRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewDanmakuRepository(db), mock
}

func TestCreate(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`INSERT INTO danmaku`).
		WithArgs(int64(10), int64(20), int64(30), "test", 1.5, "#ffffff", int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	d := &Danmaku{VideoID: 10, ManuscriptID: 20, UserID: 30, Content: "test", Time: 1.5, Color: "#ffffff", Mode: 1}
	id, err := repo.Create(context.Background(), d)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_Error(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`INSERT INTO danmaku`).WillReturnError(errors.New("db down"))
	_, err := repo.Create(context.Background(), &Danmaku{})
	assert.Error(t, err)
}

func TestListByVideo(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, video_id, manuscript_id, user_id, content, time, color, mode, created_at`).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at"}).
			AddRow(1, 10, 20, 30, "a", 1.0, "#fff", int32(0), now).
			AddRow(2, 10, 20, 31, "b", 2.0, "#000", int32(1), now))

	list, err := repo.ListByVideo(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, int64(1), list[0].ID)
	assert.Equal(t, "b", list[1].Content)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListByVideo_Empty(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id, video_id, manuscript_id`).
		WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at"}))

	list, err := repo.ListByVideo(context.Background(), 999)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestListByTimeRange(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, video_id, manuscript_id, user_id, content, time, color, mode, created_at`).
		WithArgs(int64(10), 1.0, 5.0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at"}).
			AddRow(1, 10, 20, 30, "a", 2.0, "#fff", int32(0), now))

	list, err := repo.ListByTimeRange(context.Background(), 10, 1.0, 5.0)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), list[0].ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListByTimeRange_Empty(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id, video_id, manuscript_id`).
		WithArgs(int64(10), 100.0, 200.0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at"}))

	list, err := repo.ListByTimeRange(context.Background(), 10, 100.0, 200.0)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestDelete(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectExec(`DELETE FROM danmaku WHERE id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(30)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Delete(context.Background(), 1, 30))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_NoRows(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectExec(`DELETE FROM danmaku WHERE id = \$1 AND user_id = \$2`).
		WithArgs(int64(999), int64(30)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	require.NoError(t, repo.Delete(context.Background(), 999, 30))
}

func TestCountByVideo(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku WHERE video_id`).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))
	count, err := repo.CountByVideo(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, int64(42), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCountByManuscriptIDs(t *testing.T) {
	repo, mock := newMockRepo(t)
	ids := []int64{1, 2, 3}

	mock.ExpectQuery(`SELECT manuscript_id, COUNT\(\*\) FROM danmaku WHERE manuscript_id = ANY`).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "count"}).
			AddRow(1, 5).
			AddRow(3, 10))

	result, err := repo.CountByManuscriptIDs(context.Background(), ids)
	require.NoError(t, err)
	assert.Equal(t, int64(5), result[1])
	assert.Equal(t, int64(0), result[2])
	assert.Equal(t, int64(10), result[3])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCountByManuscriptIDs_Empty(t *testing.T) {
	repo, _ := newMockRepo(t)
	result, err := repo.CountByManuscriptIDs(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestTrendByDate(t *testing.T) {
	repo, mock := newMockRepo(t)
	ids := []int64{1, 2}

	mock.ExpectQuery(`SELECT TO_CHAR\(created_at`).
		WillReturnRows(sqlmock.NewRows([]string{"day", "count"}).
			AddRow("2025-01-01", 10).
			AddRow("2025-01-02", 20))

	result, err := repo.TrendByDate(context.Background(), ids, "2025-01-01", "2025-01-02")
	require.NoError(t, err)
	assert.Equal(t, 10, result["2025-01-01"])
	assert.Equal(t, 20, result["2025-01-02"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTrendByDate_Empty(t *testing.T) {
	repo, _ := newMockRepo(t)
	result, err := repo.TrendByDate(context.Background(), nil, "2025-01-01", "2025-01-02")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListByCreator(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT danmaku.id, danmaku.video_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at"}).
			AddRow(1, 10, 20, 30, "a", 1.0, "#fff", int32(0), now))

	list, total, err := repo.ListByCreator(context.Background(), 30, 10, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), list[0].ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListByCreator_Empty(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT danmaku.id, danmaku.video_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at"}))

	list, total, err := repo.ListByCreator(context.Background(), 30, 0, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, list)
}

func TestDeleteByCreator(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectExec(`DELETE FROM danmaku`).
		WithArgs(int64(1), int64(30)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.DeleteByCreator(context.Background(), 1, 30))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertDailyMetric(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WithArgs(int64(10), int64(30), 1).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpsertDailyMetric(context.Background(), 10, 30, "danmaku_count", 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDanmakuBroadcaster_PubSub(t *testing.T) {
	b := NewDanmakuBroadcaster()
	ch := b.Subscribe(1)

	event := &DanmakuEvent{ID: 1, VideoID: 1, Content: "hello"}
	b.Broadcast(1, event)

	received := <-ch
	assert.Equal(t, int64(1), received.ID)
	assert.Equal(t, "hello", received.Content)
}

func TestDanmakuBroadcaster_BroadcastNoSubscriber(t *testing.T) {
	b := NewDanmakuBroadcaster()
	b.Broadcast(1, &DanmakuEvent{ID: 1, VideoID: 1})
}

func TestDanmakuBroadcaster_MultipleSubscribers(t *testing.T) {
	b := NewDanmakuBroadcaster()
	ch1 := b.Subscribe(1)
	ch2 := b.Subscribe(1)
	assert.Equal(t, ch1, ch2)
}

func TestDanmakuBroadcaster_DifferentVideos(t *testing.T) {
	b := NewDanmakuBroadcaster()
	ch1 := b.Subscribe(1)
	ch2 := b.Subscribe(2)

	b.Broadcast(1, &DanmakuEvent{ID: 1, VideoID: 1, Content: "a"})
	b.Broadcast(2, &DanmakuEvent{ID: 2, VideoID: 2, Content: "b"})

	assert.Equal(t, "a", (<-ch1).Content)
	assert.Equal(t, "b", (<-ch2).Content)
}
