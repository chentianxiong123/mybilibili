package coreapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"mybilibili/pkg/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noFlusher wraps ResponseWriter but does NOT implement http.Flusher.
type noFlusher struct {
	http.ResponseWriter
}

// wrappedRecorder captures the status code written via WriteHeader.
type wrappedRecorder struct {
	http.ResponseWriter
	code int
}

func (w *wrappedRecorder) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

// ---------- NewHTTPHandler / SetJWT ----------

func TestNewHTTPHandler(t *testing.T) {
	h := NewHTTPHandler()
	assert.NotNil(t, h)
	assert.Nil(t, h.jwt)
}

func TestSetJWT(t *testing.T) {
	h := NewHTTPHandler()
	j := auth.NewJWT("secret")
	h.SetJWT(j)
	assert.Equal(t, j, h.jwt)
}

// ---------- handleHealth ----------

func TestHandleHealth(t *testing.T) {
	h := NewHTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	h.handleHealth(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, `{"status":"ok"}`, rec.Body.String())
}

// ---------- handleVideoProcessSSE ----------

func TestHandleVideoProcessSSE_InvalidID(t *testing.T) {
	h := NewHTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/video/process/sse/notanumber", nil)
	rec := httptest.NewRecorder()
	h.handleVideoProcessSSE(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid video id")
}

func TestHandleVideoProcessSSE_EmptyID(t *testing.T) {
	h := NewHTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/video/process/sse/", nil)
	rec := httptest.NewRecorder()
	h.handleVideoProcessSSE(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleVideoProcessSSE_NegativeID(t *testing.T) {
	h := NewHTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/video/process/sse/-1", nil)
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		h.handleVideoProcessSSE(rec, req)
		close(done)
	}()

	cancel()
	<-done

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandleVideoProcessSSE_NonFlusher(t *testing.T) {
	h := NewHTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/video/process/sse/42", nil)
	inner := httptest.NewRecorder()
	nf := &noFlusher{ResponseWriter: inner}
	rec := &wrappedRecorder{ResponseWriter: nf}
	h.handleVideoProcessSSE(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.code)
	assert.Contains(t, inner.Body.String(), "streaming not supported")
}

func TestHandleVideoProcessSSE_StreamAndCancel(t *testing.T) {
	h := NewHTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/video/process/sse/123", nil)
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		h.handleVideoProcessSSE(rec, req)
		close(done)
	}()

	time.Sleep(6 * time.Second)
	cancel()
	<-done

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", rec.Header().Get("Connection"))
	body := rec.Body.String()
	assert.Contains(t, body, `"video_id":123`)
	assert.Contains(t, body, `"stage":"progress"`)
	assert.Contains(t, body, `"progress":0`)
}

func TestHandleVideoProcessSSE_CancelImmediately(t *testing.T) {
	h := NewHTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/video/process/sse/500", nil)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.handleVideoProcessSSE(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
}

// ---------- handleAuthVerify ----------

func TestHandleAuthVerify_NoJWTConfigured(t *testing.T) {
	h := NewHTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify", nil)
	rec := httptest.NewRecorder()
	h.handleAuthVerify(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "auth not configured")
}

func TestHandleAuthVerify_MissingToken(t *testing.T) {
	h := NewHTTPHandler()
	h.SetJWT(auth.NewJWT("secret"))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify", nil)
	rec := httptest.NewRecorder()
	h.handleAuthVerify(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing token")
}

func TestHandleAuthVerify_InvalidToken(t *testing.T) {
	h := NewHTTPHandler()
	h.SetJWT(auth.NewJWT("secret"))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify", nil)
	req.Header.Set("Authorization", "Bearer badtoken")
	rec := httptest.NewRecorder()
	h.handleAuthVerify(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid token")
}

func TestHandleAuthVerify_ValidUser(t *testing.T) {
	j := auth.NewJWT("testsecret")
	token, err := j.Generate(42)
	require.NoError(t, err)

	h := NewHTTPHandler()
	h.SetJWT(j)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.handleAuthVerify(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "42", rec.Header().Get("X-User-Id"))
	assert.Equal(t, "user", rec.Header().Get("X-User-Role"))
	assert.Empty(t, rec.Header().Get("X-Admin-Id"))
}

func TestHandleAuthVerify_ValidAdmin(t *testing.T) {
	j := auth.NewJWT("testsecret")
	token, err := j.GenerateAdmin(99)
	require.NoError(t, err)

	h := NewHTTPHandler()
	h.SetJWT(j)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.handleAuthVerify(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "99", rec.Header().Get("X-User-Id"))
	assert.Equal(t, "admin", rec.Header().Get("X-User-Role"))
	assert.Equal(t, "99", rec.Header().Get("X-Admin-Id"))
}

func TestHandleAuthVerify_EmptyRole(t *testing.T) {
	j := auth.NewJWT("testsecret")
	token, err := j.GenerateWithRole(10, "")
	require.NoError(t, err)

	h := NewHTTPHandler()
	h.SetJWT(j)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.handleAuthVerify(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, auth.RoleUser, rec.Header().Get("X-User-Role"))
}

// ---------- isImageKey ----------

func TestIsImageKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"manuscripts/cover.webp", true},
		{"manuscripts/page.png", true},
		{"manuscripts/photo.jpg", true},
		{"manuscripts/photo.jpeg", true},
		{"manuscripts/image.gif", true},
		{"manuscripts/image.avif", true},
		{"avatars/avatar.webp", true},
		{"avatars/avatar.png", true},
		{"avatars/avatar.jpg", true},
		{"avatars/avatar.jpeg", true},
		{"manuscripts/video.mp4", false},
		{"manuscripts/file.txt", false},
		{"other/file.webp", false},
		{"avatars/file.gif", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			assert.Equal(t, tt.want, isImageKey(tt.key))
		})
	}
}

// ---------- cacheControlWriter ----------

func TestCacheControlWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	ccw := &cacheControlWriter{ResponseWriter: rec, cacheControl: "public, max-age=600"}

	ccw.WriteHeader(http.StatusOK)
	assert.Equal(t, "public, max-age=600", rec.Header().Get("Cache-Control"))
	assert.Equal(t, http.StatusOK, rec.Code)

	rec2 := httptest.NewRecorder()
	ccw2 := &cacheControlWriter{ResponseWriter: rec2, cacheControl: "immutable"}
	ccw2.WriteHeader(http.StatusNotFound)
	assert.Equal(t, "immutable", rec2.Header().Get("Cache-Control"))
	assert.Equal(t, http.StatusNotFound, rec2.Code)
}

// ---------- LiveProxy ----------

func TestNewLiveProxyDefault(t *testing.T) {
	os.Unsetenv("LIVE_SERVICE_ADDR")
	lp := NewLiveProxy()
	assert.NotNil(t, lp)
	assert.NotNil(t, lp.proxy)
	assert.Equal(t, "/api/v1/live/", lp.prefix)
}

func TestNewLiveProxyCustom(t *testing.T) {
	t.Setenv("LIVE_SERVICE_ADDR", "http://127.0.0.1:9999")
	lp := NewLiveProxy()
	assert.NotNil(t, lp)
}

func TestLiveProxyServes(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend-ok"))
	}))
	defer backend.Close()

	t.Setenv("LIVE_SERVICE_ADDR", backend.URL)
	lp := NewLiveProxy()
	mux := http.NewServeMux()
	lp.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/live/test", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "backend-ok", rec.Body.String())
}

// ---------- BiliProxy ----------

func TestNewBiliProxyDefault(t *testing.T) {
	os.Unsetenv("BILI_PROXY_ADDR")
	bp := NewBiliProxy()
	assert.NotNil(t, bp)
	assert.NotNil(t, bp.proxy)
	assert.Equal(t, "/api/v1/bili/", bp.prefix)
}

func TestNewBiliProxyCustom(t *testing.T) {
	t.Setenv("BILI_PROXY_ADDR", "http://127.0.0.1:9999")
	bp := NewBiliProxy()
	assert.NotNil(t, bp)
}

func TestBiliProxyServes(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("bili-backend:" + r.URL.Path))
	}))
	defer backend.Close()

	t.Setenv("BILI_PROXY_ADDR", backend.URL)
	bp := NewBiliProxy()
	mux := http.NewServeMux()
	bp.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/456?qn=1080", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "/stream/456")
}

func TestBiliProxyStripsPrefix(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("path:" + r.URL.Path + "?" + r.URL.RawQuery))
	}))
	defer backend.Close()

	t.Setenv("BILI_PROXY_ADDR", backend.URL)
	bp := NewBiliProxy()
	mux := http.NewServeMux()
	bp.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/1?qn=64", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "path:/stream/1?qn=64")
}

func TestBiliProxyWithRawPath(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("raw=" + r.URL.RawPath))
	}))
	defer backend.Close()

	t.Setenv("BILI_PROXY_ADDR", backend.URL)
	bp := NewBiliProxy()
	mux := http.NewServeMux()
	bp.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/abc", nil)
	req.URL.RawPath = "/api/v1/bili/stream/abc"
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ---------- LiveHandler interface ----------

func TestLiveHandlerInterface(t *testing.T) {
	var _ LiveHandler = &LiveProxy{}
	var _ LiveHandler = &BiliProxy{}
}

// ---------- IdentityMiddleware integration ----------

func TestAuthVerifyThroughMiddleware(t *testing.T) {
	j := auth.NewJWT("testsecret")
	token, err := j.Generate(77)
	require.NoError(t, err)

	h := NewHTTPHandler()
	h.SetJWT(j)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/verify", h.handleAuthVerify)
	mux.HandleFunc("/api/v1/health", h.handleHealth)

	handler := auth.IdentityMiddleware(j)(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthVerifyMiddleware_AlreadyHasUserID(t *testing.T) {
	j := auth.NewJWT("testsecret")
	h := NewHTTPHandler()
	h.SetJWT(j)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", h.handleHealth)

	handler := auth.IdentityMiddleware(j)(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("X-User-Id", "123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthVerifyMiddleware_BadToken(t *testing.T) {
	j := auth.NewJWT("testsecret")
	h := NewHTTPHandler()
	h.SetJWT(j)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", h.handleHealth)

	handler := auth.IdentityMiddleware(j)(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("Authorization", "Bearer badtoken")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// ---------- StartHTTPServer ----------

func TestStartHTTPServer(t *testing.T) {
	os.Unsetenv("MYBILIBILI_UPLOAD_DIR")
	os.Unsetenv("MINIO_ENDPOINT")
	os.Unsetenv("MINIO_BUCKET")

	j := auth.NewJWT("testsecret")
	go StartHTTPServer("127.0.0.1:18234", j)
	time.Sleep(200 * time.Millisecond)

	token, _ := j.Generate(10)
	client := &http.Client{Timeout: 3 * time.Second}

	// /api/v1/health
	req, _ := http.NewRequest("GET", "http://127.0.0.1:18234/api/v1/health", nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// /api/v1/auth/verify - valid token
	req2, _ := http.NewRequest("GET", "http://127.0.0.1:18234/api/v1/auth/verify", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	resp2, err := client.Do(req2)
	require.NoError(t, err)
	resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	// /uploads/ (empty key -> local file server)
	req3, _ := http.NewRequest("GET", "http://127.0.0.1:18234/uploads/", nil)
	resp3, err := client.Do(req3)
	require.NoError(t, err)
	resp3.Body.Close()

	// /uploads/nonexistent.webp (image key -> immutable cache)
	req4, _ := http.NewRequest("GET", "http://127.0.0.1:18234/uploads/manuscripts/cover.webp", nil)
	resp4, err := client.Do(req4)
	require.NoError(t, err)
	resp4.Body.Close()
	assert.Equal(t, "public, max-age=31536000, immutable", resp4.Header.Get("Cache-Control"))

	// /uploads/video.mp4 (non-image -> 600s cache)
	req5, _ := http.NewRequest("GET", "http://127.0.0.1:18234/uploads/video.mp4", nil)
	resp5, err := client.Do(req5)
	require.NoError(t, err)
	resp5.Body.Close()
	assert.Equal(t, "public, max-age=600", resp5.Header.Get("Cache-Control"))

	// /uploads/ with custom env
	t.Setenv("MINIO_ENDPOINT", "http://127.0.0.1:9999")
	t.Setenv("MINIO_BUCKET", "test-bucket")
	t.Setenv("MYBILIBILI_UPLOAD_DIR", "/tmp/test-uploads")
}

// ---------- Register + health (via mux) ----------

func TestHealthViaMux(t *testing.T) {
	h := NewHTTPHandler()
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, `{"status":"ok"}`, rec.Body.String())
}

func TestSSEInvalidIDViaMux(t *testing.T) {
	h := NewHTTPHandler()
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/video/process/sse/bad", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
