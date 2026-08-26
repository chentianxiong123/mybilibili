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

	"mybilibili/core-service/internal/admin"
	"mybilibili/core-service/internal/clients"
	"mybilibili/core-service/internal/comment"
	"mybilibili/core-service/internal/coreapi"
	"mybilibili/core-service/internal/events"
	"mybilibili/core-service/internal/favorite"
	"mybilibili/core-service/internal/manuscript"
	"mybilibili/core-service/internal/moderation"
	"mybilibili/core-service/internal/social"
	"mybilibili/core-service/internal/support"
	"mybilibili/core-service/internal/user"
	"mybilibili/core-service/internal/video"
	"mybilibili/pkg/abstraction"
	"mybilibili/pkg/auth"
	"mybilibili/pkg/common/middleware"
	pb "mybilibili/pkg/pb"
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

	interactionRepo := social.NewInteractionRepository(db)
	interactionSvc := social.NewInteractionService(interactionRepo)
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
	followH := social.NewFollowHandler(followSvc, db, auth.NewJWT(jwtSecret))
	interactionSvc.SetFollowService(followSvc)

	dynamicRepo := social.NewDynamicRepository(db)
	dynamicSvc := social.NewDynamicService(dynamicRepo)
	collectRepo := social.NewCollectionRepository(db)
	collectSvc := social.NewCollectionService(collectRepo)
	shareRepo := social.NewShareRepository(db)
	socialH := social.NewSocialHandler(followSvc, dynamicSvc, collectSvc, shareRepo, db, auth.NewJWT(jwtSecret))

	videoRepo := video.NewRepository(db)
	videoSvc := video.NewService(videoRepo)
	videoH := video.NewHandler(videoSvc)

	adminRepo := admin.NewRepository(db)
	adminSvc := admin.NewService(adminRepo)
	adminH := admin.NewHandler(adminSvc, auth.NewJWT(jwtSecret))
	if err := adminRepo.InitPermissions(context.Background()); err != nil {
		log.Printf("WARN: init permissions: %v", err)
	}
	scheduler := admin.NewScheduler(adminSvc)
	adminH.SetScheduler(scheduler)
	go scheduler.Run(context.Background())
	userAdminH := user.NewUserAdminHandler(db, adminSvc)
	videoAdminH := video.NewAdminHandler(db)
	commentAdminH := comment.NewAdminHandler(db)
	moderationAdminH := moderation.NewAdminHandler(db)

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
	creatorCommentH := comment.NewCreatorCommentHTTPHandler(commentRepo, commentSvc, db)
	favoriteH := favorite.NewFavoriteHandler(db)
	genericInteractionH := social.NewGenericInteractionHandler(interactionRepo)

	mqType := os.Getenv("MQ_TYPE")
	if mqType == "" {
		mqType = "nats"
	}
	mqPath := os.Getenv("MQ_PATH")
	if mqPath == "" {
		mqPath = "/tmp/mybilibili-mq"
	}
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://127.0.0.1:4222"
	}
	mq, err := abstraction.NewMessageQueue(abstraction.MessageQueueConfig{Type: mqType, Path: mqPath, NATSURL: natsURL})
	if err != nil {
		log.Printf("message queue %q unavailable (fallback to file): %v", mqType, err)
		mq, _ = abstraction.NewMessageQueue(abstraction.MessageQueueConfig{Type: "file", Path: mqPath})
	}

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

	manuscriptAdminH := manuscript.NewManuscriptAdminHandler(db)
	videoProcessAdminH := manuscript.NewVideoProcessAdminHandler(db)
	manuscriptAdminH.SetEventPublisher(eventPublisher)
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

	// 订阅 media worker 的处理进度，回写 videos 表（process_status/stage/progress/has_subtitle/has_summary）
	go func() {
		ch, err := mq.Subscribe(context.Background(), "video-process-progress-topic", "core-writer")
		if err != nil {
			log.Printf("progress subscribe: %v", err)
			return
		}
		for msg := range ch {
			var evt struct {
				VideoID      int64  `json:"video_id"`
				ManuscriptID int64  `json:"manuscript_id"`
				Stage        string `json:"stage"`
				StageText    string `json:"stage_text"`
				Progress     int32  `json:"progress"`
				Status       int32  `json:"status"`
				Done         bool   `json:"done"`
				Error        string `json:"error"`
				IsVertical   int32  `json:"is_vertical"`
			}
			json.Unmarshal(msg.Payload, &evt)
			if evt.VideoID == 0 {
				continue
			}
			// 转码时探测出的横竖屏方向（>=0 表示已设置）
			if evt.IsVertical >= 0 {
				_, _ = db.Exec(`UPDATE videos SET is_vertical = $1 WHERE id = $2`, evt.IsVertical, evt.VideoID)
			}
			stage := evt.Stage
			if stage == "failed" {
				_, _ = db.Exec(`UPDATE videos SET process_status = $1, process_stage = 'FAILED', process_error = $2, updated_at = NOW() WHERE id = $3`,
					evt.Status, evt.Error, evt.VideoID)
				continue
			}
			_, _ = db.Exec(`UPDATE videos SET process_status = $1, process_stage = $2, process_progress = $3, updated_at = NOW() WHERE id = $4`,
				evt.Status, stage, evt.Progress, evt.VideoID)
			if evt.Done {
				switch stage {
				case "subtitle":
					_, _ = db.Exec(`UPDATE videos SET has_subtitle = 1 WHERE id = $1`, evt.VideoID)
				case "summary":
					_, _ = db.Exec(`UPDATE videos SET has_summary = 1 WHERE id = $1`, evt.VideoID)
				}
			}
		}
	}()

	// 订阅同一进度主题（独立消费组），转发给转码流水线看板的 SSE 客户端
	go func() {
		ch, err := mq.Subscribe(context.Background(), "video-process-progress-topic", "admin-sse")
		if err != nil {
			log.Printf("admin-sse subscribe: %v", err)
			return
		}
		for msg := range ch {
			var evt struct {
				VideoID      int64  `json:"video_id"`
				ManuscriptID int64  `json:"manuscript_id"`
				Title        string `json:"title"`
				Stage        string `json:"stage"`
				StageText    string `json:"stage_text"`
				Progress     int32  `json:"progress"`
				Status       int32  `json:"status"`
				Done         bool   `json:"done"`
				Error        string `json:"error"`
			}
			json.Unmarshal(msg.Payload, &evt)
			if evt.VideoID == 0 {
				continue
			}
			videoProcessAdminH.Hub().Broadcast(manuscript.ProgressEvt(evt))
		}
	}()


	interactionSvc.SetProfileRecorder(clients.NewHTTPProfileRecorder())


	commentSvc.SetReviewService(aiClient)

	publicAPIH := comment.NewPublicAPIHandler(commentSvc, db)

	coreapi.StartHTTPServer(httpAddr, auth.NewJWT(jwtSecret),
		liveProxy, followH, socialH, videoH, adminH, modH,
		supportH, userExtH, favoriteH,
		creatorCommentH, manuscriptHTTPH, publicAPIH, genericInteractionH,
		userAdminH, manuscriptAdminH, videoAdminH, commentAdminH, moderationAdminH,
		videoProcessAdminH,
		social.NewSearchHistoryHandler(cacheStore, auth.NewJWT(jwtSecret)))

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
