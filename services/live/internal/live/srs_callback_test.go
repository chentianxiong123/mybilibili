package live

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleSRSCallback_OnUnpublishIgnored 验证 FINDING-027 修复：
// on_unpublish 现应查询 stream key 并重置房间为 idle（不再被忽略）。
func TestHandleSRSCallback_OnUnpublishIgnored(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE stream_key = \$1`).
		WithArgs("some-stream-key").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "房", 5001, "some-stream-key", "", "", RoomStatusLive, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET status=`).
		WithArgs(RoomStatusIdle, int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.HandleSRSCallback(t.Context(), "on_unpublish", "some-stream-key")
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet(), "FINDING-027 修复: on_unpublish 应触发状态重置")
}

// TestHandleSRSCallback_OnPublishSetsLive 对照：on_publish 应查询 stream key 并置 live。
func TestHandleSRSCallback_OnPublishSetsLive(t *testing.T) {
	svc, mock := newTestService(t)

	now := time.Now()
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE stream_key = \$1`).
		WithArgs("skey").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "房", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))
	mock.ExpectExec(`UPDATE live_rooms SET status=`).
		WithArgs(RoomStatusLive, int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.HandleSRSCallback(t.Context(), "on_publish", "skey")
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
