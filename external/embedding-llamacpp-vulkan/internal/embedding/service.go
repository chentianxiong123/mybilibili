package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Request OpenAI 兼容的 embedding 请求。
type Request struct {
	Model string   `json:"model"`
	Input any      `json:"input"` // string or []string
}

// Result 单条 embedding 结果。
type Result struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// Response OpenAI 兼容的 embedding 响应。
type Response struct {
	Object  string   `json:"object"`
	Data    []Result `json:"data"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
}

type Usage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// Service embedding 服务：管理 llama-server 进程 + 代理 embedding 请求。
type Service struct {
	llamaURL    string
	llamaBin    string
	modelPath   string
	httpc       *http.Client
	cmd         *exec.Cmd
	started     bool
}

// NewService 创建 embedding 服务。
//   - llamaURL: llama-server 地址（如 http://127.0.0.1:8081），非空则直接代理
//   - llamaBin: llama-server 二进制路径（如 /path/to/llama-server），为空则只做代理
//   - modelPath: GGUF 模型路径
func NewService(llamaURL, llamaBin, modelPath string) *Service {
	return &Service{
		llamaURL:  llamaURL,
		llamaBin:  llamaBin,
		modelPath: modelPath,
		httpc:     &http.Client{Timeout: 120 * time.Second},
	}
}

// Start 启动 llama-server 子进程（仅当 llamaURL 为空且 llamaBin 不为空时）。
func (s *Service) Start(ctx context.Context) error {
	if s.llamaURL != "" {
		log.Printf("embedding: proxy mode, upstream=%s", s.llamaURL)
		return nil
	}
	if s.llamaBin == "" {
		return fmt.Errorf("no llama-server URL or binary configured")
	}

	args := []string{
		"--model", s.modelPath,
		"--host", "127.0.0.1",
		"--port", "8081",
		"--embedding",
		"--pooling", "last",
		"--embd-normalize", "2",
		"--batch-size", "512",
		"--ctx-size", "8192",
	}

	// GPU layers: 从环境变量读取，默认全部上 GPU
	gpuLayers := os.Getenv("LLAMA_N_GPU_LAYERS")
	if gpuLayers == "" {
		gpuLayers = "9999"
	}
	args = append(args, "--n-gpu-layers", gpuLayers)

	// VULKAN 设备索引
	if device := os.Getenv("LLAMA_VULKAN_DEVICE"); device != "" {
		args = append(args, "--device", device)
	}

	log.Printf("embedding: starting %s %v", s.llamaBin, args)
	s.cmd = exec.Command(s.llamaBin, args...)
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("start llama-server: %w", err)
	}
	s.started = true
	s.llamaURL = "http://127.0.0.1:8081"

	// 等待 health check
	go s.waitForHealth()
	return nil
}

func (s *Service) waitForHealth() {
	for i := 0; i < 60; i++ {
		time.Sleep(time.Second)
		resp, err := s.httpc.Get(s.llamaURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				log.Printf("embedding: llama-server ready")
				return
			}
		}
	}
	log.Printf("embedding: WARNING llama-server health check timed out")
}

// Stop 停止 llama-server 子进程。
func (s *Service) Stop() {
	if s.cmd != nil && s.started {
		log.Printf("embedding: stopping llama-server")
		s.cmd.Process.Signal(os.Interrupt)
		s.cmd.Wait()
		s.started = false
	}
}

// Process 代理 embedding 请求到 llama-server。
func (s *Service) Process(ctx context.Context, req Request) (*Response, error) {
	if s.llamaURL == "" {
		return nil, fmt.Errorf("llama-server not configured")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := s.llamaURL + "/v1/embeddings"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpc.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("llama-server request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("llama-server status %d: %s", resp.StatusCode, string(respBody))
	}

	var result Response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result, nil
}

// Health 检查 llama-server 健康状态。
func (s *Service) Health(ctx context.Context) bool {
	if s.llamaURL == "" {
		return false
	}
	resp, err := s.httpc.Get(s.llamaURL + "/health")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

// resolveModelPath 解析模型路径（支持环境变量展开）。
func ResolveModelPath(p string) string {
	if p == "" {
		p = os.Getenv("LLAMA_MODEL")
	}
	if p == "" {
		// 默认路径
		home, _ := os.UserHomeDir()
		p = filepath.Join(home, "models", "qwen3-embedding-0.6b-q8_0.gguf")
	}
	return p
}
