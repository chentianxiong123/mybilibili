package httputil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserIDFromHeader_Valid(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-User-Id", "42")
	assert.Equal(t, int64(42), GetUserIDFromHeader(r))
}

func TestGetUserIDFromHeader_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	assert.Equal(t, int64(0), GetUserIDFromHeader(r))
}

func TestGetUserIDFromHeader_NonNumeric(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-User-Id", "abc")
	assert.Equal(t, int64(0), GetUserIDFromHeader(r))
}

func TestGetAdminIDFromHeader_AdminHeader(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Admin-Id", "7")
	r.Header.Set("X-User-Id", "99")
	assert.Equal(t, int64(7), GetAdminIDFromHeader(r))
}

func TestGetAdminIDFromHeader_FallbackToUser(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-User-Id", "99")
	assert.Equal(t, int64(99), GetAdminIDFromHeader(r))
}

func TestGetAdminIDFromHeader_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	assert.Equal(t, int64(0), GetAdminIDFromHeader(r))
}

func TestParsePageParams_Defaults(t *testing.T) {
	r := httptest.NewRequest("GET", "/?x=1", nil)
	page, size := ParsePageParams(r)
	assert.Equal(t, int32(1), page)
	assert.Equal(t, int32(20), size)
}

func TestParsePageParams_PageSizeAliases(t *testing.T) {
	for _, q := range []string{"?page=2&page_size=5", "?page=3&pageSize=7", "?page=4&size=9", "?page=5&limit=11"} {
		r := httptest.NewRequest("GET", "/"+q, nil)
		page, size := ParsePageParams(r)
		assert.NotZero(t, size, "query=%s", q)
		assert.True(t, page >= 1, "query=%s", q)
	}
}

func TestParsePageParams_PageClampedTo1(t *testing.T) {
	r := httptest.NewRequest("GET", "/?page=0", nil)
	page, _ := ParsePageParams(r)
	assert.Equal(t, int32(1), page)
}

func TestParsePageParams_SizeOutOfRange(t *testing.T) {
	for _, q := range []string{"?page=1&size=0", "?page=1&size=999", "?page=1&size=-1"} {
		r := httptest.NewRequest("GET", "/"+q, nil)
		_, size := ParsePageParams(r)
		assert.Equal(t, int32(20), size, "query=%s", q)
	}
}

func TestWriteOK(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteOK(rec, map[string]string{"k": "v"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	body := rec.Body.String()
	assert.Contains(t, body, `"code":200`)
	assert.Contains(t, body, `"k":"v"`)
	assert.Contains(t, body, `"message":"ok"`)
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, 404, "nope")
	assert.Equal(t, http.StatusNotFound, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, `"code":404`)
	assert.Contains(t, body, `"message":"nope"`)
}

func TestRequireUser_Present(t *testing.T) {
	rec := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-User-Id", "123")
	uid, ok := RequireUser(rec, r)
	assert.True(t, ok)
	assert.Equal(t, int64(123), uid)
	assert.Empty(t, rec.Body.String())
}

func TestRequireUser_Missing(t *testing.T) {
	rec := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	uid, ok := RequireUser(rec, r)
	assert.False(t, ok)
	assert.Zero(t, uid)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":401`)
}

func TestPathValue_Empty(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	assert.Equal(t, "", PathValue(r, "id"))
}

func TestWriteJSON_EmptyBody(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusOK, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "null", strings.TrimSpace(rec.Body.String()))
}
