package abstraction

import (
	"context"
	"sync"
	"testing"
)

// TestOllamaCaller_FallbackDataRace pins FINDING: the package-level mutable
// ollamaFallback flag is written and read concurrently with no synchronization.
//
// Call reads ollamaFallback (to decide whether to probe), writes it (when the probe
// fails), and doCall reads it again. Concurrent Call invocations therefore race on
// this shared global. Run with `go test -race` to observe the race; this test is a
// no-op under normal runs (kept in the suite so it is re-runnable under -race).
func TestOllamaCaller_FallbackDataRace(t *testing.T) {
	// Probe fails fast (closed port) so many goroutines hit the fallback write/read.
	c := newOllamaCallerWithURL(t, "http://127.0.0.1:1")

	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp := map[string]interface{}{}
			_ = c.Call(context.Background(), "ai", "Summary", map[string]interface{}{"video_id": 1.0}, &resp)
		}()
	}
	wg.Wait()

	// Reset the shared flag so later tests in the package are not poisoned.
	ollamaFallback.Store(false)
}
