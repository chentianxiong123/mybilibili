package subtitle

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/pkg/abstraction"
)

// mockStore 实现 abstraction.DocumentStore。
type mockStore struct {
	insertID string
	err      error
}

func (m *mockStore) Insert(ctx context.Context, collection string, doc any) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.insertID != "" {
		return m.insertID, nil
	}
	return "subtitle-1", nil
}

func (m *mockStore) FindByID(ctx context.Context, collection, id string, result any) error {
	if r, ok := result.(*Subtitle); ok {
		*r = Subtitle{ID: id, VideoID: 1, Language: "zh-CN", Content: "[]"}
	}
	return nil
}

func (m *mockStore) Update(ctx context.Context, collection, id string, doc any) error { return nil }

func (m *mockStore) Delete(ctx context.Context, collection, id string) error { return nil }

func (m *mockStore) Query(ctx context.Context, collection string, filter abstraction.QueryFilter, result any) error {
	if r, ok := result.(*[]*Subtitle); ok {
		if videoID, has := filter.Filters["video_id"].(int64); has && videoID == 42 {
			*r = []*Subtitle{{ID: "subtitle-1", VideoID: 42, Language: "zh-CN", LanguageName: "中文", Content: "[]"}}
		}
	}
	return nil
}

// mockStorage 实现 abstraction.StorageService。
type mockStorage struct {
	audio []byte
}

func (m *mockStorage) Put(ctx context.Context, bucket, key string, body io.Reader, contentType string) error {
	return nil
}

func (m *mockStorage) Get(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(m.audio)), nil
}

func (m *mockStorage) Delete(ctx context.Context, bucket, key string) error { return nil }

func (m *mockStorage) Head(ctx context.Context, bucket, key string) (*abstraction.FileInfo, error) {
	return &abstraction.FileInfo{Key: key}, nil
}

func (m *mockStorage) List(ctx context.Context, bucket, prefix string) ([]abstraction.FileInfo, error) {
	return nil, nil
}

func (m *mockStorage) SignedURL(ctx context.Context, bucket, key string, expire time.Duration) (string, error) {
	return "http://minio/" + key, nil
}

func newTestSubtitleHandler(t *testing.T) (*Handler, *mockStore, *mockStorage) {
	t.Helper()
	store := &mockStore{}
	svc := NewService(NewRepository(store))
	h := NewHandler(svc)
	return h, store, &mockStorage{audio: []byte("fake-audio")}
}

func doSubtitle(t *testing.T, h *Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	h.Register(mux)
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestHandleSubtitleGenerate_200(t *testing.T) {
	h, store, storage := newTestSubtitleHandler(t)
	store.insertID = "subtitle-gen-1"
	h.SetGenerator(NewWhisperGenerator(NewRepository(store), storage))

	rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/generate", `{"manuscript_id":10,"video_id":99}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int    `json:"code"`
		Data struct {
			SubtitleID string                   `json:"subtitle_id"`
			Cues       []map[string]interface{} `json:"cues"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "subtitle-gen-1", resp.Data.SubtitleID)
	assert.NotEmpty(t, resp.Data.Cues)
}

func TestHandleSubtitleVideo_200(t *testing.T) {
	h, _, _ := newTestSubtitleHandler(t)

	// 已存在字幕
	rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/video/42", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int       `json:"code"`
		Data []json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)

	// 不存在 -> 空数组
	rr = doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/video/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp2 struct {
		Data []json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp2))
	assert.Empty(t, resp2.Data)
}

func TestParseSRT(t *testing.T) {
	srt := `1
00:00:01,000 --> 00:00:04,000
第一句字幕

2
00:00:05,000 --> 00:00:08,500
第二句字幕 带标点、数字 123

3
00:01:00,000 --> 00:01:02,000
第三句
`
	cues, err := ParseSRT(srt)
	require.NoError(t, err)
	require.Len(t, cues, 3)

	assert.Equal(t, 1, cues[0].Index)
	assert.Equal(t, "第一句字幕", cues[0].Text)
	assert.Equal(t, time.Second, cues[0].Start)
	assert.Equal(t, 4*time.Second, cues[0].End)

	assert.Equal(t, "第二句字幕 带标点、数字 123", cues[1].Text)
	assert.Equal(t, 5*time.Second, cues[1].Start)
	assert.Equal(t, 8*time.Second+500*time.Millisecond, cues[1].End)

	assert.Equal(t, 60*time.Second, cues[2].Start)
	assert.Equal(t, 62*time.Second, cues[2].End)
}