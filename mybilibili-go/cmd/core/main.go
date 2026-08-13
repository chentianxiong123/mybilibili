package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"mybilibili/internal/abstraction"
	"mybilibili/internal/admin"
	"mybilibili/internal/ai"
	"mybilibili/internal/analytics"
	"mybilibili/internal/common/middleware"
	"mybilibili/internal/core"
	pb "mybilibili/internal/core/pb"
	"mybilibili/internal/live"
	"mybilibili/internal/meeting"
	"mybilibili/internal/moderation"
	"mybilibili/internal/profile"
	"mybilibili/internal/search"
	"mybilibili/internal/social"
	"mybilibili/internal/studio"
	"mybilibili/internal/subtitle"
	"mybilibili/internal/support"
	"mybilibili/internal/video"
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
	aiH := ai.NewHandler(aiSvc)

	searchRepo := search.NewRepository(db)
	searchSvc := search.NewService(searchRepo)

	supportRepo := support.NewRepository(db)
	supportSvc := support.NewService(supportRepo)
	supportH := support.NewHandler(supportSvc)

	userExtH := core.NewUserExtendHandler(userSvc)
	manuscriptHTTPH := core.NewManuscriptHTTPHandler(db, manuscriptSvc, commentSvc, interactionSvc)
	messageH := core.NewMessageHTTPHandler(messageRepo, notifBroadcaster)
	creatorCommentH := core.NewCreatorCommentHTTPHandler(commentRepo, commentSvc)
	favoriteH := core.NewFavoriteHandler(db)
	adminManuscriptH := admin.NewManuscriptAdminHandler(db)

	abstractionCfg := abstraction.Config{}
	docStore, _ := abstraction.NewDocumentStore(abstraction.DocumentStoreConfig{Type: "memory"})
	searchEngine, _ := abstraction.NewSearchEngine(abstraction.SearchEngineConfig{Type: "memory"})
	storageSvc, _ := abstraction.NewStorageService(abstraction.StorageServiceConfig{Type: "memory"})
	caller, _ := abstraction.NewServiceCaller(abstraction.ServiceCallerConfig{Type: "memory"})
	mq, _ := abstraction.NewMessageQueue(abstraction.MessageQueueConfig{Type: "memory"})
	_ = abstractionCfg
	_ = storageSvc

	searchH := search.NewHandler(searchSvc).WithEngine(searchEngine)

	eventPublisher := core.NewEventPublisher(mq)
	_ = eventPublisher

	indexMgr := search.NewIndexManager(searchEngine, mq)
	go indexMgr.Start(context.Background())

	subtitleRepo := subtitle.NewRepository(docStore)
	subtitleSvc := subtitle.NewService(subtitleRepo)
	subtitleH := subtitle.NewHandler(subtitleSvc)

	analyticsRepo := analytics.NewRepository(db)
	analyticsSvc := analytics.NewService(analyticsRepo)
	analyticsH := analytics.NewHandler(analyticsSvc)

	profileRepo := profile.NewRepository(docStore)
	profileSvc := profile.NewService(profileRepo)
	profileH := profile.NewHandler(profileSvc)

	studioRepo := studio.NewRepository(db)
	studioSvc := studio.NewService(studioRepo)
	studioH := studio.NewHandler(studioSvc)

	summarySvc := ai.NewSummaryService(caller)
	reviewSvc := ai.NewReviewService(caller)
	customerSvc := ai.NewCustomerService(caller)
	_ = summarySvc
	aiChatH := ai.NewAIChatHandler(reviewSvc, customerSvc)

	publicAPIH := core.NewPublicAPIHandler(commentSvc)

	httpH := core.NewHTTPHandler(danmakuSvc, messageRepo, notifBroadcaster)
	core.StartHTTPServer(httpAddr, httpH, core.NewJWT(jwtSecret),
		liveH, linkmicH, followH, socialH, videoH, adminH, adminDataH, adminManuscriptH, modH,
		meetingH, aiH, searchH, supportH, userExtH, messageH, favoriteH,
		subtitleH, analyticsH, studioH, profileH, creatorCommentH, aiChatH, manuscriptHTTPH, publicAPIH)

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
