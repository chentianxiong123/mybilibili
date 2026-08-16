package admin

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"mybilibili/pkg/abstraction"
	"mybilibili/core-service/internal/search"
)

// TestPublishManuscriptIndexReachesConsumer 验证发布稿件索引事件后，search.IndexManager 能消费并写入内存索引。
func TestPublishManuscriptIndexReachesConsumer(t *testing.T) {
	ctx := context.Background()
	engine, _ := abstraction.NewSearchEngine(abstraction.SearchEngineConfig{Type: "memory"})
	mq, _ := abstraction.NewMessageQueue(abstraction.MessageQueueConfig{Type: "memory"})

	indexMgr := search.NewIndexManager(engine, mq)
	done := make(chan error, 1)
	go func() { done <- indexMgr.Start(ctx) }()

	time.Sleep(50 * time.Millisecond)

	pub := &coreEventPublisherAdapter{mq: mq}
	if err := pub.PublishManuscriptIndex(ctx, 777, "UPSERT", "PUBLISH"); err != nil {
		t.Fatalf("publish: %v", err)
	}

	// Start 消费是异步的，轮询索引
	var found bool
	for i := 0; i < 50; i++ {
		res, _ := engine.Search(ctx, "manuscripts", "", abstraction.SearchOptions{})
		if res != nil && res.Total > 0 {
			found = true
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !found {
		t.Fatal("index manager did not consume manuscript-index-topic")
	}
	select {
	case err := <-done:
		t.Fatalf("index manager exited: %v", err)
	default:
	}
}

// coreEventPublisherAdapter 在 admin 测试内直接用 abstraction.MessageQueue 发布，
// 复用 core.EventPublisher 的 payload 结构（避免 admin→core 依赖）。
type coreEventPublisherAdapter struct {
	mq abstraction.MessageQueue
}

func (a *coreEventPublisherAdapter) PublishManuscriptIndex(ctx context.Context, manuscriptID int64, operation, trigger string) error {
	evt := map[string]interface{}{"manuscript_id": manuscriptID, "operation": operation, "trigger": trigger}
	data, _ := json.Marshal(evt)
	return a.mq.Publish(ctx, "manuscript-index-topic", abstraction.Message{Topic: "manuscript-index-topic", Payload: data})
}
