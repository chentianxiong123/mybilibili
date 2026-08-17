package abstraction

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

// natsQueue 基于 NATS JetStream 的持久化消息队列。
// 每个 topic 映射一个 JetStream stream（file storage，保留 24h）；
// Subscribe 的 group 映射为 durable consumer，各消费组独立游标互不干扰，
// 同一组内多个订阅者按 NATS 负载均衡分摊消费。
type natsQueue struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func newNATSQueue(cfg MessageQueueConfig) (*natsQueue, error) {
	url := cfg.NATSURL
	if url == "" {
		url = "nats://127.0.0.1:4222"
	}
	nc, err := nats.Connect(url, nats.Timeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	js, err := nc.JetStream(nats.MaxWait(5 * time.Second))
	if err != nil {
		nc.Close()
		return nil, err
	}
	return &natsQueue{nc: nc, js: js}, nil
}

// ensureStream 懒创建 stream：不存在则建，已存在（并发竞态）则忽略。
func (q *natsQueue) ensureStream(topic string) error {
	name := natsStreamName(topic)
	if _, err := q.js.StreamInfo(name); err == nil {
		return nil
	}
	_, err := q.js.AddStream(&nats.StreamConfig{
		Name:              name,
		Subjects:          []string{topic},
		Storage:           nats.FileStorage,
		Retention:         nats.LimitsPolicy,
		MaxAge:            24 * time.Hour,
		MaxMsgsPerSubject: 100000,
	})
	if err != nil && strings.Contains(err.Error(), "already exists") {
		return nil
	}
	return err
}

func natsStreamName(topic string) string {
	return strings.NewReplacer("/", "_", ".", "_", " ", "_").Replace(topic)
}

func (q *natsQueue) Publish(ctx context.Context, topic string, msg Message) error {
	if msg.Topic == "" {
		msg.Topic = topic
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if err := q.ensureStream(topic); err != nil {
		return err
	}
	_, err = q.js.Publish(topic, data)
	return err
}

func (q *natsQueue) Subscribe(ctx context.Context, topic, group string) (<-chan Message, error) {
	if err := q.ensureStream(topic); err != nil {
		return nil, err
	}
	sub, err := q.js.SubscribeSync(topic, nats.Durable(group), nats.ManualAck(), nats.BindStream(natsStreamName(topic)))
	if err != nil {
		return nil, err
	}
	ch := make(chan Message, 100)
	go func() {
		defer close(ch)
		defer sub.Unsubscribe()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			m, err := sub.NextMsg(1 * time.Second)
			if err != nil {
				if err == nats.ErrTimeout {
					continue
				}
				return
			}
			_ = m.Ack()
			var msg Message
			if err := json.Unmarshal(m.Data, &msg); err != nil {
				continue
			}
			select {
			case ch <- msg:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func (q *natsQueue) Ack(ctx context.Context, topic string, msg Message) error  { return nil }
func (q *natsQueue) Nack(ctx context.Context, topic string, msg Message) error { return nil }
func (q *natsQueue) Enqueue(ctx context.Context, queue string, msg Message, delay time.Duration) error {
	return q.Publish(ctx, queue, msg)
}

func (q *natsQueue) Close() error {
	q.nc.Close()
	return nil
}