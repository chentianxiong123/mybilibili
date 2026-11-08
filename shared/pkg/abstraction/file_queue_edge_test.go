package abstraction

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 本文件覆盖 file_queue.go 的错误/边界分支:
// - newFileQueue 目录创建失败
// - Publish 的 MkdirAll / WriteFile 失败
// - Subscribe 的 ReadDir / 条目过滤 / consumeDir 创建失败 /
//   rename 失败 / 坏 JSON / 发送阻塞时 ctx 取消

func TestFileQueue_NewWithUnwritablePath(t *testing.T) {
	// cfg.Path 指向"一个普通文件下的子路径" → os.MkdirAll 报错
	f := filepath.Join(t.TempDir(), "afile")
	require.NoError(t, os.WriteFile(f, []byte("x"), 0o644))
	_, err := newFileQueue(MessageQueueConfig{Path: filepath.Join(f, "sub")})
	require.Error(t, err)
}

func TestFileQueue_Publish_TopicDirError(t *testing.T) {
	q := newTestFileQueue(t)
	// 让 topic 的父路径与已存在的文件冲突 → MkdirAll 失败
	require.NoError(t, os.WriteFile(filepath.Join(q.dir, "clash"), []byte("x"), 0o644))
	err := q.Publish(context.Background(), "clash/sub", Message{Payload: []byte("x")})
	require.Error(t, err)
}

func TestFileQueue_Publish_WriteFileError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: chmod cannot block writes")
	}
	q := newTestFileQueue(t)
	topicDir := filepath.Join(q.dir, "ro")
	require.NoError(t, os.MkdirAll(topicDir, 0o755))
	require.NoError(t, os.Chmod(topicDir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(topicDir, 0o700) })

	err := q.Publish(context.Background(), "ro", Message{Payload: []byte("x")})
	require.Error(t, err)
}

func TestFileQueue_Subscribe_ReadDirError(t *testing.T) {
	q := newTestFileQueue(t)
	// topic 目录实际是一个普通文件 → os.ReadDir 报错 → continue
	require.NoError(t, os.WriteFile(filepath.Join(q.dir, "rd"), []byte("x"), 0o644))

	ctx, cancel := context.WithCancel(context.Background())
	ch, err := q.Subscribe(ctx, "rd", "g1")
	require.NoError(t, err)

	time.Sleep(700 * time.Millisecond) // 至少跑几个 tick 触发 ReadDir 错误
	cancel()
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "channel should be closed after cancel")
	case <-time.After(2 * time.Second):
		t.Fatal("channel not closed")
	}
}

func TestFileQueue_Subscribe_SkipsInvalidEntries(t *testing.T) {
	q := newTestFileQueue(t)
	topic := "mix"
	topicDir := filepath.Join(q.dir, topic)
	consumePath := filepath.Join(q.dir, topic+".g1")
	require.NoError(t, os.MkdirAll(topicDir, 0o755))

	// consumeDir 位置上放一个文件 → MkdirAll(consumeDir) 失败 → continue
	require.NoError(t, os.WriteFile(consumePath, []byte("x"), 0o644))

	// 有效消息（会被 rename / read / unmarshal 处理）
	good := Message{ID: "good", Payload: []byte("ok")}
	gb, err := json.Marshal(good)
	require.NoError(t, err)
	name := fmt.Sprintf("%d.%d.json", time.Now().UnixNano(), rand.Int63())
	require.NoError(t, os.WriteFile(filepath.Join(topicDir, name), gb, 0o644))

	// 需要跳过的条目: 子目录 + 非 .json 文件
	require.NoError(t, os.Mkdir(filepath.Join(topicDir, "subdir"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(topicDir, "notes.txt"), []byte("x"), 0o644))

	// 由于 consumeDir 无法创建, 所有 .json 条目都会被 continue 跳过,
	// 包括这条 "good" 消息(不会被投递, 文件保留在 topicDir)
	ctx, cancel := context.WithCancel(context.Background())
	ch, err := q.Subscribe(ctx, topic, "g1")
	require.NoError(t, err)

	time.Sleep(700 * time.Millisecond)
	cancel()
	select {
	case m, ok := <-ch:
		if ok {
			t.Fatalf("no message should be delivered, got %+v", m)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("channel not closed after cancel")
	}

	// 未被消费的文件应仍留在 topicDir (good.json + notes.txt + subdir)
	entries, err := os.ReadDir(topicDir)
	require.NoError(t, err)
	require.Len(t, entries, 3)
}

func TestFileQueue_Subscribe_ReadFileError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: chmod cannot block reads")
	}
	q := newTestFileQueue(t)
	topic := "readfail"
	require.NoError(t, q.Publish(context.Background(), topic, Message{ID: "x", Payload: []byte("x")}))

	topicDir := filepath.Join(q.dir, topic)
	entries, err := os.ReadDir(topicDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	msgFile := filepath.Join(topicDir, entries[0].Name())
	require.NoError(t, os.Chmod(msgFile, 0o000))
	t.Cleanup(func() { _ = os.Chmod(msgFile, 0o600) })

	// rename 成功(consumeDir 可写), 但读已改名文件时权限不足 → continue
	ctx, cancel := context.WithCancel(context.Background())
	ch, err := q.Subscribe(ctx, topic, "g1")
	require.NoError(t, err)

	time.Sleep(700 * time.Millisecond)
	cancel()
	select {
	case m, ok := <-ch:
		if ok {
			t.Fatalf("unreadable message should be skipped, got %+v", m)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("channel not closed after cancel")
	}
}

func TestFileQueue_Subscribe_BadJSONMessage(t *testing.T) {
	q := newTestFileQueue(t)
	topic := "badjson"
	topicDir := filepath.Join(q.dir, topic)
	require.NoError(t, os.MkdirAll(topicDir, 0o755))
	// 文件内容是非法 JSON → rename/read 成功后 unmarshal 失败 → continue
	require.NoError(t, os.WriteFile(filepath.Join(topicDir, "bad.json"), []byte("{not-json"), 0o644))

	ctx, cancel := context.WithCancel(context.Background())
	ch, err := q.Subscribe(ctx, topic, "g1")
	require.NoError(t, err)

	time.Sleep(700 * time.Millisecond)
	cancel()
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "bad json must be skipped and channel closed after cancel")
	case <-time.After(2 * time.Second):
		t.Fatal("channel not closed")
	}
}

func TestFileQueue_Subscribe_RenameError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: chmod cannot block writes")
	}
	q := newTestFileQueue(t)
	topic := "renamefail"
	topicDir := filepath.Join(q.dir, topic)
	require.NoError(t, os.MkdirAll(topicDir, 0o755))

	// 先写入消息, 再把 topicDir 改成只读 → ReadDir 可读(r-x)但 rename 移出失败
	msg := Message{ID: "r1", Payload: []byte("x")}
	require.NoError(t, q.Publish(context.Background(), topic, msg))
	require.NoError(t, os.Chmod(topicDir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(topicDir, 0o700) })

	ctx, cancel := context.WithCancel(context.Background())
	ch, err := q.Subscribe(ctx, topic, "g1")
	require.NoError(t, err)

	time.Sleep(400 * time.Millisecond) // 先让 goroutine 跑: 不要取消太快
	cancel()
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "rename failure must be skipped and channel closed after cancel")
	case <-time.After(2 * time.Second):
		t.Fatal("channel not closed")
	}
}

func TestFileQueue_Subscribe_SendBlockedCancels(t *testing.T) {
	q := newTestFileQueue(t)
	topic := "full"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := q.Subscribe(ctx, topic, "g1")
	require.NoError(t, err)

	// 发布超过缓冲容量(100)的消息; 前100条填满缓冲, 之后 send 阻塞
	for i := 0; i < 120; i++ {
		require.NoError(t, q.Publish(ctx, topic, Message{ID: fmt.Sprintf("m%d", i), Payload: []byte("x")}))
	}
	// 等 goroutine 消费文件井填满缓冲
	time.Sleep(900 * time.Millisecond)
	cancel()
	// 先排空缓冲中的消息, 再确认 channel 关闭
	deadline := time.After(3 * time.Second)
	closed := false
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				closed = true
			}
		case <-deadline:
			t.Fatal("channel not closed")
		}
		if closed {
			break
		}
	}
}