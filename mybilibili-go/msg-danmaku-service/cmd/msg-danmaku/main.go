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

	"mybilibili/msg-danmaku-service/internal/danmaku"
	"mybilibili/msg-danmaku-service/internal/message"
	pb "mybilibili/pkg/pb"
)

func main() {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/mybilibili?sslmode=disable"
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8086"
	}

	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":9086"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	danmakuRepo := danmaku.NewDanmakuRepository(db)
	danmakuBroadcaster := danmaku.NewDanmakuBroadcaster()
	danmakuSvc := danmaku.NewDanmakuService(danmakuRepo, danmakuBroadcaster)
	danmakuH := danmaku.NewHTTPHandler(danmakuSvc, danmakuBroadcaster)

	messageRepo := message.NewMessageRepository(db)
	notifBroadcaster := message.NewNotificationBroadcaster()
	messageH := message.NewMessageHTTPHandler(messageRepo, notifBroadcaster)

	go func() {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("failed to listen gRPC: %v", err)
		}
		srv := grpc.NewServer()
		pb.RegisterMsgDanmakuServiceServer(srv, message.NewGrpcServer(messageRepo, notifBroadcaster))
		reflection.Register(srv)
		log.Printf("MsgDanmaku gRPC listening on %s", grpcAddr)
		log.Fatal(srv.Serve(lis))
	}()

	mux := http.NewServeMux()
	danmakuH.Register(mux)
	messageH.Register(mux)
	log.Printf("MsgDanmaku HTTP listening on %s", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, mux))
}