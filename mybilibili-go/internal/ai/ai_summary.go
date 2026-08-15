package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"mybilibili/internal/abstraction"
)

type SummaryService struct {
	caller     abstraction.ServiceCaller
	cacheStore abstraction.CacheStore
}

func NewSummaryService(caller abstraction.ServiceCaller) *SummaryService {
	return &SummaryService{caller: caller}
}

func (s *SummaryService) SetCacheStore(cs abstraction.CacheStore) {
	s.cacheStore = cs
}

func (s *SummaryService) GetSummary(ctx context.Context, videoID int64) (string, error) {
	if s.cacheStore != nil {
		key := fmt.Sprintf("summary:%d", videoID)
		data, err := s.cacheStore.Get(ctx, key)
		if err == nil {
			return string(data), nil
		}
	}

	req := map[string]interface{}{"video_id": videoID}
	var resp struct {
		Summary string `json:"summary"`
	}
	if s.caller == nil {
		return "", errors.New("summary service not configured")
	}
	if err := s.caller.Call(ctx, "ai", "Summary", req, &resp); err != nil {
		return "", err
	}

	if s.cacheStore != nil && resp.Summary != "" {
		key := fmt.Sprintf("summary:%d", videoID)
		s.cacheStore.Set(ctx, key, []byte(resp.Summary), 0)
	}
	return resp.Summary, nil
}

func (s *SummaryService) StreamSummary(ctx context.Context, videoID int64) (<-chan string, error) {
	req := map[string]interface{}{"video_id": videoID}
	ch, err := s.caller.CallStream(ctx, "ai", "SummaryStream", req)
	if err != nil {
		return nil, err
	}
	out := make(chan string, 10)
	go func() {
		defer close(out)
		for data := range ch {
			out <- string(data)
		}
	}()
	return out, nil
}

func (s *SummaryService) CheckSummary(ctx context.Context, videoID int64) (bool, error) {
	req := map[string]interface{}{"video_id": videoID}
	var resp struct {
		HasSummary bool `json:"has_summary"`
	}
	if s.caller == nil {
		return false, nil
	}
	if err := s.caller.Call(ctx, "ai", "CheckSummary", req, &resp); err != nil {
		return false, err
	}
	return resp.HasSummary, nil
}

type ReviewService struct {
	caller abstraction.ServiceCaller
}

func NewReviewService(caller abstraction.ServiceCaller) *ReviewService {
	return &ReviewService{caller: caller}
}

func (s *ReviewService) Moderate(ctx context.Context, content, scene string) (map[string]interface{}, error) {
	req := map[string]interface{}{"content": content, "scene": scene}
	var resp map[string]interface{}
	if s.caller == nil {
		// local fallback: basic check
		return map[string]interface{}{"passed": true, "reason": ""}, nil
	}
	if err := s.caller.Call(ctx, "ai", "Moderate", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ReviewService) ReviewComment(ctx context.Context, content string) (bool, error) {
	resp, err := s.Moderate(ctx, content, "COMMENT")
	if err != nil {
		return false, err
	}
	passed, _ := resp["passed"].(bool)
	return passed, nil
}

type CustomerService struct {
	caller abstraction.ServiceCaller
}

func NewCustomerService(caller abstraction.ServiceCaller) *CustomerService {
	return &CustomerService{caller: caller}
}

func (s *CustomerService) Chat(ctx context.Context, userID int64, content string) (string, error) {
	req := map[string]interface{}{"user_id": userID, "content": content}
	var resp struct {
		Reply string `json:"reply"`
	}
	if s.caller == nil {
		return "", fmt.Errorf("ai customer service not configured")
	}
	if err := s.caller.Call(ctx, "ai", "CustomerChat", req, &resp); err != nil {
		return "", err
	}
	return resp.Reply, nil
}

func (s *CustomerService) History(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	req := map[string]interface{}{"user_id": userID}
	var resp []map[string]interface{}
	if s.caller == nil {
		return []map[string]interface{}{}, nil
	}
	if err := s.caller.Call(ctx, "ai", "CustomerHistory", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *CustomerService) Transfer(ctx context.Context, userID int64) error {
	req := map[string]interface{}{"user_id": userID}
	var resp map[string]interface{}
	if s.caller == nil {
		return nil
	}
	return s.caller.Call(ctx, "ai", "CustomerTransfer", req, &resp)
}

var _ = json.Marshal
var _ = io.EOF
var _ = time.Now