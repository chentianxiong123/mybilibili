package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"mybilibili/pkg/abstraction"
	"mybilibili/transcoder-service/internal/transcoder"
)

func main() {
	addr := getEnv("HTTP_ADDR", ":8092")

	// MinIO 存储（对象引用方式：从 bucket 读源、写产物）
	// minio-go 不接受带 scheme 的 endpoint（会报 fully qualified paths），
	// 统一剥掉 http(s):// 前缀；core 的 /uploads 反代才需要带 scheme。
	minioCfg := abstraction.DefaultMinioConfig()
	if v := os.Getenv("MINIO_ENDPOINT"); v != "" {
		minioCfg.Endpoint = v
		minioCfg.PublicEndpoint = v
	}
	minioCfg.Endpoint = stripScheme(minioCfg.Endpoint)
	minioCfg.PublicEndpoint = stripScheme(minioCfg.PublicEndpoint)
	if v := os.Getenv("MINIO_ACCESS_KEY"); v != "" {
		minioCfg.AccessKey = v
	}
	if v := os.Getenv("MINIO_SECRET_KEY"); v != "" {
		minioCfg.SecretKey = v
	}
	if v := os.Getenv("MINIO_BUCKET"); v != "" {
		minioCfg.BucketName = v
	}
	storage, err := abstraction.NewMinioStorageService(minioCfg)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	encoder := getEnv("TRANSCODE_ENCODER", "auto")
	svc := transcoder.NewService(storage, encoder)

	mux := http.NewServeMux()
	transcoder.RegisterRoutes(mux, svc)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	server := &http.Server{Addr: addr, Handler: mux}
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		server.Close()
	}()

	log.Printf("transcoder service listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// stripScheme 去掉 http(s):// 前缀，供 minio-go 客户端使用（其不接受带 scheme 的 endpoint）。
func stripScheme(s string) string {
	for _, p := range []string{"https://", "http://"} {
		if len(s) > len(p) && s[:len(p)] == p {
			return s[len(p):]
		}
	}
	return s
}