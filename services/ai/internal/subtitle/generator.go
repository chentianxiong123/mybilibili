package subtitle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"mybilibili/pkg/abstraction"
)

// WhisperGenerator 从 MinIO 取已提取的音频 -> 调 Cloudflare Whisper -> 生成字幕对象入库。
// 承接原 work-service 的 whisper 逻辑（work 不再接触 ffmpeg/音频）。
type WhisperGenerator struct {
	repo    *Repository
	storage abstraction.StorageService
	httpc   *http.Client
}

func NewWhisperGenerator(repo *Repository, storage abstraction.StorageService) *WhisperGenerator {
	return &WhisperGenerator{
		repo:    repo,
		storage: storage,
		httpc:   &http.Client{Timeout: 300 * time.Second},
	}
}

// audioKey 与 transcoder-service 提取音频的落盘 key 对齐。
func audioKey(manuscriptID, videoID int64) string {
	return fmt.Sprintf("manuscripts/%d/videos/%d/audio/audio.mp3", manuscriptID, videoID)
}

// GenerateFromAudio 读取 MinIO 音频 -> whisper 转写 -> 写一个字幕对象（source=whisper, status=1）。
// 返回字幕 ID 与转写 cue 列表。
func (g *WhisperGenerator) GenerateFromAudio(ctx context.Context, manuscriptID, videoID int64) (string, []map[string]interface{}, error) {
	rc, err := g.storage.Get(ctx, "mybilibili", audioKey(manuscriptID, videoID))
	if err != nil {
		return "", nil, fmt.Errorf("subtitle: get audio: %w", err)
	}
	audioData, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		return "", nil, fmt.Errorf("subtitle: read audio: %w", err)
	}

	cues, err := callWhisperAPI(ctx, audioData, g.httpc)
	if err != nil {
		return "", nil, err
	}
	cuesJSON, _ := json.Marshal(cues)

	st := &Subtitle{
		VideoID:      videoID,
		Language:     "zh-CN",
		LanguageName: "中文",
		Format:       "json",
		Content:      string(cuesJSON),
		IsDefault:    true,
		Status:       1,
		Source:       "whisper",
	}
	id, err := g.repo.Create(ctx, st)
	if err != nil {
		return "", nil, fmt.Errorf("subtitle: save: %w", err)
	}
	return id, cues, nil
}

// callWhisperAPI 请求 Cloudflare Workers AI 的 whisper 转写接口。
// 未配置 CLOUDFLARE_AI_ACCOUNT_ID / CLOUDFLARE_AI_API_TOKEN 时返回占位 cue（便于无密钥环境跑通链路）。
func callWhisperAPI(ctx context.Context, audioData []byte, httpc *http.Client) ([]map[string]interface{}, error) {
	accountID := os.Getenv("CLOUDFLARE_AI_ACCOUNT_ID")
	apiToken := os.Getenv("CLOUDFLARE_AI_API_TOKEN")

	if accountID == "" || apiToken == "" {
		return []map[string]interface{}{
			{"index": 1, "startTime": 0.0, "endTime": 1.0, "text": "..."},
		}, nil
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := fw.Write(audioData); err != nil {
		return nil, fmt.Errorf("write audio data: %w", err)
	}
	w.Close()

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/@cf/openai/whisper", accountID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api status %d: %s", resp.StatusCode, string(body))
	}

	var whisperResp struct {
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
	if err := json.Unmarshal(body, &whisperResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if !whisperResp.Success {
		return nil, fmt.Errorf("whisper API returned success=false: %s", string(body))
	}

	cues := make([]map[string]interface{}, 0, len(whisperResp.Result.Segments))
	for _, seg := range whisperResp.Result.Segments {
		cues = append(cues, map[string]interface{}{
			"index":     seg.ID + 1,
			"startTime": seg.Start,
			"endTime":   seg.End,
			"text":      strings.TrimSpace(seg.Text),
		})
	}
	return cues, nil
}