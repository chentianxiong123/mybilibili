package httputil

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGetUserIDFromHeader_NegativeAccepted pins FINDING: GetUserIDFromHeader accepts
// negative IDs, so RequireUser treats "X-User-Id: -5" as a valid authenticated user.
//
// The identity header is normally rebuilt by auth.IdentityMiddleware from validated
// claims (which deletes client-supplied headers first), so this is only reachable when
// a handler reads the header without that middleware — e.g. a public route that fails
// to be wrapped. A negative value must be rejected, not passed through.
func TestGetUserIDFromHeader_NegativeAccepted(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-User-Id", "-5")

	uid := GetUserIDFromHeader(r)
	// FINDING-082 修复：负 X-User-Id 应被拒绝（0）。
	require.Equal(t, int64(0), uid, "negative X-User-Id should be rejected")
}

// TestParsePageParams_OverflowPins FINDING: ParsePageParams clamps an overflowing
// page value to int32 max instead of resetting to 1, so a caller-supplied huge page
// yields an absurd offset rather than page 1.
func TestParsePageParams_Overflow(t *testing.T) {
	r := httptest.NewRequest("GET", "/?page=99999999999999&page_size=20", nil)
	page, size := ParsePageParams(r)
	require.Equal(t, int32(1), page, "FINDING-083 修复: 溢出 page 应重置为 1")
	require.Equal(t, int32(20), size)
}
