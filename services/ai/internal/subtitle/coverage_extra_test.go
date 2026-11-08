package subtitle

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/pkg/abstraction"
)

// ---------------------------------------------------------------------------
// 辅助 mock：带 error 注入的 DocumentStore / StorageService
// ---------------------------------------------------------------------------

// errStore 允许按方法注入错误，用于覆盖 Repository/Service 的错误分支。
type errStore struct {
	insertErr error
	findErr   error
	updateErr error
	deleteErr error
	queryErr  error

	insertID string
}

func (m *errStore) Insert(ctx context.Context, collection string, doc any) (string, error) {
	if m.insertErr != nil {
		return "", m.insertErr
	}
	if m.insertID != "" {
		return m.insertID, nil
	}
	return "err-store-id", nil
}

func (m *errStore) FindByID(ctx context.Context, collection, id string, result any) error {
	if m.findErr != nil {
		return m.findErr
	}
	if r, ok := result.(*Subtitle); ok {
		*r = Subtitle{ID: id, VideoID: 1, Language: "zh-CN", Content: "[]"}
	}
	return nil
}

func (m *errStore) Update(ctx context.Context, collection, id string, doc any) error { return m.updateErr }
func (m *errStore) Delete(ctx context.Context, collection, id string) error           { return m.deleteErr }
func (m *errStore) Query(ctx context.Context, collection string, filter abstraction.QueryFilter, result any) error {
	if m.queryErr != nil {
		return m.queryErr
	}
	if r, ok := result.(*[]*Subtitle); ok {
		*r = []*Subtitle{}
	}
	return nil
}

// errReadCloser 在 Read 时返回错误，用于覆盖 GenerateFromAudio 的读音频错误分支。
type errReadCloser struct{ err error }

func (e errReadCloser) Read([]byte) (int, error) { return 0, e.err }
func (e errReadCloser) Close() error             { return nil }

// failReadStorage 返回一个读时出错的对象。
type failReadStorage struct{}

func (f failReadStorage) Put(ctx context.Context, bucket, key string, body io.Reader, contentType string) error {
	return nil
}
func (f failReadStorage) Get(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return errReadCloser{err: errors.New("read boom")}, nil
}
func (f failReadStorage) Delete(ctx context.Context, bucket, key string) error { return nil }
func (f failReadStorage) Head(ctx context.Context, bucket, key string) (*abstraction.FileInfo, error) {
	return nil, nil
}
func (f failReadStorage) List(ctx context.Context, bucket, prefix string) ([]abstraction.FileInfo, error) {
	return nil, nil
}
func (f failReadStorage) SignedURL(ctx context.Context, bucket, key string, expire time.Duration) (string, error) {
	return "", nil
}

// roundTripFunc 使任意函数满足 http.RoundTripper，可拦截 Cloudflare 的硬编码 URL。
type roundTripFunc func(r *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

// ===========================================================================
// generator.go
// ===========================================================================

func TestGenerateFromAudio_ReadError(t *testing.T) {
	gen := NewWhisperGenerator(&Repository{store: &errStore{}}, failReadStorage{})
	_, _, err := gen.GenerateFromAudio(context.Background(), 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read audio")
}

func TestGenerateFromAudio_WhisperAPIError(t *testing.T) {
	os.Setenv("WHISPER_API_URL", "http://127.0.0.1:1")
	defer os.Unsetenv("WHISPER_API_URL")

	gen := NewWhisperGenerator(&Repository{store: &errStore{}}, &mockStorage{audio: []byte("audio")})
	_, _, err := gen.GenerateFromAudio(context.Background(), 1, 1)
	require.Error(t, err)
}

func TestGenerateFromAudio_SaveError(t *testing.T) {
	os.Unsetenv("WHISPER_API_URL")
	os.Unsetenv("CLOUDFLARE_AI_ACCOUNT_ID")

	store := &errStore{insertErr: errors.New("db down")}
	gen := NewWhisperGenerator(&Repository{store: store}, &mockStorage{audio: []byte("audio")})
	_, _, err := gen.GenerateFromAudio(context.Background(), 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save")
}

func TestCallWhisperOpenAI_InvalidURL(t *testing.T) {
	_, err := callWhisperOpenAI(context.Background(), []byte("audio"), &http.Client{}, "://bad")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create request")
}

func TestCallWhisperCloudflare_SuccessWithSegments(t *testing.T) {
	httpc := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Host, "api.cloudflare.com")
			assert.Contains(t, r.Header.Get("Authorization"), "Bearer tok")
			assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
			return jsonResponse(200, `{"success":true,"result":{"text":"hello","segments":[
				{"id":0,"start":0,"end":1.5,"text":"  first  "},
				{"id":1,"start":1.5,"end":3.0,"text":"second"}
			]}}`), nil
		}),
	}
	cues, err := callWhisperCloudflare(context.Background(), []byte("audio"), httpc, "acct", "tok")
	require.NoError(t, err)
	require.Len(t, cues, 2)
	assert.Equal(t, 1, cues[0]["index"])
	assert.Equal(t, "first", cues[0]["text"])
	assert.Equal(t, 2, cues[1]["index"])
	assert.Equal(t, "second", cues[1]["text"])
}

func TestCallWhisperCloudflare_EmptySegments(t *testing.T) {
	httpc := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return jsonResponse(200, `{"success":true,"result":{"text":"hi","segments":[]}}`), nil
		}),
	}
	cues, err := callWhisperCloudflare(context.Background(), []byte("audio"), httpc, "acct", "tok")
	require.NoError(t, err)
	assert.Empty(t, cues)
}

func TestCallWhisperCloudflare_Non200_WithTransport(t *testing.T) {
	httpc := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return jsonResponse(500, "boom"), nil
		}),
	}
	_, err := callWhisperCloudflare(context.Background(), []byte("audio"), httpc, "acct", "tok")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestCallWhisperCloudflare_BadJSON_WithTransport(t *testing.T) {
	httpc := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return jsonResponse(200, "not json"), nil
		}),
	}
	_, err := callWhisperCloudflare(context.Background(), []byte("audio"), httpc, "acct", "tok")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse response")
}

func TestCallWhisperCloudflare_SuccessFalse_WithTransport(t *testing.T) {
	httpc := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return jsonResponse(200, `{"success":false,"result":{}}`), nil
		}),
	}
	_, err := callWhisperCloudflare(context.Background(), []byte("audio"), httpc, "acct", "tok")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "success=false")
}

func TestCallWhisperCloudflare_TransportError(t *testing.T) {
	httpc := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return nil, errors.New("dial fail")
		}),
	}
	_, err := callWhisperCloudflare(context.Background(), []byte("audio"), httpc, "acct", "tok")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api request")
}

func TestCallWhisperCloudflare_InvalidAccountID(t *testing.T) {
	httpc := &http.Client{}
	_, err := callWhisperCloudflare(context.Background(), []byte("audio"), httpc, "bad\nacct", "tok")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create request")
}

// ===========================================================================
// srt_parser.go
// ===========================================================================

func TestParseTimeRange_StartError(t *testing.T) {
	_, _, err := parseTimeRange("bad --> 00:00:01,000")
	require.Error(t, err)
}

func TestParseTimeRange_EndError(t *testing.T) {
	_, _, err := parseTimeRange("00:00:01,000 --> bad")
	require.Error(t, err)
}

// ===========================================================================
// subtitle_handler.go
// ===========================================================================

func TestHandleGenerate_GeneratorError(t *testing.T) {
	store := &errStore{}
	h := NewHandler(NewService(NewRepository(store)))
	h.SetGenerator(NewWhisperGenerator(NewRepository(store), &failingStorage{err: errors.New("storage down")}))
	rr := doSubtitleAsUser(t, h, http.MethodPost, "/api/v1/subtitle/generate", `{"manuscript_id":1,"video_id":99}`, 42)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandleUpload_Anonymous401(t *testing.T) {
	h := newFlexHandler(&flexMockStore{})
	// 不带 X-User-Id → RequireUser 拒绝
	rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/upload", `{"video_id":10,"content":"test"}`)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandleUpload_UploadError(t *testing.T) {
	h := NewHandler(NewService(NewRepository(&errStore{insertErr: errors.New("db fail")})))
	rr := doSubtitleAsUser(t, h, http.MethodPost, "/api/v1/subtitle/upload", `{"video_id":10,"content":"test"}`, 7)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandleUploadSRT_Anonymous401(t *testing.T) {
	h := newFlexHandler(&flexMockStore{})
	boundary := "----TestBoundary"
	var body strings.Builder
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"video_id\"\r\n\r\n10\r\n")
	body.WriteString("--" + boundary + "--\r\n")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/subtitle/upload-srt", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.Register(mux)
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestHandleUploadSRT_ParseFormError(t *testing.T) {
	h := newFlexHandler(&flexMockStore{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/subtitle/upload-srt", strings.NewReader("not a multipart body"))
	req.Header.Set("Content-Type", "multipart/form-data") // 缺 boundary → ParseMultipartForm 报错
	req.Header.Set("X-User-Id", "4")
	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.Register(mux)
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleUploadSRT_UploadError(t *testing.T) {
	h := NewHandler(NewService(NewRepository(&errStore{insertErr: errors.New("db fail")})))
	srt := "1\n00:00:01,000 --> 00:00:04,000\nHello\n"
	rr := doMultipartAsUser(t, h, "/api/v1/subtitle/upload-srt",
		map[string]string{"video_id": "10"}, "file", "test.srt", srt, 4)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandleAllVideos_ListError(t *testing.T) {
	h := NewHandler(NewService(NewRepository(&errStore{queryErr: errors.New("db fail")})))
	rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/videos", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 200, resp.Code)
}

func TestHandleScan_ListError(t *testing.T) {
	h := NewHandler(NewService(NewRepository(&errStore{queryErr: errors.New("db fail")})))
	rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/scan/55", "")
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleImport_UploadError(t *testing.T) {
	h := NewHandler(NewService(NewRepository(&errStore{insertErr: errors.New("db fail")})))
	srt := "1\n00:00:01,000 --> 00:00:04,000\nHello\n"
	body, _ := json.Marshal(map[string]interface{}{"video_id": 10, "srt": srt})
	rr := doSubtitleAsUser(t, h, http.MethodPost, "/api/v1/subtitle/import-srt", string(body), 42)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandleImportSystem_UploadError(t *testing.T) {
	h := NewHandler(NewService(NewRepository(&errStore{insertErr: errors.New("db fail")})))
	srt := "1\n00:00:01,000 --> 00:00:04,000\nHello\n"
	body, _ := json.Marshal(map[string]interface{}{"video_id": 10, "srt": srt})
	rr := doSubtitle(t, h, http.MethodPost, "/api/v1/subtitle/import-system", string(body))
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandleSubtitleByID_GetNotFound(t *testing.T) {
	h := NewHandler(NewService(NewRepository(&errStore{findErr: errors.New("not found")})))
	rr := doSubtitle(t, h, http.MethodGet, "/api/v1/subtitle/missing", "")
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ===========================================================================
// subtitle_repository.go / Service 错误分支
// ===========================================================================

func TestRepository_GetByID_Error(t *testing.T) {
	repo := NewRepository(&errStore{findErr: errors.New("boom")})
	_, err := repo.GetByID(context.Background(), "x")
	require.Error(t, err)
}

func TestRepository_ListByVideo_Error(t *testing.T) {
	repo := NewRepository(&errStore{queryErr: errors.New("boom")})
	_, err := repo.ListByVideo(context.Background(), 1)
	require.Error(t, err)
}

func TestRepository_ListPending_Error(t *testing.T) {
	repo := NewRepository(&errStore{queryErr: errors.New("boom")})
	_, err := repo.ListPending(context.Background())
	require.Error(t, err)
}

func TestRepository_ListAll_Error(t *testing.T) {
	repo := NewRepository(&errStore{queryErr: errors.New("boom")})
	_, err := repo.ListAll(context.Background())
	require.Error(t, err)
}

func TestRepository_GetByLanguage_ListError(t *testing.T) {
	repo := NewRepository(&errStore{queryErr: errors.New("boom")})
	_, err := repo.GetByLanguage(context.Background(), 1, "zh-CN")
	require.Error(t, err)
}

func TestRepository_SetDefault_ListError(t *testing.T) {
	repo := NewRepository(&errStore{queryErr: errors.New("boom")})
	err := repo.SetDefault(context.Background(), 1, "sub-1")
	require.Error(t, err)
}

func TestRepository_SetDefault_UpdateError(t *testing.T) {
	// 用 queryFn 风格：直接构建一个返回有数据的 store，再注入 updateErr。
	flex := &flexMockStore{
		queryFn: func(_ context.Context, _ abstraction.QueryFilter, result any) error {
			if r, ok := result.(*[]*Subtitle); ok {
				*r = []*Subtitle{{ID: "sd-1", VideoID: 10, Language: "zh-CN", Content: "[]"}}
			}
			return nil
		},
		updateErr: errors.New("boom"),
	}
	repo := NewRepository(flex)
	err := repo.SetDefault(context.Background(), 10, "sd-1")
	require.Error(t, err)
}

func TestService_Approve_GetError(t *testing.T) {
	svc := NewService(NewRepository(&errStore{findErr: errors.New("boom")}))
	err := svc.Approve(context.Background(), "x")
	require.Error(t, err)
}

func TestService_Reject_GetError(t *testing.T) {
	svc := NewService(NewRepository(&errStore{findErr: errors.New("boom")}))
	err := svc.Reject(context.Background(), "x")
	require.Error(t, err)
}

func TestService_Preview_GetError(t *testing.T) {
	svc := NewService(NewRepository(&errStore{findErr: errors.New("boom")}))
	_, err := svc.Preview(context.Background(), "x")
	require.Error(t, err)
}

func TestService_SetDefault_Error(t *testing.T) {
	svc := NewService(NewRepository(&errStore{queryErr: errors.New("boom")}))
	err := svc.SetDefault(context.Background(), 1, "sub-1")
	require.Error(t, err)
}

// ===========================================================================
// Repository.Create 边界
// ===========================================================================

func TestRepository_Create_SetsDefaults(t *testing.T) {
	store := &errStore{insertID: "created-1"}
	repo := NewRepository(store)
	sub := &Subtitle{VideoID: 5, Language: "en", Content: "[]", Format: "", Status: 0}
	id, err := repo.Create(context.Background(), sub)
	require.NoError(t, err)
	assert.Equal(t, "created-1", id)
	assert.Equal(t, "srt", sub.Format)
	assert.False(t, sub.UploadTime.IsZero())
}

// ===========================================================================
// 并发安全冒烟（race detector 目标）
// ===========================================================================

func TestParseSRT_Concurrent(t *testing.T) {
	content := "1\n00:00:01,000 --> 00:00:04,000\nHello\n\n2\n00:00:05,000 --> 00:00:08,500\nWorld\n"
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cues, err := ParseSRT(content)
			if err != nil || len(cues) != 2 {
				t.Errorf("unexpected parse result: %v %v", cues, err)
			}
		}()
	}
	wg.Wait()
}

func TestSubtitleToMap_Concurrent(t *testing.T) {
	sub := &Subtitle{ID: "c1", VideoID: 1, Language: "zh-CN", Content: `[{"index":1,"startTime":0,"endTime":1,"text":"x"}]`}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m := subtitleToMap(sub)
			if m["id"] != "c1" {
				t.Errorf("unexpected map: %v", m)
			}
		}()
	}
	wg.Wait()
}
