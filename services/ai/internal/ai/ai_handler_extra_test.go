package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "mybilibili/pkg/pb"
)

// ==================== WithReview / WithCustomer ====================

func TestHandler_WithReview(t *testing.T) {
	h, _ := newTestHandler(t)
	svc := NewReviewService(nil)
	assert.Equal(t, h, h.WithReview(svc))
	assert.Equal(t, svc, h.reviewSvc)
}

func TestHandler_WithCustomer(t *testing.T) {
	h, _ := newTestHandler(t)
	svc := NewCustomerService(nil)
	assert.Equal(t, h, h.WithCustomer(svc))
	assert.Equal(t, svc, h.customerSvc)
}

// ==================== maskKey short branch ====================

func TestMaskKey_Short(t *testing.T) {
	assert.Equal(t, "****", maskKey("short"))
	assert.Equal(t, "****", maskKey(""))
	assert.Equal(t, "abcdef****ijkl", maskKey("abcdefghijkl"))
}

// ==================== handleConfigTest / assistant / route 405 ====================

func TestHandleConfigTest_405(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/config/test", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleAssistantSend_405(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/assistant/send", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleRouteTest_405(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/skills/customer-service/route-test", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleRouteSkills_405(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/skills/customer-service/route", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleCustomerDefaults_POST(t *testing.T) {
	h, mock := newTestHandler(t)
	for i := 0; i < 6; i++ {
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM ai_skills WHERE name`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
		mock.ExpectQuery(`INSERT INTO ai_skills`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(i + 1)))
	}
	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/skills/customer-service/defaults", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(6), resp.Data["created"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleSkills POST create single skill ====================

func TestHandleSkills_POST_CreateSingle(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`INSERT INTO ai_skills`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(9)))
	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/skills",
		`{"name":"新技能","description":"d","type":"LLM"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== handleBindingsByPath edges ====================

func TestHandleBindingsByPath_NoFeature(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/bindings/", "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleBindingsByPath_MethodNotAllowed(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPut, "/api/v1/ai/bindings/chat", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

// ==================== handleSummary generate 500 ====================

func TestHandleSummary_Generate_500(t *testing.T) {
	h, _ := newTestHandler(t)
	h.WithSummary(NewSummaryService(&mockCaller{summary: "x"}))
	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/summary/generate",
		`{"manuscript_id":1,"video_id":123}`)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// ==================== handleSummaryStream (SSE) ====================

func TestHandleSummaryStream_ErrorEvent(t *testing.T) {
	h, _ := newTestHandler(t)
	h.WithSummary(NewSummaryService(&mockCallerErr{err: errors.New("stream failed")}))
	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/summary/stream/123", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "event: error")
	assert.Contains(t, rr.Body.String(), "生成摘要失败")
}

func TestHandleSummaryStream_StreamChunks(t *testing.T) {
	h, _ := newTestHandler(t)
	h.WithSummary(NewSummaryService(&mockCaller{summary: "流式摘要数据"}))
	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/summary/stream/123", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(t, body, "event: start")
	assert.Contains(t, body, "event: done")
	assert.Contains(t, body, base64.StdEncoding.EncodeToString([]byte("流式摘要数据")))
}

// streamFailCaller 模拟流式接口失败但一次性接口成功，覆盖 GetSummary 回退分支。
type streamFailCaller struct {
	summary string
}

func (c *streamFailCaller) Call(ctx context.Context, target, method string, req, resp interface{}) error {
	if s, ok := resp.(*struct {
		Summary string `json:"summary"`
	}); ok {
		s.Summary = c.summary
	}
	return nil
}

func (c *streamFailCaller) CallStream(ctx context.Context, target, method string, req interface{}) (<-chan []byte, error) {
	return nil, errors.New("stream unavailable")
}

func (c *streamFailCaller) Close() error { return nil }

func TestHandleSummaryStream_GetSummaryFallback(t *testing.T) {
	h, _ := newTestHandler(t)
	h.WithSummary(NewSummaryService(&streamFailCaller{summary: "一次性摘要"}))
	mux := http.NewServeMux()
	h.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/summary/stream/123", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(t, body, base64.StdEncoding.EncodeToString([]byte("一次性摘要")))
	assert.Contains(t, body, "event: done")
}

// ==================== AIChatHandler review/customer endpoints ====================

func TestHandleReviewComment_405(t *testing.T) {
	chatH := NewAIChatHandler(nil, nil)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/review/comment", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleReviewReply_200(t *testing.T) {
	review := NewReviewService(&mockCaller{moderate: map[string]interface{}{"passed": false, "reason": "违规"}})
	chatH := NewAIChatHandler(review, nil)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/review/reply?content=hello", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Passed bool   `json:"passed"`
		Reason string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.False(t, resp.Passed)
	assert.Equal(t, "违规", resp.Reason)
}

func TestHandleReviewReply_405(t *testing.T) {
	chatH := NewAIChatHandler(nil, nil)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/review/reply", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleReviewReport_200(t *testing.T) {
	review := NewReviewService(&mockCaller{moderate: map[string]interface{}{"passed": true, "reason": ""}})
	chatH := NewAIChatHandler(review, nil)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/review/report?content=x", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Passed bool `json:"passed"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp.Passed)
}

func TestHandleReviewReport_405(t *testing.T) {
	chatH := NewAIChatHandler(nil, nil)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodDelete, "/api/v1/ai/review/report", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleReviewContent_405(t *testing.T) {
	chatH := NewAIChatHandler(nil, nil)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/review/content?content=x", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleReviewContent_WithTags(t *testing.T) {
	review := NewReviewService(&mockCaller{moderate: map[string]interface{}{"passed": false, "tags": []interface{}{"violence", "spam"}}})
	chatH := NewAIChatHandler(review, nil)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/review/content?content=x&scene=VIDEO", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var tags []string
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &tags))
	assert.Equal(t, []string{"violence", "spam"}, tags)
}

func TestModerateWith_ErrorFallback(t *testing.T) {
	review := NewReviewService(&errorCaller{err: errors.New("down")})
	chatH := NewAIChatHandler(review, nil)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/review/reply?content=x", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Passed bool `json:"passed"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp.Passed)
}

func TestHandleCustomerChat_200(t *testing.T) {
	customer := NewCustomerService(&fakeCaller{customerReply: "您好"})
	chatH := NewAIChatHandler(nil, customer)
	mux := http.NewServeMux()
	chatH.Register(mux)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/customer/chat", strings.NewReader(`{"content":"你好"}`))
	req.Header.Set("X-User-Id", "42")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Reply string `json:"reply"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "您好", resp.Reply)
}

func TestHandleCustomerChat_Error(t *testing.T) {
	customer := NewCustomerService(&errorCaller{err: errors.New("down")})
	chatH := NewAIChatHandler(nil, customer)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/customer/chat", `{"content":"hi"}`)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandleCustomerChat_405(t *testing.T) {
	chatH := NewAIChatHandler(nil, nil)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/customer/chat", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleCustomerHistory_200(t *testing.T) {
	customer := NewCustomerService(&fakeCaller{customerHist: []map[string]interface{}{{"role": "user", "content": "hi"}}})
	chatH := NewAIChatHandler(nil, customer)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/customer/history/7", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
}

func TestHandleCustomerHistory_Error(t *testing.T) {
	customer := NewCustomerService(&errorCaller{err: errors.New("down")})
	chatH := NewAIChatHandler(nil, customer)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/customer/history/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
}

func TestHandleCustomerHistory_405(t *testing.T) {
	chatH := NewAIChatHandler(nil, nil)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/customer/history/1", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleCustomerTransfer_200(t *testing.T) {
	customer := NewCustomerService(&fakeCaller{})
	chatH := NewAIChatHandler(nil, customer)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/customer/transfer", "")
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleCustomerTransfer_Error(t *testing.T) {
	customer := NewCustomerService(&errorCaller{err: errors.New("down")})
	chatH := NewAIChatHandler(nil, customer)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodPost, "/api/v1/ai/customer/transfer", "")
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandleCustomerTransfer_405(t *testing.T) {
	chatH := NewAIChatHandler(nil, nil)
	mux := http.NewServeMux()
	chatH.Register(mux)
	rr := doAI(t, mux, http.MethodGet, "/api/v1/ai/customer/transfer", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

// ==================== ReviewComment error path ====================

func TestReviewComment_ModerateError(t *testing.T) {
	svc := NewReviewService(&errorCaller{err: errors.New("down")})
	passed, err := svc.ReviewComment(context.Background(), "x")
	assert.False(t, passed)
	assert.Error(t, err)
}

// ==================== gRPC server ====================

func TestNewGrpcServer(t *testing.T) {
	s := NewGrpcServer(NewReviewService(nil), NewSummaryService(nil))
	assert.NotNil(t, s)
}

func TestGrpcReviewContent_DefaultScene(t *testing.T) {
	s := NewGrpcServer(NewReviewService(&fakeCaller{moderate: map[string]interface{}{"passed": true, "reason": ""}}), nil)
	resp, err := s.ReviewContent(context.Background(), &pb.ReviewContentRequest{Content: "hello"})
	require.NoError(t, err)
	assert.True(t, resp.Passed)
}

func TestGrpcReviewContent_WithSceneRejected(t *testing.T) {
	s := NewGrpcServer(NewReviewService(&fakeCaller{moderate: map[string]interface{}{"passed": false, "reason": "违规"}}), nil)
	resp, err := s.ReviewContent(context.Background(), &pb.ReviewContentRequest{Content: "x", Scene: "VIDEO"})
	require.NoError(t, err)
	assert.False(t, resp.Passed)
	assert.Equal(t, "违规", resp.Reason)
}

func TestGrpcReviewContent_Error(t *testing.T) {
	s := NewGrpcServer(NewReviewService(&errorCaller{err: errors.New("down")}), nil)
	resp, err := s.ReviewContent(context.Background(), &pb.ReviewContentRequest{Content: "x"})
	require.NoError(t, err)
	assert.True(t, resp.Passed)
}

func TestGrpcGetSummary_NilSvc(t *testing.T) {
	s := NewGrpcServer(nil, nil)
	resp, err := s.GetSummary(context.Background(), &pb.GetSummaryRequest{VideoId: 1})
	require.NoError(t, err)
	assert.False(t, resp.HasSummary)
}

func TestGrpcGetSummary_Ok(t *testing.T) {
	s := NewGrpcServer(nil, NewSummaryService(&fakeCaller{summary: "s"}))
	resp, err := s.GetSummary(context.Background(), &pb.GetSummaryRequest{VideoId: 1})
	require.NoError(t, err)
	assert.True(t, resp.HasSummary)
	assert.Equal(t, "s", resp.Summary)
}

func TestGrpcGetSummary_Error(t *testing.T) {
	s := NewGrpcServer(nil, NewSummaryService(&errorCaller{err: errors.New("down")}))
	resp, err := s.GetSummary(context.Background(), &pb.GetSummaryRequest{VideoId: 1})
	require.NoError(t, err)
	assert.False(t, resp.HasSummary)
}
