package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"mybilibili/bili-proxy-service/internal/importer"
)

func main() {
	var (
		chrome   = flag.String("chrome", "/usr/bin/google-chrome", "chrome 可执行文件路径")
		target   = flag.Int("target", 30, "目标导入条数")
		sessdata = flag.String("sessdata", os.Getenv("BILI_SESSDATA"), "B站 SESSDATA（可选）")
		prefix   = flag.String("prefix", "/api/v1/bili/stream", "写入DB的play_url前缀")
		partitions = flag.String("partitions",
			"https://www.bilibili.com/v/popular/rank/all,"+
				"https://www.bilibili.com/v/knowledge,"+
				"https://www.bilibili.com/v/tech,"+
				"https://www.bilibili.com/v/game,"+
				"https://www.bilibili.com/v/life,"+
				"https://www.bilibili.com/v/ent",
			"分区页面URL列表（逗号分隔，dump DOM 提 BV）")
	)
	flag.Parse()

	pgDSN := os.Getenv("PG_DSN")
	if pgDSN == "" {
		pgDSN = "postgres://postgres:postgres@localhost:5432/mybilibili?sslmode=disable"
	}
	db, err := sql.Open("postgres", pgDSN)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	imp := importer.New(db, importer.Config{
		ChromePath: *chrome,
		Partitions: splitCSV(*partitions),
		Target:     *target,
		Sessdata:   *sessdata,
		PlayPrefix: *prefix,
	})

	log.Printf("collecting bare BVs from %d partition pages ...", len(splitCSV(*partitions)))
	bvs, err := imp.CollectBVs()
	if err != nil {
		log.Fatalf("collect bvs: %v", err)
	}
	log.Printf("got %d unique bvs", len(bvs))

	imported, skipped, err := imp.Run(bvs)
	if err != nil {
		log.Fatalf("import: %v", err)
	}
	log.Printf("done. imported=%d skipped=%d", imported, skipped)
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}