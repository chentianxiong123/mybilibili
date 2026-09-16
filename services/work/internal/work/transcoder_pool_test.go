package work

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranscoderPool_NewFromNonExistentFile(t *testing.T) {
	p, err := NewTranscoderPool(filepath.Join(t.TempDir(), "missing.yaml"))
	require.NoError(t, err)
	assert.Empty(t, p.List())
}

func TestTranscoderPool_LoadYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pool.yaml")
	cfg := PoolConfig{
		Transcoders: []Node{
			{Name: "p400", Addr: "http://a:8080", Capabilities: []string{"nvenc"}, Weight: 3},
			{Name: "iGPU", Addr: "http://b:8080", Capabilities: []string{"vaapi"}, Weight: 1},
		},
	}
	require.NoError(t, os.WriteFile(path, mustYAML(t, cfg), 0o644))

	p, err := NewTranscoderPool(path)
	require.NoError(t, err)
	nodes := p.List()
	require.Len(t, nodes, 2)
	// List 按 name 排序 (ASCII: 小写 > 大写)
	assert.Equal(t, "iGPU", nodes[0].Name)
	assert.Equal(t, "p400", nodes[1].Name)
	assert.Equal(t, []string{"nvenc"}, findNode(p, "p400").Capabilities)
	assert.Equal(t, StatusHealthy, findNode(p, "p400").Status, "新节点默认 healthy")
	assert.Equal(t, "round_robin", p.cfg.LoadBalance)
}

func TestTranscoderPool_AddValidation(t *testing.T) {
	p, err := NewTranscoderPool(filepath.Join(t.TempDir(), "pool.yaml"))
	require.NoError(t, err)

	require.Error(t, p.Add(Node{Name: "", Addr: "http://x"}))
	require.Error(t, p.Add(Node{Name: "n", Addr: ""}))

	require.NoError(t, p.Add(Node{Name: "n", Addr: "http://x"}))
	// weight 默认 1
	assert.Equal(t, 1, findNode(p, "n").Weight)
	// 重名报错
	assert.Error(t, p.Add(Node{Name: "n", Addr: "http://y"}))
}

func TestTranscoderPool_AddPersistsYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pool.yaml")
	p, err := NewTranscoderPool(path)
	require.NoError(t, err)

	require.NoError(t, p.Add(Node{Name: "p400", Addr: "http://a", Capabilities: []string{"nvenc"}, Weight: 2}))
	require.NoError(t, p.Add(Node{Name: "iGPU", Addr: "http://b", Weight: 0})) // weight 0 → 1

	// 从磁盘重新加载
	p2, err := NewTranscoderPool(path)
	require.NoError(t, err)
	assert.Len(t, p2.List(), 2)
	assert.Equal(t, 2, findNode(p2, "p400").Weight)
	assert.Equal(t, 1, findNode(p2, "iGPU").Weight, "weight<=0 归一为 1")
}

func TestTranscoderPool_Update(t *testing.T) {
	p, err := NewTranscoderPool(filepath.Join(t.TempDir(), "pool.yaml"))
	require.NoError(t, err)
	require.NoError(t, p.Add(Node{Name: "n", Addr: "http://a", Weight: 1}))

	// 不存在的节点
	assert.Error(t, p.Update("ghost", Node{Addr: "http://z"}))

	// 只改部分字段
	require.NoError(t, p.Update("n", Node{Addr: "http://new", Capabilities: []string{"vaapi"}, Weight: 4}))
	n := findNode(p, "n")
	assert.Equal(t, "http://new", n.Addr)
	assert.Equal(t, []string{"vaapi"}, n.Capabilities)
	assert.Equal(t, 4, n.Weight)
}

func TestTranscoderPool_Delete(t *testing.T) {
	p, err := NewTranscoderPool(filepath.Join(t.TempDir(), "pool.yaml"))
	require.NoError(t, err)
	require.NoError(t, p.Add(Node{Name: "n", Addr: "http://a"}))

	assert.Error(t, p.Delete("ghost"))
	require.NoError(t, p.Delete("n"))
	_, ok := p.Get("n")
	assert.False(t, ok)
}

func TestTranscoderPool_DisableEnable(t *testing.T) {
	p, err := NewTranscoderPool(filepath.Join(t.TempDir(), "pool.yaml"))
	require.NoError(t, err)
	require.NoError(t, p.Add(Node{Name: "n", Addr: "http://a"}))

	require.NoError(t, p.Disable("n"))
	assert.Equal(t, StatusDisabled, findNode(p, "n").Status)

	// disabled 节点不参与 Pick
	_, _, err = p.Pick(context.Background())
	assert.Error(t, err, "disabled 后应无可用节点")

	require.NoError(t, p.Enable("n"))
	n := findNode(p, "n")
	assert.Equal(t, StatusHealthy, n.Status)
	assert.Zero(t, n.FailCount, "Enable 重置 FailCount")
}

func TestTranscoderPool_PickWeightedRoundRobin(t *testing.T) {
	p, err := NewTranscoderPool(filepath.Join(t.TempDir(), "pool.yaml"))
	require.NoError(t, err)
	require.NoError(t, p.Add(Node{Name: "heavy", Addr: "http://a", Weight: 3}))
	require.NoError(t, p.Add(Node{Name: "light", Addr: "http://b", Weight: 1}))

	// 10 次 Pick: heavy 应占 75%
	seen := map[string]int{}
	for i := 0; i < 40; i++ {
		node, _, err := p.Pick(context.Background())
		require.NoError(t, err)
		seen[node.Name]++
	}
	assert.Greater(t, seen["heavy"], seen["light"], "weight=3 应多于 weight=1")
	// 加权轮询是确定性的（非概率），3:1 权重下 40 次应为 30:10
	assert.InDelta(t, 30, seen["heavy"], 5, "heavy 应接近 30")
	assert.InDelta(t, 10, seen["light"], 5, "light 应接近 10")
}

func TestTranscoderPool_PickNoHealthy(t *testing.T) {
	p, err := NewTranscoderPool(filepath.Join(t.TempDir(), "pool.yaml"))
	require.NoError(t, err)
	_, _, err = p.Pick(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no healthy transcoder")
}

func TestTranscoderPool_MarkResultAndAvgLatency(t *testing.T) {
	p, err := NewTranscoderPool(filepath.Join(t.TempDir(), "pool.yaml"))
	require.NoError(t, err)
	require.NoError(t, p.Add(Node{Name: "n", Addr: "http://a"}))

	assert.Zero(t, findNode(p, "n").AvgLatencyMs(), "无任务时 0")
	p.MarkResult("n", true, 200*time.Millisecond)
	p.MarkResult("n", false, 100*time.Millisecond)
	p.MarkResult("unknown", true, time.Second) // 未知节点忽略

	n := findNode(p, "n")
	assert.Equal(t, int64(2), n.TotalJobs)
	assert.Equal(t, int64(1), n.FailedJobs)
	assert.Equal(t, int64(150), n.AvgLatencyMs(), "avg = (200+100)/2")
}

func TestTranscoderPool_GetDeepCopy(t *testing.T) {
	p, err := NewTranscoderPool(filepath.Join(t.TempDir(), "pool.yaml"))
	require.NoError(t, err)
	require.NoError(t, p.Add(Node{Name: "n", Addr: "http://a"}))

	a, ok := p.Get("n")
	require.True(t, ok)
	_, ok = p.Get("ghost")
	assert.False(t, ok)
	// 修改拷贝不影响池内节点
	a.Addr = "http://changed"
	assert.Equal(t, "http://a", findNode(p, "n").Addr)
}

func TestTranscoderPool_ProbeAddrSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/capabilities", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"encoder": "nvenc", "capabilities": []string{"h264", "hevc"}})
	}))
	defer srv.Close()

	p, err := NewTranscoderPool(filepath.Join(t.TempDir(), "pool.yaml"))
	require.NoError(t, err)

	encoder, caps, err := p.ProbeAddr(srv.URL)
	require.NoError(t, err)
	assert.Equal(t, "nvenc", encoder)
	assert.ElementsMatch(t, []string{"h264", "hevc"}, caps)
}

func TestTranscoderPool_ProbeAddrFailStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	p, _ := NewTranscoderPool(filepath.Join(t.TempDir(), "pool.yaml"))
	_, _, err := p.ProbeAddr(srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 503")
}

func TestTranscoderPool_ProbeAddrUnreachable(t *testing.T) {
	p, _ := NewTranscoderPool(filepath.Join(t.TempDir(), "pool.yaml"))
	// 指向一个立即拒绝的地址
	_, _, err := p.ProbeAddr("http://127.0.0.1:1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connect")
}

func TestTranscoderPool_HealthProbeUnhealthyTransition(t *testing.T) {
	p, err := NewTranscoderPool(filepath.Join(t.TempDir(), "pool.yaml"))
	require.NoError(t, err)
	require.NoError(t, p.Add(Node{Name: "n", Addr: "http://127.0.0.1:1"}))

	// 手动探测 3 次（默认 threshold=3）
	for i := 0; i < 3; i++ {
		require.NoError(t, p.Probe("n"))
	}
	n := findNode(p, "n")
	assert.Equal(t, StatusUnhealthy, n.Status, "3 次失败后 unhealthy")
	assert.Equal(t, 3, n.FailCount)
	assert.NotEmpty(t, n.LastError)
}

func TestTranscoderPool_HealthProbeRecovery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p, err := NewTranscoderPool(filepath.Join(t.TempDir(), "pool.yaml"))
	require.NoError(t, err)
	require.NoError(t, p.Add(Node{Name: "n", Addr: srv.URL}))
	// 先打 unhealthy
	p.mu.Lock()
	p.nodes["n"].Status = StatusUnhealthy
	p.nodes["n"].FailCount = 2
	p.mu.Unlock()

	require.NoError(t, p.Probe("n"))
	n := findNode(p, "n")
	assert.Equal(t, StatusHealthy, n.Status, "探测成功恢复 healthy")
	assert.Zero(t, n.FailCount)
	assert.Empty(t, n.LastError)
}

func TestTranscoderPool_Reload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pool.yaml")
	p, err := NewTranscoderPool(path)
	require.NoError(t, err)
	require.NoError(t, p.Add(Node{Name: "old", Addr: "http://a"}))

	// 直接写一个新配置
	newCfg := PoolConfig{Transcoders: []Node{
		{Name: "fresh", Addr: "http://c"},
	}}
	require.NoError(t, os.WriteFile(path, mustYAML(t, newCfg), 0o644))

	require.NoError(t, p.Reload())
	_, ok := p.Get("old")
	assert.False(t, ok, "旧节点消失")
	_, ok = p.Get("fresh")
	assert.True(t, ok)
}

func TestTranscoderPool_ConfigYAML(t *testing.T) {
	p, err := NewTranscoderPool(filepath.Join(t.TempDir(), "pool.yaml"))
	require.NoError(t, err)
	require.NoError(t, p.Add(Node{Name: "b", Addr: "http://b"}))
	require.NoError(t, p.Add(Node{Name: "a", Addr: "http://a"}))

	out, err := p.ConfigYAML()
	require.NoError(t, err)
	var cfg PoolConfig
	require.NoError(t, yaml.Unmarshal([]byte(out), &cfg))
	assert.Len(t, cfg.Transcoders, 2)
	// 按 name 排序
	names := make([]string, 0, 2)
	for _, n := range cfg.Transcoders {
		names = append(names, n.Name)
	}
	sort.Strings(names)
	assert.Equal(t, names, []string{"a", "b"})
}

// --- helpers ---

func findNode(p *TranscoderPool, name string) *Node {
	n, ok := p.Get(name)
	if !ok {
		return nil
	}
	return n
}

func mustYAML(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := yaml.Marshal(v)
	require.NoError(t, err)
	return b
}
