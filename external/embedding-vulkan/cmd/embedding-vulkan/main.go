package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"mybilibili/embedding_vulkan/internal/embedding"
)

func main() {
	addr := getEnv("HTTP_ADDR", ":8081")

	llamaURL := os.Getenv("LLAMA_SERVER_URL")   // 外部 llama-server 地址
	llamaBin := os.Getenv("LLAMA_SERVER_BIN")     // llama-server 二进制路径
	modelPath := embedding.ResolveModelPath("")

	svc := embedding.NewService(llamaURL, llamaBin, modelPath)
	if err := svc.Start(context.Background()); err != nil {
		log.Printf("embedding: start warning: %v (proxy-only mode)", err)
	}
	defer svc.Stop()

	mux := http.NewServeMux()
	embedding.RegisterRoutes(mux, svc)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if svc.Health(r.Context()) {
			w.Write([]byte(`{"status":"ok"}`))
		} else {
			http.Error(w, `{"status":"unavailable"}`, 503)
		}
	})
	mux.HandleFunc("/api/v1/capabilities", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"service":"embedding","model":%q,"api":"openai-compatible"}`,
			getEnv("LLAMA_MODEL", "qwen3-embedding-0.6b"))
	})

	server := &http.Server{Addr: addr, Handler: mux}
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		server.Close()
	}()

	log.Printf("embedding-vulkan service listening on %s", addr)
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

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
