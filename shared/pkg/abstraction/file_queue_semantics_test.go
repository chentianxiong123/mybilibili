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

// TestFileQueue_PublishTopicTraversal pins FINDING: file_queue does not sanitize
// topic/group names, so a crafted topic escapes the queue root directory.
//
// Publish joins q.dir with the raw topic (filepath.Join(q.dir, topic)) and then
// MkdirAll + writes a message there. A topic of "../escaped" writes the message one
// level above the queue root, allowing a caller-controlled topic to plant files
// outside the queue directory (and Subscribe's consumeDir has the same flaw).
func TestFileQueue_PublishTopicTraversal(t *testing.T) {
	base := filepath.Join(t.TempDir(), "queuedir")
	q, err := newFileQueue(MessageQueueConfig{Path: base})
	require.NoError(t, err)
	defer q.Close()

	// FINDING-075 修复：topic "../escaped" 应被拒绝（返回 error），不写出队列根之外。
	assert.Error(t, q.Publish(context.Background(), "../escaped", Message{ID: "x"}), "FINDING-075: 路径遍历 topic 应被拒绝")
	outside := filepath.Join(base, "..", "escaped")
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatalf("FINDING-075 修复失败: 消息仍被写出队列根之外 (%v)", outside)
	}
}

// TestFileQueue_AckIsNoopAndConsumeDirGrows pins FINDING: file_queue Ack/Nack are
// no-ops and the consume directory is never pruned, so delivered messages accumulate
// forever and are orphaned (never redelivered) after a restart.
//
// After Subscribe delivers a message, the renamed file remains in {topic}.{group}/
// permanently. There is no redelivery on startup, so a crash after the rename (but
// before the channel send) loses the message, and a normal delivery is never acked
// away — unbounded disk growth.
func TestFileQueue_AckIsNoopAndConsumeDirGrows(t *testing.T) {
	base := t.TempDir()
	q, err := newFileQueue(MessageQueueConfig{Path: base})
	require.NoError(t, err)
	defer q.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := q.Subscribe(ctx, "topic", "g1")
	require.NoError(t, err)
	require.NoError(t, q.Publish(ctx, "topic", Message{ID: "m1", Payload: []byte("p")}))

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for delivery")
	}

	// Ack is a no-op; the delivered file must remain in the consume dir.
	require.NoError(t, q.Ack(ctx, "topic", Message{ID: "m1"}))
	entries, err := os.ReadDir(filepath.Join(base, "topic.g1"))
	require.NoError(t, err)
	require.NotEmpty(t, entries, "Ack is a no-op: delivered message should be removed from consume dir")
}
