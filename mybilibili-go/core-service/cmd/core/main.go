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

	"mybilibili/core-service/internal/ai"
	"mybilibili/core-service/internal/admin"
	"mybilibili/core-service/internal/analytics"
	"mybilibili/core-service/internal/comment"
	"mybilibili/core-service/internal/coreapi"
	"mybilibili/core-service/internal/danmaku"
	"mybilibili/core-service/internal/events"
	"mybilibili/core-service/internal/favorite"
	"mybilibili/core-service/internal/interaction"
	"mybilibili/core-service/internal/manuscript"
	"mybilibili/core-service/internal/live"
	"mybilibili/core-service/internal/meeting"
	"mybilibili/core-service/internal/message"
	"mybilibili/core-service/internal/moderation"
	"mybilibili/core-service/internal/profile"
	"mybilibili/core-service/internal/studio"
	"mybilibili/core-service/internal/subtitle"
	"mybilibili/core-service/internal/support"
	"mybilibili/core-service/internal/user"
	"mybilibili/core-service/internal/video"
	"mybilibili/pkg/abstraction"
	"mybilibili/pkg/common/middleware"
	pb "mybilibili/pkg/pb"
	"mybilibili/core-service/internal/search"
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

	danmakuRepo := danmaku.NewDanmakuRepository(db)
	danmakuBroadcaster := danmaku.NewDanmakuBroadcaster()
	danmakuSvc := danmaku.NewDanmakuService(danmakuRepo, danmakuBroadcaster)

	messageRepo := message.NewMessageRepository(db)
	notifBroadcaster := message.NewNotificationBroadcaster()
	commentSvc.SetMessageRepo(messageRepo)
	interactionSvc.SetMessageRepo(messageRepo)

	liveRepo := live.NewRepository(db)
	liveHub := live.NewHub()
	liveSvc := live.NewService(liveRepo, liveHub)
	liveH := live.NewHTTPHandler(liveSvc, liveHub)

	linkmicRepo := live.NewLinkmicRepository(db)
	linkmicSvc := live.NewLinkmicService(linkmicRepo, liveHub, liveRepo)
	linkmicH := live.NewLinkmicHandler(linkmicSvc)

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

	adminRepo := admin.NewRepository(db)
	adminSvc := admin.NewService(adminRepo)
	adminH := admin.NewHandler(adminSvc)
	adminDataH := admin.NewAdminDataHandler(db)

	modRepo := moderation.NewRepository(db)
	modSvc := moderation.NewService(modRepo)
	modH := moderation.NewHandler(modSvc)

	meetingRepo := meeting.NewRepository(db)
	meetingSvc := meeting.NewService(meetingRepo)
	meetingH := meeting.NewHandler(meetingSvc)

	aiRepo := ai.NewRepository(db)
	aiSvc := ai.NewService(aiRepo)

	caller, _ := abstraction.NewServiceCaller(abstraction.ServiceCallerConfig{Type: "ollama"})
	summarySvc := ai.NewSummaryService(caller)
	reviewSvc := ai.NewReviewService(caller)
	customerSvc := ai.NewCustomerService(caller)

	aiH := ai.NewHandler(aiSvc).WithSummary(summarySvc).WithReview(reviewSvc).WithCustomer(customerSvc)

	searchRepo := search.NewRepository(db)
	searchSvc := search.NewService(searchRepo)

	supportRepo := support.NewRepository(db)
	supportSvc := support.NewService(supportRepo)
	supportH := support.NewHandler(supportSvc)

	userExtH := user.NewUserExtendHandler(userSvc)
	manuscriptHTTPH := manuscript.NewManuscriptHTTPHandler(db, manuscriptSvc, commentSvc, interactionSvc)
	messageH := message.NewMessageHTTPHandler(messageRepo, notifBroadcaster)
	creatorCommentH := comment.NewCreatorCommentHTTPHandler(commentRepo, commentSvc)
	favoriteH := favorite.NewFavoriteHandler(db)
	genericInteractionH := interaction.NewGenericInteractionHandler(interactionRepo)

	docStore, _ := abstraction.NewDocumentStore(abstraction.DocumentStoreConfig{Type: "pg-jsonb", DSN: dsn})
	searchEngine, _ := abstraction.NewSearchEngine(abstraction.SearchEngineConfig{Type: "pg-fts", DSN: dsn})
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
	summarySvc.SetCacheStore(cacheStore)

	searchH := search.NewHandler(searchSvc).WithEngine(searchEngine)

	eventPublisher := events.NewEventPublisher(mq)

	adminManuscriptH := admin.NewManuscriptAdminHandler(db)
	adminManuscriptH.SetEventPublisher(eventPublisher)
	interactionSvc.SetEventPublisher(eventPublisher)

	indexMgr := search.NewIndexManager(searchEngine, mq)
	go indexMgr.Start(context.Background())

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

	analyticsRepo := analytics.NewRepository(db)
	analyticsSvc := analytics.NewService(analyticsRepo)
	analyticsH := analytics.NewHandler(analyticsSvc)

	profileRepo := profile.NewRepository(docStore)
	profileSvc := profile.NewService(profileRepo)
	profileH := profile.NewHandler(profileSvc)
	interactionSvc.SetProfileRecorder(profileSvc)

	studioRepo := studio.NewRepository(db)
	studioSvc := studio.NewService(studioRepo)
	studioH := studio.NewHandler(studioSvc)

	commentSvc.SetReviewService(reviewSvc)
	aiChatH := ai.NewAIChatHandler(reviewSvc, customerSvc)

	publicAPIH := coreapi.NewPublicAPIHandler(commentSvc)

	httpH := coreapi.NewHTTPHandler(danmakuSvc, messageRepo, notifBroadcaster)
	coreapi.StartHTTPServer(httpAddr, httpH, user.NewJWT(jwtSecret),
		liveH, linkmicH, followH, socialH, videoH, adminH, adminDataH, adminManuscriptH, modH,
		meetingH, aiH, searchH, supportH, userExtH, messageH, favoriteH,
		subtitleH, analyticsH, studioH, profileH, creatorCommentH, aiChatH, manuscriptHTTPH, publicAPIH, genericInteractionH)

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
