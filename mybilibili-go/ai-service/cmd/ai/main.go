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

	caller, _ := abstraction.NewServiceCaller(abstraction.ServiceCallerConfig{Type: "ollama"})
	summarySvc := ai.NewSummaryService(caller)
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
	aiH.Register(mux)
	aiChatH.Register(mux)
	log.Printf("AI HTTP listening on %s", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, mux))
}