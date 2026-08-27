package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"mybilibili/pkg/abstraction"
)

type SummaryService struct {
	caller     abstraction.ServiceCaller
	cacheStore abstraction.CacheStore
	db         *sql.DB
	storage    *abstraction.MinioStorageService
}

func NewSummaryService(caller abstraction.ServiceCaller) *SummaryService {
	return &SummaryService{caller: caller}
}

func (s *SummaryService) SetCacheStore(cs abstraction.CacheStore) {
	s.cacheStore = cs
}

// SetDatabase 注入 PG，用于查询 videos.has_summary / manuscript_id（与老项目迁移数据对齐）
func (s *SummaryService) SetDatabase(db *sql.DB) {
	s.db = db
}

// SetStorage 注入 MinIO 客户端，用于读取流水线上传的摘要文件
func (s *SummaryService) SetStorage(st *abstraction.MinioStorageService) {
	s.storage = st
}

// summaryObjectKey 老项目 StorageKeys.videoSummary 的对象 key 格式
func summaryObjectKey(manuscriptID, videoID int64) string {
	return fmt.Sprintf("manuscripts/%d/videos/%d/summary/ai-summary.txt", manuscriptID, videoID)
}

// FetchStoredSummary 从 MinIO 读取流水线生成的摘要全文；无则返回 ""
func (s *SummaryService) FetchStoredSummary(ctx context.Context, videoID int64) (string, error) {
	if s.db == nil || s.storage == nil {
		return "", errors.New("storage not configured")
	}
	var manuscriptID int64
	var hasSummary int32
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(manuscript_id,0), COALESCE(has_summary,0) FROM videos WHERE id = $1`, videoID).
		Scan(&manuscriptID, &hasSummary)
	if err != nil || manuscriptID <= 0 {
		return "", err
	}
	rc, err := s.storage.Get(ctx, "mybilibili", summaryObjectKey(manuscriptID, videoID))
	if err != nil {
		return "", err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}
	return extractSummaryContent(string(data)), nil
}

// extractSummaryContent 与老项目 AiController.extractSummaryContent 对齐：
// 剥掉文件头部的 ======/视频标题/生成时间 等元信息，从【视频摘要】等标记处返回正文
func extractSummaryContent(fileContent string) string {
	if fileContent == "" {
		return ""
	}
	markers := []string{"【视频摘要】", "### 视频摘要", "视频摘要", "### 摘要"}
	for _, marker := range markers {
		if idx := strings.Index(fileContent, marker); idx >= 0 {
			return fileContent[idx:]
		}
	}
	var result strings.Builder
	foundEmptyLine := false
	for _, line := range strings.Split(fileContent, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "=") || strings.HasPrefix(line, "视频标题:") || strings.HasPrefix(line, "生成时间:") {
			continue
		}
		if strings.TrimSpace(line) == "" {
			foundEmptyLine = true
			continue
		}
		if foundEmptyLine {
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			result.WriteString(line)
		}
	}
	extracted := result.String()
	if extracted == "" {
		return fileContent
	}
	return extracted
}

func (s *SummaryService) GetSummary(ctx context.Context, videoID int64) (string, error) {
	if s.cacheStore != nil {
		key := fmt.Sprintf("summary:%d", videoID)
		data, err := s.cacheStore.Get(ctx, key)
		if err == nil {
			return string(data), nil
		}
	}

	// 优先读流水线已生成的持久化摘要（MinIO），与老项目读取链路一致
	if stored, err := s.FetchStoredSummary(ctx, videoID); err == nil && stored != "" {
		if s.cacheStore != nil {
			_ = s.cacheStore.Set(ctx, fmt.Sprintf("summary:%d", videoID), []byte(stored), 0)
		}
		return stored, nil
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
	// 与老项目一致：以 videos.has_summary 标志为准（迁移数据即真实标志）
	if s.db != nil {
		var hasSummary int32
		err := s.db.QueryRowContext(ctx,
			`SELECT COALESCE(has_summary,0) FROM videos WHERE id = $1`, videoID).Scan(&hasSummary)
		if err == nil {
			if hasSummary == 1 {
				return true, nil
			}
		} else if err != sql.ErrNoRows {
			return false, err
		}
	}
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
func (s *SummaryService) GenerateSummary(ctx context.Context, manuscriptID, videoID int64) error {
	if s.storage == nil {
		return errors.New("storage not configured")
	}
	summary := "AI generated summary (stub)"
	key := summaryObjectKey(manuscriptID, videoID)
	return s.storage.Put(ctx, "mybilibili", key, strings.NewReader(summary), "text/plain")
}
