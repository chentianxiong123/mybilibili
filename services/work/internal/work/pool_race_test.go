package work

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestPool_ListConcurrentRecord_Race 验证 FINDING-013：
// List()/Get()/Pick() 读节点运行时字段（Status/TotalJobs/TotalLatMs/FailedJobs）
// 仅在 p.mu.RLock/Lock 下，未取节点 n.mu；而 recordResult/probeOne 写这些字段
// 在 n.mu。二者并发 = 数据竞态（go test -race 应报告）。
func TestPool_ListConcurrentRecord_Race(t *testing.T) {
	p, err := NewTranscoderPool(filepath.Join(t.TempDir(), "transcoders.yaml"))
	if err != nil {
		t.Fatalf("NewTranscoderPool: %v", err)
	}
	if err := p.Add(Node{Name: "n1", Addr: "http://127.0.0.1:1"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	n := p.nodes["n1"]
	if n == nil {
		t.Fatal("node not registered")
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func() { defer wg.Done(); _ = p.List() }()
		go func() { defer wg.Done(); n.recordResult(true, time.Millisecond) }()
		// Pick 读 n.Status 于 p.mu，recordResult/probeOne 写于 n.mu —— 同竞态模式
		go func() { defer wg.Done(); _, _, _ = p.Pick(context.Background()) }()
	}
	wg.Wait()
}
