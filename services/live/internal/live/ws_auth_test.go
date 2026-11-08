package live

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUserIDFromQuery_QueryWinsOverHeader 钉住 FINDING-028：
// userIDFromQuery 从 query ?user_id= 读身份（live_ws.go:114），【优先于】
// IdentityMiddleware 注入的 X-User-Id 头。若客户端带合法 JWT（X-User-Id=5）
// 但 query 伪造 user_id=9999，WS 连接身份=9999（伪造成 victim）。
// 修复后应改为读 X-User-Id 头（或校验 query==JWT 身份）。
func TestUserIDFromQuery_QueryWinsOverHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/live/ws?user_id=9999&room_id=1", nil)
	// 模拟 IdentityMiddleware 已校验 JWT 并注入真实身份 5
	req.Header.Set("X-User-Id", "5")

	got := userIDFromQuery(req)
	assert.Equal(t, int64(9999), got, "FINDING-028: WS 信任 query user_id=9999 而非 JWT 身份 5（应使用头）")
}

// TestUserIDFromQuery_NoQuery 无 query user_id 时返回 0（未认证）。
func TestUserIDFromQuery_NoQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/live/ws?room_id=1", nil)
	assert.Equal(t, int64(0), userIDFromQuery(req))
}
