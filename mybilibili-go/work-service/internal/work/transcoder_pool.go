package work

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// NodeStatus 节点运行时状态。
type NodeStatus string

const (
	StatusHealthy   NodeStatus = "healthy"
	StatusUnhealthy NodeStatus = "unhealthy"
	StatusDisabled  NodeStatus = "disabled"
)

// Node 描述一个 transcoder 节点 + 运行时统计。
// 持久化字段: Name / Addr / Capabilities / Weight
// 运行时字段: Status / LastCheck / LastError / FailCount / 统计
type Node struct {
	Name         string     `yaml:"name" json:"name"`
	Addr         string     `yaml:"addr" json:"addr"`
	Capabilities []string   `yaml:"capabilities" json:"capabilities"`
	Weight       int        `yaml:"weight" json:"weight"`

	// 运行时（不持久化）
	Status       NodeStatus `yaml:"-" json:"status"`
	LastCheck    time.Time  `yaml:"-" json:"last_check"`
	LastError    string     `yaml:"-" json:"last_error"`
	FailCount    int        `yaml:"-" json:"-"`
	TotalJobs    int64      `yaml:"-" json:"total_jobs"`
	FailedJobs   int64      `yaml:"-" json:"failed_jobs"`
	TotalLatMs   int64      `yaml:"-" json:"-"`
	Concurrency  int32      `yaml:"-" json:"concurrency"`

	mu sync.Mutex `yaml:"-" json:"-"`
}

// AvgLatencyMs 平均转码耗时（毫秒），运行时计算。
func (n *Node) AvgLatencyMs() int64 {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.TotalJobs == 0 {
		return 0
	}
	return n.TotalLatMs / n.TotalJobs
}

func (n *Node) recordResult(success bool, latency time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.TotalJobs++
	if !success {
		n.FailedJobs++
	}
	n.TotalLatMs += latency.Milliseconds()
}

// HealthCheck 健康检查配置。
type HealthCheck struct {
	Interval         time.Duration `yaml:"interval"`
	Timeout          time.Duration `yaml:"timeout"`
	FailureThreshold int           `yaml:"failure_threshold"`
}

func defaultHealthCheck() HealthCheck {
	return HealthCheck{
		Interval:         10 * time.Second,
		Timeout:          3 * time.Second,
		FailureThreshold: 3,
	}
}

// PoolConfig transcoder 池配置（持久化到 YAML）。
type PoolConfig struct {
	Transcoders []Node       `yaml:"transcoders"`
	LoadBalance string       `yaml:"load_balance"` // round_robin (当前唯一实现)
	HealthCheck HealthCheck  `yaml:"health_check"`
}

// TranscoderPool 多 transcoder 节点池：调度 + 健康检查 + 统计 + 配置持久化。
// 设计: 配置是真理来源 (transcoders.yaml), 内存是运行时镜像。
//       任何修改通过 API → 内存 → 落盘 (顺序) → 下次 SIGHUP/reload 自动应用文件版本。
type TranscoderPool struct {
	mu      sync.RWMutex
	cfgPath string
	cfg     PoolConfig
	nodes   map[string]*Node // by name
	rrIdx   int               // round-robin 游标

	// HTTP 客户端缓存（每个节点一个，复用 TCP 连接）
	httpc *http.Client

	stopCh chan struct{}
}

// NewTranscoderPool 从 YAML 文件加载配置；文件不存在则用空配置启动。
func NewTranscoderPool(cfgPath string) (*TranscoderPool, error) {
	p := &TranscoderPool{
		cfgPath: cfgPath,
		nodes:   map[string]*Node{},
		httpc:   &http.Client{Timeout: 30 * time.Second},
		stopCh:  make(chan struct{}),
		cfg: PoolConfig{
			LoadBalance: "round_robin",
			HealthCheck: defaultHealthCheck(),
		},
	}
	if err := p.loadConfig(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return p, nil
}

// Start 启动后台健康检查 goroutine。
func (p *TranscoderPool) Start(ctx context.Context) {
	go p.healthLoop(ctx)
}

// Stop 停止健康检查。
func (p *TranscoderPool) Stop() { close(p.stopCh) }

// loadConfig 从 YAML 文件读配置（不重置统计，保留运行时状态）。
func (p *TranscoderPool) loadConfig() error {
	b, err := os.ReadFile(p.cfgPath)
	if err != nil {
		return err
	}
	var cfg PoolConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf("parse %s: %w", p.cfgPath, err)
	}
	if cfg.HealthCheck.Interval == 0 {
		cfg.HealthCheck = defaultHealthCheck()
	}
	if cfg.LoadBalance == "" {
		cfg.LoadBalance = "round_robin"
	}
	p.applyConfig(cfg)
	return nil
}

// applyConfig 把 cfg 应用到内存池（保留已有节点的运行时状态）。
func (p *TranscoderPool) applyConfig(cfg PoolConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cfg = cfg

	// 重建索引：保留旧节点的状态，按 name 复用
	newNodes := map[string]*Node{}
	for i := range cfg.Transcoders {
		n := cfg.Transcoders[i]
		if existing, ok := p.nodes[n.Name]; ok {
			// 保留运行时状态
			n.Status = existing.Status
			n.LastCheck = existing.LastCheck
			n.LastError = existing.LastError
			n.FailCount = existing.FailCount
			n.TotalJobs = existing.TotalJobs
			n.FailedJobs = existing.FailedJobs
			n.TotalLatMs = existing.TotalLatMs
			n.Concurrency = existing.Concurrency
		} else {
			n.Status = StatusHealthy // 新节点初始健康
		}
		newNodes[n.Name] = &n
	}
	p.nodes = newNodes
}

// saveConfig 写当前内存池状态回 YAML（排除运行时字段）。
func (p *TranscoderPool) saveConfig() error {
	p.mu.RLock()
	cfg := p.cfg
	nodes := make([]Node, 0, len(p.nodes))
	for _, n := range p.nodes {
		nodes = append(nodes, Node{
			Name:         n.Name,
			Addr:         n.Addr,
			Capabilities: n.Capabilities,
			Weight:       n.Weight,
		})
	}
	p.mu.RUnlock()

	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Name < nodes[j].Name })
	cfg.Transcoders = nodes

	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	// 原子写：临时文件 + rename
	dir := filepath.Dir(p.cfgPath)
	tmp := filepath.Join(dir, ".transcoders.yaml.tmp")
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p.cfgPath)
}

// healthLoop 周期性对所有 enabled 节点做健康检查。
func (p *TranscoderPool) healthLoop(ctx context.Context) {
	// 启动后立即探一次（不等首个 interval）
	p.probeAll()
	for {
		p.mu.RLock()
		interval := p.cfg.HealthCheck.Interval
		p.mu.RUnlock()
		if interval <= 0 {
			interval = 10 * time.Second
		}
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-time.After(interval):
			p.probeAll()
		}
	}
}

func (p *TranscoderPool) probeAll() {
	p.mu.RLock()
	nodes := make([]*Node, 0, len(p.nodes))
	for _, n := range p.nodes {
		nodes = append(nodes, n)
	}
	p.mu.RUnlock()
	for _, n := range nodes {
		p.probeOne(n)
	}
}

func (p *TranscoderPool) probeOne(n *Node) {
	p.mu.RLock()
	timeout := p.cfg.HealthCheck.Timeout
	threshold := p.cfg.HealthCheck.FailureThreshold
	p.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	url := n.Addr + "/health"
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := p.httpc.Do(req)
	ok := err == nil && resp != nil && resp.StatusCode == 200
	if resp != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	n.mu.Lock()
	n.LastCheck = time.Now()
	if !ok {
		n.FailCount++
		if err != nil {
			n.LastError = err.Error()
		} else {
			n.LastError = fmt.Sprintf("status %d", resp.StatusCode)
		}
		if n.FailCount >= threshold && n.Status == StatusHealthy {
			n.Status = StatusUnhealthy
			log.Printf("transcoder[%s] unhealthy after %d fails: %s", n.Name, n.FailCount, n.LastError)
		}
	} else {
		if n.FailCount > 0 || n.Status == StatusUnhealthy {
			log.Printf("transcoder[%s] recovered after %d fails", n.Name, n.FailCount)
		}
		n.FailCount = 0
		n.LastError = ""
		if n.Status != StatusDisabled {
			n.Status = StatusHealthy
		}
	}
	n.mu.Unlock()
}

// Pick 按 round-robin 策略选一个健康的 enabled 节点。
// 返回的 Node 内部 + *TranscoderClient (HTTP client 复用)。
// 全部节点不可用时返回 error。
func (p *TranscoderPool) Pick(ctx context.Context) (*Node, *TranscoderClient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	candidates := make([]*Node, 0, len(p.nodes))
	for _, n := range p.nodes {
		if n.Status == StatusDisabled {
			continue
		}
		if n.Status == StatusUnhealthy {
			continue
		}
		candidates = append(candidates, n)
	}
	if len(candidates) == 0 {
		return nil, nil, errors.New("no healthy transcoder available")
	}

	// round-robin: 按 Weight 加权扩展后轮流
	expanded := make([]*Node, 0, len(candidates)*10)
	for _, n := range candidates {
		w := n.Weight
		if w <= 0 {
			w = 1
		}
		for i := 0; i < w; i++ {
			expanded = append(expanded, n)
		}
	}
	p.rrIdx = (p.rrIdx + 1) % len(expanded)
	chosen := expanded[p.rrIdx]
	return chosen, NewTranscoderClient(chosen.Addr), nil
}

// MarkResult 反馈一次转码结果（用于统计 + 健康度判断）。
func (p *TranscoderPool) MarkResult(name string, success bool, latency time.Duration) {
	p.mu.RLock()
	n, ok := p.nodes[name]
	p.mu.RUnlock()
	if !ok {
		return
	}
	n.recordResult(success, latency)
	if !success {
		// 单次失败不立即判 unhealthy，由 healthLoop 周期性探测决定
	}
}

// List 列出所有节点（深拷贝），给 admin API 用。
func (p *TranscoderPool) List() []*Node {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]*Node, 0, len(p.nodes))
	for _, n := range p.nodes {
		copy := *n
		copy.mu = sync.Mutex{}
		out = append(out, &copy)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Get 取单个节点。
func (p *TranscoderPool) Get(name string) (*Node, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	n, ok := p.nodes[name]
	if !ok {
		return nil, false
	}
	copy := *n
	copy.mu = sync.Mutex{}
	return &copy, true
}

// Add 新增节点，立即持久化到 YAML。
func (p *TranscoderPool) Add(n Node) error {
	if n.Name == "" || n.Addr == "" {
		return errors.New("name and addr required")
	}
	p.mu.Lock()
	if _, exists := p.nodes[n.Name]; exists {
		p.mu.Unlock()
		return fmt.Errorf("node %q already exists", n.Name)
	}
	if n.Weight <= 0 {
		n.Weight = 1
	}
	n.Status = StatusHealthy
	n.LastCheck = time.Now()
	p.nodes[n.Name] = &n
	p.mu.Unlock()
	return p.saveConfig()
}

// Update 更新节点 (addr / capabilities / weight)，持久化。
func (p *TranscoderPool) Update(name string, upd Node) error {
	p.mu.Lock()
	n, ok := p.nodes[name]
	if !ok {
		p.mu.Unlock()
		return fmt.Errorf("node %q not found", name)
	}
	if upd.Addr != "" {
		n.Addr = upd.Addr
	}
	if upd.Capabilities != nil {
		n.Capabilities = upd.Capabilities
	}
	if upd.Weight > 0 {
		n.Weight = upd.Weight
	}
	p.mu.Unlock()
	return p.saveConfig()
}

// Delete 删除节点，持久化。
func (p *TranscoderPool) Delete(name string) error {
	p.mu.Lock()
	if _, ok := p.nodes[name]; !ok {
		p.mu.Unlock()
		return fmt.Errorf("node %q not found", name)
	}
	delete(p.nodes, name)
	p.mu.Unlock()
	return p.saveConfig()
}

// Enable 启用节点（恢复参与调度）。
func (p *TranscoderPool) Enable(name string) error {
	p.mu.Lock()
	n, ok := p.nodes[name]
	if !ok {
		p.mu.Unlock()
		return fmt.Errorf("node %q not found", name)
	}
	n.Status = StatusHealthy
	n.FailCount = 0
	p.mu.Unlock()
	return nil
}

// Disable 禁用节点（不参与调度，但保留配置）。
func (p *TranscoderPool) Disable(name string) error {
	p.mu.Lock()
	n, ok := p.nodes[name]
	if !ok {
		p.mu.Unlock()
		return fmt.Errorf("node %q not found", name)
	}
	n.Status = StatusDisabled
	p.mu.Unlock()
	return nil
}

// Probe 主动探活一次（不等周期）。
func (p *TranscoderPool) Probe(name string) error {
	p.mu.RLock()
	n, ok := p.nodes[name]
	p.mu.RUnlock()
	if !ok {
		return fmt.Errorf("node %q not found", name)
	}
	p.probeOne(n)
	return nil
}

// ProbeAddr 探测未保存的地址，返回该地址的能力 (用于新增前的"测试连接"按钮)。
func (p *TranscoderPool) ProbeAddr(addr string) (encoder string, caps []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url := addr + "/api/v1/capabilities"
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := p.httpc.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("connect: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var out struct {
		Encoder      string   `json:"encoder"`
		Capabilities []string `json:"capabilities"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", nil, fmt.Errorf("decode: %w", err)
	}
	return out.Encoder, out.Capabilities, nil
}

// Reload 重新从 YAML 文件加载配置。
func (p *TranscoderPool) Reload() error { return p.loadConfig() }

// ConfigYAML 返回当前配置的 YAML 表示（用于 admin 页面查看）。
func (p *TranscoderPool) ConfigYAML() (string, error) {
	p.mu.RLock()
	cfg := p.cfg
	nodes := make([]Node, 0, len(p.nodes))
	for _, n := range p.nodes {
		nodes = append(nodes, Node{
			Name: n.Name, Addr: n.Addr,
			Capabilities: n.Capabilities, Weight: n.Weight,
		})
	}
	p.mu.RUnlock()
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Name < nodes[j].Name })
	cfg.Transcoders = nodes
	b, err := yaml.Marshal(cfg)
	return string(b), err
}

// writeFileAtomic 原子写入文件：临时文件 + rename。
// 供 admin_api.go 在 PUT /api/v1/admin/transcoders/config 时使用。
func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp := filepath.Join(dir, ".config.yaml.tmp")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
