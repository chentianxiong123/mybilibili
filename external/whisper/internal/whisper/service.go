package whisper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"mybilibili/pkg/abstraction"
)

// Request whisper 转写请求（MinIO 对象引用）。
type Request struct {
	Bucket       string `json:"bucket"`
	AudioKey     string `json:"audio_key"`
	ManuscriptID int64  `json:"manuscript_id"`
	VideoID      int64  `json:"video_id"`
	Language     string `json:"language"`
}

// Cue 单条字幕 cue。
type Cue struct {
	Index     int     `json:"index"`
	StartTime float64 `json:"startTime"`
	EndTime   float64 `json:"endTime"`
	Text      string  `json:"text"`
}

// Result whisper 转写结果。
type Result struct {
	Cues  []Cue  `json:"cues"`
	Text  string `json:"text"`
	Lang  string `json:"language"`
}

// Service whisper 服务：从 MinIO 取音频 → 调 whisper API → 返回字幕 cue。
type Service struct {
	storage abstraction.StorageService
	httpc   *http.Client
}

func NewService(storage abstraction.StorageService) *Service {
	return &Service{
		storage: storage,
		httpc:   &http.Client{Timeout: 300 * time.Second},
	}
}

// Process 执行一次 whisper 转写任务。
func (s *Service) Process(ctx context.Context, req Request) (*Result, error) {
	if req.Bucket == "" {
		req.Bucket = "mybilibili"
	}
	if req.AudioKey == "" {
		return nil, fmt.Errorf("audio_key required")
	}
	if req.Language == "" {
		req.Language = "zh-CN"
	}

	// 1. 从 MinIO 下载音频
	rc, err := s.storage.Get(ctx, req.Bucket, req.AudioKey)
	if err != nil {
		return nil, fmt.Errorf("download audio: %w", err)
	}
	audioData, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		return nil, fmt.Errorf("read audio: %w", err)
	}

	// 2. 调 whisper API
	cues, text, err := s.callWhisper(ctx, audioData)
	if err != nil {
		return nil, fmt.Errorf("whisper: %w", err)
	}

	return &Result{
		Cues: cues,
		Text: text,
		Lang: req.Language,
	}, nil
}

// callWhisper 根据环境变量选择 whisper 后端。
func (s *Service) callWhisper(ctx context.Context, audioData []byte) ([]Cue, string, error) {
	if apiURL := os.Getenv("WHISPER_API_URL"); apiURL != "" {
		return s.callWhisperOpenAI(ctx, audioData, apiURL)
	}
	if accountID := os.Getenv("CLOUDFLARE_AI_ACCOUNT_ID"); accountID != "" {
		return s.callWhisperCloudflare(ctx, audioData, accountID, os.Getenv("CLOUDFLARE_AI_API_TOKEN"))
	}
	return nil, "", fmt.Errorf("no whisper backend configured (set WHISPER_API_URL or CLOUDFLARE_AI_*)")
}

// callWhisperOpenAI 调用 OpenAI 兼容的 /v1/audio/transcriptions 端点。
func (s *Service) callWhisperOpenAI(ctx context.Context, audioData []byte, apiURL string) ([]Cue, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return nil, "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := fw.Write(audioData); err != nil {
		return nil, "", fmt.Errorf("write audio data: %w", err)
	}
	w.WriteField("model", "whisper-1")
	w.WriteField("response_format", "verbose_json")
	w.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, &buf)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := s.httpc.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("whisper api status %d: %s", resp.StatusCode, string(body))
	}

	// 尝试解析 verbose_json 格式（带 segments）
	var verboseResp struct {
		Text     string `json:"text"`
		Segments []struct {
			ID    int     `json:"id"`
			Start float64 `json:"start"`
			End   float64 `json:"end"`
			Text  string  `json:"text"`
		} `json:"segments"`
	}
	if err := json.Unmarshal(body, &verboseResp); err == nil && len(verboseResp.Segments) > 0 {
		cues := make([]Cue, 0, len(verboseResp.Segments))
		for _, seg := range verboseResp.Segments {
			cues = append(cues, Cue{
				Index:     seg.ID + 1,
				StartTime: seg.Start,
				EndTime:   seg.End,
				Text:      strings.TrimSpace(seg.Text),
			})
		}
		return cues, verboseResp.Text, nil
	}

	// fallback: 简单 text 格式
	var simpleResp struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(body, &simpleResp); err != nil {
		return nil, "", fmt.Errorf("parse response: %w", err)
	}
	text := strings.TrimSpace(simpleResp.Text)
	if text == "" {
		return nil, "", fmt.Errorf("whisper returned empty text")
	}
	cues := []Cue{{Index: 1, StartTime: 0, EndTime: 0, Text: text}}
	return cues, text, nil
}

// callWhisperCloudflare 直连 Cloudflare Workers AI 的 whisper 接口。
func (s *Service) callWhisperCloudflare(ctx context.Context, audioData []byte, accountID, apiToken string) ([]Cue, string, error) {
	if apiToken == "" {
		return nil, "", fmt.Errorf("CLOUDFLARE_AI_API_TOKEN required")
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return nil, "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := fw.Write(audioData); err != nil {
		return nil, "", fmt.Errorf("write audio data: %w", err)
	}
	w.Close()

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/@cf/openai/whisper", accountID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := s.httpc.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("api status %d: %s", resp.StatusCode, string(body))
	}

	var cfResp struct {
		Success bool `json:"success"`
		Result  struct {
			Text     string `json:"text"`
			Segments []struct {
				ID    int     `json:"id"`
				Start float64 `json:"start"`
				End   float64 `json:"end"`
				Text  string  `json:"text"`
			} `json:"segments"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &cfResp); err != nil {
		return nil, "", fmt.Errorf("parse response: %w", err)
	}
	if !cfResp.Success {
		return nil, "", fmt.Errorf("whisper API returned success=false: %s", string(body))
	}

	cues := make([]Cue, 0, len(cfResp.Result.Segments))
	for _, seg := range cfResp.Result.Segments {
		cues = append(cues, Cue{
			Index:     seg.ID + 1,
			StartTime: seg.Start,
			EndTime:   seg.End,
			Text:      strings.TrimSpace(seg.Text),
		})
	}
	log.Printf("whisper: %d cues, %d chars", len(cues), len(cfResp.Result.Text))
	return cues, cfResp.Result.Text, nil
}
