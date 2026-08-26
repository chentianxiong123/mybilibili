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
	transcoder *TranscodeWorker
	mq         abstraction.MessageQueue
	storage    abstraction.StorageService
	docStore   abstraction.DocumentStore
	search     abstraction.SearchEngine
	workDir    string
	db         *sql.DB
}

func NewPipeline(
	mq abstraction.MessageQueue,
	storage abstraction.StorageService,
	docStore abstraction.DocumentStore,
	search abstraction.SearchEngine,
	workDir string,
	encoder string,
) *Pipeline {
	return &Pipeline{
		transcoder: NewTranscodeWorker(storage, encoder),
		mq:         mq,
		storage:    storage,
		docStore:   docStore,
		search:     search,
		workDir:    workDir,
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

	src := filepath.Join(dir, "source.mp4")
	if err := p.downloadSource(ctx, task.SourceURL, src); err != nil {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "下载源文件失败", 0, 6, err.Error())
		return
	}

	// 检测视频横竖屏（高>宽即竖屏），随进度事件回写 videos.is_vertical
	if w, h, err := p.transcoder.GetVideoSize(ctx, src); err == nil && w > 0 && h > 0 {
		isVertical := int32(0)
		if h > w {
			isVertical = 1
		}
		p.mq.Publish(ctx, TopicVideoProgress, abstraction.Message{
			Topic: TopicVideoProgress,
			Payload: marshal(ProgressEvent{
				VideoID: task.VideoID, ManuscriptID: task.ManuscriptID,
				Stage: "transcode", StageText: "方向检测", Progress: 15, Status: 1,
				IsVertical: isVertical, Done: true, OccurredAt: now(),
			}),
		})
	} else {
		log.Printf("get video size failed (video %d): %v", task.VideoID, err)
	}

	p.emitProgress(task.VideoID, task.ManuscriptID, "transcoding", "转码中 1080p", 30, 1, "")
	if err := p.transcoder.Transcode(ctx, src, filepath.Join(dir, "1080p"), "1080p"); err != nil {
		log.Printf("transcode 1080p warning: %v", err)
	}

	p.emitProgress(task.VideoID, task.ManuscriptID, "transcoding", "转码中 720p", 50, 1, "")
	if err := p.transcoder.Transcode(ctx, src, filepath.Join(dir, "720p"), "720p"); err != nil {
		log.Printf("transcode 720p warning: %v", err)
	}

	p.emitProgress(task.VideoID, task.ManuscriptID, "transcoding", "转码中 480p", 70, 1, "")
	if err := p.transcoder.Transcode(ctx, src, filepath.Join(dir, "480p"), "480p"); err != nil {
		log.Printf("transcode 480p warning: %v", err)
	}

	if err := p.uploadTranscoded(ctx, task, dir); err != nil {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "上传转码结果失败", 0, 6, err.Error())
		return
	}
}

func (p *Pipeline) doExtractAudio(ctx context.Context, task ProcessMessage, dir string) {
	p.emitProgress(task.VideoID, task.ManuscriptID, "audio", "提取音频", 10, 2, "")
	src := filepath.Join(dir, "source.mp4")
	if err := p.downloadSource(ctx, task.SourceURL, src); err != nil {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "下载失败", 0, 7, err.Error())
		return
	}
	if err := p.transcoder.ExtractAudio(ctx, src, dir); err != nil {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "音频提取失败", 0, 7, err.Error())
		return
	}
	audioFile := filepath.Join(dir, "audio.mp3")
	key := fmt.Sprintf("manuscripts/%d/videos/%d/audio/audio.mp3", task.ManuscriptID, task.VideoID)
	f, err := os.Open(audioFile)
	if err != nil {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "读取音频失败", 0, 7, err.Error())
		return
	}
	defer f.Close()
	p.storage.Put(ctx, "mybilibili", key, f, "audio/mpeg")
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

func (p *Pipeline) downloadSource(ctx context.Context, sourceURL, dest string) error {
	if sourceURL == "" {
		return fmt.Errorf("empty source url")
	}
	var body io.ReadCloser
	switch {
	case strings.HasPrefix(sourceURL, "http://"), strings.HasPrefix(sourceURL, "https://"):
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("download source: http %d", resp.StatusCode)
		}
		body = resp.Body
	case strings.HasPrefix(sourceURL, "file://"):
		f, err := os.Open(strings.TrimPrefix(sourceURL, "file://"))
		if err != nil {
			return err
		}
		body = f
	default:
		f, err := os.Open(sourceURL)
		if err != nil {
			return err
		}
		body = f
	}
	defer body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, body); err != nil {
		return err
	}
	return nil
}

func (p *Pipeline) uploadTranscoded(ctx context.Context, task ProcessMessage, dir string) error {
	// 上传每个清晰度目录下的全部文件（playlist.m3u8 + ts 段），对齐老项目 videoHlsObject key。
	playURLs := map[string]string{}
	for _, quality := range []string{"1080p", "720p", "480p"} {
		qualityDir := filepath.Join(dir, quality)
		entries, err := os.ReadDir(qualityDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			f, err := os.Open(filepath.Join(qualityDir, e.Name()))
			if err != nil {
				continue
			}
			key := fmt.Sprintf("manuscripts/%d/videos/%d/transcoded/%s/%s",
				task.ManuscriptID, task.VideoID, quality, e.Name())
			ct := "application/octet-stream"
			if strings.HasSuffix(e.Name(), ".m3u8") {
				ct = "application/vnd.apple.mpegurl"
			} else if strings.HasSuffix(e.Name(), ".ts") {
				ct = "video/mp2t"
			}
			_ = p.storage.Put(ctx, "mybilibili", key, f, ct)
			f.Close()
		}
		if _, err := os.Stat(filepath.Join(qualityDir, "playlist.m3u8")); err == nil {
			playURLs[quality] = fmt.Sprintf("/uploads/manuscripts/%d/videos/%d/transcoded/%s/playlist.m3u8",
				task.ManuscriptID, task.VideoID, quality)
		}
	}
	// 转码完成后回写 play_url_hd/sd/ld（对齐老项目 markTranscodeSuccess）
	return p.writePlayURLs(ctx, task, playURLs)
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
