package clients

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"mybilibili/pkg/auth"
)

const recorderTestSecret = "recorder-test-secret"

// newTestRecorder 构造带签名能力的 recorder（与 NewHTTPProfileRecorder 等价，
// 仅 baseURL/client 指向测试服务器）。
func newTestRecorder(baseURL string, client *http.Client) *HTTPProfileRecorder {
	return &HTTPProfileRecorder{baseURL: baseURL, client: client, jwt: auth.NewJWT(recorderTestSecret)}
}

// assertBearerUserID 断言请求以 Bearer 携带了目标用户身份（而非裸 X-User-Id 头）。
func assertBearerUserID(t *testing.T, r *http.Request, want int64) {
	t.Helper()
	assert.Equal(t, "", r.Header.Get("X-User-Id"), "不得再直塞 X-User-Id，该头对下游不可信")
	tok, ok := auth.BearerToken(r.Header.Get("Authorization"))
	if !assert.True(t, ok, "应携带 Bearer 凭证") {
		return
	}
	claims, err := auth.NewJWT(recorderTestSecret).Parse(tok)
	assert.NoError(t, err)
	if assert.NotNil(t, claims) {
		assert.Equal(t, want, claims.UserId)
	}
}

func TestNewHTTPProfileRecorder_DefaultAddr(t *testing.T) {
	t.Setenv("SEARCH_SERVICE_ADDR", "")
	r := NewHTTPProfileRecorder()
	assert.Equal(t, "http://127.0.0.1:8084", r.baseURL)
	assert.NotNil(t, r.client)
	assert.NotNil(t, r.jwt, "必须具备签发能力，否则身份无法送达下游")
}

func TestNewHTTPProfileRecorder_EnvAddr(t *testing.T) {
	t.Setenv("SEARCH_SERVICE_ADDR", "http://example.com:9999")
	r := NewHTTPProfileRecorder()
	assert.Equal(t, "http://example.com:9999", r.baseURL)
}

func TestRecordWatch_Success(t *testing.T) {
	var mu sync.Mutex
	var capturedReq *http.Request
	var capturedBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		capturedReq = r
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r := newTestRecorder(srv.URL, srv.Client())
	err := r.RecordWatch(context.Background(), 123, 456, []string{"gaming", "funny"}, 300)
	assert.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "POST", capturedReq.Method)
	assert.Equal(t, "/api/v1/profile/record/watch", capturedReq.URL.Path)
	assertBearerUserID(t, capturedReq, 123)
	assert.Equal(t, "application/json", capturedReq.Header.Get("Content-Type"))
	assert.Equal(t, float64(456), capturedBody["categoryId"])
	assert.Equal(t, float64(300), capturedBody["durationSeconds"])
	tags := capturedBody["tags"].([]interface{})
	assert.Len(t, tags, 2)
}

func TestRecordLike_Success(t *testing.T) {
	var mu sync.Mutex
	var capturedReq *http.Request
	var capturedBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		capturedReq = r
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r := newTestRecorder(srv.URL, srv.Client())
	err := r.RecordLike(context.Background(), 100, 200, []string{"tech"})
	assert.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "/api/v1/profile/record/like", capturedReq.URL.Path)
	assertBearerUserID(t, capturedReq, 100)
	assert.Equal(t, float64(200), capturedBody["categoryId"])
	assert.Equal(t, float64(0), capturedBody["durationSeconds"])
}

func TestRecordCollect_Success(t *testing.T) {
	var mu sync.Mutex
	var capturedReq *http.Request

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		capturedReq = r
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r := newTestRecorder(srv.URL, srv.Client())
	err := r.RecordCollect(context.Background(), 77, 88, []string{"coding", "go"})
	assert.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "/api/v1/profile/record/collect", capturedReq.URL.Path)
	assertBearerUserID(t, capturedReq, 77)
}

func TestRecordWatch_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	r := newTestRecorder(srv.URL, srv.Client())
	err := r.RecordWatch(context.Background(), 1, 2, nil, 0)
	assert.NoError(t, err)
}

func TestRecordWatch_NetworkError(t *testing.T) {
	r := &HTTPProfileRecorder{baseURL: "http://127.0.0.1:1", client: &http.Client{}, jwt: auth.NewJWT(recorderTestSecret)}
	err := r.record(context.Background(), "watch", 1, 2, nil, 0)
	assert.Error(t, err)
}

func TestRecordWatch_EmptyTags(t *testing.T) {
	var capturedBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r := newTestRecorder(srv.URL, srv.Client())
	err := r.RecordWatch(context.Background(), 1, 2, nil, 100)
	assert.NoError(t, err)

	tags := capturedBody["tags"]
	if tags == nil {
		assert.Nil(t, tags)
	} else {
		arr, ok := tags.([]interface{})
		assert.True(t, ok)
		assert.Empty(t, arr)
	}
}
