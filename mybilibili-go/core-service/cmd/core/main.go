package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"mybilibili/core-service/internal/clients"
	"mybilibili/core-service/internal/comment"
	"mybilibili/core-service/internal/coreapi"
	"mybilibili/core-service/internal/events"
	"mybilibili/core-service/internal/favorite"
	"mybilibili/core-service/internal/interaction"
	"mybilibili/core-service/internal/manuscript"
	"mybilibili/core-service/internal/moderation"
	"mybilibili/core-service/internal/studio"
	"mybilibili/core-service/internal/subtitle"
	"mybilibili/core-service/internal/support"
	"mybilibili/core-service/internal/user"
	"mybilibili/core-service/internal/video"
	"mybilibili/pkg/abstraction"
	"mybilibili/pkg/auth"
	"mybilibili/pkg/common/middleware"
	pb "mybilibili/pkg/pb"
	"mybilibili/core-service/internal/social"
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

	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo, jwtSecret)
	userH := user.NewHandler(userSvc)

	manuscriptRepo := manuscript.NewManuscriptRepository(db)
	manuscriptSvc := manuscript.NewManuscriptService(manuscriptRepo, userRepo)
	manuscriptH := manuscript.NewManuscriptHandler(manuscriptSvc)

	commentRepo := comment.NewCommentRepository(db)
	commentSvc := comment.NewCommentService(commentRepo)
	commentH := comment.NewCommentHandler(commentSvc)

	interactionRepo := interaction.NewInteractionRepository(db)
	interactionSvc := interaction.NewInteractionService(interactionRepo)
	interactionH := interactionSvc

	msgDanmakuClient, err := clients.NewMsgDanmakuClient()
	if err != nil {
		log.Printf("msg-danmaku gRPC client unavailable: %v", err)
		msgDanmakuClient = nil
	}
	defer func() {
		if msgDanmakuClient != nil {
			msgDanmakuClient.Close()
		}
	}()

	commentSvc.SetDB(db)
	commentSvc.SetNotifier(msgDanmakuClient)
	interactionSvc.SetDB(db)
	interactionSvc.SetNotifier(msgDanmakuClient)

	liveProxy := coreapi.NewLiveProxy()

	followRepo := social.NewFollowRepository(db)
	followSvc := social.NewFollowService(followRepo)
	followH := social.NewFollowHandler(followSvc)

	dynamicRepo := social.NewDynamicRepository(db)
	dynamicSvc := social.NewDynamicService(dynamicRepo)
	collectRepo := social.NewCollectionRepository(db)
	collectSvc := social.NewCollectionService(collectRepo)
	shareRepo := social.NewShareRepository(db)
	socialH := social.NewSocialHandler(followSvc, dynamicSvc, collectSvc, shareRepo)

	videoRepo := video.NewRepository(db)
	videoSvc := video.NewService(videoRepo)
	videoH := video.NewHandler(videoSvc)

	modRepo := moderation.NewRepository(db)
	modSvc := moderation.NewService(modRepo)
	modH := moderation.NewHandler(modSvc)

	aiClient, err := clients.NewAiClient()
	if err != nil {
		log.Printf("ai gRPC client unavailable: %v", err)
		aiClient = nil
	}
	defer func() {
		if aiClient != nil {
			aiClient.Close()
		}
	}()

	searchClient, err := clients.NewSearchClient()
	if err != nil {
		log.Printf("search gRPC client unavailable: %v", err)
		searchClient = nil
	}
	defer func() {
		if searchClient != nil {
			searchClient.Close()
		}
	}()

	supportRepo := support.NewRepository(db)
	supportSvc := support.NewService(supportRepo)
	supportH := support.NewHandler(supportSvc)

	userExtH := user.NewUserExtendHandler(userSvc)
	manuscriptHTTPH := manuscript.NewManuscriptHTTPHandler(db, manuscriptSvc, commentSvc, interactionSvc)
	creatorCommentH := comment.NewCreatorCommentHTTPHandler(commentRepo, commentSvc)
	favoriteH := favorite.NewFavoriteHandler(db)
	genericInteractionH := interaction.NewGenericInteractionHandler(interactionRepo)

	docStore, _ := abstraction.NewDocumentStore(abstraction.DocumentStoreConfig{Type: "pg-jsonb", DSN: dsn})
	mq, _ := abstraction.NewMessageQueue(abstraction.MessageQueueConfig{Type: "memory"})

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	cacheStore, err := abstraction.NewCacheStore(abstraction.CacheStoreConfig{Type: "redis", Addr: redisAddr})
	if err != nil {
		log.Printf("redis cache unavailable (fallback to memory): %v", err)
		cacheStore, _ = abstraction.NewCacheStore(abstraction.CacheStoreConfig{Type: "memory"})
	}
	userSvc.SetCacheStore(cacheStore)
	manuscriptSvc.SetCacheStore(cacheStore)
	commentSvc.SetCacheStore(cacheStore)

	eventPublisher := events.NewEventPublisher(mq)

	interactionSvc.SetEventPublisher(eventPublisher)

	go func() {
		ch, err := mq.Subscribe(context.Background(), "video-publish-topic", "auto-publish")
		if err != nil {
			log.Printf("auto-publish subscribe: %v", err)
			return
		}
		for msg := range ch {
			var evt struct {
				ManuscriptID int64 `json:"manuscript_id"`
				VideoID      int64 `json:"video_id"`
			}
			json.Unmarshal(msg.Payload, &evt)
			if evt.ManuscriptID == 0 {
				continue
			}
			var remaining int
			_ = db.QueryRow(`SELECT COUNT(*) FROM videos WHERE manuscript_id = $1 AND process_status != 5`, evt.ManuscriptID).Scan(&remaining)
			if remaining == 0 {
				_, _ = db.Exec(`UPDATE manuscripts SET status = 3 WHERE id = $1 AND status = 0`, evt.ManuscriptID)
				log.Printf("auto-published manuscript %d (all videos completed)", evt.ManuscriptID)
			}
		}
	}()

	subtitleRepo := subtitle.NewRepository(docStore)
	subtitleSvc := subtitle.NewService(subtitleRepo)
	subtitleH := subtitle.NewHandler(subtitleSvc)

	interactionSvc.SetProfileRecorder(interaction.NewHTTPProfileRecorder())

	studioRepo := studio.NewRepository(db)
	studioSvc := studio.NewService(studioRepo)
	studioH := studio.NewHandler(studioSvc)

	commentSvc.SetReviewService(aiClient)

	publicAPIH := coreapi.NewPublicAPIHandler(commentSvc)

	coreapi.StartHTTPServer(httpAddr, auth.NewJWT(jwtSecret),
		liveProxy, followH, socialH, videoH, modH,
		supportH, userExtH, favoriteH,
		subtitleH, studioH, creatorCommentH, manuscriptHTTPH, publicAPIH, genericInteractionH)

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
