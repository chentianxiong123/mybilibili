package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// HTTPProfileRecorder 通过 HTTP 调用 search-service 的 profile/record/ 端点记录用户行为。
// 当 interaction 发生点赞/收藏/观看时，异步记录用户画像用于推荐算法。
type HTTPProfileRecorder struct {
	baseURL string
	client  *http.Client
}

func NewHTTPProfileRecorder() *HTTPProfileRecorder {
	addr := os.Getenv("SEARCH_SERVICE_ADDR")
	if addr == "" {
		addr = "http://127.0.0.1:8084"
	}
	return &HTTPProfileRecorder{baseURL: addr, client: &http.Client{}}
}

func (r *HTTPProfileRecorder) RecordWatch(ctx context.Context, userID, categoryID int64, tags []string, duration int64) error {
	return r.record(ctx, "watch", userID, categoryID, tags, duration)
}

func (r *HTTPProfileRecorder) RecordLike(ctx context.Context, userID, categoryID int64, tags []string) error {
	return r.record(ctx, "like", userID, categoryID, tags, 0)
}

func (r *HTTPProfileRecorder) RecordCollect(ctx context.Context, userID, categoryID int64, tags []string) error {
	return r.record(ctx, "collect", userID, categoryID, tags, 0)
}

func (r *HTTPProfileRecorder) record(ctx context.Context, action string, userID, categoryID int64, tags []string, duration int64) error {
	body, _ := json.Marshal(map[string]interface{}{
		"categoryId":      categoryID,
		"tags":            tags,
		"durationSeconds": duration,
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/profile/record/%s", r.baseURL, action), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", fmt.Sprintf("%d", userID))
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
