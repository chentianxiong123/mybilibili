package message

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func recvTimeout(ch <-chan *NotificationEvent) *NotificationEvent {
	select {
	case e := <-ch:
		return e
	case <-time.After(200 * time.Millisecond):
		return nil
	}
}

// AUDIT-5 (data race): NotificationBroadcaster keeps an unsynchronized
// map[int64]chan (message_repository.go:248-285). Concurrent SSE Subscribe and
// request-path Send/Unsubscribe race on the map. Under `go test -race` this is a
// data race; under plain `go test` it can crash with "concurrent map writes".
// Unsubscribe (line 272-276) also does a non-atomic read-then-delete that can
// drop a just-recreated subscriber channel.
func TestAuditNotificationBroadcaster_ConcurrentAccess(t *testing.T) {
	if !auditRaceEnabled() {
		t.Skip("run with -race to detect the unsynchronized map race")
	}
	notif := NewNotificationBroadcaster()

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int64) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				ch := notif.Subscribe(n)
				notif.Send(n, &NotificationEvent{Type: "test", Content: "x"})
				notif.Unsubscribe(n, ch)
				_ = ch
			}
		}(int64(i))
	}
	wg.Wait()
	assert.NotNil(t, notif)
}

// AUDIT-6 (fan-out): two SSE connections for the same user share one channel, so
// a single notification is delivered to only one of the two subscribers.
func TestAuditNotificationBroadcaster_Fanout(t *testing.T) {
	notif := NewNotificationBroadcaster()
	ch1 := notif.Subscribe(7)
	ch2 := notif.Subscribe(7) // same channel

	notif.Send(7, &NotificationEvent{Type: "message", Content: "hello"})

	// FINDING-044 钉住：同用户两订阅者共享单通道——仅 1 人收到。
	got1 := recvTimeout(ch1)
	got2 := recvTimeout(ch2)
	assert.NotNil(t, got1, "subscriber 1 should receive")
	assert.Nil(t, got2, "FINDING-044: 共享通道仅投递到 1 个接收者，subscriber 2 收不到")
}
