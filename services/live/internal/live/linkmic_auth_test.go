package live

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLinkmicAccept_NoAuth 钉住 FINDING-026（已修复）：
// POST /api/v1/live/linkmic/accept/{id} 无用户身份 → 401。
func TestLinkmicAccept_NoAuth(t *testing.T) {
	h, _ := newTestLinkmicHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/linkmic/accept/1", strings.NewReader(""))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "FINDING-026 修复：无鉴权 accept 应 401")
}

// TestLinkmicToggleAudio_NoAuth 钉住 FINDING-026（已修复）：
// POST /api/v1/live/linkmic/toggle-audio/{id} 无用户身份 → 401。
func TestLinkmicToggleAudio_NoAuth(t *testing.T) {
	h, _ := newTestLinkmicHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/live/linkmic/toggle-audio/1", strings.NewReader(""))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "FINDING-026 修复：无鉴权 toggle-audio 应 401")
}
