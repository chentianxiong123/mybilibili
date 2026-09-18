package work

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewTranscoderClient(t *testing.T) {
	c := NewTranscoderClient("http://localhost:9999")
	if c.baseURL != "http://localhost:9999" {
		t.Errorf("expected baseURL http://localhost:9999, got %s", c.baseURL)
	}
}

func TestTranscoderClient_Transcode_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/transcode" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":200,"data":{"play_urls":{"480p":"http://x/480"},"audio_key":"audio/123","is_vertical":0}}`))
	}))
	defer server.Close()

	c := NewTranscoderClient(server.URL)
	result, err := c.Transcode(context.Background(), TranscodeRequest{
		Bucket:       "mybucket",
		SourceKey:    "source/video.mp4",
		ManuscriptID: 1,
		VideoID:      2,
		Qualities:    []string{"480p"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PlayURLs["480p"] != "http://x/480" {
		t.Errorf("unexpected play_urls: %v", result.PlayURLs)
	}
	if result.AudioKey != "audio/123" {
		t.Errorf("unexpected audio_key: %s", result.AudioKey)
	}
}

func TestTranscoderClient_Transcode_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500}`))
	}))
	defer server.Close()

	c := NewTranscoderClient(server.URL)
	_, err := c.Transcode(context.Background(), TranscodeRequest{Bucket: "b", SourceKey: "k"})
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestTranscoderClient_Transcode_BadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid`))
	}))
	defer server.Close()

	c := NewTranscoderClient(server.URL)
	_, err := c.Transcode(context.Background(), TranscodeRequest{Bucket: "b", SourceKey: "k"})
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestTranscoderClient_Transcode_Unreachable(t *testing.T) {
	c := NewTranscoderClient("http://127.0.0.1:1")
	_, err := c.Transcode(context.Background(), TranscodeRequest{Bucket: "b", SourceKey: "k"})
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestTranscoderClient_Transcode_NilData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":200,"data":null}`))
	}))
	defer server.Close()

	c := NewTranscoderClient(server.URL)
	_, err := c.Transcode(context.Background(), TranscodeRequest{Bucket: "b", SourceKey: "k"})
	if err == nil {
		t.Fatal("expected error for nil data")
	}
}
