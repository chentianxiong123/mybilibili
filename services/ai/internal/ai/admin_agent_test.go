package ai

import (
	"context"
	"testing"
)

// TestAdminAssistant_Chat 多轮 agent 回复（fake LLM 直接终局答案）。
func TestAdminAssistant_Chat(t *testing.T) {
	assistant := NewAdminAssistant(&fakeLLM{reply: "已查询稿件列表"}, nil)
	reply, err := assistant.Chat(context.Background(), 7, "帮我查稿件列表")
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if reply.Reply == "" {
		t.Fatalf("reply 为空")
	}
}

// TestAdminAssistant_Memory 多轮：第二轮仍能工作。
func TestAdminAssistant_Memory(t *testing.T) {
	assistant := NewAdminAssistant(&fakeLLM{reply: "收到"}, nil)
	if _, err := assistant.Chat(context.Background(), 1, "你好"); err != nil {
		t.Fatalf("第1轮: %v", err)
	}
	if _, err := assistant.Chat(context.Background(), 1, "再问一次"); err != nil {
		t.Fatalf("第2轮: %v", err)
	}
}

// TestAdminTools_RenderCapture 工具调用后写入会话渲染状态（表格/指标卡）。
func TestAdminTools_RenderCapture(t *testing.T) {
	tools := AdminTools(&AdminToolEnv{
		ListManuscripts: func(ctx context.Context, q string, limit int) ([]map[string]any, error) {
			return []map[string]any{{"ID": 1, "标题": "测试稿", "播放": 100}}, nil
		},
	})
	if len(tools) == 0 {
		t.Fatalf("工具集为空")
	}
	state := &AdminState{}
	ctx := context.WithValue(context.Background(), adminStateKey{}, state)
	if _, err := tools[0].Fn(ctx, `{"query":"","limit":10}`); err != nil { // 唯一工具=admin_list_manuscripts
		t.Fatalf("tool: %v", err)
	}
	render := state.getRender()
	if render == nil {
		t.Fatalf("工具应写入渲染状态")
	}
	if render.ChartType != "table" {
		t.Fatalf("chartType=%q, want table", render.ChartType)
	}
	data, _ := render.Data.(map[string]any)
	if len(data["columns"].([]string)) != 3 || len(data["rows"].([][]any)) != 1 {
		t.Fatalf("表格列/行错误: %+v", data)
	}
}

// TestAdminTools_StatisticsRender 统计指标卡渲染。
func TestAdminTools_StatisticsRender(t *testing.T) {
	tools := AdminTools(&AdminToolEnv{
		Statistics: func(ctx context.Context) (map[string]any, error) {
			return map[string]any{"稿件总数": 100, "用户总数": 20}, nil
		},
	})
	state := &AdminState{}
	ctx := context.WithValue(context.Background(), adminStateKey{}, state)
	if _, err := tools[0].Fn(ctx, ""); err != nil { // tools[0]=admin_statistics
		t.Fatalf("tool: %v", err)
	}
	render := state.getRender()
	if render == nil || render.ChartType != "number" {
		t.Fatalf("期望 number 渲染，得到 %+v", render)
	}
}

// TestAdminTools_TableReply 表格渲染列/行。
func TestAdminTools_TableReply(t *testing.T) {
	rows := []map[string]any{{"ID": 1, "标题": "A"}, {"ID": 2, "标题": "B"}}
	text, render, err := tableReply("稿件", rows)
	if err != nil {
		t.Fatalf("tableReply: %v", err)
	}
	if text == "" {
		t.Fatalf("text 为空")
	}
	data, _ := render.Data.(map[string]any)
	if len(data["columns"].([]string)) != 2 || len(data["rows"].([][]any)) != 2 {
		t.Fatalf("表格列/行错误: %+v", data)
	}
}

// TestAdminTools_EmptyEnv 空 env 时工具集为空（不崩）。
func TestAdminTools_EmptyEnv(t *testing.T) {
	tools := AdminTools(nil)
	if len(tools) != 0 {
		t.Fatalf("空 env 应返回空工具集, got %d", len(tools))
	}
}
