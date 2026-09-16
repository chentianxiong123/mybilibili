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

func TestHandleHealth_200(t *testing.T) {
	h := NewHandler(nil, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestHandleResolve_400_InvalidID(t *testing.T) {
	h := NewHandler(nil, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/resolve/abc", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":400`)
}

func TestHandleResolve_404_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	h := NewHandler(db, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	// loadVideo: SELECT v.id ... WHERE v.id = $1 → no rows
	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(sqlmock.NewRows([]string{"id", "cid"}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/resolve/1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code, "body=%s", rec.Body.String())
}

func TestHandleStream_400_InvalidID(t *testing.T) {
	h := NewHandler(nil, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/abc", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":400`)
}

func TestHandleStream_404_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	h := NewHandler(db, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	// loadVideo: video doesn't exist → 404
	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(sqlmock.NewRows([]string{"id", "cid"}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code, "body=%s", rec.Body.String())
}

func TestHandleResolve_400_MissingID(t *testing.T) {
	h := NewHandler(nil, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/resolve/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":400`)
}

func TestHandleStream_400_MissingID(t *testing.T) {
	h := NewHandler(nil, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":400`)
}

func TestHandleResolve_404_NoCID(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(42, 0),
	)

	h := NewHandler(db, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/resolve/42", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code, "body=%s", rec.Body.String())
}

func TestHandleStream_404_NoCID(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(1, 0),
	)

	h := NewHandler(db, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code, "body=%s", rec.Body.String())
}

func TestHandleResolve_404_NoManuscript(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(1, 100),
	)
	mock.ExpectQuery(`SELECT bvid FROM manuscripts`).WillReturnRows(
		sqlmock.NewRows([]string{"bvid"}),
	)

	h := NewHandler(db, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/resolve/1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code, "body=%s", rec.Body.String())
}

func TestHandleStream_404_NoManuscript(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(1, 100),
	)
	mock.ExpectQuery(`SELECT m.bvid FROM manuscripts`).WillReturnRows(
		sqlmock.NewRows([]string{"bvid"}),
	)

	h := NewHandler(db, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code, "body=%s", rec.Body.String())
}

func TestHandleResolve_200_CacheHit(t *testing.T) {
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

	// Pre-populate CDN cache
	key := "BV1xx411c7Xx/999/64"
	h.mu.Lock()
	h.cache[key] = cdnCacheEntry{url: "https://cached.example.com/v.mp4", kind: "video/mp4", ts: time.Now()}
	h.mu.Unlock()

	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/resolve/42?qn=64", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	var result map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, "https://cached.example.com/v.mp4", result["url"])
	assert.Equal(t, "video/mp4", result["kind"])
}

func TestHandleStream_200_CacheHit(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(42, 999),
	)
	mock.ExpectQuery(`SELECT m.bvid FROM manuscripts`).WillReturnRows(
		sqlmock.NewRows([]string{"bvid"}).AddRow("BV1xx411c7Xx"),
	)

	h := NewHandler(db, bilibili.NewClient("test"))

	// Pre-populate CDN cache
	key := "BV1xx411c7Xx/999/64"
	h.mu.Lock()
	h.cache[key] = cdnCacheEntry{url: "https://cached.example.com/v.mp4", kind: "video/mp4", ts: time.Now()}
	h.mu.Unlock()

	// mock upstream CDN to return video bytes
	cdnSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake-video-bytes"))
	}))
	defer cdnSrv.Close()

	// Override cache URL to point to our mock CDN
	h.mu.Lock()
	h.cache[key] = cdnCacheEntry{url: cdnSrv.URL, kind: "video/mp4", ts: time.Now()}
	h.mu.Unlock()

	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/42?qn=64", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	assert.Equal(t, "fake-video-bytes", rec.Body.String())
}

func TestHandleLegacyPaths(t *testing.T) {
	h := NewHandler(nil, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	// /resolve/ (legacy path) with invalid id
	req := httptest.NewRequest(http.MethodGet, "/resolve/abc", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// /stream/ (legacy path) with invalid id
	req = httptest.NewRequest(http.MethodGet, "/stream/abc", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestProxyRange_206(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(42, 999),
	)
	mock.ExpectQuery(`SELECT m.bvid FROM manuscripts`).WillReturnRows(
		sqlmock.NewRows([]string{"bvid"}).AddRow("BV1xx411c7Xx"),
	)

	h := NewHandler(db, bilibili.NewClient("test"))

	cdnSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "bytes=0-1023", r.Header.Get("Range"))
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Range", "bytes 0-1023/10240")
		w.Header().Set("Content-Length", "1024")
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte("partial-video-bytes"))
	}))
	defer cdnSrv.Close()

	key := "BV1xx411c7Xx/999/64"
	h.mu.Lock()
	h.cache[key] = cdnCacheEntry{url: cdnSrv.URL, kind: "video/mp4", ts: time.Now()}
	h.mu.Unlock()

	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/42?qn=64", nil)
	req.Header.Set("Range", "bytes=0-1023")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusPartialContent, rec.Code, "body=%s", rec.Body.String())
	assert.Equal(t, "partial-video-bytes", rec.Body.String())
	assert.Equal(t, "bytes 0-1023/10240", rec.Header().Get("Content-Range"))
}

func TestProxyRange_InvalidRange(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(42, 999),
	)
	mock.ExpectQuery(`SELECT m.bvid FROM manuscripts`).WillReturnRows(
		sqlmock.NewRows([]string{"bvid"}).AddRow("BV1xx411c7Xx"),
	)

	h := NewHandler(db, bilibili.NewClient("test"))

	cdnSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		w.Write([]byte("range not satisfiable"))
	}))
	defer cdnSrv.Close()

	key := "BV1xx411c7Xx/999/64"
	h.mu.Lock()
	h.cache[key] = cdnCacheEntry{url: cdnSrv.URL, kind: "video/mp4", ts: time.Now()}
	h.mu.Unlock()

	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/42?qn=64", nil)
	req.Header.Set("Range", "bytes=999999-1000000")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, rec.Code)
}

func TestHandleStream_Redirect(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(42, 999),
	)
	mock.ExpectQuery(`SELECT m.bvid FROM manuscripts`).WillReturnRows(
		sqlmock.NewRows([]string{"bvid"}).AddRow("BV1xx411c7Xx"),
	)

	h := NewHandler(db, bilibili.NewClient("test"))

	finalSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("redirected-video"))
	}))
	defer finalSrv.Close()

	redirectSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, finalSrv.URL+"/video.mp4", http.StatusFound)
	}))
	defer redirectSrv.Close()

	key := "BV1xx411c7Xx/999/64"
	h.mu.Lock()
	h.cache[key] = cdnCacheEntry{url: redirectSrv.URL + "/video.mp4", kind: "video/mp4", ts: time.Now()}
	h.mu.Unlock()

	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/42?qn=64", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "redirected-video", rec.Body.String())
}

func TestHandleStream_DirectPlay(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(42, 999),
	)
	mock.ExpectQuery(`SELECT m.bvid FROM manuscripts`).WillReturnRows(
		sqlmock.NewRows([]string{"bvid"}).AddRow("BV1xx411c7Xx"),
	)

	h := NewHandler(db, bilibili.NewClient("test"))

	cdnSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", "2048")
		w.Header().Set("ETag", `"abc123"`)
		w.Header().Set("Last-Modified", "Mon, 01 Sep 2026 00:00:00 GMT")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("direct-play-bytes"))
	}))
	defer cdnSrv.Close()

	key := "BV1xx411c7Xx/999/64"
	h.mu.Lock()
	h.cache[key] = cdnCacheEntry{url: cdnSrv.URL, kind: "video/mp4", ts: time.Now()}
	h.mu.Unlock()

	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/42?qn=64", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "direct-play-bytes", rec.Body.String())
	assert.Equal(t, "video/mp4", rec.Header().Get("Content-Type"))
	assert.Equal(t, "2048", rec.Header().Get("Content-Length"))
	assert.Equal(t, `"abc123"`, rec.Header().Get("ETag"))
	assert.Equal(t, "no-store, must-revalidate", rec.Header().Get("Cache-Control"))
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
