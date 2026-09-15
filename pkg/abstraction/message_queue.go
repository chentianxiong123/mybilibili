package abstraction

import (
	"context"
	"time"
)

type Message struct {
	ID        string
	Topic     string
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
