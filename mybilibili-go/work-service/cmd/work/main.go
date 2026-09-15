package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// transcoder 池: 从 YAML 加载配置 (文件不存在则空池启动), 后台 goroutine 周期性探活
	cfgPath := getEnv("WORK_CONFIG_PATH", "/etc/mybilibili/transcoders.yaml")
	pool, err := work.NewTranscoderPool(cfgPath)
	if err != nil {
		log.Fatalf("transcoder pool: %v", err)
	}
	log.Printf("transcoder pool loaded: %d nodes from %s", len(pool.List()), cfgPath)

	aiBase := getEnv("AI_ADDR", "http://127.0.0.1:8088")
	aiClient := work.NewAIClient(aiBase)

	pipeline := work.NewPipeline(mq, storage, docStore, search, pool, aiClient, workDir)
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

	// 启动 transcoder 池后台健康检查
	pool.Start(ctx)
	defer pool.Stop()

	// 启动 admin HTTP server (transcoder 节点管理 API)
	adminAddr := getEnv("ADMIN_HTTP_ADDR", ":8090")
	adminAPI := work.NewAdminAPI(pool)
	adminServer := &http.Server{
		Addr:              adminAddr,
		Handler:           adminAPI.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("work admin API listening on %s", adminAddr)
		if err := adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("admin server: %v", err)
		}
	}()
	defer func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutCancel()
		adminServer.Shutdown(shutCtx)
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
