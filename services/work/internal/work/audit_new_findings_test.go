package work

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// FINDING A (NEW, correctness / state machine):
//   pipeline.go:92-94 — process() calls chainNext() unconditionally when
//   ProcessMode==AUTO_CHAIN, even when the current step FAILED. doTranscode
//   only emits a "failed" progress event and returns without an error; process
//   has no failure signal to gate the chain. So a failed TRANSCODE still chains
//   to EXTRACT_AUDIO -> ... -> PUBLISH, publishing a video with no play URLs.
//
// This is distinct from FINDING-020 (the unconditional "done" emit): here the
// *chain itself* advances past a failed step.
// ---------------------------------------------------------------------------
func TestAudit_ChainAdvancesAfterTranscodeFailure(t *testing.T) {
	// transcoder always returns 500 -> doTranscode fails on all 3 retries
	transcodeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 500})
	}))
	defer transcodeSrv.Close()

	mq := &mockMQ{}
	cfgFile := filepath.Join(t.TempDir(), "pool.yaml")
	pool, err := NewTranscoderPool(cfgFile)
	require.NoError(t, err)
	require.NoError(t, pool.Add(Node{Name: "t1", Addr: transcodeSrv.URL, Weight: 1}))

	p := NewPipeline(mq, nil, nil, nil, pool, nil, t.TempDir())

	task := ProcessMessage{
		ManuscriptID: 10,
		VideoID:      1,
		SourceURL:    "/uploads/manuscripts/10/videos/1/source/video.mp4",
		ProcessType:  ProcessTypeTranscode,
		ProcessMode:  ProcessModeAutoChain,
	}
	p.process(context.Background(), task)

	// (1) transcode must have failed
	sawFailed := false
	for _, msg := range mq.published {
		if msg.Topic != TopicVideoProgress {
			continue
		}
		var evt ProgressEvent
		_ = json.Unmarshal(msg.Payload, &evt)
		if evt.Stage == "failed" {
			sawFailed = true
		}
	}
	require.True(t, sawFailed, "precondition: transcode emitted 'failed'")

	// (2) FINDING-020 修复：失败后不再 chain 到下一步（此前 BUG 在 TRANSCODE 失败后仍发布 EXTRACT_AUDIO）。
	chained := false
	for _, msg := range mq.published {
		if msg.Topic != TopicVideoProcess {
			continue
		}
		var next ProcessMessage
		_ = json.Unmarshal(msg.Payload, &next)
		if next.ProcessType == ProcessTypeExtractAudio {
			chained = true
		}
	}
	assert.False(t, chained,
		"FINDING-020 修复：TRANSCODE 失败后不应 chain 到 EXTRACT_AUDIO")
}

// ---------------------------------------------------------------------------
// FINDING B (NEW, data race):
//   transcoder_pool.go:163-178 — applyConfig() reads node runtime fields
//   (Status/LastCheck/LastError/FailCount/TotalJobs/FailedJobs/TotalLatMs/
//   Concurrency) from the existing node holding ONLY p.mu, never n.mu.
//   probeOne() (transcoder_pool.go:266-289) and recordResult() (61-69) write
//   those same fields holding ONLY n.mu. applyConfig is reachable via
//   Reload()/handleConfigReload while healthLoop() concurrently probes.
//   => data race on shared Node state (distinct code path from FINDING-013
//   which covers List/Get/Pick).
// ---------------------------------------------------------------------------
func TestAudit_ApplyConfigRaceWithProbe(t *testing.T) {
	pool, err := NewTranscoderPool(filepath.Join(t.TempDir(), "transcoders.yaml"))
	require.NoError(t, err)
	require.NoError(t, pool.Add(Node{Name: "n1", Addr: "http://127.0.0.1:1"}))

	cfg := PoolConfig{Transcoders: []Node{{Name: "n1", Addr: "http://127.0.0.1:1", Weight: 1}}}

	var wg sync.WaitGroup
	for i := 0; i < 300; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			// probeOne writes Status/FailCount/LastError under n.mu
			pool.mu.RLock()
			n := pool.nodes["n1"]
			pool.mu.RUnlock()
			if n != nil {
				pool.probeOne(n)
			}
		}()
		go func() {
			defer wg.Done()
			// applyConfig reads those same fields without n.mu (only p.mu)
			pool.applyConfig(cfg)
		}()
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// FINDING C (NEW, channel close / panic):
//   transcoder_pool.go:133 — Stop() does close(p.stopCh) with no sync.Once.
//   Calling Stop() twice panics "close of closed channel". healthLoop selects
//   on stopCh and ctx.Done(); Stop() can be reached from shutdown paths / tests
//   that call it more than once.
// ---------------------------------------------------------------------------
func TestAudit_StopTwicePanics(t *testing.T) {
	pool, err := NewTranscoderPool(filepath.Join(t.TempDir(), "transcoders.yaml"))
	require.NoError(t, err)

	pool.Stop()
	defer func() {
		if r := recover(); r != nil {
			t.Fatal("FINDING-048 修复失败: 第二次 Stop() 不应 panic (close of closed channel)")
		}
	}()
	pool.Stop() // FINDING-048 修复: sync.Once 使二次 Stop 幂等，不 panic
}

// ---------------------------------------------------------------------------
// FINDING D (NEW, correctness / timestamp):
//   events.go:68-70 — now() does time.Now().Format("2006-01-02T15:04:05Z").
//   The literal "Z" labels the value as UTC, but time.Now() returns LOCAL time.
//   When the server runs in a non-UTC timezone, every ProgressEvent.OccurredAt
//   (and PublishEvent timestamp) is wrong by the UTC offset — and can even
//   carry the wrong DATE. Fix: use time.Now().UTC().Format(...).
// ---------------------------------------------------------------------------
func TestAudit_NowLabeledUTCButUsesLocalTime(t *testing.T) {
	saved := time.Local
	time.Local = time.FixedZone("CST", 8*3600) // UTC+8
	defer func() { time.Local = saved }()

	got := now()
	parsed, err := time.Parse("2006-01-02T15:04:05Z", got)
	require.NoError(t, err, "now() should be RFC3339-parseable")

	realUTC := time.Now().UTC().Truncate(time.Second)
	delta := parsed.Sub(realUTC).Abs()
	// FINDING-049 修复：now() 返回真实 UTC（此前本地时间标注 "Z"，UTC+8 差 8h）。
	assert.True(t, delta < time.Minute,
		"FINDING-049 修复: now() 应为 UTC；got %q (delta=%v)", got, delta)
}
