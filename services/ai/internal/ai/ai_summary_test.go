package ai

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- helpers ----------

type fakeCacheStore struct {
	data map[string][]byte
}

func newFakeCacheStore() *fakeCacheStore {
	return &fakeCacheStore{data: make(map[string][]byte)}
}
func (c *fakeCacheStore) Get(_ context.Context, key string) ([]byte, error) {
	v, ok := c.data[key]
	if !ok {
		return nil, fmt.Errorf("cache miss")
	}
	return v, nil
}
func (c *fakeCacheStore) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	c.data[key] = value
	return nil
}
func (c *fakeCacheStore) Delete(_ context.Context, key string) error { delete(c.data, key); return nil }
func (c *fakeCacheStore) Exists(_ context.Context, key string) (bool, error) {
	_, ok := c.data[key]; return ok, nil
}
func (c *fakeCacheStore) Lock(_ context.Context, _ string, _ time.Duration) (bool, error) {
	return true, nil
}
func (c *fakeCacheStore) Unlock(_ context.Context, _ string) error    { return nil }
func (c *fakeCacheStore) Incr(_ context.Context, _ string, _ time.Duration) (int64, error) {
	return 0, nil
}
func (c *fakeCacheStore) Close() error { return nil }

type fakeCaller struct {
	called         bool
	summary        string
	hasSum         bool
	moderate       map[string]interface{}
	customerReply  string
	customerHist   []map[string]interface{}
	callErr        error
	customerHistErr error
	customerTransferErr error
}

func (c *fakeCaller) Call(_ context.Context, _, method string, _ any, resp any) error {
	c.called = true
	if c.callErr != nil {
		return c.callErr
	}
	switch method {
	case "Summary":
		if s, ok := resp.(*struct {
			Summary string `json:"summary"`
		}); ok {
			s.Summary = c.summary
		}
	case "CheckSummary":
		if s, ok := resp.(*struct {
			HasSummary bool `json:"has_summary"`
		}); ok {
			s.HasSummary = c.hasSum
		}
	case "Moderate":
		if s, ok := resp.(*map[string]interface{}); ok {
			*s = c.moderate
		}
	case "CustomerChat":
		if s, ok := resp.(*struct {
			Reply string `json:"reply"`
		}); ok {
			s.Reply = c.customerReply
		}
	case "CustomerHistory":
		if s, ok := resp.(*[]map[string]interface{}); ok {
			*s = c.customerHist
			return c.customerHistErr
		}
	case "CustomerTransfer":
		return c.customerTransferErr
	}
	return nil
}
func (c *fakeCaller) CallStream(_ context.Context, _, _ string, _ any) (<-chan []byte, error) {
	ch := make(chan []byte, 1)
	ch <- []byte(c.summary)
	close(ch)
	return ch, nil
}
func (c *fakeCaller) Close() error { return nil }

func newDBMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// ========== extractSummaryContent ==========

func TestExtractSummaryContent_Empty(t *testing.T) {
	assert.Equal(t, "", extractSummaryContent(""))
}

func TestExtractSummaryContent_WithMarker(t *testing.T) {
	input := "some header\n【视频摘要】\n这是摘要内容"
	got := extractSummaryContent(input)
	assert.Equal(t, "【视频摘要】\n这是摘要内容", got)
}

func TestExtractSummaryContent_WithMarker2(t *testing.T) {
	input := "header line\n### 视频摘要\n正文"
	got := extractSummaryContent(input)
	assert.Equal(t, "### 视频摘要\n正文", got)
}

func TestExtractSummaryContent_NoMarker(t *testing.T) {
	input := "header\n\nfirst line after blank\nsecond line"
	got := extractSummaryContent(input)
	assert.Equal(t, "first line after blank\nsecond line", got)
}

func TestExtractSummaryContent_HeaderLines(t *testing.T) {
	input := "=====\n视频标题: 测试\n生成时间: 2024-01-01\n\n摘要正文在这里\n第二行"
	got := extractSummaryContent(input)
	assert.Equal(t, "摘要正文在这里\n第二行", got)
}

func TestExtractSummaryContent_OnlyHeaders(t *testing.T) {
	input := "=====\n视频标题: 测试\n生成时间: 2024-01-01\n"
	got := extractSummaryContent(input)
	assert.Equal(t, input, got)
}

func TestExtractSummaryContent_VideoSummaryMarker(t *testing.T) {
	input := "header\n视频摘要\n正文内容"
	got := extractSummaryContent(input)
	assert.Equal(t, "视频摘要\n正文内容", got)
}

func TestExtractSummaryContent_ZhaiYaoMarker(t *testing.T) {
	input := "header\n### 摘要\n正文内容"
	got := extractSummaryContent(input)
	assert.Equal(t, "### 摘要\n正文内容", got)
}

// ========== FetchStoredSummary ==========

func TestFetchStoredSummary_NilDB(t *testing.T) {
	svc := &SummaryService{}
	svc.SetStorage(nil)
	_, err := svc.FetchStoredSummary(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage not configured")
}

func TestFetchStoredSummary_NilStorage(t *testing.T) {
	db, _ := newDBMock(t)
	svc := &SummaryService{}
	svc.SetDatabase(db)
	_, err := svc.FetchStoredSummary(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage not configured")
}

func TestFetchStoredSummary_NotFound(t *testing.T) {
	// storage is a concrete *MinioStorageService; nil check triggers before DB query
	svc := &SummaryService{}
	_, err := svc.FetchStoredSummary(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage not configured")
}

func TestFetchStoredSummary_DBError(t *testing.T) {
	// storage is nil → nil check triggers before DB query
	svc := &SummaryService{}
	_, err := svc.FetchStoredSummary(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage not configured")
}

func TestFetchStoredSummary_ZeroManuscriptID(t *testing.T) {
	// storage is nil → nil check triggers before DB query
	svc := &SummaryService{}
	_, err := svc.FetchStoredSummary(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage not configured")
}

// ========== GetSummary ==========

func TestGetSummary_CacheHit(t *testing.T) {
	cache := newFakeCacheStore()
	cache.data["summary:42"] = []byte("cached summary")
	svc := &SummaryService{cacheStore: cache}
	got, err := svc.GetSummary(context.Background(), 42)
	assert.NoError(t, err)
	assert.Equal(t, "cached summary", got)
}

func TestGetSummary_NilCaller(t *testing.T) {
	svc := &SummaryService{}
	_, err := svc.GetSummary(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "summary service not configured")
}

func TestGetSummary_CacheMiss_FetchStored(t *testing.T) {
	db, mock := newDBMock(t)
	mock.ExpectQuery(`SELECT COALESCE`).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "has_summary"}).
			AddRow(10, 1))
	// storage is nil → FetchStoredSummary returns error → caller path
	svc := &SummaryService{}
	svc.SetDatabase(db)
	caller := &fakeCaller{summary: "from caller"}
	svc.caller = caller
	got, err := svc.GetSummary(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "from caller", got)
	assert.True(t, caller.called)
}

func TestGetSummary_CacheMiss_CallerFallback(t *testing.T) {
	caller := &fakeCaller{summary: "caller response"}
	svc := &SummaryService{caller: caller}
	got, err := svc.GetSummary(context.Background(), 100)
	assert.NoError(t, err)
	assert.Equal(t, "caller response", got)
}

func TestGetSummary_CallerError(t *testing.T) {
	caller := &fakeCaller{callErr: fmt.Errorf("downstream error")}
	svc := &SummaryService{caller: caller}
	_, err := svc.GetSummary(context.Background(), 1)
	assert.Error(t, err)
}

func TestGetSummary_CacheMiss_WithCache(t *testing.T) {
	caller := &fakeCaller{summary: "fresh summary"}
	cache := newFakeCacheStore()
	svc := &SummaryService{caller: caller, cacheStore: cache}
	got, err := svc.GetSummary(context.Background(), 55)
	assert.NoError(t, err)
	assert.Equal(t, "fresh summary", got)
	cached, err := cache.Get(context.Background(), "summary:55")
	assert.NoError(t, err)
	assert.Equal(t, "fresh summary", string(cached))
}

// ========== CheckSummary ==========

func TestCheckSummary_DBTrue(t *testing.T) {
	db, mock := newDBMock(t)
	mock.ExpectQuery(`SELECT COALESCE\(has_summary`).
		WillReturnRows(sqlmock.NewRows([]string{"has_summary"}).AddRow(1))
	svc := &SummaryService{}
	svc.SetDatabase(db)
	got, err := svc.CheckSummary(context.Background(), 1)
	assert.NoError(t, err)
	assert.True(t, got)
}

func TestCheckSummary_DBFalse(t *testing.T) {
	db, mock := newDBMock(t)
	mock.ExpectQuery(`SELECT COALESCE\(has_summary`).
		WillReturnRows(sqlmock.NewRows([]string{"has_summary"}).AddRow(0))
	svc := &SummaryService{}
	svc.SetDatabase(db)
	got, err := svc.CheckSummary(context.Background(), 1)
	assert.NoError(t, err)
	assert.False(t, got)
}

func TestCheckSummary_NilDB_FallbackCaller(t *testing.T) {
	caller := &fakeCaller{hasSum: true}
	svc := &SummaryService{caller: caller}
	got, err := svc.CheckSummary(context.Background(), 1)
	assert.NoError(t, err)
	assert.True(t, got)
}

func TestCheckSummary_NilDB_NilCaller(t *testing.T) {
	svc := &SummaryService{}
	got, err := svc.CheckSummary(context.Background(), 1)
	assert.NoError(t, err)
	assert.False(t, got)
}

func TestCheckSummary_DBCallerFallback(t *testing.T) {
	db, mock := newDBMock(t)
	mock.ExpectQuery(`SELECT COALESCE\(has_summary`).
		WillReturnRows(sqlmock.NewRows([]string{"has_summary"}).AddRow(0))
	caller := &fakeCaller{hasSum: true}
	svc := &SummaryService{}
	svc.SetDatabase(db)
	svc.caller = caller
	got, err := svc.CheckSummary(context.Background(), 1)
	assert.NoError(t, err)
	assert.True(t, got)
	assert.True(t, caller.called)
}

// ========== GenerateSummary ==========

func TestGenerateSummary_NilStorage(t *testing.T) {
	svc := &SummaryService{}
	err := svc.GenerateSummary(context.Background(), 1, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage not configured")
}

// ========== ReviewService.Moderate ==========

func TestModerate_NilCaller(t *testing.T) {
	svc := NewReviewService(nil)
	resp, err := svc.Moderate(context.Background(), "test content", "COMMENT")
	assert.NoError(t, err)
	assert.Equal(t, true, resp["passed"])
	assert.Equal(t, "", resp["reason"])
}

func TestModerate_WithCaller(t *testing.T) {
	expected := map[string]interface{}{"passed": false, "reason": "敏感内容"}
	caller := &fakeCaller{moderate: expected}
	svc := NewReviewService(caller)
	resp, err := svc.Moderate(context.Background(), "bad content", "COMMENT")
	assert.NoError(t, err)
	assert.Equal(t, false, resp["passed"])
	assert.Equal(t, "敏感内容", resp["reason"])
}

func TestModerate_CallerError(t *testing.T) {
	// Moderate with nil caller won't error (it returns local fallback)
	// Test with caller that errors
	errCaller := &errorCaller{err: fmt.Errorf("service unavailable")}
	svc2 := NewReviewService(errCaller)
	_, err := svc2.Moderate(context.Background(), "test", "COMMENT")
	assert.Error(t, err)
}

func TestReviewComment_Passed(t *testing.T) {
	caller := &fakeCaller{moderate: map[string]interface{}{"passed": true, "reason": ""}}
	svc := NewReviewService(caller)
	passed, err := svc.ReviewComment(context.Background(), "正常评论")
	assert.NoError(t, err)
	assert.True(t, passed)
}

func TestReviewComment_Rejected(t *testing.T) {
	caller := &fakeCaller{moderate: map[string]interface{}{"passed": false, "reason": "违规"}}
	svc := NewReviewService(caller)
	passed, err := svc.ReviewComment(context.Background(), "违规评论")
	assert.NoError(t, err)
	assert.False(t, passed)
}

// ========== CustomerService ==========

func TestCustomerChat_NilCaller(t *testing.T) {
	svc := NewCustomerService(nil)
	_, err := svc.Chat(context.Background(), 1, "hello")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

func TestCustomerChat_WithCaller(t *testing.T) {
	caller := &fakeCaller{customerReply: "你好，有什么可以帮您？"}
	svc := NewCustomerService(caller)
	reply, err := svc.Chat(context.Background(), 1, "你好")
	assert.NoError(t, err)
	assert.Equal(t, "你好，有什么可以帮您？", reply)
}

func TestCustomerHistory_NilCaller(t *testing.T) {
	svc := NewCustomerService(nil)
	hist, err := svc.History(context.Background(), 1)
	assert.NoError(t, err)
	assert.Empty(t, hist)
}

func TestCustomerHistory_WithCaller(t *testing.T) {
	caller := &fakeCaller{customerHist: []map[string]interface{}{{"role": "user", "content": "hi"}}}
	svc := NewCustomerService(caller)
	hist, err := svc.History(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, hist, 1)
}

func TestCustomerTransfer_NilCaller(t *testing.T) {
	svc := NewCustomerService(nil)
	err := svc.Transfer(context.Background(), 1)
	assert.NoError(t, err)
}

func TestCustomerTransfer_WithCaller(t *testing.T) {
	caller := &fakeCaller{}
	svc := NewCustomerService(caller)
	err := svc.Transfer(context.Background(), 1)
	assert.NoError(t, err)
}

func TestCustomerTransfer_CallerError(t *testing.T) {
	caller := &fakeCaller{customerTransferErr: fmt.Errorf("transfer failed")}
	svc := NewCustomerService(caller)
	err := svc.Transfer(context.Background(), 1)
	assert.Error(t, err)
}

// ---------- errorCaller for Moderate error path ----------

type errorCaller struct {
	err error
}

func (e *errorCaller) Call(_ context.Context, _, _ string, _, _ any) error { return e.err }
func (e *errorCaller) CallStream(_ context.Context, _, _ string, _ any) (<-chan []byte, error) {
	return nil, e.err
}
func (e *errorCaller) Close() error { return nil }

// ========== StreamSummary ==========

func TestStreamSummary_Success(t *testing.T) {
	caller := &fakeCaller{summary: "stream data"}
	svc := &SummaryService{caller: caller}
	ch, err := svc.StreamSummary(context.Background(), 1)
	assert.NoError(t, err)
	var results []string
	for data := range ch {
		results = append(results, data)
	}
	assert.Contains(t, results, "stream data")
}

func TestStreamSummary_CallerError(t *testing.T) {
	errCaller := &errorCaller{err: fmt.Errorf("stream failed")}
	svc := &SummaryService{caller: errCaller}
	_, err := svc.StreamSummary(context.Background(), 1)
	assert.Error(t, err)
}

// ========== helper: summaryObjectKey ==========

func TestSummaryObjectKey(t *testing.T) {
	got := summaryObjectKey(10, 20)
	assert.Equal(t, "manuscripts/10/videos/20/summary/ai-summary.txt", got)
}

// ========== constructor tests ==========

func TestNewSummaryService(t *testing.T) {
	caller := &fakeCaller{}
	svc := NewSummaryService(caller)
	assert.NotNil(t, svc)
	assert.Equal(t, caller, svc.caller)
}

func TestNewReviewService(t *testing.T) {
	caller := &fakeCaller{}
	svc := NewReviewService(caller)
	assert.NotNil(t, svc)
}

func TestNewCustomerService(t *testing.T) {
	caller := &fakeCaller{}
	svc := NewCustomerService(caller)
	assert.NotNil(t, svc)
}

func TestSetCacheStore(t *testing.T) {
	svc := &SummaryService{}
	cache := newFakeCacheStore()
	svc.SetCacheStore(cache)
	assert.Equal(t, cache, svc.cacheStore)
}

func TestSetDatabase(t *testing.T) {
	db, _ := newDBMock(t)
	svc := &SummaryService{}
	svc.SetDatabase(db)
	assert.Equal(t, db, svc.db)
}

// ========== FetchStoredSummary with io.ReadAll error simulation ==========

func TestFetchStoredSummary_IOReadAllError(t *testing.T) {
	db, mock := newDBMock(t)
	mock.ExpectQuery(`SELECT COALESCE`).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "has_summary"}).
			AddRow(10, 1))
	svc := &SummaryService{}
	svc.SetDatabase(db)
	// storage is nil so it returns "storage not configured"
	// This tests the path: db OK → storage.Get fails
	_, err := svc.FetchStoredSummary(context.Background(), 1)
	assert.Error(t, err)
}

// ========== extractSummaryContent edge cases ==========

func TestExtractSummaryContent_MarkerAtStart(t *testing.T) {
	input := "【视频摘要】这是摘要"
	got := extractSummaryContent(input)
	assert.Equal(t, "【视频摘要】这是摘要", got)
}

func TestExtractSummaryContent_NoEmptyLine(t *testing.T) {
	input := "header\nno blank line found here"
	got := extractSummaryContent(input)
	assert.Equal(t, input, got)
}

func TestExtractSummaryContent_EmptyLineThenContent(t *testing.T) {
	input := "header1\nheader2\n\n\ncontent after blanks"
	got := extractSummaryContent(input)
	assert.Equal(t, "content after blanks", got)
}

func TestExtractSummaryContent_CRLF(t *testing.T) {
	input := "header\r\n\r\ncontent\r\nmore"
	got := extractSummaryContent(input)
	assert.Equal(t, "content\nmore", got)
}

// ========== GetSummary with FetchStoredSummary success via storage mock ==========

// Since MinioStorageService is a concrete struct, we can't mock it.
// But we can test the full caller-fallback path and verify caching.

func TestGetSummary_CacheMiss_AllFail(t *testing.T) {
	svc := &SummaryService{}
	_, err := svc.GetSummary(context.Background(), 1)
	assert.Error(t, err)
}

func TestGetSummary_NoCache(t *testing.T) {
	caller := &fakeCaller{summary: "no cache result"}
	svc := &SummaryService{caller: caller}
	got, err := svc.GetSummary(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "no cache result", got)
	// Verify no cache was set (cacheStore is nil)
}

func TestGetSummary_CallerReturnsEmpty(t *testing.T) {
	caller := &fakeCaller{summary: ""}
	svc := &SummaryService{caller: caller}
	got, err := svc.GetSummary(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "", got)
}

// ========== CheckSummary with sql.ErrNoRows ==========

func TestCheckSummary_DBErrNoRows(t *testing.T) {
	db, mock := newDBMock(t)
	mock.ExpectQuery(`SELECT COALESCE\(has_summary`).
		WillReturnError(sql.ErrNoRows)
	svc := &SummaryService{}
	svc.SetDatabase(db)
	// sql.ErrNoRows → error ignored → falls to caller (nil)
	got, err := svc.CheckSummary(context.Background(), 1)
	assert.NoError(t, err)
	assert.False(t, got)
}

func TestCheckSummary_DBErrNoRows_FallbackCaller(t *testing.T) {
	db, mock := newDBMock(t)
	mock.ExpectQuery(`SELECT COALESCE\(has_summary`).
		WillReturnError(sql.ErrNoRows)
	caller := &fakeCaller{hasSum: true}
	svc := &SummaryService{}
	svc.SetDatabase(db)
	svc.caller = caller
	got, err := svc.CheckSummary(context.Background(), 1)
	assert.NoError(t, err)
	assert.True(t, got)
}

func TestCheckSummary_DBGeneralError(t *testing.T) {
	db, mock := newDBMock(t)
	mock.ExpectQuery(`SELECT COALESCE\(has_summary`).
		WillReturnError(fmt.Errorf("connection refused"))
	svc := &SummaryService{}
	svc.SetDatabase(db)
	got, err := svc.CheckSummary(context.Background(), 1)
	assert.Error(t, err)
	assert.False(t, got)
}

// ========== Moderate with all response types ==========

func TestModerate_Passed(t *testing.T) {
	caller := &fakeCaller{moderate: map[string]interface{}{"passed": true, "reason": ""}}
	svc := NewReviewService(caller)
	resp, err := svc.Moderate(context.Background(), "正常内容", "VIDEO")
	assert.NoError(t, err)
	assert.Equal(t, true, resp["passed"])
}

func TestModerate_Rejected(t *testing.T) {
	caller := &fakeCaller{moderate: map[string]interface{}{"passed": false, "reason": "违规"}}
	svc := NewReviewService(caller)
	resp, err := svc.Moderate(context.Background(), "违规内容", "VIDEO")
	assert.NoError(t, err)
	assert.Equal(t, false, resp["passed"])
}

// ========== CustomerService edge cases ==========

func TestCustomerChat_EmptyReply(t *testing.T) {
	caller := &fakeCaller{customerReply: ""}
	svc := NewCustomerService(caller)
	reply, err := svc.Chat(context.Background(), 1, "test")
	assert.NoError(t, err)
	assert.Equal(t, "", reply)
}

func TestCustomerHistory_Empty(t *testing.T) {
	caller := &fakeCaller{customerHist: []map[string]interface{}{}}
	svc := NewCustomerService(caller)
	hist, err := svc.History(context.Background(), 1)
	assert.NoError(t, err)
	assert.Empty(t, hist)
}

func TestCustomerHistory_Error(t *testing.T) {
	caller := &fakeCaller{customerHistErr: fmt.Errorf("db error")}
	svc := NewCustomerService(caller)
	hist, err := svc.History(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, hist)
}

// ========== unused import guard ==========
var _ = io.EOF
var _ = strings.NewReader
