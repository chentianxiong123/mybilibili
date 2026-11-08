package support

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionProcess_NoAuth 钉住 FINDING-031：
// PUT /api/v1/operation/internal/tickets/session/{sessionID} 在【无鉴权】下仍执行
// UPDATE support_tickets SET admin_reply=... status='processed' —— 未认证者可伪造
// admin 回复并"处理"任意工单。handleSessionProcess 无鉴权，路径 /internal/ 不受
// AdminPathGuard 保护。
func TestSessionProcess_NoAuth(t *testing.T) {
	h, mock := newMockSupport(t)
	mock.ExpectExec(`UPDATE support_tickets SET admin_reply = \$2, status = 'processed', processed_at = NOW\(\) WHERE session_id = \$1`).
		WithArgs(int64(9), "hacked reply").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	// 无 X-Admin-Id / X-User-Id —— 模拟未认证请求
	req := httptest.NewRequest(http.MethodPut, "/api/v1/operation/internal/tickets/session/9", strings.NewReader(`{"adminReply":"hacked reply"}`))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet(), "FINDING-031: 无鉴权仍执行 UPDATE support_tickets（应被拒绝）")
}

// TestCustomerSession_SpoofUserID 钉住 FINDING-031：
// POST /api/v1/operation/internal/tickets/customer-session 信任 body userId ——
// 客户端可伪造成任意用户建单（此处 victim=777）。
func TestCustomerSession_SpoofUserID(t *testing.T) {
	h, mock := newMockSupport(t)
	mock.ExpectQuery(`INSERT INTO support_tickets`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, time.Now()))

	mux := http.NewServeMux()
	h.Register(mux)
	// 无鉴权头，仅 body userId=777（伪装 victim）
	req := httptest.NewRequest(http.MethodPost, "/api/v1/operation/internal/tickets/customer-session", strings.NewReader(`{"userId":777,"title":"t","content":"c"}`))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"ticketId"`)
	require.NoError(t, mock.ExpectationsWereMet(), "FINDING-031: body userId 被信任（应使用认证身份）")
}
