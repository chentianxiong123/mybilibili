package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"mybilibili/pkg/abstraction"
	"mybilibili/pkg/auth"
	pb "mybilibili/pkg/pb"
	"mybilibili/search-service/internal/analytics"
	"mybilibili/search-service/internal/hot"
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

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(ctxBackground()).Err(); err != nil {
		log.Printf("warning: failed to connect redis at %s: %v", redisAddr, err)
	} else {
		log.Printf("connected to redis at %s", redisAddr)
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
	hotRepo := hot.NewRepository(rdb)
	searchSvc := search.NewService(searchRepo, hotRepo)

	searchH := search.NewHandler(searchSvc)

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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"status":"ok"}`)) })
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"status":"ok"}`)) })
	searchH.Register(mux)
	analyticsH.Register(mux)
	profileH.Register(mux)
	log.Printf("Search HTTP listening on %s", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, auth.IdentityMiddleware(auth.NewJWT(jwtSecret))(mux)))
}

func ctxBackground() context.Context {
	return context.Background()
}