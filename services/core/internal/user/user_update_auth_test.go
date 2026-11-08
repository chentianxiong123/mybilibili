package user

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserUpdate_NoAuth 钉住 FINDING-029：
// PUT /api/v1/user/{id} 在【无 X-User-Id / Authorization】下仍执行
// UPDATE users SET nickname=... WHERE id=<victim> —— 未认证者可覆写任意用户资料
// （昵称/公告/简介/性别/封面）。handleUserByID 无 RequireUser/GetUserIDFromHeader/id==caller。
func TestUserUpdate_NoAuth(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE users SET signature = \$1, nickname = \$2, updated_at = NOW\(\) WHERE id = \$3`).
		WithArgs("x", "hacked", int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 无 X-User-Id / Authorization —— 模拟未认证请求
	req := httptest.NewRequest(http.MethodPut, "/api/v1/user/5", strings.NewReader(`{"nickname":"hacked","signature":"x"}`))
	w := httptest.NewRecorder()
	h.handleUserByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet(), "FINDING-029: 无鉴权仍执行 UPDATE users（应被拒绝）")
}
