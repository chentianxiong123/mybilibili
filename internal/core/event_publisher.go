package core

import (
	"context"
	"encoding/json"

	"mybilibili/internal/abstraction"
)

type EventPublisher struct {
	mq abstraction.MessageQueue
}

func NewEventPublisher(mq abstraction.MessageQueue) *EventPublisher {
	return &EventPublisher{mq: mq}
}

func (p *EventPublisher) PublishManuscriptIndex(ctx context.Context, manuscriptID int64, operation, trigger string) error {
	evt := map[string]interface{}{
		"manuscript_id": manuscriptID,
		"operation":     operation,
		"trigger":       trigger,
	}
	data, _ := json.Marshal(evt)
	return p.mq.Publish(ctx, "manuscript-index-topic", abstraction.Message{
		Topic: "manuscript-index-topic", Payload: data,
	})
}

func (p *EventPublisher) PublishAnalytics(ctx context.Context, manuscriptID, userID int64, eventType, metricType string, delta int64) error {
	evt := map[string]interface{}{
		"event_type":    eventType,
		"manuscript_id": manuscriptID,
		"user_id":       userID,
		"metric_type":   metricType,
		"delta":         delta,
		"occurred_at":   now(),
	}
	data, _ := json.Marshal(evt)
	return p.mq.Publish(ctx, "manuscript-analytics-topic", abstraction.Message{
		Topic: "manuscript-analytics-topic", Payload: data,
	})
}

func (p *EventPublisher) PublishVideoProcess(ctx context.Context, manuscriptID, videoID int64, processType string) error {
	evt := map[string]interface{}{
		"manuscript_id": manuscriptID,
		"video_id":      videoID,
		"process_type":  processType,
		"process_mode":  "AUTO_CHAIN",
	}
	data, _ := json.Marshal(evt)
	return p.mq.Publish(ctx, "video-process-topic", abstraction.Message{
		Topic: "video-process-topic", Payload: data,
	})
}

func now() string {
	return "stub-time"
}

var _ = context.Background
