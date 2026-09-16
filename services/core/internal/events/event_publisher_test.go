package events

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mybilibili/pkg/abstraction"
)

type recordingQueue struct {
	messages []abstraction.Message
}

func (r *recordingQueue) Publish(ctx context.Context, topic string, msg abstraction.Message) error {
	r.messages = append(r.messages, msg)
	return nil
}
func (r *recordingQueue) Subscribe(ctx context.Context, topic, group string) (<-chan abstraction.Message, error) {
	return nil, nil
}
func (r *recordingQueue) Ack(ctx context.Context, topic string, msg abstraction.Message) error   { return nil }
func (r *recordingQueue) Nack(ctx context.Context, topic string, msg abstraction.Message) error  { return nil }
func (r *recordingQueue) Enqueue(ctx context.Context, queue string, msg abstraction.Message, delay time.Duration) error {
	return nil
}
func (r *recordingQueue) Close() error { return nil }

func TestPublishManuscriptIndex(t *testing.T) {
	rq := &recordingQueue{}
	p := NewEventPublisher(rq)

	err := p.PublishManuscriptIndex(context.Background(), 123, "UPSERT", "APPROVE")
	require.NoError(t, err)
	require.Len(t, rq.messages, 1)
	assert.Equal(t, "manuscript-index-topic", rq.messages[0].Topic)

	var evt map[string]interface{}
	require.NoError(t, json.Unmarshal(rq.messages[0].Payload, &evt))
	assert.Equal(t, float64(123), evt["manuscript_id"])
	assert.Equal(t, "UPSERT", evt["operation"])
	assert.Equal(t, "APPROVE", evt["trigger"])
}

func TestPublishAnalytics(t *testing.T) {
	rq := &recordingQueue{}
	p := NewEventPublisher(rq)

	err := p.PublishAnalytics(context.Background(), 100, 200, "VIEW", "view_count", 5)
	require.NoError(t, err)
	require.Len(t, rq.messages, 1)
	assert.Equal(t, "manuscript-analytics-topic", rq.messages[0].Topic)

	var evt map[string]interface{}
	require.NoError(t, json.Unmarshal(rq.messages[0].Payload, &evt))
	assert.Equal(t, float64(100), evt["manuscript_id"])
	assert.Equal(t, float64(200), evt["user_id"])
	assert.Equal(t, "VIEW", evt["event_type"])
	assert.Equal(t, "view_count", evt["metric_type"])
	assert.Equal(t, float64(5), evt["delta"])
	assert.NotEmpty(t, evt["occurred_at"])
}

func TestPublishVideoProcess(t *testing.T) {
	rq := &recordingQueue{}
	p := NewEventPublisher(rq)

	err := p.PublishVideoProcess(context.Background(), 100, 300, "TRANSCODE", "http://src", 42)
	require.NoError(t, err)
	require.Len(t, rq.messages, 1)
	assert.Equal(t, "video-process-topic", rq.messages[0].Topic)

	var evt map[string]interface{}
	require.NoError(t, json.Unmarshal(rq.messages[0].Payload, &evt))
	assert.Equal(t, float64(100), evt["manuscript_id"])
	assert.Equal(t, float64(300), evt["video_id"])
	assert.Equal(t, "TRANSCODE", evt["process_type"])
	assert.Equal(t, "AUTO_CHAIN", evt["process_mode"])
	assert.Equal(t, "http://src", evt["source_url"])
	assert.Equal(t, float64(42), evt["uploader_id"])
}

func TestEventPublisher_WithMemoryQueue(t *testing.T) {
	q, err := abstraction.NewMessageQueue(abstraction.MessageQueueConfig{Type: "memory"})
	require.NoError(t, err)
	p := NewEventPublisher(q)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := q.Subscribe(ctx, "video-process-topic", "test-group")
	require.NoError(t, err)

	require.NoError(t, p.PublishVideoProcess(ctx, 1, 2, "TRANSCODE", "url", 3))

	select {
	case msg := <-ch:
		assert.Equal(t, "video-process-topic", msg.Topic)
		assert.NotEmpty(t, msg.Payload)
	default:
		t.Fatal("expected message from memory queue")
	}
}
