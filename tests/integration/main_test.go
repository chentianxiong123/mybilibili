// 集成测试需要服务运行中。
// 运行方式：
//   1. make test-integration       ← 自动启动 docker-compose + 跑测试 + 清理
//   2. 手动 docker compose up -d  ← 然后 go test -tags=integration ./tests/integration/...
//
// 测试使用 build tag "integration"，普通 go test 不会执行。

//go:build integration

package integration

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	coreURL   = "http://localhost:8080"
	searchURL = "http://localhost:8084"
	msgURL    = "http://localhost:8086"
	liveURL   = "http://localhost:8087"
	aiURL     = "http://localhost:8088"
	studioURL = "http://localhost:8089"
	workURL   = "http://localhost:8090"
	biliURL   = "http://localhost:8091"
	nuxtURL   = "http://localhost"
)

func TestMain(m *testing.M) {
	waitForServices()
	code := m.Run()
	os.Exit(code)
}

func waitForServices() {
	services := []string{coreURL + "/api/v1/health", searchURL + "/health", nuxtURL}
	client := &http.Client{Timeout: 3 * time.Second}
	for _, url := range services {
		ready := false
		for i := 0; i < 30; i++ {
			resp, err := client.Get(url)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == 200 {
					ready = true
					break
				}
			}
			time.Sleep(time.Second)
		}
		if !ready {
			log.Printf("警告: 服务未就绪 %s（测试可能失败）", url)
		}
	}
}