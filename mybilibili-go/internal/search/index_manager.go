package search

import (
	"context"
	"encoding/json"
	"log"

	"mybilibili/internal/abstraction"
)

type IndexManager struct {
	engine abstraction.SearchEngine
	mq     abstraction.MessageQueue
}

func NewIndexManager(engine abstraction.SearchEngine, mq abstraction.MessageQueue) *IndexManager {
	return &IndexManager{engine: engine, mq: mq}
}

func (m *IndexManager) Start(ctx context.Context) error {
	ch, err := m.mq.Subscribe(ctx, "manuscript-index-topic", "search-indexer")
	if err != nil {
		return err
	}
	log.Println("search index manager started")
	for msg := range ch {
		var evt struct {
			ManuscriptID int64  `json:"manuscript_id"`
			Operation    string `json:"operation"`
			Trigger      string `json:"trigger"`
		}
		if err := json.Unmarshal(msg.Payload, &evt); err != nil {
			continue
		}
		switch evt.Operation {
		case "UPSERT":
			m.engine.Index(ctx, "manuscripts", formatID(evt.ManuscriptID), map[string]interface{}{
				"id": evt.ManuscriptID, "operation": "upsert",
			})
		case "DELETE":
			m.engine.Delete(ctx, "manuscripts", formatID(evt.ManuscriptID))
		}
	}
	return nil
}

func formatID(id int64) string {
	return "ms_" + itoa(id)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

var _ = itoa
var _ = json.Marshal
