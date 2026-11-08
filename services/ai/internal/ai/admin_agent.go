package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
)

// AdminReply 是 admin 助手的结构化回复（reply 文本 + render 渲染块）。
type AdminReply struct {
	Reply  string       `json:"reply"`
	Render *AdminRender `json:"render,omitempty"`
}

// AdminAssistant 是管理后台 AI 助手（langchaingo 多轮工具调用 agent）。
// 每个 admin 会话独立 memory + 渲染状态；工具可产出表格/指标卡（AdminRender）。
// 未配置真实 LLM 时 NewAdminAssistant 构造成功但 Chat 返回显式"未配置"错误。
type AdminAssistant struct {
	mu        sync.Mutex
	base      *agents.ConversationalAgent
	executors map[int64]*agents.Executor // adminID -> executor(含独立 memory)
	states    map[int64]*AdminState      // adminID -> 渲染状态
	usageLog  func(feature string, userID int64, tokenCount, durationMs int)
}

// NewAdminAssistant 构造 admin 助手。tools 为空时仍可纯对话。
func NewAdminAssistant(llm llms.Model, tools []CustomerTool) *AdminAssistant {
	return &AdminAssistant{
		base:      agents.NewConversationalAgent(llm, toLangchainTools(tools)),
		executors: make(map[int64]*agents.Executor),
		states:    make(map[int64]*AdminState),
	}
}

// SetUsageLog 注入统一用量日志回调。
func (a *AdminAssistant) SetUsageLog(fn func(feature string, userID int64, tokenCount, durationMs int)) {
	a.usageLog = fn
}

func (a *AdminAssistant) executor(adminID int64) (*agents.Executor, *AdminState) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e, ok := a.executors[adminID]
	if !ok {
		mem := memory.NewConversationWindowBuffer(10)
		e = agents.NewExecutor(a.base, agents.WithMemory(mem))
		a.executors[adminID] = e
	}
	st, ok := a.states[adminID]
	if !ok {
		st = &AdminState{}
		a.states[adminID] = st
	}
	return e, st
}

// Chat 处理一条 admin 消息，返回结构化回复（reply + render 表格/指标卡）。
func (a *AdminAssistant) Chat(ctx context.Context, adminID int64, content string) (*AdminReply, error) {
	if a.base == nil {
		return nil, fmt.Errorf("ai admin assistant not configured")
	}
	e, st := a.executor(adminID)
	// 在 ctx 携带会话状态，供工具写入渲染结果。
	ctx = context.WithValue(ctx, adminStateKey{}, st)
	start := time.Now()
	out, err := e.Call(ctx, map[string]any{"input": content})
	if err != nil {
		return nil, fmt.Errorf("admin assistant: %w", err)
	}
	reply := ""
	if v, ok := out["output"].(string); ok {
		reply = v
	} else {
		return nil, fmt.Errorf("admin assistant: no output")
	}
	if a.usageLog != nil {
		tokens := len(reply) / 4
		if tokens < 1 {
			tokens = 1
		}
		a.usageLog("admin_assistant", adminID, tokens, int(time.Since(start).Milliseconds()))
	}
	return &AdminReply{Reply: reply, Render: st.getRender()}, nil
}

// Reset 清空会话状态。
func (a *AdminAssistant) Reset(adminID int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.executors, adminID)
	delete(a.states, adminID)
}
