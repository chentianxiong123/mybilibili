package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/tools"
)

// CustomerTool 是客服 agent 可调用的工具（见 customer_tools.go）。
type customerTool struct {
	name        string
	description string
	fn          func(ctx context.Context, input string) (string, error)
}

func (t *customerTool) Name() string              { return t.name }
func (t *customerTool) Description() string       { return t.description }
func (t *customerTool) Call(ctx context.Context, input string) (string, error) {
	return t.fn(ctx, input)
}

// CustomerAgent 是 AI 客服的多轮 agent（langchaingo conversational agent）。
// 每个会话独立 memory（多轮上下文），支持工具调用（技能检索/建工单/转人工）。
// 未配置真实 LLM 时 NewCustomerAgent 构造成功但 Chat 返回显式"未配置"错误。
type CustomerAgent struct {
	mu        sync.Mutex
	base      *agents.ConversationalAgent
	executors map[int64]*agents.Executor // sessionID -> executor(含独立 memory)
	// usageLog 统一写入 ai_usage_logs（与调用器一致）；nil 时不记录。
	usageLog func(feature string, userID int64, tokenCount, durationMs int)
}

// NewCustomerAgent 构造多轮客服 agent。
// llm 为 langchaingo 模型（openai.LLM）；tools 为空时仍可纯对话（无工具）。
func NewCustomerAgent(llm llms.Model, customerTools []CustomerTool) *CustomerAgent {
	base := agents.NewConversationalAgent(llm, toLangchainTools(customerTools))
	return &CustomerAgent{
		base:      base,
		executors: make(map[int64]*agents.Executor),
	}
}

func toLangchainTools(customerTools []CustomerTool) []tools.Tool {
	out := make([]tools.Tool, 0, len(customerTools))
	for _, t := range customerTools {
		out = append(out, &customerTool{name: t.Name, description: t.Description, fn: t.Fn})
	}
	return out
}

// executor 返回会话对应的 executor（惰性创建，含独立 window memory）。
func (a *CustomerAgent) executor(sessionID int64) *agents.Executor {
	a.mu.Lock()
	defer a.mu.Unlock()
	if e, ok := a.executors[sessionID]; ok {
		return e
	}
	// 保留最近 10 轮对话上下文。
	mem := memory.NewConversationWindowBuffer(10)
	e := agents.NewExecutor(a.base, agents.WithMemory(mem))
	a.executors[sessionID] = e
	return e
}

// SetUsageLog 注入统一用量日志回调（与调用器共享同一写入，AI 模块统一管理用量）。
func (a *CustomerAgent) SetUsageLog(fn func(feature string, userID int64, tokenCount, durationMs int)) {
	a.usageLog = fn
}

// Chat 处理一条用户消息（多轮：自动带上该会话历史 + 工具调用）。
func (a *CustomerAgent) Chat(ctx context.Context, sessionID int64, content string) (string, error) {
	if a.base == nil {
		return "", fmt.Errorf("ai customer agent not configured")
	}
	start := time.Now()
	out, err := a.executor(sessionID).Call(ctx, map[string]any{"input": content})
	if err != nil {
		return "", fmt.Errorf("customer agent: %w", err)
	}
	if v, ok := out["output"].(string); ok {
		if a.usageLog != nil {
			// 统一用量：feature=customer_service，userID=sessionID，token 按长度估算。
			tokens := len(v) / 4
			if tokens < 1 {
				tokens = 1
			}
			a.usageLog("customer_service", sessionID, tokens, int(time.Since(start).Milliseconds()))
		}
		return v, nil
	}
	return "", fmt.Errorf("customer agent: no output")
}

// Reset 清空某会话记忆（如会话关闭）。
func (a *CustomerAgent) Reset(sessionID int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.executors, sessionID)
}
