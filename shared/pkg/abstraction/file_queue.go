package abstraction

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// fileQueue 基于本地文件系统的跨进程消息队列，零外部依赖（单机部署适用）。
// 目录结构：
//
//	{path}/{topic}/         待投递消息（*.json）
//	{path}/{topic}.{group}/ 该消费组已消费的消息（rename 移入）
//
// Publish 原子写文件；Subscribe 轮询目录，每个消费组通过 rename 原子抢占消息，
// 天然支持多消费者互不重复消费。
type fileQueue struct {
	dir  string
	stop chan struct{}
}

func newFileQueue(cfg MessageQueueConfig) (*fileQueue, error) {
	dir := cfg.Path
	if dir == "" {
		dir = "/tmp/mybilibili-mq"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("file queue dir: %w", err)
	}
	return &fileQueue{dir: dir, stop: make(chan struct{})}, nil
}

// safeSegment 校验 topic/group 分段名，拒绝路径遍历（FINDING-075：
// 原始值此前直接 filepath.Join(q.dir, topic)，`topic="../escaped"` 可写出队列根之外）。
func safeSegment(seg string) (string, error) {
	seg = strings.TrimSpace(seg)
	if seg == "" {
		return "", fmt.Errorf("empty queue segment")
	}
	if filepath.IsAbs(seg) || seg == "." || seg == ".." || strings.ContainsAny(seg, `/\`) {
		return "", fmt.Errorf("invalid queue segment: %q", seg)
	}
	return seg, nil
}

func (q *fileQueue) Publish(ctx context.Context, topic string, msg Message) error {
	topic, err := safeSegment(topic)
	if err != nil {
		return err
	}
	if msg.Topic == "" {
		msg.Topic = topic
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	topicDir := filepath.Join(q.dir, topic)
	if err := os.MkdirAll(topicDir, 0o755); err != nil {
		return err
	}
	name := fmt.Sprintf("%d.%d.json", time.Now().UnixNano(), rand.Int63())
	tmp := filepath.Join(topicDir, name+".tmp")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(topicDir, name))
}

func (q *fileQueue) Subscribe(ctx context.Context, topic, group string) (<-chan Message, error) {
	topic, err := safeSegment(topic)
	if err != nil {
		return nil, err
	}
	group, err = safeSegment(group)
	if err != nil {
		return nil, err
	}
	ch := make(chan Message, 100)
	topicDir := filepath.Join(q.dir, topic)
	consumeDir := filepath.Join(q.dir, topic+"."+group)

	go func() {
		defer close(ch)
		// FINDING-076 修复：启动重读 consumeDir——上次崩溃留下的已 rename 未 Ack
		// 消息移回 topicDir 重新投递（否则永久孤儿丢失）。
		recoverConsumed(ctx, topicDir, consumeDir)
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-q.stop:
				return
			case <-ticker.C:
			}
			entries, err := os.ReadDir(topicDir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
					continue
				}
				src := filepath.Join(topicDir, e.Name())
				if err := os.MkdirAll(consumeDir, 0o755); err != nil {
					continue
				}
				dst := filepath.Join(consumeDir, e.Name())
				if err := os.Rename(src, dst); err != nil {
					continue
				}
				data, err := os.ReadFile(dst)
				if err != nil {
					continue
				}
				var msg Message
				if err := json.Unmarshal(data, &msg); err != nil {
					continue
				}
				msg.Group = group // 供 Ack/Nack 定位消费组目录
				select {
				case ch <- msg:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

// recoverConsumed 将消费组目录中未被 Ack 的遗留文件移回待投递目录（FINDING-076 崩溃恢复）。
func recoverConsumed(ctx context.Context, topicDir, consumeDir string) {
	entries, err := os.ReadDir(consumeDir)
	if err != nil {
		return
	}
	// topicDir 可能尚未创建（首次消费），先建目录。
	_ = os.MkdirAll(topicDir, 0o755)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		src := filepath.Join(consumeDir, e.Name())
		dst := filepath.Join(topicDir, e.Name())
		_ = os.Rename(src, dst)
	}
}

// consumeDirFor 定位消息所属的消费组目录：优先 msg.Group，否则扫描所有 {topic}.* 目录。
func (q *fileQueue) consumeDirFor(topic string, msg Message) (string, error) {
	if msg.Group != "" {
		seg, err := safeSegment(msg.Group)
		if err != nil {
			return "", err
		}
		return filepath.Join(q.dir, topic+"."+seg), nil
	}
	// 未带 Group：扫描匹配 topic.<group> 的目录，找到含 msg.ID 的文件的那个。
	matches, err := filepath.Glob(filepath.Join(q.dir, topic+".*"))
	if err != nil {
		return "", err
	}
	for _, dir := range matches {
		if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			var m Message
			if json.Unmarshal(data, &m) == nil && m.ID == msg.ID {
				return dir, nil
			}
		}
	}
	return "", nil // 未找到（已被 Ack/不存在）→ 调用方按 no-op 处理
}

func (q *fileQueue) Ack(ctx context.Context, topic string, msg Message) error {
	topic, err := safeSegment(topic)
	if err != nil {
		return err
	}
	dir, err := q.consumeDirFor(topic, msg)
	if err != nil {
		return err
	}
	if dir == "" {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var m Message
		if json.Unmarshal(data, &m) == nil && m.ID == msg.ID {
			// FINDING-076 修复：Ack 删除已处理消息文件（此前 no-op → 无界增长）。
			return os.Remove(filepath.Join(dir, e.Name()))
		}
	}
	return nil
}

func (q *fileQueue) Nack(ctx context.Context, topic string, msg Message) error {
	topic, err := safeSegment(topic)
	if err != nil {
		return err
	}
	dir, err := q.consumeDirFor(topic, msg)
	if err != nil {
		return err
	}
	if dir == "" {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	topicDir := filepath.Join(q.dir, topic)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var m Message
		if json.Unmarshal(data, &m) == nil && m.ID == msg.ID {
			// FINDING-076 修复：Nack 移回待投递目录重试（此前 no-op → 失败静默丢弃）。
			return os.Rename(filepath.Join(dir, e.Name()), filepath.Join(topicDir, e.Name()))
		}
	}
	return nil
}
func (q *fileQueue) Enqueue(ctx context.Context, queue string, msg Message, delay time.Duration) error {
	return q.Publish(ctx, queue, msg)
}

func (q *fileQueue) Close() error {
	select {
	case <-q.stop:
	default:
		close(q.stop)
	}
	return nil
}