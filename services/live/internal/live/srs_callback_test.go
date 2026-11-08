package live

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleSRSCallback_OnUnpublishIgnored 钉住 FINDING-027：
// HandleSRSCallback 对 action != "on_publish"（如 on_unpublish/on_stop）直接
// return nil，【不执行任何 DB 查询/状态更新】→ 房间停留在 "live" 永不重置。
// 这证明流结束回调被忽略。
func TestHandleSRSCallback_OnUnpublishIgnored(t *testing.T) {
	svc, mock := newTestService(t)

	// 不设置任何 DB 期望——若 on_unpublish 被处理会触发 GetRoomByStreamKey/UpdateRoomStatus
	err := svc.HandleSRSCallback(t.Context(), "on_unpublish", "some-stream-key")
	assert.NoError(t, err)
	// 零期望：无 DB 调用（FINDING-027：on_unpublish 被忽略）
	require.NoError(t, mock.ExpectationsWereMet(), "FINDING-027: on_unpublish 应触发状态重置，但被忽略")
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
