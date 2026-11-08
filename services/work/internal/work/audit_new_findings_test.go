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

	// (2) BUG: chain still advances to the next step after failure
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
	assert.True(t, chained,
		"BUG: AUTO_CHAIN advanced to EXTRACT_AUDIO even though TRANSCODE failed; "+
			"failed step should stop the chain, not publish a broken video")
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
			// BUG confirmed: double close panics
			return
		}
		t.Fatal("BUG: second Stop() did not panic (close of closed channel)")
	}()
	pool.Stop()
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

	got := now() // e.g. 2026-11-08T06:26:40Z
	parsed, err := time.Parse("2006-01-02T15:04:05Z", got)
	require.NoError(t, err, "now() should be RFC3339-parseable")

	realUTC := time.Now().UTC().Truncate(time.Second)
	delta := parsed.Sub(realUTC).Abs()
	// Demonstrates the bug: now() is off by the server's UTC offset (+8h here),
	// even though the "Z" suffix claims UTC. If now() were honest UTC, delta≈0.
	assert.True(t, delta > time.Hour,
		"expected now() to be mislabeled (local time claimed as UTC); got %q (delta=%v)",
		got, delta)
}
