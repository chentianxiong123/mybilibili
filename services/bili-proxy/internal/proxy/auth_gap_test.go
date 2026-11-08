package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleResolve_NoAuthStill200_IDOR pins the AUTH GAP on the
// /api/v1/bili/resolve/ endpoint (已修复): 无鉴权 → 401（此前任意未认证者可解析任意视频 CDN）。
func TestHandleResolve_NoAuthStill200_IDOR(t *testing.T) {
	db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	h := NewHandler(db, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/resolve/42?qn=64", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code, "FINDING-032 修复：无鉴权 resolve 应 401")
}

// TestHandleStream_NoAuthStill200_IDOR pins the same AUTH GAP on the
// /api/v1/bili/stream/ (and legacy /stream/) proxy endpoints (已修复):
// 无鉴权 → 401。
func TestHandleStream_NoAuthStill200_IDOR(t *testing.T) {
	db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	h := NewHandler(db, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/42?qn=64", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code, "FINDING-032 修复：无鉴权 stream 应 401")
}
