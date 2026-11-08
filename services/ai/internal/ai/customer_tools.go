package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CustomerToolEnv 注入客服工具所需的业务能力（DB 访问），由 main.go 提供闭包。
type CustomerToolEnv struct {
	// LookupSkill 语义检索客服技能（返回 JSON 数组字符串）。
	LookupSkill func(ctx context.Context, query string) (string, error)
	// CreateTicket 创建人工工单。
	CreateTicket func(ctx context.Context, userID int64, content string) (string, error)
	// TransferHuman 转人工。
	TransferHuman func(ctx context.Context, userID int64) (string, error)
}

// CustomerTool 是客服 agent 可调用的工具（技能检索/工单/转人工等）。
type CustomerTool struct {
	Name        string
	Description string
	Fn          func(ctx context.Context, input string) (string, error)
}

// Call 执行工具（实现 tools.Tool 形状，便于测试/复用）。
func (t CustomerTool) Call(ctx context.Context, input string) (string, error) {
	return t.Fn(ctx, input)
}

// CustomerTools 返回多轮客服 agent 可用的工具集。
// 未注入的 env 能力对应工具返回明确提示（agent 可据以继续对话）。
func CustomerTools(env *CustomerToolEnv) []CustomerTool {
	var out []CustomerTool

	if env != nil && env.LookupSkill != nil {
		out = append(out, CustomerTool{
			Name:        "skill_lookup",
			Description: "语义检索客服技能（知识库）。输入用户问题/关键词，返回匹配的技能说明。",
			Fn: func(ctx context.Context, input string) (string, error) {
				return env.LookupSkill(ctx, input)
			},
		})
	}
	if env != nil && env.CreateTicket != nil {
		out = append(out, CustomerTool{
			Name:        "create_ticket",
			Description: "创建人工客服工单。输入用户问题描述，返回工单号。",
			Fn: func(ctx context.Context, input string) (string, error) {
				userID := extractUserID(input)
				return env.CreateTicket(ctx, userID, input)
			},
		})
	}
	if env != nil && env.TransferHuman != nil {
		out = append(out, CustomerTool{
			Name:        "transfer_human",
			Description: "转人工客服。输入任意，返回转人工结果。",
			Fn: func(ctx context.Context, input string) (string, error) {
				userID := extractUserID(input)
				return env.TransferHuman(ctx, userID)
			},
		})
	}
	return out
}

// extractUserID 从工具输入中尽力提取 user_id（agent 工具输入通常为 JSON）。
func extractUserID(input string) int64 {
	var m map[string]any
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		return 0
	}
	switch v := m["user_id"].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	}
	return 0
}

// defaultTicketNo 生成简单工单号（真实生成在 repo，此处占位提示）。
func defaultTicketNo() string {
	return fmt.Sprintf("TK-%d", time.Now().UnixNano()%1000000)
}
