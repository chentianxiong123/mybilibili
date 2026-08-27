package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"mybilibili/pkg/abstraction"
	"mybilibili/work-service/internal/work"
)

func main() {
	mqType := getEnv("MQ_TYPE", "nats")
	mqPath := getEnv("MQ_PATH", "/tmp/mybilibili-mq")
	natsURL := getEnv("NATS_URL", "nats://127.0.0.1:4222")

	mq, err := abstraction.NewMessageQueue(abstraction.MessageQueueConfig{Type: mqType, Path: mqPath, NATSURL: natsURL})
	if err != nil {
		log.Printf("message queue %q unavailable (fallback to file): %v", mqType, err)
		mq, err = abstraction.NewMessageQueue(abstraction.MessageQueueConfig{Type: "file", Path: mqPath})
		if err != nil {
			log.Fatalf("message queue fallback: %v", err)
		}
	}
	defer mq.Close()

	// 转码产物/音频/字幕/摘要统一持久化到 MinIO（对齐老项目存储链路）；
	// factory 的 minio 分支是未实现桩，这里直接使用真实实现 NewMinioStorageService。
	// 注意：本仓库 minio-go 不接受带 scheme 的 endpoint（会报 fully qualified paths），
	// 与 ai-service 一致使用 主机:端口 形式。
	minioCfg := abstraction.DefaultMinioConfig()
	minioCfg.Endpoint = "127.0.0.1:9000"
	minioCfg.PublicEndpoint = "127.0.0.1:9000"
	if v := os.Getenv("MINIO_ENDPOINT"); v != "" {
		// minio-go 不接受带 scheme 的 endpoint，剥掉 http(s):// 前缀。
		minioCfg.Endpoint = stripScheme(v)
		minioCfg.PublicEndpoint = stripScheme(v)
	}
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

	docStore, err := abstraction.NewDocumentStore(abstraction.DocumentStoreConfig{
		Type: getEnv("DOC_STORE_TYPE", "memory"),
		DSN:  getEnv("PG_DSN", ""),
	})
	if err != nil {
		log.Fatalf("doc store: %v", err)
	}

	search, err := abstraction.NewSearchEngine(abstraction.SearchEngineConfig{Type: getEnv("SEARCH_TYPE", "memory")})
	if err != nil {
		log.Fatalf("search: %v", err)
	}

	workDir := getEnv("WORK_DIR", "/tmp/work")
	transcoderBase := getEnv("TRANSCODER_ADDR", "http://127.0.0.1:8092")
	transcoderClient := work.NewTranscoderClient(transcoderBase)

	pipeline := work.NewPipeline(mq, storage, docStore, search, transcoderClient, workDir)
	if dsn := getEnv("PG_DSN", ""); dsn != "" {
		if db, err := sql.Open("postgres", dsn); err == nil {
			pipeline.SetDatabase(db)
			defer db.Close()
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
	}()

	log.Println("work service starting (orchestrator)")
	if err := pipeline.Start(ctx); err != nil {
		log.Fatalf("pipeline: %v", err)
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
