package subtitle

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- SRTCue.ToJSON ---

func TestSRTCue_ToJSON(t *testing.T) {
	cue := SRTCue{
		Index: 1,
		Start: 5*time.Second + 500*time.Millisecond,
		End:   10 * time.Second,
		Text:  "hello world",
	}
	m := cue.ToJSON()
	assert.Equal(t, 1, m["index"])
	assert.Equal(t, "5.5s", m["start"])
	assert.Equal(t, "10s", m["end"])
	assert.Equal(t, "hello world", m["text"])
}

func TestSRTCue_ToJSON_Zero(t *testing.T) {
	cue := SRTCue{Index: 0}
	m := cue.ToJSON()
	assert.Equal(t, "0s", m["start"])
	assert.Equal(t, "0s", m["end"])
}

// --- SRTCue.ToCueMap ---

func TestSRTCue_ToCueMap(t *testing.T) {
	cue := SRTCue{
		Index: 2,
		Start: 1500 * time.Millisecond,
		End:   3 * time.Second,
		Text:  "test",
	}
	m := cue.ToCueMap()
	assert.Equal(t, 2, m["index"])
	assert.Equal(t, 1.5, m["startTime"])
	assert.Equal(t, 3.0, m["endTime"])
	assert.Equal(t, "test", m["text"])
}

// --- SRTCuesToJSON ---

func TestSRTCuesToJSON_NonEmpty(t *testing.T) {
	cues := []SRTCue{
		{Index: 1, Start: 0, End: time.Second, Text: "first"},
		{Index: 2, Start: time.Second, End: 2 * time.Second, Text: "second"},
	}
	result := SRTCuesToJSON(cues)
	assert.NotEmpty(t, result)

	var parsed []map[string]interface{}
	err := json.Unmarshal([]byte(result), &parsed)
	require.NoError(t, err)
	assert.Len(t, parsed, 2)
	assert.Equal(t, float64(1), parsed[0]["index"])
	assert.Equal(t, 0.0, parsed[0]["startTime"])
	assert.Equal(t, 1.0, parsed[0]["endTime"])
	assert.Equal(t, "first", parsed[0]["text"])
}

// Note: TestSRTCuesToJSON_Empty is in srt_parser_test.go

// --- ParseSRT edge cases ---

func TestParseSRT_Empty(t *testing.T) {
	cues, err := ParseSRT("")
	require.NoError(t, err)
	assert.Nil(t, cues)
}

func TestParseSRT_WhitespaceOnly(t *testing.T) {
	cues, err := ParseSRT("   \n\n   ")
	require.NoError(t, err)
	assert.Nil(t, cues)
}

func TestParseSRT_SingleCue(t *testing.T) {
	srt := "1\n00:00:01,000 --> 00:00:02,000\nhello"
	cues, err := ParseSRT(srt)
	require.NoError(t, err)
	assert.Len(t, cues, 1)
	assert.Equal(t, 1, cues[0].Index)
	assert.Equal(t, time.Second, cues[0].Start)
	assert.Equal(t, 2*time.Second, cues[0].End)
	assert.Equal(t, "hello", cues[0].Text)
}

func TestParseSRT_DotSeparator(t *testing.T) {
	srt := "1\n00:00:01.000 --> 00:00:02.000\ntest"
	cues, err := ParseSRT(srt)
	require.NoError(t, err)
	assert.Len(t, cues, 1)
	assert.Equal(t, time.Second, cues[0].Start)
}

func TestParseSRT_InvalidIndex_Skipped(t *testing.T) {
	srt := "not-a-number\n00:00:01,000 --> 00:00:02,000\nhello"
	cues, err := ParseSRT(srt)
	require.NoError(t, err)
	assert.Empty(t, cues)
}

func TestParseSRT_TooFewLines_Skipped(t *testing.T) {
	srt := "1\n00:00:01,000 --> 00:00:02,000"
	cues, err := ParseSRT(srt)
	require.NoError(t, err)
	assert.Empty(t, cues)
}

func TestParseSRT_InvalidTimeRange_Skipped(t *testing.T) {
	srt := "1\ninvalid time range\nhello"
	cues, err := ParseSRT(srt)
	require.NoError(t, err)
	assert.Empty(t, cues)
}

func TestParseSRT_MultipleBlocks(t *testing.T) {
	srt := "1\n00:00:00,000 --> 00:00:01,000\nfirst\n\n2\n00:00:01,000 --> 00:00:02,000\nsecond"
	cues, err := ParseSRT(srt)
	require.NoError(t, err)
	assert.Len(t, cues, 2)
	assert.Equal(t, "first", cues[0].Text)
	assert.Equal(t, "second", cues[1].Text)
}

// Note: TestParseSRT_CRLF and TestSRTCuesToJSON_Empty are in srt_parser_test.go

func TestParseSRT_TimeWithHours(t *testing.T) {
	srt := "1\n01:30:00,000 --> 02:00:00,000\nlong"
	cues, err := ParseSRT(srt)
	require.NoError(t, err)
	assert.Len(t, cues, 1)
	assert.Equal(t, 90*time.Minute, cues[0].Start)
	assert.Equal(t, 2*time.Hour, cues[0].End)
}

func TestParseSRT_MillisecondPrecision(t *testing.T) {
	srt := "1\n00:00:00,123 --> 00:00:00,456\nprecise"
	cues, err := ParseSRT(srt)
	require.NoError(t, err)
	assert.Len(t, cues, 1)
	assert.Equal(t, 123*time.Millisecond, cues[0].Start)
	assert.Equal(t, 456*time.Millisecond, cues[0].End)
}

func TestParseSRT_EmptyBlock_Skipped(t *testing.T) {
	srt := "\n\n1\n00:00:00,000 --> 00:00:01,000\nhello"
	cues, err := ParseSRT(srt)
	require.NoError(t, err)
	assert.Len(t, cues, 1)
}

// --- parseTime edge cases ---

func TestParseTime_WithoutMilliseconds(t *testing.T) {
	d, err := parseTime("00:01:30")
	require.NoError(t, err)
	assert.Equal(t, 90*time.Second, d)
}

func TestParseTime_InvalidParts(t *testing.T) {
	_, err := parseTime("00:01")
	assert.Error(t, err)
}

func TestParseTime_FourParts(t *testing.T) {
	_, err := parseTime("00:01:30:00")
	assert.Error(t, err)
}

func TestParseTimeRange_SingleArrow(t *testing.T) {
	_, _, err := parseTimeRange("00:00:00,000 -> 00:00:01,000")
	assert.Error(t, err)
}

func TestParseTimeRange_Empty(t *testing.T) {
	_, _, err := parseTimeRange("")
	assert.Error(t, err)
}

func TestParseTimeRange_ExtraSpaces(t *testing.T) {
	start, end, err := parseTimeRange("  00:00:00,000  -->  00:00:01,000  ")
	require.NoError(t, err)
	assert.Equal(t, time.Duration(0), start)
	assert.Equal(t, time.Second, end)
}
