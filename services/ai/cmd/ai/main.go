package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"mybilibili/ai/internal/ai"
	"mybilibili/ai/internal/subtitle"
	"mybilibili/pkg/abstraction"
	"mybilibili/pkg/auth"
	"mybilibili/pkg/common/logging"
	pb "mybilibili/pkg/pb"
)

func main() {
	// 统一日志: LOG_DIR 未设置则 stdout(dev 不落盘), 设置则写 <LOG_DIR>/ai.log(部署落盘)。
	logWriter := logging.New(logging.Options{Name: "ai"})
	defer logWriter.Close()
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

	docStore, _ := abstraction.NewDocumentStore(abstraction.DocumentStoreConfig{Type: "pg-jsonb", DSN: dsn})
	subtitleRepo := subtitle.NewRepository(docStore)
	subtitleSvc := subtitle.NewService(subtitleRepo)
	subtitleH := subtitle.NewHandler(subtitleSvc)
	// FINDING-067 修复：视频归属校验（Upload 此前接受任意 videoID）。
	subtitleH.SetVideoOwnerCheck(func(ctx context.Context, videoID, userID int64) (bool, error) {
		var owner int64
		err := db.QueryRowContext(ctx,
			`SELECT m.user_id FROM videos v JOIN manuscripts m ON m.id = v.manuscript_id WHERE v.id = $1`,
			videoID).Scan(&owner)
		if err != nil {
			return false, nil // 视频不存在 → 拒绝
		}
		return owner == userID, nil
	})
	// 真实 LLM 调用器（langchaingo）：端点/密钥/模型从 env 读取；
	// 未配置时 Call 返回显式 "AI LLM not configured"（不占位/不假装成功）。
	caller, cerr := ai.NewLangchaingoCaller(ai.LangchaingoConfig{})
	if cerr != nil {
		log.Printf("ai: langchaingo caller init failed: %v", cerr)
	}
	if caller == nil {
		// 极端情况：构造失败仍提供调用器（其 Call 会返回未配置错误）。
		caller, _ = abstraction.NewServiceCaller(abstraction.ServiceCallerConfig{Type: "ollama"})
	}
	summarySvc := ai.NewSummaryService(caller)
	summarySvc.SetDatabase(db)
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "127.0.0.1:9000"
	}
	// minio-go 不接受带 scheme 的 endpoint，剥掉 http(s):// 前缀。
	minioEndpoint = stripScheme(minioEndpoint)
	if storage, serr := abstraction.NewMinioStorageService(abstraction.MinioConfig{
		Endpoint:   minioEndpoint,
		AccessKey:  getEnvDefault("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey:  getEnvDefault("MINIO_SECRET_KEY", "minioadmin"),
		BucketName: "mybilibili",
		// FINDING-015 修复：UseSSL 独立开关（端点裸化后 HTTPS 无法由 scheme 前缀判定）。
		UseSSL: getEnvDefault("MINIO_USE_SSL", "") == "true",
	}); serr == nil {
		summarySvc.SetStorage(storage)
		subtitleH.SetGenerator(subtitle.NewWhisperGenerator(subtitleRepo, storage))
	} else {
		log.Printf("WARN: minio storage unavailable, summaries fallback to live generation: %v", serr)
	}
	reviewSvc := ai.NewReviewService(caller)
	customerSvc := ai.NewCustomerService(caller)

	// 真实写入 AI 用量日志（FINDING-085：ai_usage_logs 此前从未写入，用量端点恒 0）。
	ai.SetUsageLog(caller, func(feature string, userID int64, tokenCount, durationMs int) {
		if err := aiRepo.LogUsage(context.Background(), feature, userID, os.Getenv("AI_LLM_MODEL"), tokenCount, durationMs); err != nil {
			log.Printf("WARN: ai usage log: %v", err)
		}
	})

	aiH := ai.NewHandler(aiSvc).WithSummary(summarySvc).WithReview(reviewSvc).WithCustomer(customerSvc)

	// 多轮客服 agent（langchaingo conversational agent + 工具）：
	// 复用同一 langchaingo 模型；未配置真实 LLM 时 agent.Chat 返回显式"未配置"错误。
	if model := ai.ModelFromCaller(caller); model != nil {
		// 语义技能检索：配置 AI_PGVECTOR_URL 且模型支持嵌入时启用 pgvector，否则回退关键词。
		skillIndex, _ := ai.NewSkillIndex(model, os.Getenv("AI_PGVECTOR_URL"))
		if skillIndex != nil {
			log.Printf("ai: 技能语义检索已启用（langchaingo pgvector + embeddings）")
		} else {
			log.Printf("WARN: 技能检索回退关键词匹配（未配置 AI_PGVECTOR_URL 或模型不支持嵌入）")
		}
		customerAgent := ai.NewCustomerAgent(model, ai.CustomerTools(&ai.CustomerToolEnv{
			LookupSkill: func(ctx context.Context, query string) (string, error) {
				if skillIndex != nil {
					docs, err := skillIndex.Search(ctx, query, 5)
					if err == nil && len(docs) > 0 {
						out := make([]map[string]any, 0, len(docs))
						for _, d := range docs {
							out = append(out, map[string]any{"content": d.PageContent, "metadata": d.Metadata})
						}
						b, _ := json.Marshal(out)
						return string(b), nil
					}
				}
				// 关键词回退。
				skills, err := aiRepo.ListSkills(ctx)
				if err != nil {
					return "", err
				}
				out := make([]map[string]any, 0, len(skills))
				q := strings.ToLower(strings.TrimSpace(query))
				for _, s := range skills {
					if q == "" || strings.Contains(strings.ToLower(s.Name+" "+s.Description), q) {
						out = append(out, map[string]any{"name": s.Name, "description": s.Description, "type": s.Type})
					}
				}
				b, _ := json.Marshal(out)
				return string(b), nil
			},
			CreateTicket:  func(ctx context.Context, userID int64, content string) (string, error) { return aiRepo.CreateTicket(ctx, userID, content) },
			TransferHuman: func(ctx context.Context, userID int64) (string, error) { return aiRepo.CreateTicket(ctx, userID, "用户申请转人工客服") },
		}))
		// 统一用量管理：agent 与调用器共享同一 ai_usage_logs 写入回调。
		customerAgent.SetUsageLog(func(feature string, userID int64, tokenCount, durationMs int) {
			if err := aiRepo.LogUsage(context.Background(), feature, userID, os.Getenv("AI_LLM_MODEL"), tokenCount, durationMs); err != nil {
				log.Printf("WARN: ai usage log (agent): %v", err)
			}
		})
		customerSvc.SetAgent(customerAgent)
		log.Printf("ai: 多轮客服 agent 已启用（langchaingo conversational agent + 工具）")
		// 管理后台 AI 助手（FINDING-088）：多轮工具调用 agent，产出表格/指标卡渲染。
		adminAssistant := ai.NewAdminAssistant(model, ai.AdminTools(&ai.AdminToolEnv{
			Statistics:      func(ctx context.Context) (map[string]any, error) { return aiRepo.GetAdminStatistics(ctx) },
			ListManuscripts: func(ctx context.Context, q string, limit int) ([]map[string]any, error) { return aiRepo.ListAdminManuscripts(ctx, q, limit) },
			ListUsers:       func(ctx context.Context, q string, limit int) ([]map[string]any, error) { return aiRepo.ListAdminUsers(ctx, q, limit) },
			ListTickets:     func(ctx context.Context, limit int) ([]map[string]any, error) { return aiRepo.ListAdminTickets(ctx, limit) },
			GetManuscript:   func(ctx context.Context, id int64) (map[string]any, error) { return aiRepo.GetManuscriptByID(ctx, id) },
			GetUser:         func(ctx context.Context, id int64) (map[string]any, error) { return aiRepo.GetUserByID(ctx, id) },
			ListComments:    func(ctx context.Context, q string, limit int) ([]map[string]any, error) { return aiRepo.ListComments(ctx, q, limit) },
			ListReviews:     func(ctx context.Context, status string, limit int) ([]map[string]any, error) { return aiRepo.ListReviews(ctx, status, limit) },
			ListAuditLogs:   func(ctx context.Context, limit int) ([]map[string]any, error) { return aiRepo.ListAuditLogs(ctx, limit) },
			ListConfigs:     func(ctx context.Context) ([]map[string]any, error) { return aiRepo.ListAIConfigs(ctx) },
			ListSkills:      func(ctx context.Context) ([]map[string]any, error) { return aiRepo.ListAISkills(ctx) },
			ManuscriptMetrics: func(ctx context.Context, id int64, days int) (map[string]any, error) { return aiRepo.GetManuscriptMetrics(ctx, id, days) },
			UserGrowth:      func(ctx context.Context, days int) (map[string]any, error) { return aiRepo.GetUserGrowth(ctx, days) },
			UpdateManuscriptStatus: func(ctx context.Context, id, status int64) error { return aiRepo.UpdateManuscriptStatus(ctx, id, status) },
			UpdateUserStatus: func(ctx context.Context, id, status int64) error { return aiRepo.UpdateUserStatus(ctx, id, status) },
			ReplyTicket:     func(ctx context.Context, ticketNo, reply string) error { return aiRepo.ReplyTicket(ctx, ticketNo, reply) },
			CloseTicket:     func(ctx context.Context, ticketNo string) error { return aiRepo.CloseTicket(ctx, ticketNo) },
			DeleteComment:   func(ctx context.Context, id int64) error { return aiRepo.DeleteComment(ctx, id) },
		}))
		adminAssistant.SetUsageLog(func(feature string, userID int64, tokenCount, durationMs int) {
			if err := aiRepo.LogUsage(context.Background(), feature, userID, os.Getenv("AI_LLM_MODEL"), tokenCount, durationMs); err != nil {
				log.Printf("WARN: ai usage log (admin assistant): %v", err)
			}
		})
		aiH = aiH.WithAssistant(adminAssistant)
		log.Printf("ai: 管理后台 AI 助手已启用（工具调用 agent + 表格/指标卡渲染）")
	} else {
		log.Printf("WARN: 未配置真实 LLM，客服回退单轮调用器（未配置时返回显式错误）")
	}

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
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})
	subtitleH.Register(mux)
	aiH.Register(mux)
	aiChatH.Register(mux)
	log.Printf("AI HTTP listening on %s", httpAddr)
	// 吊销名单 / 刷新令牌防重放：Redis 抖动时降级为进程内，不影响可用性
	auth.LogSetupWarning(auth.SetupStoresFromEnv())

	log.Fatal(http.ListenAndServe(httpAddr, auth.IdentityMiddleware(newJWT())(auth.AdminPathGuard(mux))))
}

// newJWT 构造验签器。secret 与 core 保持一致，缺失时回退开发默认值。
func newJWT() *auth.JWT {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-in-production"
	}
	return auth.NewJWT(secret)
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// stripScheme 去掉 http(s):// 前缀，供 minio-go 客户端使用。
func stripScheme(s string) string {
	for _, p := range []string{"https://", "http://"} {
		if len(s) > len(p) && s[:len(p)] == p {
			return s[len(p):]
		}
	}
	return s
}
