package media

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mybilibili/internal/abstraction"
)

// Pipeline orchestrates the MQ-driven video processing chain.
type Pipeline struct {
	transcoder *TranscodeWorker
	mq         abstraction.MessageQueue
	storage    abstraction.StorageService
	docStore   abstraction.DocumentStore
	search     abstraction.SearchEngine
	workDir    string
}

func NewPipeline(
	mq abstraction.MessageQueue,
	storage abstraction.StorageService,
	docStore abstraction.DocumentStore,
	search abstraction.SearchEngine,
	workDir string,
) *Pipeline {
	return &Pipeline{
		transcoder: NewTranscodeWorker(storage),
		mq:         mq,
		storage:    storage,
		docStore:   docStore,
		search:     search,
		workDir:    workDir,
	}
}

func (p *Pipeline) Start(ctx context.Context) error {
	ch, err := p.mq.Subscribe(ctx, TopicVideoProcess, "media-worker")
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}
	log.Println("media worker started, waiting for tasks...")

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
	audioFile := filepath.Join(dir, "audio.wav")
	key := fmt.Sprintf("manuscripts/%d/videos/%d/audio/audio.wav", task.ManuscriptID, task.VideoID)
	f, err := os.Open(audioFile)
	if err != nil {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "读取音频失败", 0, 7, err.Error())
		return
	}
	defer f.Close()
	p.storage.Put(ctx, "mybilibili", key, f, "audio/wav")
	p.emitProgress(task.VideoID, task.ManuscriptID, "audio", "音频提取完成", 100, 2, "")
}

func (p *Pipeline) doGenerateSubtitle(ctx context.Context, task ProcessMessage, dir string) {
	p.emitProgress(task.VideoID, task.ManuscriptID, "subtitle", "生成字幕", 10, 3, "")
	audioFile := filepath.Join(dir, "audio.wav")
	_ = audioFile
	// Whisper stub: in production, call whisper.cpp or OpenAI Whisper API via ServiceCaller
	subContent := `[{"index":1,"startTime":0,"endTime":1,"text":"..."}]`
	sub := map[string]interface{}{
		"video_id":      task.VideoID,
		"language":      "zh-CN",
		"language_name": "中文",
		"format":        "json",
		"content":       subContent,
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
	}
	p.mq.Publish(context.Background(), TopicVideoProgress, abstraction.Message{
		Topic: TopicVideoProgress, Payload: marshal(evt),
	})
}

func (p *Pipeline) downloadSource(ctx context.Context, sourceURL, dest string) error {
	// For now, stub: assumes source is already local or downloadable
	_ = sourceURL
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func (p *Pipeline) uploadTranscoded(ctx context.Context, task ProcessMessage, dir string) error {
	for _, quality := range []string{"1080p", "720p", "480p"} {
		playlist := filepath.Join(dir, quality, "index.m3u8")
		f, err := os.Open(playlist)
		if err != nil {
			continue
		}
		key := fmt.Sprintf("manuscripts/%d/videos/%d/transcoded/%s/index.m3u8", task.ManuscriptID, task.VideoID, quality)
		p.storage.Put(ctx, "mybilibili", key, f, "application/vnd.apple.mpegurl")
		f.Close()
	}
	return nil
}

var _ = os.RemoveAll
