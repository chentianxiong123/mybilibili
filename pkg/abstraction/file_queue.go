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

func (q *fileQueue) Publish(ctx context.Context, topic string, msg Message) error {
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
	ch := make(chan Message, 100)
	topicDir := filepath.Join(q.dir, topic)
	consumeDir := filepath.Join(q.dir, topic+"."+group)

	go func() {
		defer close(ch)
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

func (q *fileQueue) Ack(ctx context.Context, topic string, msg Message) error  { return nil }
func (q *fileQueue) Nack(ctx context.Context, topic string, msg Message) error { return nil }
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