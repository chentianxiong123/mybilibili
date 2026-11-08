package subtitle

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestImportSRT_NoAuth 钉住 FINDING-034（已修复）：
// POST /api/v1/subtitle/import-srt 无用户身份 → 401，未认证者不可注入字幕。
func TestImportSRT_NoAuth(t *testing.T) {
	h, _, _ := newTestSubtitleHandler(t)
	srt := "1\n00:00:01,000 --> 00:00:02,000\nhello\n"
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/subtitle/import-srt", strings.NewReader(`{"video_id":5,"srt":"`+strings.ReplaceAll(srt, "\n", "\\n")+`"}`))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "FINDING-034 修复：无鉴权 import-srt 应 401")
}
