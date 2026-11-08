package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"mybilibili/bili-proxy/internal/bilibili"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "unsafe" // required by //go:linkname
)

// biliHTTPClient aliases the unexported package-level http client used by
// bilibili.Client.doGet, so proxy tests can inject a deterministic mock
// transport and exercise the getCDN cache-miss (real ResolveStream) paths
// without touching network.
//
//go:linkname biliHTTPClient mybilibili/bili-proxy/internal/bilibili.httpClient
var biliHTTPClient *http.Client

type proxyRoundTripFunc func(req *http.Request) (*http.Response, error)

func (f proxyRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func installBiliHTTPClient(t *testing.T, handler proxyRoundTripFunc) {
	t.Helper()
	orig := biliHTTPClient
	biliHTTPClient = &http.Client{Transport: handler}
	t.Cleanup(func() { biliHTTPClient = orig })
}

func biliJSONResponse(t *testing.T, status int, body string) *http.Response {
	t.Helper()
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

const mockPlayURLDurl = `{"code":0,"data":{"durl":[{"url":"https://cdn.example.com/real.mp4","size":100,"length":10}]}}`

// TestGetCDN_CacheMiss_Success exercises the cache-miss success path: the
// ResolveStream call is made, the result is written to the cache, and a
// subsequent getCDN serves from cache without hitting upstream again.
func TestGetCDN_CacheMiss_Success(t *testing.T) {
	calls := 0
	installBiliHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		calls++
		assert.Equal(t, "/x/player/playurl", req.URL.Path)
		assert.Equal(t, "BV1xx", req.URL.Query().Get("bvid"))
		assert.Equal(t, "42", req.URL.Query().Get("cid"))
		assert.Equal(t, "64", req.URL.Query().Get("qn"))
		return biliJSONResponse(t, http.StatusOK, mockPlayURLDurl), nil
	})

	h := NewHandler(nil, bilibili.NewClient(""))
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	u, kind, err := h.getCDN(req, "BV1xx", 42, 64)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/real.mp4", u)
	assert.Equal(t, "video/mp4", kind)
	assert.Equal(t, 1, calls)

	// Cache should now be populated.
	h.mu.Lock()
	entry, ok := h.cache["BV1xx/42/64"]
	h.mu.Unlock()
	require.True(t, ok)
	assert.Equal(t, "https://cdn.example.com/real.mp4", entry.url)
	assert.Equal(t, "video/mp4", entry.kind)

	// Second call hits the cache and does not resolve again.
	u2, kind2, err := h.getCDN(req, "BV1xx", 42, 64)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/real.mp4", u2)
	assert.Equal(t, "video/mp4", kind2)
	assert.Equal(t, 1, calls)
}

// TestGetCDN_CacheMiss_Error exercises the cache-miss error path: when
// ResolveStream fails, getCDN returns the error and does not poison the cache.
func TestGetCDN_CacheMiss_Error(t *testing.T) {
	installBiliHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	})

	h := NewHandler(nil, bilibili.NewClient(""))
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	u, kind, err := h.getCDN(req, "BV1xx", 42, 64)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Empty(t, u)
	assert.Empty(t, kind)

	// The failed attempt must not populate the cache.
	h.mu.Lock()
	_, ok := h.cache["BV1xx/42/64"]
	h.mu.Unlock()
	assert.False(t, ok)
}

// TestGetCDN_CacheExpired verifies that a stale entry (older than the TTL)
// triggers a fresh ResolveStream instead of being served.
func TestGetCDN_CacheExpired(t *testing.T) {
	calls := 0
	installBiliHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		calls++
		return biliJSONResponse(t, http.StatusOK, mockPlayURLDurl), nil
	})

	h := NewHandler(nil, bilibili.NewClient(""))
	h.mu.Lock()
	h.cache["BV1xx/42/64"] = cdnCacheEntry{
		url:  "https://cdn.example.com/stale.mp4",
		kind: "video/mp4",
		ts:   time.Now().Add(-cdnCacheTTL - time.Minute),
	}
	h.mu.Unlock()

	u, kind, err := h.getCDN(httptest.NewRequest(http.MethodGet, "/", nil), "BV1xx", 42, 64)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/real.mp4", u)
	assert.Equal(t, "video/mp4", kind)
	assert.Equal(t, 1, calls)
}

// TestGetCDN_CacheFresh verifies a fresh entry is served without resolving.
func TestGetCDN_CacheFresh(t *testing.T) {
	calls := 0
	installBiliHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		calls++
		return biliJSONResponse(t, http.StatusOK, mockPlayURLDurl), nil
	})

	h := NewHandler(nil, bilibili.NewClient(""))
	h.mu.Lock()
	h.cache["BV1xx/42/64"] = cdnCacheEntry{
		url:  "https://cdn.example.com/fresh.mp4",
		kind: "video/mp4",
		ts:   time.Now(),
	}
	h.mu.Unlock()

	u, kind, err := h.getCDN(httptest.NewRequest(http.MethodGet, "/", nil), "BV1xx", 42, 64)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/fresh.mp4", u)
	assert.Equal(t, "video/mp4", kind)
	assert.Equal(t, 0, calls)
}

// TestGetCDN_ConcurrentCacheHit hammers getCDN from many goroutines while the
// cache is warm, verifying the proxy's cache map is properly guarded by the
// mutex (run with -race). Note: a concurrent cache-MISS test that shares one
// bilibili.Client would trip a data race inside bilibili.Client.throttle, which
// is a separate real bug tracked as a finding.
func TestGetCDN_ConcurrentCacheHit(t *testing.T) {
	h := NewHandler(nil, bilibili.NewClient(""))
	h.mu.Lock()
	h.cache["BV1xx/42/64"] = cdnCacheEntry{
		url:  "https://cdn.example.com/cached.mp4",
		kind: "video/mp4",
		ts:   time.Now(),
	}
	h.mu.Unlock()

	const n = 16
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u, _, err := h.getCDN(httptest.NewRequest(http.MethodGet, "/", nil), "BV1xx", 42, 64)
			if err != nil {
				t.Errorf("concurrent cache-hit error: %v", err)
				return
			}
			if u != "https://cdn.example.com/cached.mp4" {
				t.Errorf("concurrent cache-hit unexpected url: %q", u)
			}
		}()
	}
	wg.Wait()
}

// TestHandleResolve_502_GetCDNError covers the resolve BadGateway branch
// (upstream ResolveStream failure).
func TestHandleResolve_502_GetCDNError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(42, 999),
	)
	mock.ExpectQuery(`SELECT bvid FROM manuscripts`).WillReturnRows(
		sqlmock.NewRows([]string{"bvid"}).AddRow("BV1xx411c7Xx"),
	)
	installBiliHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("upstream boom")
	})

	h := NewHandler(db, bilibili.NewClient(""))
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/resolve/42?qn=64", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadGateway, rec.Code, "body=%s", rec.Body.String())
	assert.Contains(t, rec.Body.String(), "upstream boom")
}

// TestHandleResolve_200_GetCDNCacheMiss_DefaultQn covers the full resolve
// success path with a real (mocked) ResolveStream and the qn defaulting to 64.
func TestHandleResolve_200_GetCDNCacheMiss_DefaultQn(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(42, 999),
	)
	mock.ExpectQuery(`SELECT bvid FROM manuscripts`).WillReturnRows(
		sqlmock.NewRows([]string{"bvid"}).AddRow("BV1xx411c7Xx"),
	)
	installBiliHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "64", req.URL.Query().Get("qn"), "qn should default to 64")
		return biliJSONResponse(t, http.StatusOK, mockPlayURLDurl), nil
	})

	h := NewHandler(db, bilibili.NewClient(""))
	mux := http.NewServeMux()
	h.Register(mux)

	// No ?qn parameter -> default 64.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/resolve/42", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	var result map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, "https://cdn.example.com/real.mp4", result["url"])
	assert.Equal(t, "video/mp4", result["kind"])
}

// TestHandleStream_502_GetCDNError covers the stream BadGateway branch
// (upstream ResolveStream failure, logged then 502).
func TestHandleStream_502_GetCDNError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(42, 999),
	)
	mock.ExpectQuery(`SELECT m.bvid FROM manuscripts`).WillReturnRows(
		sqlmock.NewRows([]string{"bvid"}).AddRow("BV1xx411c7Xx"),
	)
	installBiliHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("resolve boom")
	})

	h := NewHandler(db, bilibili.NewClient(""))
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/42?qn=64", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadGateway, rec.Code, "body=%s", rec.Body.String())
	assert.Contains(t, rec.Body.String(), "resolve stream failed")
}

// TestHandleStream_QnDefault_CacheHit covers the qn==0 -> 64 default branch in
// handleStream, using a cache entry keyed by the default qn.
func TestHandleStream_QnDefault_CacheHit(t *testing.T) {
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
		w.Write([]byte("default-qn-video"))
	}))
	defer cdnSrv.Close()

	h := NewHandler(db, bilibili.NewClient(""))
	key := "BV1xx411c7Xx/999/64"
	h.mu.Lock()
	h.cache[key] = cdnCacheEntry{url: cdnSrv.URL, kind: "video/mp4", ts: time.Now()}
	h.mu.Unlock()

	mux := http.NewServeMux()
	h.Register(mux)

	// No ?qn -> default 64 -> cache key uses 64.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/42", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	assert.Equal(t, "default-qn-video", rec.Body.String())
}

// TestHandleStream_200_GetCDNCacheMiss covers the full stream success path with
// a real (mocked) ResolveStream followed by proxying the resolved CDN URL.
func TestHandleStream_200_GetCDNCacheMiss(t *testing.T) {
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
		assert.Equal(t, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36", r.Header.Get("User-Agent"))
		assert.Equal(t, "https://www.bilibili.com/", r.Header.Get("Referer"))
		w.Header().Set("Content-Type", "video/mp4")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("resolved-cdn-video"))
	}))
	defer cdnSrv.Close()

	installBiliHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		// Point ResolveStream at the mock CDN so proxyRange has a live target.
		body := strings.Replace(mockPlayURLDurl, "https://cdn.example.com/real.mp4", cdnSrv.URL+"/v.mp4", 1)
		return biliJSONResponse(t, http.StatusOK, body), nil
	})

	h := NewHandler(db, bilibili.NewClient(""))
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/42?qn=64", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	assert.Equal(t, "resolved-cdn-video", rec.Body.String())
}

// TestProxyRange_BuildRequestError covers the 502 branch when the upstream
// target URL is malformed (http.NewRequestWithContext fails).
func TestProxyRange_BuildRequestError(t *testing.T) {
	h := NewHandler(nil, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	h.proxyRange(rec, req, "://malformed-url", "video/mp4")

	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.Contains(t, rec.Body.String(), "build request failed")
}

// TestProxyRange_DoError covers the 502 branch when the upstream client.Do
// fails (target server is unreachable).
func TestProxyRange_DoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	target := srv.URL
	srv.Close()

	h := NewHandler(nil, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	h.proxyRange(rec, req, target, "video/mp4")

	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.Contains(t, rec.Body.String(), "upstream request failed")
}

// TestProxyRange_NoContentTypeFromUpstream covers the branch where the
// upstream response carries no Content-Type, so the handler falls back to the
// resolved kind.
func TestProxyRange_NoContentTypeFromUpstream(t *testing.T) {
	cdnSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Suppress Go's automatic Content-Type sniffing so the upstream
		// response genuinely carries no Content-Type header.
		w.Header()["Content-Type"] = nil
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("no-content-type-body"))
	}))
	defer cdnSrv.Close()

	h := NewHandler(nil, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	h.proxyRange(rec, req, cdnSrv.URL, "video/mp4")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "video/mp4", rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-store, must-revalidate", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "no-content-type-body", rec.Body.String())
}

// TestProxyRange_CancelledContext verifies a cancelled request context aborts
// the upstream call and yields a 502.
func TestProxyRange_CancelledContext(t *testing.T) {
	blocked := make(chan struct{})
	defer close(blocked)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocked
	}))
	defer srv.Close()

	h := NewHandler(nil, nil)
	rec := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	cancel() // cancel before the request is made

	h.proxyRange(rec, req, srv.URL, "video/mp4")

	// client.Do with an already-cancelled context returns an error -> 502.
	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.Contains(t, rec.Body.String(), "upstream request failed")
}

// TestStripPrefix exercises the path-prefix helper directly.
func TestStripPrefix(t *testing.T) {
	assert.Equal(t, "123", stripPrefix("/api/v1/bili/stream/123", "/api/v1/bili/stream/"))
	assert.Equal(t, "123", stripPrefix("/stream/123", "/stream/"))
	assert.Equal(t, "abc", stripPrefix("/resolve/abc", "/resolve/"))
	// Prefix not present: path returned unchanged.
	assert.Equal(t, "/api/v1/bili/stream/123", stripPrefix("/api/v1/bili/stream/123", "/nope/"))
	// Empty prefix.
	assert.Equal(t, "/api/v1/bili/stream/123", stripPrefix("/api/v1/bili/stream/123", ""))
}

// TestHandleLegacyPaths_ValidID covers the legacy /resolve/ and /stream/ paths
// with a valid id that resolves successfully through the cache.
func TestHandleLegacyPaths_ValidID(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(42, 999),
	)
	mock.ExpectQuery(`SELECT bvid FROM manuscripts`).WillReturnRows(
		sqlmock.NewRows([]string{"bvid"}).AddRow("BV1xx411c7Xx"),
	)

	h := NewHandler(db, bilibili.NewClient(""))
	key := "BV1xx411c7Xx/999/64"
	h.mu.Lock()
	h.cache[key] = cdnCacheEntry{url: "https://cached.example.com/v.mp4", kind: "video/mp4", ts: time.Now()}
	h.mu.Unlock()

	mux := http.NewServeMux()
	h.Register(mux)

	// Legacy /resolve/ path with valid id -> 200 (cache hit).
	req := httptest.NewRequest(http.MethodGet, "/resolve/42", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())

	var result map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, "https://cached.example.com/v.mp4", result["url"])
}

// TestWriteErrDirect covers writeErr directly.
func TestWriteErrDirect(t *testing.T) {
	rec := httptest.NewRecorder()
	writeErr(rec, http.StatusBadRequest, "bad input")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), `"code":400`)
	assert.Contains(t, rec.Body.String(), `"message":"bad input"`)
}

// TestHandleStream_404_DBError covers the DB query error branches returning 404.
func TestHandleStream_404_DBError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(42, 999),
	)
	mock.ExpectQuery(`SELECT m.bvid FROM manuscripts`).WillReturnError(errors.New("db down"))

	h := NewHandler(db, bilibili.NewClient(""))
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/42", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code, "body=%s", rec.Body.String())
	assert.Contains(t, rec.Body.String(), "manuscript not found")
}

// TestHandleResolve_404_DBError covers the resolve manuscript query error.
func TestHandleResolve_404_DBError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "cid"}).AddRow(42, 999),
	)
	mock.ExpectQuery(`SELECT bvid FROM manuscripts`).WillReturnError(errors.New("db down"))

	h := NewHandler(db, bilibili.NewClient(""))
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/resolve/42", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code, "body=%s", rec.Body.String())
	assert.Contains(t, rec.Body.String(), "manuscript not found")
}

// TestLoadVideo_QueryError covers loadVideo returning a non-ErrNoRows DB error.
func TestLoadVideo_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT v.id`).WillReturnError(errors.New("scan boom"))

	h := NewHandler(db, nil)
	row, err := h.loadVideo(httptest.NewRequest(http.MethodGet, "/", nil), 42)
	assert.Error(t, err)
	assert.Nil(t, row)
}
