package coreapi

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mybilibili/pkg/auth"
)

type JWT = auth.JWT

type HTTPHandler struct {
	jwt *auth.JWT
}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{}
}

func (h *HTTPHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/video/process/sse/", h.handleVideoProcessSSE)
	mux.HandleFunc("/api/v1/health", h.handleHealth)
}

// SetJWT 注入验签用的 JWT 工具（验证端点 handleAuthVerify 使用）。
func (h *HTTPHandler) SetJWT(j *auth.JWT) {
	h.jwt = j
}

// handleAuthVerify 供 Traefik forwardAuth 全权验签调用。
// 读 Authorization: Bearer <token>，验签后通过响应头向 Traefik 返回身份，
// Traefik 将 X-User-Id / X-User-Role / X-Admin-Id 注入转发的请求头（下游全信）。
func (h *HTTPHandler) handleAuthVerify(w http.ResponseWriter, r *http.Request) {
	j := h.jwt
	if j == nil {
		http.Error(w, "auth not configured", http.StatusInternalServerError)
		return
	}
	tokenStr, ok := auth.BearerToken(r.Header.Get("Authorization"))
	if !ok {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	claims, err := j.Parse(tokenStr)
	if err != nil || claims == nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	w.Header().Set("X-User-Id", strconv.FormatInt(claims.UserId, 10))
	if claims.Role == "" {
		claims.Role = auth.RoleUser
	}
	w.Header().Set("X-User-Role", claims.Role)
	if claims.IsAdmin {
		w.Header().Set("X-Admin-Id", strconv.FormatInt(claims.UserId, 10))
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

func (h *HTTPHandler) handleVideoProcessSSE(w http.ResponseWriter, r *http.Request) {
	videoIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/video/process/sse/")
	videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid video id", 400)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", 500)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			data := fmt.Sprintf(`{"video_id":%d,"stage":"progress","progress":0}`, videoID)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

type LiveHandler interface {
	Register(mux *http.ServeMux)
}

type LiveProxy struct {
	proxy  *httputil.ReverseProxy
	prefix string
}

func NewLiveProxy() *LiveProxy {
	target := os.Getenv("LIVE_SERVICE_ADDR")
	if target == "" {
		target = "http://127.0.0.1:8087"
	}
	u, _ := url.Parse(target)
	return &LiveProxy{proxy: httputil.NewSingleHostReverseProxy(u), prefix: "/api/v1/live/"}
}

func isImageKey(key string) bool {
	return strings.HasPrefix(key, "manuscripts/") &&
		(strings.HasSuffix(key, ".webp") || strings.HasSuffix(key, ".png") ||
			strings.HasSuffix(key, ".jpg") || strings.HasSuffix(key, ".jpeg") ||
			strings.HasSuffix(key, ".gif") || strings.HasSuffix(key, ".avif")) ||
		(strings.HasPrefix(key, "avatars/") && (strings.HasSuffix(key, ".webp") ||
			strings.HasSuffix(key, ".png") || strings.HasSuffix(key, ".jpg") || strings.HasSuffix(key, ".jpeg")))
}

type cacheControlWriter struct {
	http.ResponseWriter
	cacheControl string
}

func (c *cacheControlWriter) WriteHeader(code int) {
	c.Header().Set("Cache-Control", c.cacheControl)
	c.ResponseWriter.WriteHeader(code)
}

func (p *LiveProxy) Register(mux *http.ServeMux) {
	mux.HandleFunc(p.prefix, p.proxy.ServeHTTP)
}

func StartHTTPServer(addr string, jwt *JWT, extras ...LiveHandler) {
	mux := http.NewServeMux()
	h := NewHTTPHandler()
	h.SetJWT(jwt)
	mux.HandleFunc("/api/v1/auth/verify", h.handleAuthVerify)
	h.Register(mux)
	for _, e := range extras {
		e.Register(mux)
	}

	uploadDir := os.Getenv("MYBILIBILI_UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = filepath.Join(os.TempDir(), "mybilibili-uploads")
	}
	_ = os.MkdirAll(uploadDir, 0o755)

	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "http://127.0.0.1:9000"
	}
	minioBucket := os.Getenv("MINIO_BUCKET")
	if minioBucket == "" {
		minioBucket = "mybilibili"
	}
	localHandler := http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir)))
	minioURL, _ := url.Parse(strings.TrimRight(minioEndpoint, "/"))
	minioProxy := httputil.NewSingleHostReverseProxy(minioURL)
	mux.Handle("/uploads/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/uploads/")
		if key == "" {
			localHandler.ServeHTTP(w, r)
			return
		}
		// 图片类（封面/头像）强缓存；m3u8/ts 视频沿用默认（会被转码替换）
		cacheControl := "public, max-age=600"
		if isImageKey(key) {
			cacheControl = "public, max-age=31536000, immutable"
		}
		r.URL.Path = "/" + minioBucket + "/" + key
		r.Host = minioURL.Host
		minioProxy.ServeHTTP(&cacheControlWriter{ResponseWriter: w, cacheControl: cacheControl}, r)
	}))

	var handler http.Handler = mux
	if jwt != nil {
		handler = auth.IdentityMiddleware(jwt)(mux)
	}

	log.Printf("HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}