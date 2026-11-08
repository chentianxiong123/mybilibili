package danmaku

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupAuditHandler builds the handler and service sharing ONE broadcaster,
// exactly like production main.go (danmakuBroadcaster passed to both).
func setupAuditHandler(t *testing.T) (*HTTPHandler, *DanmakuBroadcaster, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	repo := NewDanmakuRepository(db)
	bc := NewDanmakuBroadcaster()
	svc := NewDanmakuService(repo, bc)
	h := NewHTTPHandler(svc, bc)
	return h, bc, mock
}

// AUDIT-1 (duplication): a single POST /api/v1/danmaku/send is broadcast TWICE.
// Root cause: DanmakuService.Send() broadcasts internally (danmaku_service.go:39)
// AND the HTTP handler broadcasts again (http_handler.go:66).
func TestAuditDanmakuSend_DoubleBroadcast(t *testing.T) {
	h, bc, mock := setupAuditHandler(t)
	ch := bc.Subscribe(99)

	mock.ExpectQuery(`INSERT INTO danmaku`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/danmaku/send",
		strings.NewReader(`{"video_id":99,"manuscript_id":0,"content":"hi","time":1.0,"color":"#fff","mode":1}`))
	req.Header.Set("X-User-Id", "1001")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	// FINDING-039 钉住：bug 产生 2 次广播（svc.Send 广播 + handler 再广播）。
	var got []*DanmakuEvent
	timeout := time.After(200 * time.Millisecond)
drain:
	for {
		select {
		case e := <-ch:
			got = append(got, e)
		case <-timeout:
			break drain
		}
	}
	assert.Equalf(t, 2, len(got),
		"FINDING-039: 每 send 应广播 1 次但实际 %d（svc.Send + handler 双广播）", len(got))
}

// AUDIT-2 (fan-out): two SSE clients on the same video share ONE channel; a
// single broadcast is delivered to only ONE of them, so concurrent viewers miss
// danmaku. Root cause: Subscribe() returns the shared per-video channel
// (danmaku_repository.go:229-234) instead of a per-subscriber channel.
func TestAuditDanmakuBroadcaster_Fanout(t *testing.T) {
	bc := NewDanmakuBroadcaster()

	ch1 := bc.Subscribe(42)
	ch2 := bc.Subscribe(42) // same shared channel returned

	bc.Broadcast(42, &DanmakuEvent{ID: 1, VideoID: 42, Content: "hello"})

	// FINDING-041 钉住：两订阅者共享单通道——仅 1 人收到（got1!=nil 且 got2==nil）。
	got1 := readWithTimeout(ch1)
	got2 := readWithTimeout(ch2)
	assert.NotNil(t, got1, "subscriber 1 should receive the event")
	assert.Nil(t, got2, "FINDING-041: fan-out broken — 共享通道仅投递到 1 个接收者，subscriber 2 收不到")
}

// AUDIT-3 (data race): concurrent Subscribe/Broadcast/Unsubscribe on the
// DanmakuBroadcaster touch an unsynchronized map (danmaku_repository.go:206-247).
// Detected deterministically by `go test -race`. Under plain `go test` this may
// crash with "fatal error: concurrent map writes" (which is itself the bug).
func TestAuditDanmakuBroadcaster_ConcurrentAccess(t *testing.T) {
	if !auditRaceEnabled() {
		t.Skip("run with -race to detect the unsynchronized map race")
	}
	bc := NewDanmakuBroadcaster()

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				ch := bc.Subscribe(int64(n))
				bc.Broadcast(int64(n), &DanmakuEvent{ID: int64(j), VideoID: int64(n)})
				bc.Unsubscribe(int64(n), ch)
				_ = ch
			}
		}(i)
	}
	wg.Wait()
}

func readWithTimeout(ch <-chan *DanmakuEvent) *DanmakuEvent {
	select {
	case e := <-ch:
		return e
	case <-time.After(200 * time.Millisecond):
		return nil
	}
}

// AUDIT-4 (unauthenticated SSE): /sse/danmaku accepts an anonymous connection
// with only a video_id — no token, no X-User-Id. A user can read the live
// danmaku stream of ANY video without authentication. (Danmaku is public content,
// but the endpoint is completely unauthenticated and unrate-limited.)
func TestAuditDanmakuSSE_AnonymousConnect(t *testing.T) {
	h, _, _ := setupAuditHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	// No auth header, no token — just video_id.
	req := httptest.NewRequest(http.MethodGet, "/sse/danmaku?video_id=777", nil)
	rr := httptest.NewRecorder()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req = req.WithContext(ctx)

	done := make(chan struct{})
	go func() {
		defer close(done)
		mux.ServeHTTP(rr, req)
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done

	// FINDING-042 钉住：匿名连接被接受（200）——无 token/X-User-Id 仍可流式弹幕。
	assert.Equalf(t, http.StatusOK, rr.Code,
		"FINDING-042: SSE 匿名连接应 401 但实际 %d（仅 video_id 即可连接）", rr.Code)
	_ = json.Valid(rr.Body.Bytes())
}
