package work

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMarshal(t *testing.T) {
	data := marshal(map[string]int{"a": 1})
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}
	var m map[string]int
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["a"] != 1 {
		t.Errorf("expected a=1, got %d", m["a"])
	}
}

func TestMarshal_Nil(t *testing.T) {
	data := marshal(nil)
	if string(data) != "null" {
		t.Errorf("expected null, got %s", string(data))
	}
}

func TestNow(t *testing.T) {
	s := now()
	_, err := time.Parse("2006-01-02T15:04:05Z", s)
	if err != nil {
		t.Fatalf("invalid time format: %v (value: %s)", err, s)
	}
}

func TestConstants(t *testing.T) {
	if ProcessTypeTranscode != "TRANSCODE" {
		t.Errorf("unexpected ProcessTypeTranscode: %s", ProcessTypeTranscode)
	}
	if ProcessTypeExtractAudio != "EXTRACT_AUDIO" {
		t.Errorf("unexpected ProcessTypeExtractAudio: %s", ProcessTypeExtractAudio)
	}
	if ProcessTypeGenerateSub != "GENERATE_SUBTITLE" {
		t.Errorf("unexpected ProcessTypeGenerateSub: %s", ProcessTypeGenerateSub)
	}
	if ProcessTypeAISummary != "AI_SUMMARY" {
		t.Errorf("unexpected ProcessTypeAISummary: %s", ProcessTypeAISummary)
	}
	if ProcessModeAutoChain != "AUTO_CHAIN" {
		t.Errorf("unexpected ProcessModeAutoChain: %s", ProcessModeAutoChain)
	}
	if ProcessModeManualSingle != "MANUAL_SINGLE" {
		t.Errorf("unexpected ProcessModeManualSingle: %s", ProcessModeManualSingle)
	}
	if TopicVideoProcess != "video-process-topic" {
		t.Errorf("unexpected TopicVideoProcess: %s", TopicVideoProcess)
	}
	if TopicVideoPublish != "video-publish-topic" {
		t.Errorf("unexpected TopicVideoPublish: %s", TopicVideoPublish)
	}
	if TopicVideoProgress != "video-process-progress-topic" {
		t.Errorf("unexpected TopicVideoProgress: %s", TopicVideoProgress)
	}
	if TopicManuscriptIndex != "manuscript-index-topic" {
		t.Errorf("unexpected TopicManuscriptIndex: %s", TopicManuscriptIndex)
	}
}

func TestProcessMessage_Marshal(t *testing.T) {
	msg := ProcessMessage{
		ManuscriptID: 1,
		VideoID:      2,
		UploaderID:   3,
		SourceURL:    "http://x",
		ProcessType:  ProcessTypeTranscode,
		ProcessMode:  ProcessModeAutoChain,
		Priority:     1,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded ProcessMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.ProcessType != ProcessTypeTranscode {
		t.Errorf("expected TRANSCODE, got %s", decoded.ProcessType)
	}
}

func TestProgressEvent_Marshal(t *testing.T) {
	event := ProgressEvent{
		VideoID:      1,
		ManuscriptID: 2,
		Title:        "test",
		Stage:        "transcode",
		StageText:    "transcoding",
		Progress:     50,
		Status:       1,
		Done:         false,
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded ProgressEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.Progress != 50 {
		t.Errorf("expected progress 50, got %d", decoded.Progress)
	}
}

func TestPublishEvent_Marshal(t *testing.T) {
	event := PublishEvent{
		ManuscriptID: 1,
		VideoID:      2,
		Trigger:      "auto",
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded PublishEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.Trigger != "auto" {
		t.Errorf("expected trigger auto, got %s", decoded.Trigger)
	}
}

func TestIndexEvent_Marshal(t *testing.T) {
	event := IndexEvent{
		ManuscriptID: 1,
		Operation:    "index",
		Trigger:      "publish",
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded IndexEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.Operation != "index" {
		t.Errorf("expected operation index, got %s", decoded.Operation)
	}
}
