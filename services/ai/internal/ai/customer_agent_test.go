package ai

import (
	"context"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

// fakeLLM 实现 llms.Model，返回固定的 Final Answer 文本（供 agent 解析）。
type fakeLLM struct{ reply string }

func (f *fakeLLM) GenerateContent(_ context.Context, _ []llms.MessageContent, _ ...llms.CallOption) (*llms.ContentResponse, error) {
	// conversational agent 结束标记为 "AI:"。
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "AI: " + f.reply}}}, nil
}
func (f *fakeLLM) Call(_ context.Context, _ string, _ ...llms.CallOption) (string, error) {
	return f.reply, nil
}

// TestCustomerAgent_Chat 多轮 agent：fake LLM 返回 Final Answer，Chat 应提取回复。
func TestCustomerAgent_Chat(t *testing.T) {
	agent := NewCustomerAgent(&fakeLLM{reply: "您好，请问有什么可以帮您？"}, nil)
	reply, err := agent.Chat(context.Background(), 42, "你好")
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if !strings.Contains(reply, "您好") {
		t.Fatalf("Chat reply = %q, want 含'您好'", reply)
	}
}

// TestCustomerAgent_MultiTurnMemory 多轮：同一会话保留上下文（第二次调用不报错）。
func TestCustomerAgent_MultiTurnMemory(t *testing.T) {
	agent := NewCustomerAgent(&fakeLLM{reply: "这是第2轮回复"}, nil)
	if _, err := agent.Chat(context.Background(), 7, "第一轮"); err != nil {
		t.Fatalf("round1: %v", err)
	}
	reply, err := agent.Chat(context.Background(), 7, "第二轮")
	if err != nil {
		t.Fatalf("round2: %v", err)
	}
	if !strings.Contains(reply, "第2轮") {
		t.Fatalf("round2 reply = %q", reply)
	}
}

// TestCustomerAgent_Reset 清空会话记忆。
func TestCustomerAgent_Reset(t *testing.T) {
	agent := NewCustomerAgent(&fakeLLM{reply: "ok"}, nil)
	agent.Chat(context.Background(), 3, "hi")
	agent.Reset(3)
	if _, ok := agent.executors[3]; ok {
		t.Fatalf("Reset 后会话 3 记忆应被清空")
	}
}

// TestCustomerAgent_UsageLog 统一用量：agent.Chat 后调用 usageLog 回调。
func TestCustomerAgent_UsageLog(t *testing.T) {
	agent := NewCustomerAgent(&fakeLLM{reply: "统一用量回复"}, nil)
	called := false
	var gotFeature string
	var gotUser int64
	agent.SetUsageLog(func(feature string, userID int64, tokenCount, durationMs int) {
		called = true
		gotFeature = feature
		gotUser = userID
	})
	if _, err := agent.Chat(context.Background(), 99, "hi"); err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if !called {
		t.Fatalf("agent 应调用 usageLog（统一用量管理）")
	}
	if gotFeature != "customer_service" || gotUser != 99 {
		t.Fatalf("usageLog feature=%q user=%d, want customer_service/99", gotFeature, gotUser)
	}
}

// TestCustomerTools_LookupSkill 技能检索工具。
func TestCustomerTools_LookupSkill(t *testing.T) {
	env := &CustomerToolEnv{
		LookupSkill: func(_ context.Context, query string) (string, error) {
			return `[{"name":"退款","description":"退款流程说明"}]`, nil
		},
	}
	tools := CustomerTools(env)
	if len(tools) != 1 {
		t.Fatalf("tools = %d, want 1（仅注入 LookupSkill）", len(tools))
	}
	out, err := tools[0].Call(context.Background(), "怎么退款")
	if err != nil {
		t.Fatalf("LookupSkill: %v", err)
	}
	if !strings.Contains(out, "退款") {
		t.Fatalf("LookupSkill out = %q", out)
	}
}

// TestCustomerTools_EmptyEnv 无 env：不产生任何工具（agent 纯对话）。
func TestCustomerTools_EmptyEnv(t *testing.T) {
	if n := len(CustomerTools(nil)); n != 0 {
		t.Fatalf("nil env 应 0 工具，got %d", n)
	}
}
