package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"

	"mybilibili/studio-service/internal/studio"
)

func main() {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/mybilibili?sslmode=disable"
	}
	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8089"
	}
	dataDir := os.Getenv("STUDIO_DATA_DIR")
	if dataDir == "" {
		dataDir = "/tmp/studio-data"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	repo := studio.NewRepository(db)
	svc := studio.NewService(repo)
	h := studio.NewHandler(svc)

	assetsDir := filepath.Join(dataDir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		log.Fatalf("create assets dir: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})
	h.Register(mux)
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))

	log.Printf("studio HTTP listening on %s, data dir %s", httpAddr, dataDir)
	log.Fatal(http.ListenAndServe(httpAddr, mux))
}