package abstraction

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileQueue_AckNack_Noop(t *testing.T) {
	q, err := newFileQueue(MessageQueueConfig{Path: t.TempDir()})
	require.NoError(t, err)
	assert.NoError(t, q.Ack(context.Background(), "t", Message{}))
	assert.NoError(t, q.Nack(context.Background(), "t", Message{}))
	assert.NoError(t, q.Close())
}

func TestFileQueue_Close_Idempotent(t *testing.T) {
	q, err := newFileQueue(MessageQueueConfig{Path: t.TempDir()})
	require.NoError(t, err)
	assert.NoError(t, q.Close())
	// 第二次关闭会因 close(stop) 已关闭而 panic; 在子 goroutine 中保护
	require.NotPanics(t, func() {
		_ = q.Close()
	})
}

func TestFileQueue_Enqueue_CallsPublish(t *testing.T) {
	q, err := newFileQueue(MessageQueueConfig{Path: t.TempDir()})
	require.NoError(t, err)
	defer q.Close()

	ctx := context.Background()
	msg := Message{Topic: "t", Key: "k", Payload: []byte("hi")}
	// Enqueue -> Publish, 写入文件
	assert.NoError(t, q.Enqueue(ctx, "t", msg, 0))

	// Subscribe 拉取
	ch, err := q.Subscribe(ctx, "t", "g1")
	require.NoError(t, err)
	select {
	case got := <-ch:
		assert.Equal(t, "k", got.Key)
		assert.Equal(t, []byte("hi"), got.Payload)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}