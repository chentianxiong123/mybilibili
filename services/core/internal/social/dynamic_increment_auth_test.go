package social

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDynamicCommentIncrement_NoAuth 钉住 FINDING-030：
// POST /api/v1/dynamic/comment/increment/{id} 在【无 X-User-Id / Authorization】下
// 仍执行 UPDATE user_dynamics SET comment_count=comment_count+1 —— 未认证者可
// 无限膨胀任意动态的评论数。handleDynamicCommentIncrement 无鉴权。
func TestDynamicCommentIncrement_NoAuth(t *testing.T) {
	h, mock := newMockSocialHandler(t)
	mock.ExpectExec(`UPDATE user_dynamics SET comment_count = GREATEST\(comment_count \+ \$1, 0\) WHERE id = \$2`).
		WithArgs(int64(1), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := muxForSocial(h)
	// 无 X-User-Id —— 模拟未认证请求
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dynamic/comment/increment/7", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet(), "FINDING-030: 无鉴权仍执行 UPDATE comment_count（应被拒绝）")
}
