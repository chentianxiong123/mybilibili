package work

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAIClient(t *testing.T) {
	c := NewAIClient("http://localhost:9999")
	if c.baseURL != "http://localhost:9999" {
		t.Errorf("expected baseURL http://localhost:9999, got %s", c.baseURL)
	}
}

func TestAIClient_GenerateSubtitle_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/subtitle/generate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]int64
		json.NewDecoder(r.Body).Decode(&body)
		if body["manuscript_id"] != 100 {
			t.Errorf("expected manuscript_id=100, got %d", body["manuscript_id"])
		}
		if body["video_id"] != 200 {
			t.Errorf("expected video_id=200, got %d", body["video_id"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":200}`))
	}))
	defer server.Close()

	c := NewAIClient(server.URL)
	err := c.GenerateSubtitle(context.Background(), 100, 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAIClient_GenerateSummary_200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ai/summary/generate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":200}`))
	}))
	defer server.Close()

	c := NewAIClient(server.URL)
	err := c.GenerateSummary(context.Background(), 10, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAIClient_post_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"code":500}`))
	}))
	defer server.Close()

	c := NewAIClient(server.URL)
	err := c.GenerateSubtitle(context.Background(), 1, 2)
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestAIClient_post_BadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid`))
	}))
	defer server.Close()

	c := NewAIClient(server.URL)
	err := c.GenerateSubtitle(context.Background(), 1, 2)
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestAIClient_post_Unreachable(t *testing.T) {
	c := NewAIClient("http://127.0.0.1:1")
	err := c.GenerateSubtitle(context.Background(), 1, 2)
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
