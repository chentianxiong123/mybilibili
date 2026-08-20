package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"mybilibili/pkg/abstraction"
	pb "mybilibili/pkg/pb"
	"mybilibili/search-service/internal/analytics"
	"mybilibili/search-service/internal/profile"
	"mybilibili/search-service/internal/search"
)

func main() {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/mybilibili?sslmode=disable"
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8084"
	}

	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":9084"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	searchRepo := search.NewRepository(db)
	searchSvc := search.NewService(searchRepo)

	searchEngine, _ := abstraction.NewSearchEngine(abstraction.SearchEngineConfig{Type: "pg-fts", DSN: dsn})
	mq, _ := abstraction.NewMessageQueue(abstraction.MessageQueueConfig{Type: "memory"})

	searchH := search.NewHandler(searchSvc).WithEngine(searchEngine)
	indexMgr := search.NewIndexManager(searchEngine, mq)
	go indexMgr.Start(context.Background())

	analyticsRepo := analytics.NewRepository(db)
	analyticsSvc := analytics.NewService(analyticsRepo)
	analyticsH := analytics.NewHandler(analyticsSvc)

	docStore, _ := abstraction.NewDocumentStore(abstraction.DocumentStoreConfig{Type: "pg-jsonb", DSN: dsn})
	profileRepo := profile.NewRepository(docStore)
	profileSvc := profile.NewService(profileRepo)
	profileH := profile.NewHandler(profileSvc)

	go func() {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("failed to listen gRPC: %v", err)
		}
		srv := grpc.NewServer()
		pb.RegisterSearchServiceServer(srv, search.NewGrpcServer(searchSvc))
		reflection.Register(srv)
		log.Printf("Search gRPC listening on %s", grpcAddr)
		log.Fatal(srv.Serve(lis))
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"status":"ok"}`)) })
	searchH.Register(mux)
	analyticsH.Register(mux)
	profileH.Register(mux)
	log.Printf("Search HTTP listening on %s", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, mux))
}