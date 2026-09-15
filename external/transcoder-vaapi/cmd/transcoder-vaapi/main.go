package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"mybilibili/pkg/abstraction"
	"mybilibili/transcoder_vaapi/internal/transcoder"
)

func main() {
	addr := getEnv("HTTP_ADDR", ":8092")

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

	svc := transcoder.NewService(storage, "vaapi")

	mux := http.NewServeMux()
	transcoder.RegisterRoutes(mux, svc)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/api/v1/capabilities", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"encoder":%q,"capabilities":%s}`,
			svc.Encoder(),
			mustJSON(svc.Capabilities()),
		)
	})

	server := &http.Server{Addr: addr, Handler: mux}
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		server.Close()
	}()

	log.Printf("transcoder-vaapi listening on %s", addr)
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

func stripScheme(s string) string {
	for _, p := range []string{"https://", "http://"} {
		if len(s) > len(p) && s[:len(p)] == p {
			return s[len(p):]
		}
	}
	return s
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
