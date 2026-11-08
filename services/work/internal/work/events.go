package work

import (
	"context"
	"encoding/json"
	"time"
)

// Event types mirroring the Java MQ pipeline
const (
	ProcessTypeTranscode    = "TRANSCODE"
	ProcessTypeExtractAudio = "EXTRACT_AUDIO"
	ProcessTypeGenerateSub  = "GENERATE_SUBTITLE"
	ProcessTypeAISummary    = "AI_SUMMARY"

	ProcessModeAutoChain    = "AUTO_CHAIN"
	ProcessModeManualSingle = "MANUAL_SINGLE"

	TopicVideoProcess    = "video-process-topic"
	TopicVideoPublish    = "video-publish-topic"
	TopicVideoProgress   = "video-process-progress-topic"
	TopicManuscriptIndex = "manuscript-index-topic"
)

type ProcessMessage struct {
	ManuscriptID int64  `json:"manuscript_id"`
	VideoID      int64  `json:"video_id"`
	UploaderID   int64  `json:"uploader_id"`
	SourceURL    string `json:"source_url"`
	ProcessType  string `json:"process_type"`
	ProcessMode  string `json:"process_mode"`
	Priority     int32  `json:"priority"`
	// Retry 记录已重试次数（FINDING-020 修复：失败任务有限重试，避免 at-most-once 丢失）。
	Retry int `json:"retry"`
}

type ProgressEvent struct {
	VideoID      int64  `json:"video_id"`
	ManuscriptID int64  `json:"manuscript_id"`
	Title        string `json:"title"`
	Stage        string `json:"stage"`
	StageText    string `json:"stage_text"`
	Progress     int32  `json:"progress"`
	Status       int32  `json:"status"`
	StatusText   string `json:"status_text"`
	Error        string `json:"error"`
	Done         bool   `json:"done"`
	OccurredAt   string `json:"occurred_at"`
	// IsVertical 视频方向：0=横屏 1=竖屏；-1 表示未设置（非转码类事件默认值）
	IsVertical   int32  `json:"is_vertical"`
}

type PublishEvent struct {
	ManuscriptID int64  `json:"manuscript_id"`
	VideoID      int64  `json:"video_id"`
	Trigger      string `json:"trigger"`
}

type IndexEvent struct {
	ManuscriptID int64  `json:"manuscript_id"`
	Operation    string `json:"operation"`
	Trigger      string `json:"trigger"`
}

func marshal(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}

func now() string {
	// FINDING-049 修复：用 UTC 时间（此前 time.Now() 本地时间标注 "Z" 声明 UTC，
	// 非 UTC 服务器时间戳偏差（UTC+8 差 8h））。
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

var _ = marshal
var _ = now
var _ = context.Background
