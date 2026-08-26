package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"mybilibili/bili-proxy-service/internal/bilibili"
	"mybilibili/bili-proxy-service/internal/proxy"
)

const defaultListen = ":8091"

func main() {
	listen := os.Getenv("LISTEN")
	if listen == "" {
		listen = defaultListen
	}
	pgDSN := os.Getenv("PG_DSN")
	if pgDSN == "" {
		pgDSN = "postgres://postgres:postgres@localhost:5432/mybilibili?sslmode=disable"
	}
	db, err := sql.Open("postgres", pgDSN)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	client := bilibili.NewClient(os.Getenv("BILI_SESSDATA"))

	mux := http.NewServeMux()
	proxy.NewHandler(db, client).Register(mux)

	log.Printf("bili-proxy listening on %s", listen)
	if err := http.ListenAndServe(listen, mux); err != nil {
		log.Fatalf("http server: %v", err)
	}
}