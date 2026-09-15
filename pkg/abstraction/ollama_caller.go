package abstraction

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type ollamaCaller struct {
	baseURL    string
	httpClient *http.Client
	model      string
}

func newOllamaCaller() *ollamaCaller {
	baseURL := os.Getenv("OLLAMA_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "deepseek-r1:8b"
	}
	return &ollamaCaller{
		baseURL: baseURL,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		model:   model,
	}
}

func (c *ollamaCaller) checkOllama() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", c.baseURL, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

func (c *ollamaCaller) chat(ctx context.Context, system, user string) (string, error) {
	body := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"stream": false,
	}
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("ollama parse: %w: %s", err, string(respBody))
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("ollama: no choices in response: %s", string(respBody))
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

var ollamaFallback = false

func (c *ollamaCaller) Call(ctx context.Context, target, method string, req, resp any) error {
	if !ollamaFallback && !c.checkOllama() {
		ollamaFallback = true
		log.Printf("ollama not reachable at %s, using fallback responses", c.baseURL)
	}
	return c.doCall(ctx, target, method, req, resp)
}

func (c *ollamaCaller) doCall(ctx context.Context, target, method string, req, resp any) error {
	reqData, _ := req.(map[string]interface{})
	if reqData == nil {
		b, _ := json.Marshal(req)
		json.Unmarshal(b, &reqData)
	}

	switch method {
	case "Summary":
		videoID, _ := reqData["video_id"].(float64)
		reply, err := c.chat(ctx,
			"你是一个视频摘要助手。请用中文简洁总结视频内容，不超过200字。",
			fmt.Sprintf("请为视频 %d 生成摘要", int64(videoID)),
		)
		if err != nil || ollamaFallback {
			reply = fmt.Sprintf("AI summary for video %d (placeholder)", int64(videoID))
		}
		if r, ok := resp.(*struct{ Summary string }); ok {
			r.Summary = reply
		} else if m, ok := resp.(*map[string]interface{}); ok {
			(*m)["summary"] = reply
		}
		return nil

	case "CheckSummary":
		if r, ok := resp.(*struct{ HasSummary bool }); ok {
			r.HasSummary = false
		}
		return nil

	case "Moderate":
		content, _ := reqData["content"].(string)
		scene, _ := reqData["scene"].(string)
		reply, err := c.chat(ctx,
			"你是一个内容审核助手。请判断以下内容是否违规。只返回JSON：{\"passed\":true/false,\"reason\":\"原因\"}",
			fmt.Sprintf("场景: %s\n内容: %s", scene, content),
		)
		passed := true
		reason := ""
		if err == nil && !ollamaFallback {
			var result struct {
				Passed bool   `json:"passed"`
				Reason string `json:"reason"`
			}
			reply = extractJSON(reply)
			if json.Unmarshal([]byte(reply), &result) == nil {
				passed = result.Passed
				reason = result.Reason
			}
		}
		if m, ok := resp.(*map[string]interface{}); ok {
			(*m)["passed"] = passed
			(*m)["reason"] = reason
		}
		return nil

	case "CustomerChat":
		content, _ := reqData["content"].(string)
		reply, err := c.chat(ctx,
			"你是一个B站客服助手。请用中文友好地回答用户问题，回答简洁不超过200字。",
			content,
		)
		if err != nil || ollamaFallback {
			reply = "感谢您的咨询，客服助手暂时无法连接AI模型，请稍后再试。"
		}
		if r, ok := resp.(*struct{ Reply string }); ok {
			r.Reply = reply
		} else if m, ok := resp.(*map[string]interface{}); ok {
			(*m)["reply"] = reply
		}
		return nil

	case "CustomerHistory":
		if m, ok := resp.(*[]map[string]interface{}); ok {
			*m = []map[string]interface{}{}
		}
		return nil

	case "CustomerTransfer":
		return nil

	default:
		return fmt.Errorf("ollama caller: unknown method %s", method)
	}
}

func (c *ollamaCaller) CallStream(ctx context.Context, target, method string, req any) (<-chan []byte, error) {
	ch := make(chan []byte, 1)
	go func() {
		defer close(ch)
		reply := "stream not supported in ollama caller"
		ch <- []byte(reply)
	}()
	return ch, nil
}

func (c *ollamaCaller) Close() error { return nil }

func extractJSON(s string) string {
	start := strings.Index(s, "{")
	if start == -1 {
		return s
	}
	end := strings.LastIndex(s, "}")
	if end == -1 || end < start {
		return s
	}
	return s[start : end+1]
}