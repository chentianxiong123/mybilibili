package work

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSourceKeyFromURL 验证 sourceKeyFromURL 的核心解析逻辑。
func TestSourceKeyFromURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"raw key", "manuscripts/10/videos/25/source/video.mp4", "manuscripts/10/videos/25/source/video.mp4"},
		{"with /uploads/ prefix", "/uploads/manuscripts/10/videos/25/source/video.mp4", "manuscripts/10/videos/25/source/video.mp4"},
		{"with /uploads/ prefix (no leading slash)", "uploads/manuscripts/x.mp4", "manuscripts/x.mp4"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sourceKeyFromURL(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestProcessTypes_Constants 验证 ProcessType 常量值。
func TestProcessTypes_Constants(t *testing.T) {
	assert.Equal(t, "TRANSCODE", ProcessTypeTranscode)
	assert.Equal(t, "EXTRACT_AUDIO", ProcessTypeExtractAudio)
	assert.Equal(t, "GENERATE_SUBTITLE", ProcessTypeGenerateSub)
	assert.Equal(t, "AI_SUMMARY", ProcessTypeAISummary)
}

func TestProcessModes_Constants(t *testing.T) {
	assert.Equal(t, "AUTO_CHAIN", ProcessModeAutoChain)
	assert.Equal(t, "MANUAL_SINGLE", ProcessModeManualSingle)
}

func TestTopics_Constants(t *testing.T) {
	assert.Equal(t, "video-process-topic", TopicVideoProcess)
	assert.Equal(t, "video-publish-topic", TopicVideoPublish)
	assert.Equal(t, "video-process-progress-topic", TopicVideoProgress)
	assert.Equal(t, "manuscript-index-topic", TopicManuscriptIndex)
}

// 防 unused 警告
var _ = filepath.Join