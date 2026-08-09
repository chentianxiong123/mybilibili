package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mybilibili/internal/abstraction"
	"mybilibili/internal/media"
)

func main() {
	mqType := getEnv("MQ_TYPE", "memory")

	mq, err := abstraction.NewMessageQueue(abstraction.MessageQueueConfig{Type: mqType})
	if err != nil {
		log.Fatalf("message queue: %v", err)
	}
	defer mq.Close()

	storage, err := abstraction.NewStorageService(abstraction.StorageServiceConfig{Type: getEnv("STORAGE_TYPE", "memory")})
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	docStore, err := abstraction.NewDocumentStore(abstraction.DocumentStoreConfig{Type: getEnv("DOC_STORE_TYPE", "memory")})
	if err != nil {
		log.Fatalf("doc store: %v", err)
	}

	search, err := abstraction.NewSearchEngine(abstraction.SearchEngineConfig{Type: getEnv("SEARCH_TYPE", "memory")})
	if err != nil {
		log.Fatalf("search: %v", err)
	}

	workDir := getEnv("MEDIA_WORK_DIR", "/tmp/media-work")

	pipeline := media.NewPipeline(mq, storage, docStore, search, workDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
	}()

	log.Println("media service starting (independent FFmpeg pipeline)")
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
