package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"mybilibili/internal/common/middleware"
	"mybilibili/internal/core"
	pb "mybilibili/internal/core/pb"
	"mybilibili/internal/live"
)

func main() {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/mybilibili?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
	}

	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":9090"
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("connected to database")

	userRepo := core.NewRepository(db)
	userSvc := core.NewService(userRepo, jwtSecret)
	userH := core.NewHandler(userSvc)

	manuscriptRepo := core.NewManuscriptRepository(db)
	manuscriptSvc := core.NewManuscriptService(manuscriptRepo, userRepo)
	manuscriptH := core.NewManuscriptHandler(manuscriptSvc)

	commentRepo := core.NewCommentRepository(db)
	commentSvc := core.NewCommentService(commentRepo)
	commentH := core.NewCommentHandler(commentSvc)

	interactionRepo := core.NewInteractionRepository(db)
	interactionSvc := core.NewInteractionService(interactionRepo)
	interactionH := core.NewInteractionHandler(interactionSvc)

	danmakuRepo := core.NewDanmakuRepository(db)
	danmakuBroadcaster := core.NewDanmakuBroadcaster()
	danmakuSvc := core.NewDanmakuService(danmakuRepo, danmakuBroadcaster)

	messageRepo := core.NewMessageRepository(db)
	notifBroadcaster := core.NewNotificationBroadcaster()

	liveRepo := live.NewRepository(db)
	liveHub := live.NewHub()
	liveSvc := live.NewService(liveRepo, liveHub)
	liveH := live.NewHTTPHandler(liveSvc, liveHub)

	httpH := core.NewHTTPHandler(danmakuSvc, messageRepo, notifBroadcaster)
	core.StartHTTPServer(httpAddr, httpH, liveH)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.Recovery,
			middleware.Logging,
			middleware.Timeout(10*time.Second),
		),
	)

	pb.RegisterUserServiceServer(srv, userH)
	pb.RegisterManuscriptServiceServer(srv, manuscriptH)
	pb.RegisterCommentServiceServer(srv, commentH)
	pb.RegisterInteractionServiceServer(srv, interactionH)
	reflection.Register(srv)

	log.Printf("gRPC listening on %s, HTTP/SSE listening on %s", grpcAddr, httpAddr)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
