package live

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLinkmicAccept_NoAuth 钉住 FINDING-026：
// POST /api/v1/live/linkmic/accept/{id} 在【无 X-User-Id / Authorization】下
// 仍执行 UPDATE live_linkmic SET status（接受任意连麦）—— 未认证者可强制连接
// 任意连麦。handleLinkmic accept 分支不传 userID、无鉴权；LinkmicService.Accept
// 仅 UpdateStatus（linkmic.go:132-134）无所有权检查。
func TestLinkmicAccept_NoAuth(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	mock.ExpectExec(`UPDATE live_linkmic SET status`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	// 无 X-User-Id —— 模拟未认证请求
	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/linkmic/accept/1", strings.NewReader(""))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"status":"ok"`)
	require.NoError(t, mock.ExpectationsWereMet(), "FINDING-026: 无鉴权 accept 仍执行 UPDATE（应被拒绝）")
}

// TestLinkmicToggleAudio_NoAuth 钉住 FINDING-026：
// POST /api/v1/live/linkmic/toggle-audio/{id} 在无鉴权下仍执行音视频开关。
func TestLinkmicToggleAudio_NoAuth(t *testing.T) {
	h, mock := newTestLinkmicHandler(t)
	mock.ExpectQuery(`UPDATE live_linkmic SET audio_enabled =`).
		WillReturnRows(sqlmock.NewRows([]string{"audio_enabled"}).AddRow(1))

	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/linkmic/toggle-audio/1", strings.NewReader(""))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet(), "FINDING-026: 无鉴权 toggle-audio 仍执行（应被拒绝）")
}
