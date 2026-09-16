//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func doPost(t *testing.T, url string, body interface{}) (*http.Response, apiResponse) {
	t.Helper()
	b, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	var apiResp apiResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)
	return resp, apiResp
}

func doGet(t *testing.T, url string, token string) (*http.Response, apiResponse) {
	t.Helper()
	req, _ := http.NewRequest("GET", url, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	var apiResp apiResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)
	return resp, apiResp
}

func randomUsername() string {
	return fmt.Sprintf("test_%d", time.Now().UnixNano())
}