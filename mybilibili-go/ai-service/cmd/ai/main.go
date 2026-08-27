package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"mybilibili/ai-service/internal/ai"
	"mybilibili/ai-service/internal/subtitle"
	"mybilibili/pkg/abstraction"
	pb "mybilibili/pkg/pb"
)

func main() {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/mybilibili?sslmode=disable"
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8088"
	}

	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":9088"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	aiRepo := ai.NewRepository(db)
	aiSvc := ai.NewService(aiRepo)

	docStore, _ := abstraction.NewDocumentStore(abstraction.DocumentStoreConfig{Type: "pg-jsonb", DSN: dsn})
	subtitleRepo := subtitle.NewRepository(docStore)
	subtitleSvc := subtitle.NewService(subtitleRepo)
	subtitleH := subtitle.NewHandler(subtitleSvc)
	caller, _ := abstraction.NewServiceCaller(abstraction.ServiceCallerConfig{Type: "ollama"})
	summarySvc := ai.NewSummaryService(caller)
	summarySvc.SetDatabase(db)
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "127.0.0.1:9000"
	}
	// minio-go 不接受带 scheme 的 endpoint，剥掉 http(s):// 前缀。
	minioEndpoint = stripScheme(minioEndpoint)
	if storage, serr := abstraction.NewMinioStorageService(abstraction.MinioConfig{
		Endpoint:   minioEndpoint,
		AccessKey:  getEnvDefault("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey:  getEnvDefault("MINIO_SECRET_KEY", "minioadmin"),
		BucketName: "mybilibili",
	}); serr == nil {
		summarySvc.SetStorage(storage)
		subtitleH.SetGenerator(subtitle.NewWhisperGenerator(subtitleRepo, storage))
	} else {
		log.Printf("WARN: minio storage unavailable, summaries fallback to live generation: %v", serr)
	}
	reviewSvc := ai.NewReviewService(caller)
	customerSvc := ai.NewCustomerService(caller)

	aiH := ai.NewHandler(aiSvc).WithSummary(summarySvc).WithReview(reviewSvc).WithCustomer(customerSvc)
	aiChatH := ai.NewAIChatHandler(reviewSvc, customerSvc)

	go func() {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("failed to listen gRPC: %v", err)
		}
		srv := grpc.NewServer()
		pb.RegisterAiServiceServer(srv, ai.NewGrpcServer(reviewSvc, summarySvc))
		reflection.Register(srv)
		log.Printf("AI gRPC listening on %s", grpcAddr)
		log.Fatal(srv.Serve(lis))
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})
	subtitleH.Register(mux)
	aiH.Register(mux)
	aiChatH.Register(mux)
	log.Printf("AI HTTP listening on %s", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, mux))
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// stripScheme 去掉 http(s):// 前缀，供 minio-go 客户端使用。
func stripScheme(s string) string {
	for _, p := range []string{"https://", "http://"} {
		if len(s) > len(p) && s[:len(p)] == p {
			return s[len(p):]
		}
	}
	return s
}