package work

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPipeline_FailedTaskMaskedByDone 钉住 FINDING-020：
// process 的 defer emitProgress("done") 无条件触发，掩盖子任务 doTranscode
// 发出的 "failed"。空 SourceURL → doTranscode :102 发 "failed"（status 6），
// 但 process 返回后 defer :79 仍发 "done"（status 5）→ 前端最终看到 "done"。
func TestPipeline_FailedTaskMaskedByDone(t *testing.T) {
	mq := &mockSubscribeMQ{}
	p := NewPipeline(mq, nil, nil, nil, nil, nil, t.TempDir())

	task := ProcessMessage{
		ManuscriptID: 10,
		VideoID:      1,
		SourceURL:    "", // 空 → doTranscode 发 "failed"（status 6）
		ProcessType:  ProcessTypeTranscode,
		ProcessMode:  ProcessModeManualSingle,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 直接调用 process（同包可访问 unexported）
	p.process(ctx, task)

	require.NotEmpty(t, mq.published, "expected progress events")

	var sawFailed, sawDone bool
	for _, ev := range mq.published {
		payload := string(ev.Payload)
		if payload == "" {
			continue
		}
		if containsSubstr(payload, `"failed"`) {
			sawFailed = true
		}
		if containsSubstr(payload, `"done"`) {
			sawDone = true
		}
	}
	assert.True(t, sawFailed, "expected 'failed' event for empty source key")
	assert.True(t, sawDone, "FINDING-020: expected 'done' ALSO emitted (masking bug) even after 'failed'")
}

func containsSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
