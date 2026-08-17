package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	storage, err := abstraction.NewStorageService(abstraction.StorageServiceConfig{Type: getEnv("STORAGE_TYPE", "memory")})
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

	pipeline := work.NewPipeline(mq, storage, docStore, search, workDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
	}()

	log.Println("work service starting (FFmpeg pipeline)")
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
