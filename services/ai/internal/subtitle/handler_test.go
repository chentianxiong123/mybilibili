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

func TestHandleSubtitleGenerate_400(t *testing.T) {
	h, store, storage := newTestSubtitleHandler(t)
	h.SetGenerator(NewWhisperGenerator(NewRepository(store), storage))

	// 缺少 video_id → 400
	rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/generate", `{"manuscript_id":10}`)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(400), resp["code"])
}

func TestParseSRT_Basic(t *testing.T) {
	srt := `1
00:00:00,500 --> 00:00:02,000
你好世界

2
00:00:03,000 --> 00:00:05,500
测试字幕第二句`
	cues, err := ParseSRT(srt)
	require.NoError(t, err)
	require.Len(t, cues, 2)

	assert.Equal(t, 1, cues[0].Index)
	assert.Equal(t, "你好世界", cues[0].Text)
	assert.Equal(t, 500*time.Millisecond, cues[0].Start)
	assert.Equal(t, 2*time.Second, cues[0].End)

	assert.Equal(t, 2, cues[1].Index)
	assert.Equal(t, "测试字幕第二句", cues[1].Text)
	assert.Equal(t, 3*time.Second, cues[1].Start)
	assert.Equal(t, 5*time.Second+500*time.Millisecond, cues[1].End)
}

// ---------------------------------------------------------------------------
// flexMock: 可按 filter / findByID 回调返回不同数据的 mock store
// ---------------------------------------------------------------------------

type flexMockStore struct {
	insertID  string
	insertErr error
	deleteErr error
	updateErr error
	findByID  func(ctx context.Context, id string, result any) error
	queryFn   func(ctx context.Context, filter abstraction.QueryFilter, result any) error
}

func (m *flexMockStore) Insert(ctx context.Context, collection string, doc any) (string, error) {
	if m.insertErr != nil {
		return "", m.insertErr
	}
	if m.insertID != "" {
		return m.insertID, nil
	}
	return "flex-sub-1", nil
}

func (m *flexMockStore) FindByID(ctx context.Context, collection, id string, result any) error {
	if m.findByID != nil {
		return m.findByID(ctx, id, result)
	}
	if r, ok := result.(*Subtitle); ok {
		*r = Subtitle{ID: id, VideoID: 100, Language: "zh-CN", LanguageName: "中文", Content: `[{"index":1,"startTime":0,"endTime":1,"text":"test"}]`}
	}
	return nil
}

func (m *flexMockStore) Update(ctx context.Context, collection, id string, doc any) error {
	return m.updateErr
}

func (m *flexMockStore) Delete(ctx context.Context, collection, id string) error {
	return m.deleteErr
}

func (m *flexMockStore) Query(ctx context.Context, collection string, filter abstraction.QueryFilter, result any) error {
	if m.queryFn != nil {
		return m.queryFn(ctx, filter, result)
	}
	if r, ok := result.(*[]*Subtitle); ok {
		*r = []*Subtitle{}
	}
	return nil
}

func newFlexHandler(store *flexMockStore) *Handler {
	return NewHandler(NewService(NewRepository(store)))
}

func doMultipart(t *testing.T, h *Handler, path string, fields map[string]string, fileField, fileName, fileContent string) *httptest.ResponseRecorder {
	t.Helper()
	// 手动构造 multipart body
	boundary := "----TestBoundary"
	var body bytes.Buffer
	for k, v := range fields {
		body.WriteString("--" + boundary + "\r\n")
		body.WriteString("Content-Disposition: form-data; name=\"" + k + "\"\r\n\r\n")
		body.WriteString(v + "\r\n")
	}
	if fileName != "" {
		body.WriteString("--" + boundary + "\r\n")
		body.WriteString("Content-Disposition: form-data; name=\"" + fileField + "\"; filename=\"" + fileName + "\"\r\n")
		body.WriteString("Content-Type: text/plain\r\n\r\n")
		body.WriteString(fileContent)
		body.WriteString("\r\n")
	}
	body.WriteString("--" + boundary + "--\r\n")

	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.Register(mux)
	mux.ServeHTTP(rr, req)
	return rr
}

// ==================== handlePending ====================

func TestHandlePending(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/pending", "")
		assert.Equal(t, http.StatusOK, rr.Code)
		var resp struct {
			Code int                      `json:"code"`
			Data []map[string]interface{} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, 200, resp.Code)
		assert.Empty(t, resp.Data)
	})

	t.Run("with data", func(t *testing.T) {
		store := &flexMockStore{
			queryFn: func(_ context.Context, _ abstraction.QueryFilter, result any) error {
				if r, ok := result.(*[]*Subtitle); ok {
					*r = []*Subtitle{
						{ID: "p1", VideoID: 10, Language: "zh-CN", LanguageName: "中文", Content: `[{"index":1,"startTime":0,"endTime":1,"text":"hi"}]`, Status: 0, UploadTime: time.Now()},
					}
				}
				return nil
			},
		}
		h := newFlexHandler(store)
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/pending", "")
		assert.Equal(t, http.StatusOK, rr.Code)
		var resp struct {
			Code int                      `json:"code"`
			Data []map[string]interface{} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "p1", resp.Data[0]["id"])
	})

	t.Run("nil list fallback", func(t *testing.T) {
		store := &flexMockStore{
			queryFn: func(_ context.Context, _ abstraction.QueryFilter, result any) error {
				// 不设置 r，模拟 nil
				return nil
			},
		}
		h := newFlexHandler(store)
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/pending", "")
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

// ==================== handleAllVideos ====================

func TestHandleAllVideos(t *testing.T) {
	t.Run("GET success", func(t *testing.T) {
		store := &flexMockStore{
			queryFn: func(_ context.Context, _ abstraction.QueryFilter, result any) error {
				if r, ok := result.(*[]*Subtitle); ok {
					*r = []*Subtitle{
						{ID: "a1", VideoID: 10, Language: "zh-CN", LanguageName: "中文", Content: `[{"index":1,"startTime":0,"endTime":1,"text":"hi"}]`, UploadTime: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)},
					}
				}
				return nil
			},
		}
		h := newFlexHandler(store)
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/videos", "")
		assert.Equal(t, http.StatusOK, rr.Code)
		var resp struct {
			Code int                      `json:"code"`
			Data []map[string]interface{} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Len(t, resp.Data, 1)
	})

	t.Run("GET empty", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/videos", "")
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("POST 405", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/videos", "")
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

// ==================== handleScan ====================

func TestHandleScan(t *testing.T) {
	t.Run("GET success", func(t *testing.T) {
		store := &flexMockStore{
			queryFn: func(_ context.Context, _ abstraction.QueryFilter, result any) error {
				if r, ok := result.(*[]*Subtitle); ok {
					*r = []*Subtitle{
						{ID: "s1", VideoID: 55, Language: "zh-CN", LanguageName: "中文", Content: "[]"},
					}
				}
				return nil
			},
		}
		h := newFlexHandler(store)
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/scan/55", "")
		assert.Equal(t, http.StatusOK, rr.Code)
		var resp struct {
			Code int                      `json:"code"`
			Data []map[string]interface{} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Len(t, resp.Data, 1)
	})

	t.Run("GET empty result", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/scan/99", "")
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("POST 405", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/scan/1", "")
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

// ==================== handleUpload ====================

func TestHandleUpload(t *testing.T) {
	t.Run("POST success with content", func(t *testing.T) {
		store := &flexMockStore{insertID: "upload-1"}
		h := newFlexHandler(store)
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/upload",
			`{"video_id":10,"language":"en","language_name":"English","content":"[{\"index\":1,\"startTime\":0,\"endTime\":1,\"text\":\"hello\"}]"}`)
		assert.Equal(t, http.StatusOK, rr.Code)
		var resp struct {
			Code int                    `json:"code"`
			Data map[string]interface{} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "upload-1", resp.Data["id"])
	})

	t.Run("POST srt_content fallback", func(t *testing.T) {
		store := &flexMockStore{insertID: "upload-2"}
		h := newFlexHandler(store)
		srt := "1\n00:00:01,000 --> 00:00:04,000\nHello\n"
		body, _ := json.Marshal(map[string]interface{}{
			"video_id":    10,
			"srt_content": srt,
		})
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/upload", string(body))
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("POST no content -> 400", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/upload",
			`{"video_id":10}`)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("POST default language", func(t *testing.T) {
		store := &flexMockStore{insertID: "upload-3"}
		h := newFlexHandler(store)
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/upload",
			`{"video_id":10,"content":"test"}`)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("POST is_default", func(t *testing.T) {
		store := &flexMockStore{insertID: "upload-4"}
		h := newFlexHandler(store)
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/upload",
			`{"video_id":10,"content":"test","is_default":true}`)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("POST with X-User-Id header", func(t *testing.T) {
		store := &flexMockStore{insertID: "upload-5"}
		h := newFlexHandler(store)
		mux := http.NewServeMux()
		h.Register(mux)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/subtitle/upload",
			strings.NewReader(`{"video_id":10,"content":"test"}`))
		req.Header.Set("X-User-Id", "42")
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("GET 405", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/upload", "")
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

// ==================== handleUploadSRT ====================

func TestHandleUploadSRT(t *testing.T) {
	t.Run("POST success", func(t *testing.T) {
		store := &flexMockStore{insertID: "srt-upload-1"}
		h := newFlexHandler(store)
		srt := "1\n00:00:01,000 --> 00:00:04,000\nHello\n"
		rr := doMultipart(t, h, "/api/v1/subtitle/upload-srt",
			map[string]string{"video_id": "10", "language": "en", "language_name": "English"},
			"file", "test.srt", srt)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("POST with defaults", func(t *testing.T) {
		store := &flexMockStore{insertID: "srt-upload-2"}
		h := newFlexHandler(store)
		srt := "1\n00:00:01,000 --> 00:00:04,000\nHello\n"
		rr := doMultipart(t, h, "/api/v1/subtitle/upload-srt",
			map[string]string{"video_id": "10"},
			"file", "test.srt", srt)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("POST no file -> 400", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		boundary := "----TestBoundary"
		var body bytes.Buffer
		body.WriteString("--" + boundary + "\r\n")
		body.WriteString("Content-Disposition: form-data; name=\"video_id\"\r\n\r\n")
		body.WriteString("10\r\n")
		body.WriteString("--" + boundary + "--\r\n")
		req := httptest.NewRequest(http.MethodPost, "/api/v1/subtitle/upload-srt", &body)
		req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
		rr := httptest.NewRecorder()
		mux := http.NewServeMux()
		h.Register(mux)
		mux.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("POST empty file -> 400", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doMultipart(t, h, "/api/v1/subtitle/upload-srt",
			map[string]string{"video_id": "10"},
			"file", "empty.srt", "")
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("GET 405", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/upload-srt", "")
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

// ==================== handleImport ====================

func TestHandleImport(t *testing.T) {
	t.Run("POST success", func(t *testing.T) {
		store := &flexMockStore{insertID: "import-1"}
		h := newFlexHandler(store)
		srt := "1\n00:00:01,000 --> 00:00:04,000\nHello\n"
		body, _ := json.Marshal(map[string]interface{}{
			"video_id": 10,
			"srt":      srt,
		})
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/import-srt", string(body))
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("POST missing fields -> 400", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/import-srt",
			`{"video_id":10}`)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("POST invalid srt -> 400", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/import-srt",
			`{"video_id":10,"srt":"not valid srt"}`)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("POST empty srt -> 400", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/import-srt",
			`{"video_id":10,"srt":""}`)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("POST zero video_id -> 400", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/import-srt",
			`{"video_id":0,"srt":"1\n00:00:01,000 --> 00:00:04,000\nHello\n"}`)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("GET 405", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/import-srt", "")
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

// ==================== handleImportSystem ====================

func TestHandleImportSystem(t *testing.T) {
	t.Run("POST success", func(t *testing.T) {
		store := &flexMockStore{insertID: "sys-import-1"}
		h := newFlexHandler(store)
		srt := "1\n00:00:01,000 --> 00:00:04,000\nHello\n"
		body, _ := json.Marshal(map[string]interface{}{
			"video_id": 10,
			"srt":      srt,
		})
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/import-system", string(body))
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("POST missing fields -> 400", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/import-system",
			`{"video_id":10}`)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("POST invalid srt fallback to raw", func(t *testing.T) {
		store := &flexMockStore{insertID: "sys-import-2"}
		h := newFlexHandler(store)
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/import-system",
			`{"video_id":10,"srt":"not srt"}`)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("POST empty srt -> 400", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/import-system",
			`{"video_id":10,"srt":""}`)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("GET 405", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/import-system", "")
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

// ==================== handleSetDefault ====================

func TestHandleSetDefault(t *testing.T) {
	t.Run("POST success", func(t *testing.T) {
		store := &flexMockStore{
			queryFn: func(_ context.Context, filter abstraction.QueryFilter, result any) error {
				if r, ok := result.(*[]*Subtitle); ok {
					*r = []*Subtitle{
						{ID: "sub-a", VideoID: 10, Language: "zh-CN", LanguageName: "中文", Content: "[]"},
						{ID: "sub-b", VideoID: 10, Language: "en", LanguageName: "English", Content: "[]", IsDefault: true},
					}
				}
				return nil
			},
		}
		h := newFlexHandler(store)
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/set-default",
			`{"video_id":10,"id":"sub-a"}`)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("GET 405", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/set-default", "")
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

// ==================== handleSubtitleByID ====================

func TestHandleSubtitleByID(t *testing.T) {
	t.Run("GET by id", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/sub-42", "")
		assert.Equal(t, http.StatusOK, rr.Code)
		var resp struct {
			Code int                    `json:"code"`
			Data map[string]interface{} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "sub-42", resp.Data["id"])
	})

	t.Run("GET with srt content", func(t *testing.T) {
		store := &flexMockStore{
			findByID: func(_ context.Context, id string, result any) error {
				if r, ok := result.(*Subtitle); ok {
					*r = Subtitle{
						ID: id, VideoID: 10, Language: "zh-CN", LanguageName: "中文",
						Content: "1\n00:00:01,000 --> 00:00:04,000\nHello\n",
					}
				}
				return nil
			},
		}
		h := newFlexHandler(store)
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/sub-srt", "")
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("GET with invalid content", func(t *testing.T) {
		store := &flexMockStore{
			findByID: func(_ context.Context, id string, result any) error {
				if r, ok := result.(*Subtitle); ok {
					*r = Subtitle{ID: id, VideoID: 10, Language: "zh-CN", Content: "garbage"}
				}
				return nil
			},
		}
		h := newFlexHandler(store)
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/sub-bad", "")
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("DELETE success", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodDelete, "/api/v1/subtitle/sub-1", "")
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("POST approve", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/sub-1/approve", "")
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("POST reject", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/sub-1/reject", "")
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("GET preview", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/sub-1/preview", "")
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("POST set-default via query", func(t *testing.T) {
		store := &flexMockStore{
			queryFn: func(_ context.Context, filter abstraction.QueryFilter, result any) error {
				if r, ok := result.(*[]*Subtitle); ok {
					*r = []*Subtitle{
						{ID: "sub-x", VideoID: 20, Language: "zh-CN", Content: "[]"},
					}
				}
				return nil
			},
		}
		h := newFlexHandler(store)
		rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/sub-x/set-default?video_id=20", "")
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

// ==================== handleVideoSubtitle 语言路径 ====================

func TestHandleVideoSubtitle_Language(t *testing.T) {
	t.Run("specific language found", func(t *testing.T) {
		store := &flexMockStore{
			queryFn: func(_ context.Context, filter abstraction.QueryFilter, result any) error {
				if r, ok := result.(*[]*Subtitle); ok {
					*r = []*Subtitle{
						{ID: "v1", VideoID: 42, Language: "zh-CN", LanguageName: "中文", Content: `[{"index":1,"startTime":0,"endTime":1,"text":"hi"}]`},
						{ID: "v2", VideoID: 42, Language: "en", LanguageName: "English", Content: `[{"index":1,"startTime":0,"endTime":1,"text":"hello"}]`},
					}
				}
				return nil
			},
		}
		h := newFlexHandler(store)
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/video/42/en", "")
		assert.Equal(t, http.StatusOK, rr.Code)
		var resp struct {
			Code int                    `json:"code"`
			Data map[string]interface{} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "v2", resp.Data["id"])
	})

	t.Run("specific language not found", func(t *testing.T) {
		store := &flexMockStore{
			queryFn: func(_ context.Context, filter abstraction.QueryFilter, result any) error {
				if r, ok := result.(*[]*Subtitle); ok {
					*r = []*Subtitle{
						{ID: "v1", VideoID: 42, Language: "zh-CN", LanguageName: "中文", Content: "[]"},
					}
				}
				return nil
			},
		}
		h := newFlexHandler(store)
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/video/42/ja", "")
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("nil list fallback", func(t *testing.T) {
		h := newFlexHandler(&flexMockStore{})
		rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/video/1", "")
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

// ==================== subtitleToMap 覆盖 ====================

func TestSubtitleToMap_JSONContent(t *testing.T) {
	sub := &Subtitle{
		ID: "m1", VideoID: 10, Language: "zh-CN", LanguageName: "中文",
		Format: "srt", IsDefault: false, UploadedBy: 1, Status: 0, Source: "user",
		UploadTime: time.Now(),
		Content:    `[{"index":1,"startTime":0,"endTime":1,"text":"hello"}]`,
	}
	m := subtitleToMap(sub)
	assert.Equal(t, "m1", m["id"])
	cues, ok := m["cues"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, cues, 1)
}

func TestSubtitleToMap_SRTContent(t *testing.T) {
	sub := &Subtitle{
		ID: "m2", VideoID: 10, Language: "zh-CN", LanguageName: "中文",
		Content: "1\n00:00:01,000 --> 00:00:04,000\nHello\n",
	}
	m := subtitleToMap(sub)
	cues, ok := m["cues"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, cues, 1)
}

func TestSubtitleToMap_InvalidContent(t *testing.T) {
	sub := &Subtitle{
		ID: "m3", VideoID: 10, Language: "zh-CN", LanguageName: "中文",
		Content: "garbage",
	}
	m := subtitleToMap(sub)
	cues := m["cues"]
	assert.Empty(t, cues)
}

// ==================== handleGenerate 其他路径 ====================

func TestHandleSubtitleGenerate_MethodNotAllowed(t *testing.T) {
	h := newFlexHandler(&flexMockStore{})
	rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/generate", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleSubtitleGenerate_NilGenerator(t *testing.T) {
	h := newFlexHandler(&flexMockStore{})
	rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/generate", `{"video_id":1}`)
	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}

func TestHandleSubtitleGenerate_BadJSON(t *testing.T) {
	h, store, storage := newTestSubtitleHandler(t)
	h.SetGenerator(NewWhisperGenerator(NewRepository(store), storage))
	rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/generate", `not json`)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ==================== Repository 测试 ====================

func TestRepository_Create(t *testing.T) {
	store := &flexMockStore{insertID: "repo-create-1"}
	repo := NewRepository(store)
	sub := &Subtitle{VideoID: 10, Language: "zh-CN", LanguageName: "中文", Content: "[]"}
	id, err := repo.Create(context.Background(), sub)
	require.NoError(t, err)
	assert.Equal(t, "repo-create-1", id)
	assert.Equal(t, "srt", sub.Format)
}

func TestRepository_Create_InsertError(t *testing.T) {
	store := &flexMockStore{insertErr: assert.AnError}
	repo := NewRepository(store)
	_, err := repo.Create(context.Background(), &Subtitle{})
	assert.Error(t, err)
}

func TestRepository_GetByID(t *testing.T) {
	store := &flexMockStore{}
	repo := NewRepository(store)
	sub, err := repo.GetByID(context.Background(), "test-id")
	require.NoError(t, err)
	assert.Equal(t, "test-id", sub.ID)
}

func TestRepository_ListByVideo(t *testing.T) {
	store := &flexMockStore{
		queryFn: func(_ context.Context, filter abstraction.QueryFilter, result any) error {
			if r, ok := result.(*[]*Subtitle); ok {
				if vid, ok := filter.Filters["video_id"].(int64); ok && vid == 10 {
					*r = []*Subtitle{{ID: "lbv-1", VideoID: 10, Language: "zh-CN", Content: "[]"}}
				}
			}
			return nil
		},
	}
	repo := NewRepository(store)
	list, err := repo.ListByVideo(context.Background(), 10)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestRepository_ListByVideo_Empty(t *testing.T) {
	store := &flexMockStore{}
	repo := NewRepository(store)
	list, err := repo.ListByVideo(context.Background(), 99)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestRepository_GetByLanguage(t *testing.T) {
	store := &flexMockStore{
		queryFn: func(_ context.Context, filter abstraction.QueryFilter, result any) error {
			if r, ok := result.(*[]*Subtitle); ok {
				*r = []*Subtitle{
					{ID: "gl-1", VideoID: 10, Language: "zh-CN", Content: "[]"},
					{ID: "gl-2", VideoID: 10, Language: "en", Content: "[]"},
				}
			}
			return nil
		},
	}
	repo := NewRepository(store)
	sub, err := repo.GetByLanguage(context.Background(), 10, "en")
	require.NoError(t, err)
	assert.Equal(t, "gl-2", sub.ID)
}

func TestRepository_GetByLanguage_NotFound(t *testing.T) {
	store := &flexMockStore{
		queryFn: func(_ context.Context, _ abstraction.QueryFilter, result any) error {
			if r, ok := result.(*[]*Subtitle); ok {
				*r = []*Subtitle{{ID: "gl-1", VideoID: 10, Language: "zh-CN", Content: "[]"}}
			}
			return nil
		},
	}
	repo := NewRepository(store)
	sub, err := repo.GetByLanguage(context.Background(), 10, "ja")
	require.NoError(t, err)
	assert.Nil(t, sub)
}

func TestRepository_ListPending(t *testing.T) {
	store := &flexMockStore{
		queryFn: func(_ context.Context, filter abstraction.QueryFilter, result any) error {
			if r, ok := result.(*[]*Subtitle); ok {
				if s, ok := filter.Filters["status"]; ok && s == 0 {
					*r = []*Subtitle{{ID: "lp-1", VideoID: 5, Language: "zh-CN", Content: "[]", Status: 0}}
				}
			}
			return nil
		},
	}
	repo := NewRepository(store)
	list, err := repo.ListPending(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestRepository_ListAll(t *testing.T) {
	store := &flexMockStore{
		queryFn: func(_ context.Context, _ abstraction.QueryFilter, result any) error {
			if r, ok := result.(*[]*Subtitle); ok {
				*r = []*Subtitle{
					{ID: "la-1", VideoID: 1, Language: "zh-CN", Content: "[]"},
					{ID: "la-2", VideoID: 2, Language: "en", Content: "[]"},
				}
			}
			return nil
		},
	}
	repo := NewRepository(store)
	list, err := repo.ListAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestRepository_Delete(t *testing.T) {
	store := &flexMockStore{}
	repo := NewRepository(store)
	err := repo.Delete(context.Background(), "del-1")
	assert.NoError(t, err)
}

func TestRepository_Delete_Error(t *testing.T) {
	store := &flexMockStore{deleteErr: assert.AnError}
	repo := NewRepository(store)
	err := repo.Delete(context.Background(), "del-1")
	assert.Error(t, err)
}

func TestRepository_SetDefault(t *testing.T) {
	store := &flexMockStore{
		queryFn: func(_ context.Context, _ abstraction.QueryFilter, result any) error {
			if r, ok := result.(*[]*Subtitle); ok {
				*r = []*Subtitle{
					{ID: "sd-1", VideoID: 10, Language: "zh-CN", Content: "[]"},
					{ID: "sd-2", VideoID: 10, Language: "en", Content: "[]", IsDefault: true},
				}
			}
			return nil
		},
	}
	repo := NewRepository(store)
	err := repo.SetDefault(context.Background(), 10, "sd-1")
	assert.NoError(t, err)
}

// ==================== Service 测试 ====================

func TestService_Upload(t *testing.T) {
	store := &flexMockStore{insertID: "svc-upload-1"}
	svc := NewService(NewRepository(store))
	sub, err := svc.Upload(context.Background(), 10, 1, "zh-CN", "中文", "[]")
	require.NoError(t, err)
	assert.Equal(t, "svc-upload-1", sub.ID)
	assert.Equal(t, int64(10), sub.VideoID)
	assert.Equal(t, "user", sub.Source)
}

func TestService_Upload_Error(t *testing.T) {
	store := &flexMockStore{insertErr: assert.AnError}
	svc := NewService(NewRepository(store))
	_, err := svc.Upload(context.Background(), 10, 1, "zh-CN", "中文", "[]")
	assert.Error(t, err)
}

func TestService_ListByVideo(t *testing.T) {
	store := &flexMockStore{
		queryFn: func(_ context.Context, filter abstraction.QueryFilter, result any) error {
			if r, ok := result.(*[]*Subtitle); ok {
				*r = []*Subtitle{{ID: "sl-1", VideoID: 10, Language: "zh-CN", Content: "[]"}}
			}
			return nil
		},
	}
	svc := NewService(NewRepository(store))
	list, err := svc.ListByVideo(context.Background(), 10)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestService_ListAll(t *testing.T) {
	store := &flexMockStore{
		queryFn: func(_ context.Context, _ abstraction.QueryFilter, result any) error {
			if r, ok := result.(*[]*Subtitle); ok {
				*r = []*Subtitle{{ID: "sla-1", VideoID: 1, Language: "zh-CN", Content: "[]"}}
			}
			return nil
		},
	}
	svc := NewService(NewRepository(store))
	list, err := svc.ListAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestService_ListByVideoForScan(t *testing.T) {
	store := &flexMockStore{
		queryFn: func(_ context.Context, _ abstraction.QueryFilter, result any) error {
			if r, ok := result.(*[]*Subtitle); ok {
				*r = []*Subtitle{{ID: "scan-1", VideoID: 5, Language: "zh-CN", Content: "[]"}}
			}
			return nil
		},
	}
	svc := NewService(NewRepository(store))
	list, err := svc.ListByVideoForScan(context.Background(), 5)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestService_Approve(t *testing.T) {
	store := &flexMockStore{
		findByID: func(_ context.Context, id string, result any) error {
			if r, ok := result.(*Subtitle); ok {
				*r = Subtitle{ID: id, VideoID: 1, Language: "zh-CN", Content: "[]", Status: 0}
			}
			return nil
		},
	}
	svc := NewService(NewRepository(store))
	err := svc.Approve(context.Background(), "approve-1")
	assert.NoError(t, err)
}

func TestService_Reject(t *testing.T) {
	store := &flexMockStore{
		findByID: func(_ context.Context, id string, result any) error {
			if r, ok := result.(*Subtitle); ok {
				*r = Subtitle{ID: id, VideoID: 1, Language: "zh-CN", Content: "[]", Status: 0}
			}
			return nil
		},
	}
	svc := NewService(NewRepository(store))
	err := svc.Reject(context.Background(), "reject-1")
	assert.NoError(t, err)
}

func TestService_Delete(t *testing.T) {
	svc := NewService(NewRepository(&flexMockStore{}))
	err := svc.Delete(context.Background(), "del-1")
	assert.NoError(t, err)
}

func TestService_SetDefault(t *testing.T) {
	store := &flexMockStore{
		queryFn: func(_ context.Context, _ abstraction.QueryFilter, result any) error {
			if r, ok := result.(*[]*Subtitle); ok {
				*r = []*Subtitle{{ID: "svd-1", VideoID: 10, Language: "zh-CN", Content: "[]"}}
			}
			return nil
		},
	}
	svc := NewService(NewRepository(store))
	err := svc.SetDefault(context.Background(), 10, "svd-1")
	assert.NoError(t, err)
}

func TestService_Preview_JSON(t *testing.T) {
	store := &flexMockStore{
		findByID: func(_ context.Context, id string, result any) error {
			if r, ok := result.(*Subtitle); ok {
				*r = Subtitle{ID: id, VideoID: 1, Language: "zh-CN", Content: `[{"index":1,"startTime":0,"endTime":1,"text":"hi"}]`}
			}
			return nil
		},
	}
	svc := NewService(NewRepository(store))
	cues, err := svc.Preview(context.Background(), "p1")
	require.NoError(t, err)
	assert.Len(t, cues, 1)
}

func TestService_Preview_SRT(t *testing.T) {
	store := &flexMockStore{
		findByID: func(_ context.Context, id string, result any) error {
			if r, ok := result.(*Subtitle); ok {
				*r = Subtitle{ID: id, VideoID: 1, Language: "zh-CN", Content: "1\n00:00:01,000 --> 00:00:04,000\nHello\n"}
			}
			return nil
		},
	}
	svc := NewService(NewRepository(store))
	cues, err := svc.Preview(context.Background(), "p2")
	require.NoError(t, err)
	assert.Len(t, cues, 1)
}

func TestService_Preview_Invalid(t *testing.T) {
	store := &flexMockStore{
		findByID: func(_ context.Context, id string, result any) error {
			if r, ok := result.(*Subtitle); ok {
				*r = Subtitle{ID: id, VideoID: 1, Language: "zh-CN", Content: "garbage"}
			}
			return nil
		},
	}
	svc := NewService(NewRepository(store))
	cues, err := svc.Preview(context.Background(), "p3")
	require.NoError(t, err)
	assert.Empty(t, cues)
}

// ==================== SetGenerator / Register ====================

func TestSetGenerator(t *testing.T) {
	h := newFlexHandler(&flexMockStore{})
	gen := &WhisperGenerator{}
	ret := h.SetGenerator(gen)
	assert.Equal(t, h, ret)
	assert.NotNil(t, h.gen)
}

func TestRegister(t *testing.T) {
	h := newFlexHandler(&flexMockStore{})
	mux := http.NewServeMux()
	h.Register(mux)
	// 仅验证不 panic
	assert.NotNil(t, mux)
}