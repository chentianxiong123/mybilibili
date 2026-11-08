package subtitle

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestImportSRT_NoAuth 钉住 FINDING-034：
// POST /api/v1/subtitle/import-srt 在【无 X-User-Id / Authorization】下仍成功
// 上传字幕（Upload 以 GetUserIDFromHeader=0 调用）—— 未认证者可向任意视频
// 注入字幕。handleImport（subtitle_handler.go:264）无 RequireUser。
func TestImportSRT_NoAuth(t *testing.T) {
	h, _, _ := newTestSubtitleHandler(t)
	srt := "1\n00:00:01,000 --> 00:00:02,000\nhello\n"
	mux := http.NewServeMux()
	h.Register(mux)

	// 无 X-User-Id —— 模拟未认证请求
	req := httptest.NewRequest(http.MethodPost, "/api/v1/subtitle/import-srt", strings.NewReader(`{"video_id":5,"srt":"`+strings.ReplaceAll(srt, "\n", "\\n")+`"}`))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "FINDING-034: 无鉴权 import-srt 仍返回成功（应被拒绝）")
	// uploaded_by=0 证明 GetUserIDFromHeader 无鉴权时为 0（FINDING-018 类投毒）
	assert.Contains(t, w.Body.String(), `"uploaded_by":0`)
	assert.Contains(t, w.Body.String(), `"id":"subtitle-1"`)
}
