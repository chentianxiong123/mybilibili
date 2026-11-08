package subtitle

// Pinning tests for confirmed subtitle SRT-parsing defects (FINDING-066 已修复).

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FINDING-066 修复：End < Start 负时长应报错。
func TestParseSRT_NegativeDurationAccepted_Pinned(t *testing.T) {
	_, err := ParseSRT("1\n00:00:05,000 --> 00:00:01,000\ntext")
	require.Error(t, err, "FINDING-066 修复：负时长 cue 应报错")
}

// FINDING-066 修复：乱序 cue 现按开始时间排序（不再接受重叠乱序）。
func TestParseSRT_OverlappingOutOfOrderAccepted_Pinned(t *testing.T) {
	srt := "1\n00:00:02,000 --> 00:00:04,000\nsecond\n\n2\n00:00:01,000 --> 00:00:03,000\nfirst"
	cues, err := ParseSRT(srt)
	require.NoError(t, err)
	require.Len(t, cues, 2)
	assert.True(t, cues[0].Start < cues[1].Start, "FINDING-066 修复：cue 应按开始时间排序")
}

func TestParseTime_OverLongMilliseconds_Pinned(t *testing.T) {
	_, err := parseTime("00:00:01,1234")
	require.Error(t, err, "FINDING-011 修复：ms>999 应报错")
}

func TestParseTime_NonNumericAccepted_Pinned(t *testing.T) {
	_, err := parseTime("abc:def:ghi")
	require.Error(t, err, "FINDING-011 修复：非数字应报错")
}

// FINDING-066 修复：垃圾输入返回错误（此前恒 nil + 空列表，调用方存 "[]"）。
func TestParseSRT_GarbageNeverErrors_Pinned(t *testing.T) {
	_, err := ParseSRT("this is not srt at all\n\n")
	require.Error(t, err, "FINDING-066 修复：垃圾输入应报错")
}

// FINDING-066 修复：cue 文本 HTML 转义（<script> 不再原样存储，防 stored-XSS）。
func TestParseSRT_PreservesRawHTML_Pinned(t *testing.T) {
	cues, err := ParseSRT("1\n00:00:00,000 --> 00:00:01,000\n<script>alert(1)</script>")
	require.NoError(t, err)
	require.Len(t, cues, 1)
	assert.NotContains(t, cues[0].Text, "<script>", "FINDING-066 修复：<script> 应被转义")
	assert.Contains(t, cues[0].Text, "&lt;script&gt;", "FINDING-066 修复：应转义为实体")
}
