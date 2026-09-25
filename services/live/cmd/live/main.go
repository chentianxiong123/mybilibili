package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"mybilibili/live/internal/live"
	"mybilibili/pkg/auth"
)

func main() {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/mybilibili?sslmode=disable"
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8087"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	liveRepo := live.NewRepository(db)
	liveHub := live.NewHub()
	liveSvc := live.NewService(liveRepo, liveHub)
	liveH := live.NewHTTPHandler(liveSvc, liveHub)

	linkmicRepo := live.NewLinkmicRepository(db)
	linkmicSvc := live.NewLinkmicService(linkmicRepo, liveHub, liveRepo)
	linkmicH := live.NewLinkmicHandler(linkmicSvc)

	liveAdminH := live.NewAdminHandler(db)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"status":"ok"}`)) })
	liveH.Register(mux)
	linkmicH.Register(mux)
	liveAdminH.Register(mux)

	log.Printf("Live service HTTP listening on %s", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, auth.IdentityMiddleware(newJWT())(mux)))
}

// newJWT 构造验签器。secret 与 core 保持一致，缺失时回退开发默认值。
func newJWT() *auth.JWT {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-in-production"
	}
	return auth.NewJWT(secret)
}
