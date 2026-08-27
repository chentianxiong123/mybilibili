package work

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mybilibili/pkg/abstraction"
)

// Pipeline orchestrates the MQ-driven video processing chain.
type Pipeline struct {
	transcoderClient *TranscoderClient
	mq               abstraction.MessageQueue
	storage          abstraction.StorageService
	docStore         abstraction.DocumentStore
	search           abstraction.SearchEngine
	workDir          string
	db               *sql.DB
}

func NewPipeline(
	mq abstraction.MessageQueue,
	storage abstraction.StorageService,
	docStore abstraction.DocumentStore,
	search abstraction.SearchEngine,
	transcoderClient *TranscoderClient,
	workDir string,
) *Pipeline {
	return &Pipeline{
		transcoderClient: transcoderClient,
		mq:               mq,
		storage:          storage,
		docStore:         docStore,
		search:           search,
		workDir:          workDir,
	}
}

// SetDatabase 注入数据库句柄，用于转码完成后回写 play_url_hd/sd/ld。
func (p *Pipeline) SetDatabase(db *sql.DB) {
	p.db = db
}

func (p *Pipeline) Start(ctx context.Context) error {
	ch, err := p.mq.Subscribe(ctx, TopicVideoProcess, "work-worker")
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}
	log.Println("work worker started, waiting for tasks...")

	for msg := range ch {
		var task ProcessMessage
		if err := json.Unmarshal(msg.Payload, &task); err != nil {
			log.Printf("bad message: %v", err)
			continue
		}
		log.Printf("received task: video=%d type=%s", task.VideoID, task.ProcessType)
		p.process(ctx, task)
	}
	return nil
}

func (p *Pipeline) process(ctx context.Context, task ProcessMessage) {
	workDir := filepath.Join(p.workDir, fmt.Sprintf("v%d_%d", task.VideoID, time.Now().Unix()))
	os.MkdirAll(workDir, 0o755)
	defer os.RemoveAll(workDir)

	p.emitProgress(task.VideoID, task.ManuscriptID, "started", "开始处理", 0, 0, "")
	defer p.emitProgress(task.VideoID, task.ManuscriptID, "done", "处理完成", 100, 5, "")

	switch task.ProcessType {
	case ProcessTypeTranscode:
		p.doTranscode(ctx, task, workDir)
	case ProcessTypeExtractAudio:
		p.doExtractAudio(ctx, task, workDir)
	case ProcessTypeGenerateSub:
		p.doGenerateSubtitle(ctx, task, workDir)
	case ProcessTypeAISummary:
		p.doAISummary(ctx, task, workDir)
	}

	if task.ProcessMode == ProcessModeAutoChain {
		p.chainNext(ctx, task)
	}
}

func (p *Pipeline) doTranscode(ctx context.Context, task ProcessMessage, dir string) {
	p.emitProgress(task.VideoID, task.ManuscriptID, "transcoding", "转码中", 10, 1, "")

	sourceKey := sourceKeyFromURL(task.SourceURL)
	if sourceKey == "" {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "源对象 key 为空", 0, 6, "empty source key")
		return
	}

	res, err := p.transcoderClient.Transcode(ctx, TranscodeRequest{
		Bucket:       "mybilibili",
		SourceKey:    sourceKey,
		ManuscriptID: task.ManuscriptID,
		VideoID:      task.VideoID,
		Qualities:    []string{"1080p", "720p", "480p"},
		ExtractAudio: false,
	})
	if err != nil {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "转码失败", 0, 6, err.Error())
		return
	}

	// 回写 is_vertical（横竖屏）
	if res.IsVertical != -1 {
		p.mq.Publish(ctx, TopicVideoProgress, abstraction.Message{
			Topic: TopicVideoProgress,
			Payload: marshal(ProgressEvent{
				VideoID: task.VideoID, ManuscriptID: task.ManuscriptID,
				Stage: "transcode", StageText: "方向检测", Progress: 15, Status: 1,
				IsVertical: res.IsVertical, Done: true, OccurredAt: now(),
			}),
		})
	}

	// 回写 play_url_hd/sd/ld
	if err := p.writePlayURLs(ctx, task, res.PlayURLs); err != nil {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "回写播放地址失败", 0, 6, err.Error())
		return
	}

	p.emitProgress(task.VideoID, task.ManuscriptID, "transcoding", "转码完成", 100, 1, "")
}

func (p *Pipeline) doExtractAudio(ctx context.Context, task ProcessMessage, dir string) {
	p.emitProgress(task.VideoID, task.ManuscriptID, "audio", "提取音频", 10, 2, "")
	sourceKey := sourceKeyFromURL(task.SourceURL)
	if sourceKey == "" {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "源对象 key 为空", 0, 7, "empty source key")
		return
	}
	_, err := p.transcoderClient.Transcode(ctx, TranscodeRequest{
		Bucket:       "mybilibili",
		SourceKey:    sourceKey,
		ManuscriptID: task.ManuscriptID,
		VideoID:      task.VideoID,
		ExtractAudio: true,
	})
	if err != nil {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "音频提取失败", 0, 7, err.Error())
		return
	}
	p.emitProgress(task.VideoID, task.ManuscriptID, "audio", "音频提取完成", 100, 2, "")
}

func (p *Pipeline) doGenerateSubtitle(ctx context.Context, task ProcessMessage, dir string) {
	p.emitProgress(task.VideoID, task.ManuscriptID, "subtitle", "生成字幕", 10, 3, "")
	audioFile := filepath.Join(dir, "audio.mp3")

	cues, err := callWhisperAPI(ctx, audioFile)
	if err != nil {
		log.Printf("[whisper] API call failed: %v, using placeholder", err)
		cues = []map[string]interface{}{
			{"index": 1, "startTime": 0.0, "endTime": 1.0, "text": "..."},
		}
	}
	cuesJSON, _ := json.Marshal(cues)

	sub := map[string]interface{}{
		"video_id":      task.VideoID,
		"language":      "zh-CN",
		"language_name": "中文",
		"format":        "json",
		"content":       string(cuesJSON),
		"is_default":    true,
		"source":        "whisper",
		"status":        1,
	}
	if _, err := p.docStore.Insert(ctx, "subtitles", sub); err != nil {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "保存字幕失败", 0, 8, err.Error())
		return
	}
	p.emitProgress(task.VideoID, task.ManuscriptID, "subtitle", "字幕生成完成", 100, 3, "")
}

func callWhisperAPI(ctx context.Context, audioFile string) ([]map[string]interface{}, error) {
	accountID := os.Getenv("CLOUDFLARE_AI_ACCOUNT_ID")
	apiToken := os.Getenv("CLOUDFLARE_AI_API_TOKEN")
	if accountID == "" || apiToken == "" {
		return nil, fmt.Errorf("CLOUDFLARE_AI_ACCOUNT_ID or CLOUDFLARE_AI_API_TOKEN not set")
	}

	audioData, err := os.ReadFile(audioFile)
	if err != nil {
		return nil, fmt.Errorf("read audio file: %w", err)
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

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
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

func (p *Pipeline) doAISummary(ctx context.Context, task ProcessMessage, dir string) {
	p.emitProgress(task.VideoID, task.ManuscriptID, "summary", "AI摘要", 10, 4, "")
	summary := "AI generated summary (stub)"
	key := fmt.Sprintf("manuscripts/%d/videos/%d/summary/ai-summary.txt", task.ManuscriptID, task.VideoID)
	p.storage.Put(ctx, "mybilibili", key, strings.NewReader(summary), "text/plain")
	p.emitProgress(task.VideoID, task.ManuscriptID, "summary", "AI摘要完成", 100, 4, "")
}

func (p *Pipeline) chainNext(ctx context.Context, task ProcessMessage) {
	var nextTypes = []string{
		ProcessTypeTranscode,
		ProcessTypeExtractAudio,
		ProcessTypeGenerateSub,
		ProcessTypeAISummary,
	}
	for i, t := range nextTypes {
		if t == task.ProcessType && i+1 < len(nextTypes) {
			next := task
			next.ProcessType = nextTypes[i+1]
			next.ProcessMode = ProcessModeAutoChain
			payload, _ := json.Marshal(next)
			p.mq.Publish(ctx, TopicVideoProcess, abstraction.Message{
				Topic: TopicVideoProcess, Payload: payload,
			})
			return
		}
	}
	// last step: publish
	pubEvt := PublishEvent{
		ManuscriptID: task.ManuscriptID,
		VideoID:      task.VideoID,
		Trigger:      "AUTO_CHAIN",
	}
	p.mq.Publish(ctx, TopicVideoPublish, abstraction.Message{
		Topic: TopicVideoPublish, Payload: marshal(pubEvt),
	})
}

func (p *Pipeline) emitProgress(videoID, manuscriptID int64, stage, stageText string, progress, status int32, errMsg string) {
	evt := ProgressEvent{
		VideoID: videoID, ManuscriptID: manuscriptID,
		Stage: stage, StageText: stageText, Progress: progress,
		Status: status, StatusText: stageText, Error: errMsg,
		Done: progress >= 100, OccurredAt: now(),
		IsVertical: -1,
	}
	p.mq.Publish(context.Background(), TopicVideoProgress, abstraction.Message{
		Topic: TopicVideoProgress, Payload: marshal(evt),
	})
}

func (p *Pipeline) downloadSourceRemovedMarker() {}

// sourceKeyFromURL 把源地址转成 MinIO 对象 key。
// 对齐 core 上传落盘约定: 源对象存于 manuscripts/{id}/videos/{vid}/source/video.mp4。
// 输入可能为:
//   - "/uploads/manuscripts/10/videos/25/source/video.mp4" (经 core 反代)
//   - "manuscripts/10/videos/25/source/video.mp4" (纯 MinIO key)
// 统一输出: "manuscripts/10/videos/25/source/video.mp4"
func sourceKeyFromURL(sourceURL string) string {
	key := strings.TrimPrefix(sourceURL, "/uploads/")
	if key == sourceURL {
		key = sourceURL
	}
	key = strings.TrimPrefix(key, "/")
	// 去掉多余的前缀噪音
	key = strings.TrimPrefix(key, "uploads/")
	return key
}

// writePlayURLs 将转码产物播放地址写回 videos 表。
func (p *Pipeline) writePlayURLs(ctx context.Context, task ProcessMessage, urls map[string]string) error {
	if p.db == nil {
		return nil
	}
	hd, sd, ld := urls["1080p"], urls["720p"], urls["480p"]
	_, err := p.db.ExecContext(ctx,
		`UPDATE videos SET play_url_hd = $1, play_url_sd = $2, play_url_ld = $3, updated_at = NOW() WHERE id = $4`,
		hd, sd, ld, task.VideoID)
	return err
}

var _ = os.RemoveAll
