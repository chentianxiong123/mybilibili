package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"mybilibili/bili-proxy-service/internal/importer"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/mybilibili?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	n, err := importer.New(db, importer.Config{}).ReclassifyAll()
	if err != nil {
		panic(err)
	}
	fmt.Printf("reclassified non-tech: %d\n", n)
}
