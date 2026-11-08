package subtitle

// Pinning tests for confirmed subtitle SRT-parsing defects.
// Each test asserts the CURRENT (buggy) behavior; it will fail the moment the
// bug is fixed, acting as a regression anchor documenting the defect.

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BUG: ParseSRT accepts a cue whose End < Start (negative duration). No
// validation that End > Start exists. File: internal/subtitle/srt_parser.go:35-40.
func TestParseSRT_NegativeDurationAccepted_Pinned(t *testing.T) {
	cues, err := ParseSRT("1\n00:00:05,000 --> 00:00:01,000\ntext")
	require.NoError(t, err)
	require.Len(t, cues, 1)
	// The parser returns the cue with End(1s) < Start(5s) and no error.
	assert.True(t, cues[0].End < cues[0].Start, "pinned: negative-duration cue accepted")
}

// BUG: ParseSRT accepts out-of-order / overlapping cues (cue 2 starts before
// cue 1). No ordering or overlap validation. File: srt_parser.go:35-40.
func TestParseSRT_OverlappingOutOfOrderAccepted_Pinned(t *testing.T) {
	srt := "1\n00:00:02,000 --> 00:00:04,000\nsecond\n\n2\n00:00:01,000 --> 00:00:03,000\nfirst"
	cues, err := ParseSRT(srt)
	require.NoError(t, err)
	require.Len(t, cues, 2)
	assert.True(t, cues[1].Start < cues[0].Start, "pinned: out-of-order/overlapping cues accepted")
	assert.True(t, cues[1].End > cues[0].Start, "pinned: overlapping cue ranges accepted")
}

// BUG: parseTime treats any fractional-seconds digit group as a bare integer
// of milliseconds, so "00:00:01,1234" is parsed as 1s + 1234ms = 2.234s
// FINDING-011 修复后：ms>999 → error（不再静默溢出/归一化）。
func TestParseTime_OverLongMilliseconds_Pinned(t *testing.T) {
	_, err := parseTime("00:00:01,1234")
	require.Error(t, err, "FINDING-011 修复：ms>999 应报错")
}

// FINDING-011 修复后：非数字 → error（不再静默 0s）。
func TestParseTime_NonNumericAccepted_Pinned(t *testing.T) {
	_, err := parseTime("abc:def:ghi")
	require.Error(t, err, "FINDING-011 修复：非数字应报错")
}

// BUG: ParseSRT never returns an error (signature always yields nil), so a
// completely invalid / garbage SRT is reported as a successful parse with an
// empty cue list. Callers (handleUploadSRT, handleImportSystem) then replace
// content with "[]" instead of rejecting the input.
// File: srt_parser.go:18-43.
func TestParseSRT_GarbageNeverErrors_Pinned(t *testing.T) {
	cues, err := ParseSRT("this is not srt at all\n\n")
	require.NoError(t, err, "pinned: ParseSRT never errors on invalid input")
	assert.Empty(t, cues)
}

// BUG: cue text is stored verbatim with no HTML sanitization. A subtitle line
// containing <script> survives parsing and JSON serialization unchanged
// (json.Marshal escapes '<' to \u003c but the decoded value is still raw HTML),
// a stored-XSS vector if any frontend renders text via innerHTML.
// File: srt_parser.go:39-40, 78-85; subtitle_handler.go:71-102.
func TestParseSRT_PreservesRawHTML_Pinned(t *testing.T) {
	cues, err := ParseSRT("1\n00:00:00,000 --> 00:00:01,000\n<script>alert(1)</script>")
	require.NoError(t, err)
	require.Len(t, cues, 1)
	assert.Equal(t, "<script>alert(1)</script>", cues[0].Text, "pinned: no sanitization of cue text")
	assert.Contains(t, SRTCuesToJSON(cues), "script", "pinned: raw HTML survives JSON round-trip")
}
