package work

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mybilibili/pkg/abstraction"
)

// Pipeline orchestrates the MQ-driven video processing chain.
type Pipeline struct {
	transcoderClient *TranscoderClient
	aiClient         *AIClient
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
	aiClient *AIClient,
	workDir string,
) *Pipeline {
	return &Pipeline{
		transcoderClient: transcoderClient,
		aiClient:         aiClient,
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

	// 字幕/whisper 转写已归位 ai-service（从 MinIO 读音频、写字幕对象）。
	if p.aiClient == nil {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "AI 客户端未配置", 0, 8, "ai client not configured")
		return
	}
	if err := p.aiClient.GenerateSubtitle(ctx, task.ManuscriptID, task.VideoID); err != nil {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "字幕生成失败", 0, 8, err.Error())
		return
	}
	p.emitProgress(task.VideoID, task.ManuscriptID, "subtitle", "字幕生成完成", 100, 3, "")
}

func (p *Pipeline) doAISummary(ctx context.Context, task ProcessMessage, dir string) {
	p.emitProgress(task.VideoID, task.ManuscriptID, "summary", "AI摘要", 10, 4, "")

	// AI 摘要已归位 ai-service（写 MinIO 摘要文件）。
	if p.aiClient == nil {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "AI 客户端未配置", 0, 9, "ai client not configured")
		return
	}
	if err := p.aiClient.GenerateSummary(ctx, task.ManuscriptID, task.VideoID); err != nil {
		p.emitProgress(task.VideoID, task.ManuscriptID, "failed", "AI摘要失败", 0, 9, err.Error())
		return
	}
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
