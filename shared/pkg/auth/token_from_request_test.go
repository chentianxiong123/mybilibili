package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenFromRequest_AccessTokenQuery(t *testing.T) {
	// EventSource 无法设置请求头，只能靠 query 携带凭证
	req := httptest.NewRequest("GET", "/api/v1/video/process/admin/stream?access_token=abc123", nil)
	assert.Equal(t, "abc123", TokenFromRequest(req))
}

func TestTokenFromRequest_HeaderBeatsQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/x?access_token=from-query", nil)
	req.Header.Set("Authorization", "Bearer from-header")
	assert.Equal(t, "from-header", TokenFromRequest(req))
}

func TestTokenFromRequest_CookieBeatsQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/x?access_token=from-query", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "from-cookie"})
	assert.Equal(t, "from-cookie", TokenFromRequest(req))
}

func TestTokenFromRequest_NothingReturnsEmpty(t *testing.T) {
	req := httptest.NewRequest("GET", "/x", nil)
	assert.Equal(t, "", TokenFromRequest(req))
}
