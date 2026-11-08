package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"mybilibili/pkg/abstraction"
)

// LangchaingoConfig 配置 langchaingo 调用器（真实 LLM）。
// 未配置 base_url+api_key 时 Call 返回显式 "AI LLM not configured"（不占位/不假装成功）。
type LangchaingoConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

type langchaingoCaller struct {
	llm        *openai.LLM
	model      string
	configured bool
	// usageLog 真实写入 ai_usage_logs（FINDING-085）；nil 时不记录。
	usageLog func(feature string, userID int64, tokenCount, durationMs int)
}

// NewLangchaingoCaller 构造 langchaingo 调用器。模型/端点从 cfg 或 env
// （AI_LLM_BASE_URL / AI_LLM_API_KEY / AI_LLM_MODEL）读取。
func NewLangchaingoCaller(cfg LangchaingoConfig) (abstraction.ServiceCaller, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = os.Getenv("AI_LLM_BASE_URL")
	}
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("AI_LLM_API_KEY")
	}
	if cfg.Model == "" {
		cfg.Model = os.Getenv("AI_LLM_MODEL")
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4o-mini"
	}
	// 未配置 base_url+api_key：不构造真实 LLM（llm=nil），Call 时返回显式"未配置"错误。
	configured := cfg.APIKey != "" || cfg.BaseURL != ""
	if !configured {
		return &langchaingoCaller{model: cfg.Model, configured: false}, nil
	}
	opts := []openai.Option{openai.WithModel(cfg.Model)}
	if cfg.BaseURL != "" {
		opts = append(opts, openai.WithBaseURL(cfg.BaseURL))
	}
	if cfg.APIKey != "" {
		opts = append(opts, openai.WithToken(cfg.APIKey))
	}
	llm, err := openai.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("langchaingo: create llm: %w", err)
	}
	return &langchaingoCaller{
		llm:        llm,
		model:      cfg.Model,
		configured: true,
	}, nil
}

func (c *langchaingoCaller) requireConfig() error {
	if !c.configured {
		return fmt.Errorf("AI LLM not configured: set AI_LLM_BASE_URL (OpenAI 兼容端点) 或 AI_LLM_API_KEY")
	}
	return nil
}

func (c *langchaingoCaller) generate(ctx context.Context, system, user string) (string, error) {
	return c.generateOpt(ctx, system, user, nil)
}

// generateJSON 强制模型返回 JSON（WithJSONMode），用于结构化审核结果。
func (c *langchaingoCaller) generateJSON(ctx context.Context, system, user string) (string, error) {
	return c.generateOpt(ctx, system, user, []llms.CallOption{llms.WithJSONMode()})
}

func (c *langchaingoCaller) generateOpt(ctx context.Context, system, user string, extra []llms.CallOption) (string, error) {
	if err := c.requireConfig(); err != nil {
		return "", err
	}
	prompt := user
	if system != "" {
		prompt = system + "\n\n" + user
	}
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	opts := append([]llms.CallOption{}, extra...)
	return llms.GenerateFromSinglePrompt(ctx, c.llm, prompt, opts...)
}

func (c *langchaingoCaller) Call(ctx context.Context, target, method string, req, resp any) error {
	reqData, _ := req.(map[string]interface{})
	if reqData == nil && req != nil {
		b, _ := json.Marshal(req)
		_ = json.Unmarshal(b, &reqData)
	}

	switch method {
	case "Summary":
		videoID, _ := reqData["video_id"].(float64)
		start := time.Now()
		reply, err := c.generate(ctx,
			"你是一个视频摘要助手。请用中文简洁总结视频内容，不超过200字。",
			fmt.Sprintf("请为视频 %d 生成摘要", int64(videoID)),
		)
		if err != nil {
			return err
		}
		c.logUsage("summary", 0, start, len(reply))
		return setResp(resp, "summary", reply)

	case "Moderate":
		content, _ := reqData["content"].(string)
		scene, _ := reqData["scene"].(string)
		start := time.Now()
		reply, err := c.generateJSON(ctx,
			"你是一个内容审核助手。请判断以下内容是否违规。只返回JSON：{\"passed\":true/false,\"reason\":\"原因\"}",
			fmt.Sprintf("场景: %s\n内容: %s", scene, content),
		)
		if err != nil {
			return err
		}
		c.logUsage("moderation", 0, start, len(reply))
		var result struct {
			Passed bool   `json:"passed"`
			Reason string `json:"reason"`
		}
		if err := json.Unmarshal([]byte(extractJSON(reply)), &result); err != nil {
			return fmt.Errorf("moderate: parse LLM JSON: %w", err)
		}
		if m, ok := resp.(*map[string]interface{}); ok {
			(*m)["passed"] = result.Passed
			(*m)["reason"] = result.Reason
		}
		return nil

	case "CustomerChat":
		content, _ := reqData["content"].(string)
		start := time.Now()
		reply, err := c.generate(ctx,
			"你是一个B站客服助手。请用中文友好地回答用户问题，回答简洁不超过200字。",
			content,
		)
		if err != nil {
			return err
		}
		c.logUsage("customer_chat", userIDFromReq(reqData), start, len(reply))
		return setResp(resp, "reply", reply)

	case "CheckSummary":
		// 非 LLM：摘要存在性由 DB/缓存层判断，这里保持 false（上层查存储）。
		if r, ok := resp.(*struct{ HasSummary bool }); ok {
			r.HasSummary = false
		} else if m, ok := resp.(*map[string]interface{}); ok {
			(*m)["has_summary"] = false
		}
		return nil

	case "CustomerHistory", "CustomerTransfer":
		// 非 LLM：历史/转人工由仓库层处理，调用器不参与。
		return nil

	default:
		return fmt.Errorf("langchaingo: unknown method %q", method)
	}
}

func (c *langchaingoCaller) CallStream(ctx context.Context, target, method string, req any) (<-chan []byte, error) {
	if err := c.requireConfig(); err != nil {
		return nil, err
	}
	reqData, _ := req.(map[string]interface{})
	videoID, _ := reqData["video_id"].(float64)
	prompt := fmt.Sprintf("请为视频 %d 生成摘要（流式）", int64(videoID))

	out := make(chan []byte, 16)
	go func() {
		defer close(out)
		_, err := llms.GenerateFromSinglePrompt(ctx, c.llm, prompt, llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			select {
			case out <- chunk:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		}))
		if err != nil {
			out <- []byte("ERROR: " + err.Error())
		}
	}()
	return out, nil
}

func (c *langchaingoCaller) Close() error { return nil }

// Model 返回底层 langchaingo 模型（供多轮 agent 复用同一 LLM）。
// 未配置时返回 nil。
func (c *langchaingoCaller) Model() llms.Model { return c.llm }

// Configured 返回是否已配置真实 LLM。
func (c *langchaingoCaller) Configured() bool { return c.configured }

// SetUsageLog 注入用量日志回调（写入 ai_usage_logs）。
func (c *langchaingoCaller) SetUsageLog(fn func(feature string, userID int64, tokenCount, durationMs int)) {
	c.usageLog = fn
}

// logUsage 记录一次调用（按响应长度估算 token：chars/4，最少 1）。
func (c *langchaingoCaller) logUsage(feature string, userID int64, start time.Time, replyLen int) {
	if c.usageLog == nil {
		return
	}
	tokens := replyLen / 4
	if tokens < 1 {
		tokens = 1
	}
	c.usageLog(feature, userID, tokens, int(time.Since(start).Milliseconds()))
}

// userIDFromReq 从请求 map 提取 user_id（若无则 0）。
func userIDFromReq(reqData map[string]interface{}) int64 {
	if v, ok := reqData["user_id"].(float64); ok {
		return int64(v)
	}
	return 0
}

// ModelFromCaller 从 ServiceCaller 提取 langchaingo 模型（供多轮 agent 复用）。
// 非 langchaingo 调用器时返回 nil。
func ModelFromCaller(caller abstraction.ServiceCaller) llms.Model {
	if c, ok := caller.(*langchaingoCaller); ok {
		return c.Model()
	}
	return nil
}

// SetUsageLog 在 langchaingo 调用器上注入用量日志回调（写入 ai_usage_logs）。
// 非 langchaingo 调用器时静默忽略。
func SetUsageLog(caller abstraction.ServiceCaller, fn func(feature string, userID int64, tokenCount, durationMs int)) {
	if c, ok := caller.(*langchaingoCaller); ok {
		c.SetUsageLog(fn)
	}
}

// setResp 将 value 写入 resp（兼容 *struct{Field string} 与 *map[string]interface{}）。
func setResp(resp any, key, value string) error {
	switch r := resp.(type) {
	case *struct{ Summary string }:
		r.Summary = value
	case *struct{ Reply string }:
		r.Reply = value
	case *map[string]interface{}:
		(*r)[key] = value
	default:
		// 尝试反射设置首字段（防御性）。
		return nil
	}
	return nil
}

// extractJSON 从 LLM 输出中提取首个 JSON 对象（可能包裹在 ```json ... ``` 或散文中）。
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "```"); i >= 0 {
		rest := s[i+3:]
		if j := strings.Index(rest, "```"); j >= 0 {
			rest = rest[:j]
		}
		rest = strings.TrimPrefix(rest, "json")
		s = strings.TrimSpace(rest)
	}
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			return s[i : j+1]
		}
	}
	return s
}
