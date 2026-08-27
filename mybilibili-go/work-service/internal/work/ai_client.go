package work

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AIClient 调用独立的 ai-service（whisper 字幕转写 + AI 摘要生成）。
// work 作为编排器只传 manuscript_id/video_id，不接触音频与文件内容；
// ai-service 自行从 MinIO 读取音频并写回结果。
type AIClient struct {
	baseURL string
	httpc   *http.Client
}

func NewAIClient(baseURL string) *AIClient {
	return &AIClient{
		baseURL: baseURL,
		httpc:   &http.Client{Timeout: 10 * time.Minute},
	}
}

// GenerateSubtitle 触发 ai-service 生成 whisper 字幕。
func (c *AIClient) GenerateSubtitle(ctx context.Context, manuscriptID, videoID int64) error {
	body, _ := json.Marshal(map[string]int64{
		"manuscript_id": manuscriptID,
		"video_id":      videoID,
	})
	return c.post(ctx, "/api/v1/subtitle/generate", body)
}

// GenerateSummary 触发 ai-service 生成 AI 摘要。
func (c *AIClient) GenerateSummary(ctx context.Context, manuscriptID, videoID int64) error {
	body, _ := json.Marshal(map[string]int64{
		"manuscript_id": manuscriptID,
		"video_id":      videoID,
	})
	return c.post(ctx, "/api/v1/ai/summary/generate", body)
}

func (c *AIClient) post(ctx context.Context, path string, body []byte) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpc.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ai: %w", err)
	}
	defer resp.Body.Close()
	var out struct {
		Code int `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fmt.Errorf("ai: decode: %w", err)
	}
	if resp.StatusCode != http.StatusOK || out.Code != 200 {
		return fmt.Errorf("ai: status=%d code=%d", resp.StatusCode, out.Code)
	}
	return nil
}