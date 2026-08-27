package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"mybilibili/msg-danmaku-service/internal/danmaku"
	"mybilibili/msg-danmaku-service/internal/message"
	"mybilibili/pkg/auth"
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

	// Redis 未读红点缓存（DB 仍为源真相）
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()
	unreadCache := message.NewUnreadCache(rdb, messageRepo, 0)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
	}
	jwt := auth.NewJWT(jwtSecret)

	messageH := message.NewMessageHTTPHandler(messageRepo, notifBroadcaster, unreadCache, jwt)

	go func() {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("failed to listen gRPC: %v", err)
		}
		srv := grpc.NewServer()
		pb.RegisterMsgDanmakuServiceServer(srv, message.NewGrpcServer(messageRepo, notifBroadcaster, unreadCache))
		reflection.Register(srv)
		log.Printf("MsgDanmaku gRPC listening on %s", grpcAddr)
		log.Fatal(srv.Serve(lis))
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"status":"ok"}`)) })
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"status":"ok"}`)) })
	danmakuH.Register(mux)
	messageH.Register(mux)
	var handler http.Handler = mux
	handler = auth.IdentityMiddleware(jwt)(mux)
	log.Printf("MsgDanmaku HTTP listening on %s", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, handler))
}