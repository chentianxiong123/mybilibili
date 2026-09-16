package subtitle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validSRT = `1
00:00:01,000 --> 00:00:03,200
Hello world

2
00:00:03,200 --> 00:00:05,000
第二行字幕
`

func TestParseSRT_Valid(t *testing.T) {
	cues, err := ParseSRT(validSRT)
	require.NoError(t, err)
	require.Len(t, cues, 2)
	assert.Equal(t, 1, cues[0].Index)
	assert.Equal(t, 1*time.Second, cues[0].Start)
	assert.Equal(t, 3200*time.Millisecond, cues[0].End)
	assert.Equal(t, "Hello world", cues[0].Text)
	assert.Equal(t, 2, cues[1].Index)
	assert.Equal(t, 3200*time.Millisecond, cues[1].Start)
	assert.Equal(t, 5000*time.Millisecond, cues[1].End)
	assert.Equal(t, "第二行字幕", cues[1].Text)
}

func TestParseSRT_CRLF(t *testing.T) {
	content := "1\r\n00:00:00,000 --> 00:00:01,000\r\nline\r\n\r\n2\r\n00:00:01,000 --> 00:00:02,000\r\nnext"
	cues, err := ParseSRT(content)
	require.NoError(t, err)
	require.Len(t, cues, 2)
}

func TestParseSRT_MalformedSkipped(t *testing.T) {
	content := `garbage
00:00:01,000 --> 00:00:02,000
valid block
`
	cues, err := ParseSRT(content)
	require.NoError(t, err)
	// 第一块缺 index 行 → 被跳过（只有 3 行但第一行不是数字）
	// 实际 strings.SplitN(block, "\n", 3) 对 "garbage\n00:00:01..." 会得到 ["garbage", "00:00:01,000 --> ...", "valid block"]
	// index=garbage 解析失败 → skip
	assert.Empty(t, cues)
}

func TestParseSRT_InvalidTimeRangeSkipped(t *testing.T) {
	content := `1
00:00:01,000 00:00:02,000
no arrow here
`
	cues, err := ParseSRT(content)
	require.NoError(t, err)
	assert.Empty(t, cues)
}

func TestParseSRT_MultiLineText(t *testing.T) {
	content := "1\n00:00:01,000 --> 00:00:03,000\nline A\nline B"
	cues, err := ParseSRT(content)
	require.NoError(t, err)
	require.Len(t, cues, 1)
	assert.Equal(t, "line A\nline B", cues[0].Text)
}

func TestSRTCuesToJSON_Empty(t *testing.T) {
	out := SRTCuesToJSON(nil)
	assert.Equal(t, "[]", out)
}
