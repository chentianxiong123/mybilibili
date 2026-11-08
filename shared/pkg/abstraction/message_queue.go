package abstraction

import (
	"context"
	"time"
)

type Message struct {
	ID        string
	Topic     string
	Group     string // 消费组（Subscribe 投递时填充，供 Ack/Nack 定位文件）
	Key       string
	Payload   []byte
	Timestamp time.Time
	Retry     int
}

type MessageQueue interface {
	Publish(ctx context.Context, topic string, msg Message) error
	Subscribe(ctx context.Context, topic, group string) (<-chan Message, error)
	Ack(ctx context.Context, topic string, msg Message) error
	Nack(ctx context.Context, topic string, msg Message) error
	Enqueue(ctx context.Context, queue string, msg Message, delay time.Duration) error
	Close() error
}
