package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mybilibili/bili-proxy/internal/bilibili"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleResolve_NoAuthStill200_IDOR pins the AUTH GAP on the
// /api/v1/bili/resolve/ endpoint: the handler performs NO authentication and
// NO ownership check (no RequireUser/GetUserIDFromHeader/AdminPathGuard
// anywhere in the package). Any unauthenticated caller can resolve the CDN
// URL of ANY video purely by numeric ID (IDOR / broken access control). The
// handler never inspects an Authorization header, so a cache-hit resolve
// returns 200 with zero credentials.
func TestHandleResolve_NoAuthStill200_IDOR(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(42, 999),
	)
	mock.ExpectQuery(`SELECT bvid FROM manuscripts`).WillReturnRows(
		sqlmock.NewRows([]string{"bvid"}).AddRow("BV1xx411c7Xx"),
	)

	h := NewHandler(db, bilibili.NewClient("test"))
	key := "BV1xx411c7Xx/999/64"
	h.mu.Lock()
	h.cache[key] = cdnCacheEntry{url: "https://cached.example.com/v.mp4", kind: "video/mp4", ts: time.Now()}
	h.mu.Unlock()

	mux := http.NewServeMux()
	h.Register(mux)

	// Deliberately NO Authorization / session cookie / user header.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/resolve/42?qn=64", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// AUTH GAP: 200 with zero credentials proves the endpoint is public.
	assert.Equal(t, http.StatusOK, rec.Code,
		"resolve endpoint served a CDN URL without any authentication")
	var result map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, "https://cached.example.com/v.mp4", result["url"])
}

// TestHandleStream_NoAuthStill200_IDOR pins the same AUTH GAP on the
// /api/v1/bili/stream/ (and legacy /stream/) proxy endpoints: unauthenticated
// callers can stream ANY video's content by numeric ID.
func TestHandleStream_NoAuthStill200_IDOR(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(42, 999),
	)
	mock.ExpectQuery(`SELECT m.bvid FROM manuscripts`).WillReturnRows(
		sqlmock.NewRows([]string{"bvid"}).AddRow("BV1xx411c7Xx"),
	)

	cdnSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("unauth-video-bytes"))
	}))
	defer cdnSrv.Close()

	h := NewHandler(db, bilibili.NewClient("test"))
	key := "BV1xx411c7Xx/999/64"
	h.mu.Lock()
	h.cache[key] = cdnCacheEntry{url: cdnSrv.URL, kind: "video/mp4", ts: time.Now()}
	h.mu.Unlock()

	mux := http.NewServeMux()
	h.Register(mux)

	// No auth headers at all -> still streams content (AUTH GAP / IDOR).
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/42?qn=64", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code,
		"stream endpoint served video content without any authentication")
	assert.Equal(t, "unauth-video-bytes", rec.Body.String())
}
