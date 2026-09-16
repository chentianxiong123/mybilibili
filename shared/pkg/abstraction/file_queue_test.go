package abstraction

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestFileQueue(t *testing.T) *fileQueue {
	t.Helper()
	q, err := newFileQueue(MessageQueueConfig{Path: t.TempDir()})
	require.NoError(t, err)
	t.Cleanup(func() { _ = q.Close() })
	return q
}

func TestFileQueue_NewWithDefaultDir(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("MYBILIBILI_MQ_DIR", dir)
	defer os.Unsetenv("MYBILIBILI_MQ_DIR")

	q, err := newFileQueue(MessageQueueConfig{})
	require.NoError(t, err)
	assert.NotEmpty(t, q.dir)
	_ = q.Close()
}

func TestFileQueue_PublishSubscribe(t *testing.T) {
	q := newTestFileQueue(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := q.Subscribe(ctx, "topic", "g1")
	require.NoError(t, err)

	require.NoError(t, q.Publish(ctx, "topic", Message{ID: "m1", Payload: []byte("hello")}))

	select {
	case m := <-ch:
		assert.Equal(t, "hello", string(m.Payload))
		assert.Equal(t, "m1", m.ID)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for file queue message")
	}
}

func TestFileQueue_EnqueuePublishes(t *testing.T) {
	q := newTestFileQueue(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := q.Subscribe(ctx, "dq", "g1")
	require.NoError(t, err)
	require.NoError(t, q.Enqueue(ctx, "dq", Message{ID: "d1", Payload: []byte("x")}, 10*time.Millisecond))

	select {
	case m := <-ch:
		assert.Equal(t, "x", string(m.Payload))
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for delayed message")
	}
}

func TestFileQueue_MessagesPersistAsFiles(t *testing.T) {
	q := newTestFileQueue(t)
	ctx := context.Background()

	require.NoError(t, q.Publish(ctx, "pt", Message{ID: "1", Payload: []byte("a")}))
	require.NoError(t, q.Publish(ctx, "pt", Message{ID: "2", Payload: []byte("b")}))

	entries, err := os.ReadDir(filepath.Join(q.dir, "pt"))
	require.NoError(t, err)
	assert.Len(t, entries, 2)
}

func TestFileQueue_CloseStopsSubscription(t *testing.T) {
	q := newTestFileQueue(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := q.Subscribe(ctx, "t2", "g1")
	require.NoError(t, err)

	require.NoError(t, q.Close())
	// Close 幂等
	require.NoError(t, q.Close())
}
