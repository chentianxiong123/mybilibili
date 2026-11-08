package live

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestService_UpdateRoom_CoverZeroUser 钉住 FINDING-025：
// handleCoverUpload 硬编码 userID=0 调 UpdateRoom（live_http_handler.go:270
// `h.svc.UpdateRoom(ctx, roomID, 0, "", coverURL, "")`），而 UpdateRoom 所有权
// 检查 `room.UserID != 0` 恒失败（host 永不为 0）→ 返回权限错误 → 但 handler
// 【忽略该错误】仍返回 {"status":"uploaded"}。因此直播封面【永不更新】。
func TestService_UpdateRoom_CoverZeroUser(t *testing.T) {
	svc, mock := newTestService(t)

	now := time.Now()
	// host user 5001（非 0）
	mock.ExpectQuery(`SELECT .+ FROM live_rooms WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_name", "user_id", "stream_key", "cover_url", "category", "status", "viewer_count", "created_at", "updated_at",
		}).AddRow(1, "测试", 5001, "skey", "", "", RoomStatusIdle, 0, now, now))

	// 与 handleCoverUpload 相同：userID=0
	err := svc.UpdateRoom(t.Context(), 1, 0, "", "/uploads/live-covers/x.jpg", "")
	assert.Error(t, err, "FINDING-025: userID=0 恒被所有权检查拒绝")
	assert.Contains(t, err.Error(), "only host can update room")
	require.NoError(t, mock.ExpectationsWereMet())
}
