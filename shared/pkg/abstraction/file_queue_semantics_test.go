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

// TestFileQueue_AckIsNoopAndConsumeDirGrows 验证 FINDING-076 修复：
// Ack 现删除消费组目录中的消息文件（此前 no-op → 无界增长 + 孤儿丢失）。
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

	// FINDING-076 修复：Ack 删除消息文件 → 消费组目录清空。
	require.NoError(t, q.Ack(ctx, "topic", Message{ID: "m1"}))
	entries, err := os.ReadDir(filepath.Join(base, "topic.g1"))
	require.NoError(t, err)
	require.Empty(t, entries, "FINDING-076 修复: Ack 应删除消费组目录中的消息文件")
}

// FINDING-076 修复：Nack 应把消息移回待投递目录重新投递。
func TestFileQueue_NackRequeuesToTopicDir(t *testing.T) {
	base := t.TempDir()
	q, err := newFileQueue(MessageQueueConfig{Path: base})
	require.NoError(t, err)
	defer q.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := q.Subscribe(ctx, "topic", "g1")
	require.NoError(t, err)
	require.NoError(t, q.Publish(ctx, "topic", Message{ID: "m2", Payload: []byte("p2")}))

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for delivery")
	}

	require.NoError(t, q.Nack(ctx, "topic", Message{ID: "m2"}))
	// 消息应回到 topicDir（再次可投递），consumeDir 清空。
	consumeEntries, err := os.ReadDir(filepath.Join(base, "topic.g1"))
	require.NoError(t, err)
	require.Empty(t, consumeEntries)
	topicEntries, err := os.ReadDir(filepath.Join(base, "topic"))
	require.NoError(t, err)
	require.NotEmpty(t, topicEntries, "FINDING-076 修复: Nack 应移回待投递目录")
}

// FINDING-076 修复：启动重读 consumeDir——遗留未 Ack 消息被移回 topicDir 重新投递。
func TestFileQueue_StartupRecoversConsumedOrphans(t *testing.T) {
	base := t.TempDir()
	q, err := newFileQueue(MessageQueueConfig{Path: base})
	require.NoError(t, err)
	defer q.Close()

	// 手工模拟崩溃遗留：直接在 consumeDir 放一个消息文件。
	consumeDir := filepath.Join(base, "topic.g1")
	require.NoError(t, os.MkdirAll(consumeDir, 0o755))
	orphan := filepath.Join(consumeDir, "123.456.json")
	require.NoError(t, os.WriteFile(orphan, []byte(`{"id":"m3","topic":"topic","payload":"cA=="}`), 0o644))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := q.Subscribe(ctx, "topic", "g1")
	require.NoError(t, err)

	// 启动恢复 + 重新投递应把 m3 重新送达到通道。
	select {
	case msg := <-ch:
		assert.Equal(t, "m3", msg.ID)
	case <-time.After(2 * time.Second):
		t.Fatal("FINDING-076 修复: 启动恢复应重新投递遗留消息")
	}
}
