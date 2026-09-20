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

func newMockService(t *testing.T) (*DanmakuService, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	repo := NewDanmakuRepository(db)
	broadcaster := NewDanmakuBroadcaster()
	return NewDanmakuService(repo, broadcaster), mock
}

func TestSvcSend(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery(`INSERT INTO danmaku`).
		WithArgs(int64(10), int64(20), int64(30), "hello", 1.5, "#ffffff", int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))

	event, err := svc.Send(context.Background(), 10, 20, 30, "hello", 1.5, "#ffffff", 1)
	require.NoError(t, err)
	assert.Equal(t, int64(42), event.ID)
	assert.Equal(t, int64(10), event.VideoID)
	assert.Equal(t, int64(30), event.UserID)
	assert.Equal(t, "hello", event.Content)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSvcSend_DBError(t *testing.T) {
	svc, mock := newMockService(t)
	mock.ExpectQuery(`INSERT INTO danmaku`).WillReturnError(errors.New("db down"))
	_, err := svc.Send(context.Background(), 10, 20, 30, "hello", 1.5, "#ffffff", 1)
	assert.Error(t, err)
}

func TestSvcSend_Broadcast(t *testing.T) {
	svc, mock := newMockService(t)

	ch := svc.broadcaster.Subscribe(10)

	mock.ExpectQuery(`INSERT INTO danmaku`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	event, err := svc.Send(context.Background(), 10, 0, 30, "hi", 2.0, "#000", 0)
	require.NoError(t, err)

	received := <-ch
	assert.Equal(t, event.ID, received.ID)
	assert.Equal(t, "hi", received.Content)
}

func TestSvcListByVideo(t *testing.T) {
	svc, mock := newMockService(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, video_id, manuscript_id, user_id, content, time, color, mode, created_at`).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at"}).
			AddRow(1, 10, 20, 30, "a", 1.0, "#fff", int32(0), now).
			AddRow(2, 10, 20, 31, "b", 2.0, "#000", int32(1), now))

	events, err := svc.ListByVideo(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, int64(1), events[0].ID)
	assert.Equal(t, "b", events[1].Content)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSvcListByVideo_Empty(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery(`SELECT id, video_id, manuscript_id`).
		WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at"}))

	events, err := svc.ListByVideo(context.Background(), 999)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestSvcListByTimeRange(t *testing.T) {
	svc, mock := newMockService(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, video_id, manuscript_id, user_id, content, time, color, mode, created_at`).
		WithArgs(int64(10), 1.0, 5.0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at"}).
			AddRow(1, 10, 20, 30, "a", 2.0, "#fff", int32(0), now))

	events, err := svc.ListByTimeRange(context.Background(), 10, 1.0, 5.0)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, int64(1), events[0].ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSvcDelete(t *testing.T) {
	svc, mock := newMockService(t)
	mock.ExpectExec(`DELETE FROM danmaku WHERE id = \$1 AND user_id = \$2`).
		WithArgs(int64(1), int64(30)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.Delete(context.Background(), 1, 30))
}

func TestSvcCountByVideo(t *testing.T) {
	svc, mock := newMockService(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku WHERE video_id`).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))
	count, err := svc.CountByVideo(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, int64(42), count)
}

func TestSvcCountByManuscriptIDs(t *testing.T) {
	svc, mock := newMockService(t)
	ids := []int64{1, 2}

	mock.ExpectQuery(`SELECT manuscript_id, COUNT\(\*\) FROM danmaku WHERE manuscript_id = ANY`).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "count"}).
			AddRow(1, 5).
			AddRow(2, 8))

	result, err := svc.CountByManuscriptIDs(context.Background(), ids)
	require.NoError(t, err)
	assert.Equal(t, int64(5), result[1])
	assert.Equal(t, int64(8), result[2])
}

func TestSvcTrend(t *testing.T) {
	svc, mock := newMockService(t)
	ids := []int64{1}

	mock.ExpectQuery(`SELECT TO_CHAR\(created_at`).
		WillReturnRows(sqlmock.NewRows([]string{"day", "count"}).
			AddRow("2025-01-01", 10))

	result, err := svc.Trend(context.Background(), ids, "2025-01-01", "2025-01-02")
	require.NoError(t, err)
	assert.Equal(t, 10, result["2025-01-01"])
}

func TestSvcTrend_Empty(t *testing.T) {
	svc, _ := newMockService(t)
	result, err := svc.Trend(context.Background(), nil, "2025-01-01", "2025-01-02")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestSvcCreatorList(t *testing.T) {
	svc, mock := newMockService(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT danmaku.id, danmaku.video_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at"}).
			AddRow(1, 10, 20, 30, "a", 1.0, "#fff", int32(0), now).
			AddRow(2, 11, 20, 30, "b", 2.0, "#000", int32(1), now))

	events, total, err := svc.CreatorList(context.Background(), 30, 0, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, events, 2)
}

func TestSvcCreatorList_Empty(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT danmaku.id, danmaku.video_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "user_id", "content", "time", "color", "mode", "created_at"}))

	events, total, err := svc.CreatorList(context.Background(), 30, 0, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, events)
}

func TestSvcCreatorDelete(t *testing.T) {
	svc, mock := newMockService(t)
	mock.ExpectExec(`DELETE FROM danmaku`).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.CreatorDelete(context.Background(), 1, 30))
}

func TestSvcBroadcaster(t *testing.T) {
	svc, _ := newMockService(t)
	assert.NotNil(t, svc.Broadcaster())
}

func TestSvcBroadcaster_SubscribeAndBroadcast(t *testing.T) {
	svc, mock := newMockService(t)
	_ = mock

	b := svc.Broadcaster()
	ch := b.Subscribe(5)

	mock.ExpectQuery(`INSERT INTO danmaku`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	_, err := svc.Send(context.Background(), 5, 0, 1, "test", 0.0, "#fff", 0)
	require.NoError(t, err)

	event := <-ch
	assert.Equal(t, "test", event.Content)
}
