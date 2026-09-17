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
)

func TestNewHTTPProfileRecorder_DefaultAddr(t *testing.T) {
	t.Setenv("SEARCH_SERVICE_ADDR", "")
	r := NewHTTPProfileRecorder()
	assert.Equal(t, "http://127.0.0.1:8084", r.baseURL)
	assert.NotNil(t, r.client)
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

	r := &HTTPProfileRecorder{baseURL: srv.URL, client: srv.Client()}
	err := r.RecordWatch(context.Background(), 123, 456, []string{"gaming", "funny"}, 300)
	assert.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "POST", capturedReq.Method)
	assert.Equal(t, "/api/v1/profile/record/watch", capturedReq.URL.Path)
	assert.Equal(t, "123", capturedReq.Header.Get("X-User-Id"))
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

	r := &HTTPProfileRecorder{baseURL: srv.URL, client: srv.Client()}
	err := r.RecordLike(context.Background(), 100, 200, []string{"tech"})
	assert.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "/api/v1/profile/record/like", capturedReq.URL.Path)
	assert.Equal(t, "100", capturedReq.Header.Get("X-User-Id"))
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

	r := &HTTPProfileRecorder{baseURL: srv.URL, client: srv.Client()}
	err := r.RecordCollect(context.Background(), 77, 88, []string{"coding", "go"})
	assert.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "/api/v1/profile/record/collect", capturedReq.URL.Path)
	assert.Equal(t, "77", capturedReq.Header.Get("X-User-Id"))
}

func TestRecordWatch_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	r := &HTTPProfileRecorder{baseURL: srv.URL, client: srv.Client()}
	err := r.RecordWatch(context.Background(), 1, 2, nil, 0)
	assert.NoError(t, err)
}

func TestRecordWatch_NetworkError(t *testing.T) {
	r := &HTTPProfileRecorder{baseURL: "http://127.0.0.1:1", client: &http.Client{}}
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

	r := &HTTPProfileRecorder{baseURL: srv.URL, client: srv.Client()}
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
