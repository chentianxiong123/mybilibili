package coreapi

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mybilibili/core-service/internal/user"
)

type JWT = user.JWT

type HTTPHandler struct{}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{}
}

func (h *HTTPHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/video/process/sse/", h.handleVideoProcessSSE)
	mux.HandleFunc("/api/v1/health", h.handleHealth)
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

func StartHTTPServer(addr string, jwt *JWT, extras ...LiveHandler) {
	mux := http.NewServeMux()
	h := NewHTTPHandler()
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
	minioPublicURL := strings.TrimRight(minioEndpoint, "/") + "/" + minioBucket

	localHandler := http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir)))
	mux.Handle("/uploads/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/uploads/")
		if key == "" {
			localHandler.ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, minioPublicURL+"/"+key, http.StatusMovedPermanently)
	}))

	var handler http.Handler = mux
	if jwt != nil {
		handler = user.AuthMiddleware(jwt)(mux)
	}

	log.Printf("HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}