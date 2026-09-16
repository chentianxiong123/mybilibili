package work

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"mybilibili/pkg/abstraction"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSourceKeyFromURL 验证 sourceKeyFromURL 的核心解析逻辑。
func TestSourceKeyFromURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"raw key", "manuscripts/10/videos/25/source/video.mp4", "manuscripts/10/videos/25/source/video.mp4"},
		{"with /uploads/ prefix", "/uploads/manuscripts/10/videos/25/source/video.mp4", "manuscripts/10/videos/25/source/video.mp4"},
		{"with /uploads/ prefix (no leading slash)", "uploads/manuscripts/x.mp4", "manuscripts/x.mp4"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sourceKeyFromURL(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestProcessTypes_Constants 验证 ProcessType 常量值。
func TestProcessTypes_Constants(t *testing.T) {
	assert.Equal(t, "TRANSCODE", ProcessTypeTranscode)
	assert.Equal(t, "EXTRACT_AUDIO", ProcessTypeExtractAudio)
	assert.Equal(t, "GENERATE_SUBTITLE", ProcessTypeGenerateSub)
	assert.Equal(t, "AI_SUMMARY", ProcessTypeAISummary)
}

func TestProcessModes_Constants(t *testing.T) {
	assert.Equal(t, "AUTO_CHAIN", ProcessModeAutoChain)
	assert.Equal(t, "MANUAL_SINGLE", ProcessModeManualSingle)
}

func TestTopics_Constants(t *testing.T) {
	assert.Equal(t, "video-process-topic", TopicVideoProcess)
	assert.Equal(t, "video-publish-topic", TopicVideoPublish)
	assert.Equal(t, "video-process-progress-topic", TopicVideoProgress)
	assert.Equal(t, "manuscript-index-topic", TopicManuscriptIndex)
}

// 防 unused 警告
var _ = filepath.Join

// --- mock MQ ---

type mockMQ struct {
	published []abstraction.Message
}

func (m *mockMQ) Publish(_ context.Context, _ string, msg abstraction.Message) error {
	m.published = append(m.published, msg)
	return nil
}
func (m *mockMQ) Subscribe(_ context.Context, _, _ string) (<-chan abstraction.Message, error) {
	ch := make(chan abstraction.Message)
	return ch, nil
}
func (m *mockMQ) Ack(_ context.Context, _ string, _ abstraction.Message) error { return nil }
func (m *mockMQ) Nack(_ context.Context, _ string, _ abstraction.Message) error { return nil }
func (m *mockMQ) Enqueue(_ context.Context, _ string, _ abstraction.Message, _ time.Duration) error {
	return nil
}
func (m *mockMQ) Close() error { return nil }

// --- mockMQ with subscribe support ---

type mockSubscribeMQ struct {
	published  []abstraction.Message
	tasks      []abstraction.Message
	subscribed bool
}

func (m *mockSubscribeMQ) Publish(_ context.Context, _ string, msg abstraction.Message) error {
	m.published = append(m.published, msg)
	return nil
}
func (m *mockSubscribeMQ) Subscribe(_ context.Context, _, _ string) (<-chan abstraction.Message, error) {
	m.subscribed = true
	ch := make(chan abstraction.Message, len(m.tasks))
	for _, task := range m.tasks {
		ch <- task
	}
	close(ch)
	return ch, nil
}
func (m *mockSubscribeMQ) Ack(_ context.Context, _ string, _ abstraction.Message) error { return nil }
func (m *mockSubscribeMQ) Nack(_ context.Context, _ string, _ abstraction.Message) error { return nil }
func (m *mockSubscribeMQ) Enqueue(_ context.Context, _ string, _ abstraction.Message, _ time.Duration) error {
	return nil
}
func (m *mockSubscribeMQ) Close() error { return nil }

// --- mock DB (no-op for non-UPDATE queries) ---

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestPipelineTranscode_Success(t *testing.T) {
	// mock transcoder service
	transcodeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/transcode", r.URL.Path)

		result := TranscodeResult{
			PlayURLs:   map[string]string{"1080p": "https://cdn/1080.mp4", "720p": "https://cdn/720.mp4", "480p": "https://cdn/480.mp4"},
			IsVertical: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": result})
	}))
	defer transcodeSrv.Close()

	// mock DB for writePlayURLs UPDATE
	db, dbmock := newMockDB(t)
	dbmock.ExpectExec(`UPDATE videos SET play_url_hd`).WithArgs(
		"https://cdn/1080.mp4", "https://cdn/720.mp4", "https://cdn/480.mp4", int64(1),
	).WillReturnResult(sqlmock.NewResult(0, 1))

	// mock MQ
	mq := &mockMQ{}

	// pool with one node pointing to mock transcoder
	cfgFile := filepath.Join(t.TempDir(), "pool.yaml")
	pool, err := NewTranscoderPool(cfgFile)
	require.NoError(t, err)
	require.NoError(t, pool.Add(Node{Name: "t1", Addr: transcodeSrv.URL, Weight: 1}))

	p := NewPipeline(mq, nil, nil, nil, pool, nil, t.TempDir())
	p.SetDatabase(db)

	task := ProcessMessage{
		ManuscriptID: 10,
		VideoID:      1,
		SourceURL:    "/uploads/manuscripts/10/videos/1/source/video.mp4",
		ProcessType:  ProcessTypeTranscode,
		ProcessMode:  ProcessModeManualSingle,
	}

	p.process(context.Background(), task)

	require.NoError(t, dbmock.ExpectationsWereMet())
	// Should have published progress events: started, transcoding, (maybe direction), transcoding-done, done
	assert.True(t, len(mq.published) >= 3, "expected at least 3 progress events, got %d", len(mq.published))
}

func TestPipelineSubtitle_Success(t *testing.T) {
	// mock AI service
	aiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/subtitle/generate", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"code": 200})
	}))
	defer aiSrv.Close()

	mq := &mockMQ{}
	aiClient := NewAIClient(aiSrv.URL)

	p := NewPipeline(mq, nil, nil, nil, nil, aiClient, t.TempDir())

	task := ProcessMessage{
		ManuscriptID: 10,
		VideoID:      1,
		SourceURL:    "/uploads/manuscripts/10/videos/1/source/video.mp4",
		ProcessType:  ProcessTypeGenerateSub,
		ProcessMode:  ProcessModeManualSingle,
	}

	p.process(context.Background(), task)

	assert.True(t, len(mq.published) >= 2, "expected at least 2 progress events, got %d", len(mq.published))

	// Verify the done event
	var last ProgressEvent
	for _, msg := range mq.published {
		var evt ProgressEvent
		json.Unmarshal(msg.Payload, &evt)
		if evt.Stage == "subtitle" {
			last = evt
		}
	}
	assert.Equal(t, int32(100), last.Progress)
	assert.True(t, last.Done)
}

func TestPipelineSubtitle_NoAIClient(t *testing.T) {
	mq := &mockMQ{}

	p := NewPipeline(mq, nil, nil, nil, nil, nil, t.TempDir())

	task := ProcessMessage{
		ManuscriptID: 10,
		VideoID:      1,
		SourceURL:    "/uploads/manuscripts/10/videos/1/source/video.mp4",
		ProcessType:  ProcessTypeGenerateSub,
		ProcessMode:  ProcessModeManualSingle,
	}

	p.process(context.Background(), task)

	// Should emit failed event
	found := false
	for _, msg := range mq.published {
		var evt ProgressEvent
		json.Unmarshal(msg.Payload, &evt)
		if evt.Stage == "failed" {
			found = true
			assert.Contains(t, evt.Error, "ai client not configured")
		}
	}
	assert.True(t, found, "expected failed event with ai client error")
}

func TestPipelineTranscode_EmptySourceKey(t *testing.T) {
	mq := &mockMQ{}

	cfgFile := filepath.Join(t.TempDir(), "pool.yaml")
	pool, err := NewTranscoderPool(cfgFile)
	require.NoError(t, err)

	p := NewPipeline(mq, nil, nil, nil, pool, nil, t.TempDir())

	task := ProcessMessage{
		ManuscriptID: 10,
		VideoID:      1,
		SourceURL:    "",
		ProcessType:  ProcessTypeTranscode,
		ProcessMode:  ProcessModeManualSingle,
	}

	p.process(context.Background(), task)

	found := false
	for _, msg := range mq.published {
		var evt ProgressEvent
		json.Unmarshal(msg.Payload, &evt)
		if evt.Stage == "failed" {
			found = true
			assert.Contains(t, evt.Error, "empty source key")
		}
	}
	assert.True(t, found, "expected failed event with empty source key")
}

func TestPipelineExtractAudio_EmptySourceKey(t *testing.T) {
	mq := &mockMQ{}

	p := NewPipeline(mq, nil, nil, nil, nil, nil, t.TempDir())

	task := ProcessMessage{
		ManuscriptID: 10,
		VideoID:      1,
		SourceURL:    "",
		ProcessType:  ProcessTypeExtractAudio,
		ProcessMode:  ProcessModeManualSingle,
	}

	p.process(context.Background(), task)

	found := false
	for _, msg := range mq.published {
		var evt ProgressEvent
		json.Unmarshal(msg.Payload, &evt)
		if evt.Stage == "failed" {
			found = true
			assert.Contains(t, evt.Error, "empty source key")
		}
	}
	assert.True(t, found, "expected failed event with empty source key")
}

func TestAIClient_GenerateSummary_Success(t *testing.T) {
	aiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/ai/summary/generate", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"code": 200})
	}))
	defer aiSrv.Close()

	aiClient := NewAIClient(aiSrv.URL)
	err := aiClient.GenerateSummary(context.Background(), 10, 1)
	assert.NoError(t, err)
}

func TestAIClient_GenerateSummary_ErrorStatus(t *testing.T) {
	aiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"code": 500})
	}))
	defer aiSrv.Close()

	aiClient := NewAIClient(aiSrv.URL)
	err := aiClient.GenerateSummary(context.Background(), 10, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status=500")
}

func TestAIClient_GenerateSummary_BadDecode(t *testing.T) {
	aiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer aiSrv.Close()

	aiClient := NewAIClient(aiSrv.URL)
	err := aiClient.GenerateSummary(context.Background(), 10, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decode")
}

func TestPipelineStart_SubscribesAndProcesses(t *testing.T) {
	// mock transcoder
	transcodeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := TranscodeResult{
			PlayURLs:   map[string]string{"1080p": "https://cdn/1080.mp4"},
			IsVertical: -1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": result})
	}))
	defer transcodeSrv.Close()

	db, dbmock := newMockDB(t)
	dbmock.ExpectExec(`UPDATE videos SET play_url_hd`).WillReturnResult(sqlmock.NewResult(0, 1))

	// mockMQ that sends one task then closes channel
	mq := &mockSubscribeMQ{
		tasks: []abstraction.Message{
			{
				Topic: TopicVideoProcess,
				Payload: marshal(ProcessMessage{
					ManuscriptID: 10,
					VideoID:      1,
					SourceURL:    "/uploads/manuscripts/10/videos/1/source/video.mp4",
					ProcessType:  ProcessTypeTranscode,
					ProcessMode:  ProcessModeManualSingle,
				}),
			},
		},
	}

	cfgFile := filepath.Join(t.TempDir(), "pool.yaml")
	pool, err := NewTranscoderPool(cfgFile)
	require.NoError(t, err)
	require.NoError(t, pool.Add(Node{Name: "t1", Addr: transcodeSrv.URL, Weight: 1}))

	p := NewPipeline(mq, nil, nil, nil, pool, nil, t.TempDir())
	p.SetDatabase(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = p.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, mq.subscribed, "expected Subscribe to be called")
}

func TestPipelineAISummary_Success(t *testing.T) {
	aiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/ai/summary/generate", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"code": 200})
	}))
	defer aiSrv.Close()

	mq := &mockMQ{}
	aiClient := NewAIClient(aiSrv.URL)

	p := NewPipeline(mq, nil, nil, nil, nil, aiClient, t.TempDir())

	task := ProcessMessage{
		ManuscriptID: 10,
		VideoID:      1,
		SourceURL:    "/uploads/manuscripts/10/videos/1/source/video.mp4",
		ProcessType:  ProcessTypeAISummary,
		ProcessMode:  ProcessModeManualSingle,
	}

	p.process(context.Background(), task)

	// Should have started, summary-progress, summary-done, done events
	foundSummary := false
	foundDone := false
	for _, msg := range mq.published {
		var evt ProgressEvent
		json.Unmarshal(msg.Payload, &evt)
		if evt.Stage == "summary" && evt.Progress == 100 {
			foundSummary = true
		}
		if evt.Stage == "done" {
			foundDone = true
			assert.Equal(t, int32(100), evt.Progress)
			assert.True(t, evt.Done)
		}
	}
	assert.True(t, foundSummary, "expected summary progress event with 100%")
	assert.True(t, foundDone, "expected done event")
}

func TestPipelineAISummary_NoAIClient(t *testing.T) {
	mq := &mockMQ{}
	p := NewPipeline(mq, nil, nil, nil, nil, nil, t.TempDir())

	task := ProcessMessage{
		ManuscriptID: 10,
		VideoID:      1,
		ProcessType:  ProcessTypeAISummary,
		ProcessMode:  ProcessModeManualSingle,
	}

	p.process(context.Background(), task)

	found := false
	for _, msg := range mq.published {
		var evt ProgressEvent
		json.Unmarshal(msg.Payload, &evt)
		if evt.Stage == "failed" {
			found = true
			assert.Contains(t, evt.Error, "ai client not configured")
		}
	}
	assert.True(t, found, "expected failed event")
}

func TestDoExtractAudio_200(t *testing.T) {
	transcodeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := TranscodeResult{
			AudioKey:   "manuscripts/10/videos/1/audio/audio.mp3",
			PlayURLs:   map[string]string{},
			IsVertical: -1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": result})
	}))
	defer transcodeSrv.Close()

	mq := &mockMQ{}

	cfgFile := filepath.Join(t.TempDir(), "pool.yaml")
	pool, err := NewTranscoderPool(cfgFile)
	require.NoError(t, err)
	require.NoError(t, pool.Add(Node{Name: "t1", Addr: transcodeSrv.URL, Weight: 1}))

	p := NewPipeline(mq, nil, nil, nil, pool, nil, t.TempDir())

	task := ProcessMessage{
		ManuscriptID: 10,
		VideoID:      1,
		SourceURL:    "/uploads/manuscripts/10/videos/1/source/video.mp4",
		ProcessType:  ProcessTypeExtractAudio,
		ProcessMode:  ProcessModeManualSingle,
	}

	p.process(context.Background(), task)

	found := false
	for _, msg := range mq.published {
		var evt ProgressEvent
		json.Unmarshal(msg.Payload, &evt)
		if evt.Stage == "audio" && evt.Progress == 100 {
			found = true
			assert.True(t, evt.Done)
		}
	}
	assert.True(t, found, "expected audio completion event with 100%")
}

func TestDoExtractAudio_EmptyKey(t *testing.T) {
	mq := &mockMQ{}

	p := NewPipeline(mq, nil, nil, nil, nil, nil, t.TempDir())

	task := ProcessMessage{
		ManuscriptID: 10,
		VideoID:      1,
		SourceURL:    "",
		ProcessType:  ProcessTypeExtractAudio,
		ProcessMode:  ProcessModeManualSingle,
	}

	p.process(context.Background(), task)

	found := false
	for _, msg := range mq.published {
		var evt ProgressEvent
		json.Unmarshal(msg.Payload, &evt)
		if evt.Stage == "failed" {
			found = true
			assert.Contains(t, evt.Error, "empty source key")
		}
	}
	assert.True(t, found, "expected failed event with empty source key")
}

func TestDoExtractAudio_Error(t *testing.T) {
	transcodeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"code": 500, "message": "ffmpeg error"})
	}))
	defer transcodeSrv.Close()

	mq := &mockMQ{}

	cfgFile := filepath.Join(t.TempDir(), "pool.yaml")
	pool, err := NewTranscoderPool(cfgFile)
	require.NoError(t, err)
	require.NoError(t, pool.Add(Node{Name: "t1", Addr: transcodeSrv.URL, Weight: 1}))

	p := NewPipeline(mq, nil, nil, nil, pool, nil, t.TempDir())

	task := ProcessMessage{
		ManuscriptID: 10,
		VideoID:      1,
		SourceURL:    "/uploads/manuscripts/10/videos/1/source/video.mp4",
		ProcessType:  ProcessTypeExtractAudio,
		ProcessMode:  ProcessModeManualSingle,
	}

	p.process(context.Background(), task)

	found := false
	for _, msg := range mq.published {
		var evt ProgressEvent
		json.Unmarshal(msg.Payload, &evt)
		if evt.Stage == "failed" && evt.Error != "" {
			found = true
			assert.Contains(t, evt.Error, "status=500")
		}
	}
	assert.True(t, found, "expected failed event for audio extraction error")
}

func TestPipelineAutoChain_Transcode(t *testing.T) {
	// mock transcoder service
	transcodeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := TranscodeResult{
			PlayURLs:   map[string]string{"1080p": "https://cdn/1080.mp4", "720p": "https://cdn/720.mp4", "480p": "https://cdn/480.mp4"},
			IsVertical: -1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": result})
	}))
	defer transcodeSrv.Close()

	db, dbmock := newMockDB(t)
	dbmock.ExpectExec(`UPDATE videos SET play_url_hd`).WillReturnResult(sqlmock.NewResult(0, 1))

	mq := &mockMQ{}

	cfgFile := filepath.Join(t.TempDir(), "pool.yaml")
	pool, err := NewTranscoderPool(cfgFile)
	require.NoError(t, err)
	require.NoError(t, pool.Add(Node{Name: "t1", Addr: transcodeSrv.URL, Weight: 1}))

	p := NewPipeline(mq, nil, nil, nil, pool, nil, t.TempDir())
	p.SetDatabase(db)

	task := ProcessMessage{
		ManuscriptID: 10,
		VideoID:      1,
		SourceURL:    "/uploads/manuscripts/10/videos/1/source/video.mp4",
		ProcessType:  ProcessTypeTranscode,
		ProcessMode:  ProcessModeAutoChain,
	}

	p.process(context.Background(), task)

	require.NoError(t, dbmock.ExpectationsWereMet())

	// Auto-chain should publish a next-step message (EXTRACT_AUDIO) to TopicVideoProcess
	foundChain := false
	for _, msg := range mq.published {
		if msg.Topic == TopicVideoProcess {
			var next ProcessMessage
			json.Unmarshal(msg.Payload, &next)
			if next.ProcessType == ProcessTypeExtractAudio {
				foundChain = true
				assert.Equal(t, ProcessModeAutoChain, next.ProcessMode)
			}
		}
	}
	assert.True(t, foundChain, "expected chain-next message to video-process-topic")
}