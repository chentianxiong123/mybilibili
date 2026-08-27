package work

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TranscodeRequest 与 transcoder-service 的请求体对齐（MinIO 对象引用）。
type TranscodeRequest struct {
	Bucket       string   `json:"bucket"`
	SourceKey    string   `json:"source_key"`
	ManuscriptID int64    `json:"manuscript_id"`
	VideoID      int64    `json:"video_id"`
	Qualities    []string `json:"qualities"`
	ExtractAudio bool     `json:"extract_audio"`
}

// TranscodeResult transcoder 返回：转码产物播放地址 + 音频对象 key + 方向。
type TranscodeResult struct {
	PlayURLs map[string]string `json:"play_urls"`
	AudioKey string            `json:"audio_key,omitempty"`
	// IsVertical 视频方向：0=横屏 1=竖屏 -1=未知
	IsVertical int32 `json:"is_vertical"`
}

// TranscoderClient 调用独立的 transcoder-service（ffmpeg 媒体处理）。
// transcoder 负责从 MinIO 读源、本地 ffmpeg 处理、产物写回 MinIO；
// work 作为编排器只传对象引用，不接触 ffmpeg 与文件内容。
type TranscoderClient struct {
	baseURL string
	httpc   *http.Client
}

func NewTranscoderClient(baseURL string) *TranscoderClient {
	return &TranscoderClient{
		baseURL: baseURL,
		httpc:   &http.Client{Timeout: 30 * time.Minute},
	}
}

// Transcode 提交一次转码（可选含提音频）。sourceKey 为 MinIO 对象 key。
func (c *TranscoderClient) Transcode(ctx context.Context, req TranscodeRequest) (*TranscodeResult, error) {
	url := c.baseURL + "/api/v1/transcode"
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpc.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("transcoder: %w", err)
	}
	defer resp.Body.Close()

	var out struct {
		Code int             `json:"code"`
		Data *TranscodeResult `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("transcoder: decode: %w", err)
	}
	if resp.StatusCode != http.StatusOK || out.Code != 200 || out.Data == nil {
		return nil, fmt.Errorf("transcoder: status=%d code=%d", resp.StatusCode, out.Code)
	}
	return out.Data, nil
}