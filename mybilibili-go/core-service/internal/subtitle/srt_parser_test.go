package subtitle

import (
	"testing"
)

func TestParseSRT(t *testing.T) {
	srt := `1
00:00:01,000 --> 00:00:04,000
Hello world

2
00:00:05,000 --> 00:00:08,500
This is a test
Second line`

	cues, err := ParseSRT(srt)
	if err != nil {
		t.Fatalf("ParseSRT: %v", err)
	}
	if len(cues) != 2 {
		t.Fatalf("expected 2 cues, got %d", len(cues))
	}
	if cues[0].Index != 1 {
		t.Fatalf("expected cue 0 index 1, got %d", cues[0].Index)
	}
	if cues[0].Text != "Hello world" {
		t.Fatalf("expected 'Hello world', got '%s'", cues[0].Text)
	}
	if cues[1].Text != "This is a test\nSecond line" {
		t.Fatalf("expected multi-line text, got '%s'", cues[1].Text)
	}
}

func TestParseSRTInvalid(t *testing.T) {
	_, err := ParseSRT("not valid srt content")
	if err != nil {
		t.Fatalf("expected no error for empty/parseable, got %v", err)
	}
}