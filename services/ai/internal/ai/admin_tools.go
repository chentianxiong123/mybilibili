package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// AdminRender 前端渲染描述（匹配 web/admin AdminAiChatPanel.vue 的 render 结构）。
type AdminRender struct {
	ChartType string `json:"chartType"` // "number"|"table"|"line"|"bar"|"pie"
	Title     string `json:"title"`
	Data      any    `json:"data"`
}

// AdminState 是每个 admin 会话的状态（工具写入渲染结果，Chat 读取）。
type AdminState struct {
	mu     sync.Mutex
	render *AdminRender
}

type adminStateKey struct{}

func adminStateFrom(ctx context.Context) *AdminState {
	if s, ok := ctx.Value(adminStateKey{}).(*AdminState); ok {
		return s
	}
	return nil
}

func (s *AdminState) setRender(r *AdminRender) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.render = r
}

func (s *AdminState) getRender() *AdminRender {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.render
}

// AdminToolEnv 注入 admin 助手工具所需的后台能力（统计/CRUD），由 main.go 提供闭包。
type AdminToolEnv struct {
	Statistics      func(ctx context.Context) (map[string]any, error)
	ListManuscripts func(ctx context.Context, q string, limit int) ([]map[string]any, error)
	ListUsers       func(ctx context.Context, q string, limit int) ([]map[string]any, error)
	ListTickets     func(ctx context.Context, limit int) ([]map[string]any, error)
}

// adminToolFn 是带渲染能力的工具函数（写入会话 state.render）。
type adminToolFn func(ctx context.Context, input string) (string, error)

// AdminTools 返回 admin 助手工具集（统计/列表，渲染为 number/table）。
func AdminTools(env *AdminToolEnv) []CustomerTool {
	var out []CustomerTool

	if env != nil && env.Statistics != nil {
		out = append(out, CustomerTool{
			Name:        "admin_statistics",
			Description: "查询管理后台核心统计（稿件数/用户数/工单数等），返回指标卡片数据。",
			Fn: adminToolWrapper(func(ctx context.Context, _ string) (string, *AdminRender, error) {
				data, err := env.Statistics(ctx)
				if err != nil {
					return "", nil, err
				}
				stats := make([]map[string]any, 0)
				for k, v := range data {
					stats = append(stats, map[string]any{"label": k, "value": v})
				}
				b, _ := json.Marshal(stats)
				return string(b), &AdminRender{ChartType: "number", Title: "后台核心统计", Data: stats}, nil
			}),
		})
	}
	if env != nil && env.ListManuscripts != nil {
		out = append(out, CustomerTool{
			Name:        "admin_list_manuscripts",
			Description: "分页/关键词查询稿件列表，返回表格数据。",
			Fn: adminToolWrapper(func(ctx context.Context, input string) (string, *AdminRender, error) {
				q, limit := parseAdminQuery(input)
				rows, err := env.ListManuscripts(ctx, q, limit)
				if err != nil {
					return "", nil, err
				}
				return tableReply("稿件列表", rows)
			}),
		})
	}
	if env != nil && env.ListUsers != nil {
		out = append(out, CustomerTool{
			Name:        "admin_list_users",
			Description: "查询用户列表，返回表格数据。",
			Fn: adminToolWrapper(func(ctx context.Context, input string) (string, *AdminRender, error) {
				q, limit := parseAdminQuery(input)
				rows, err := env.ListUsers(ctx, q, limit)
				if err != nil {
					return "", nil, err
				}
				return tableReply("用户列表", rows)
			}),
		})
	}
	if env != nil && env.ListTickets != nil {
		out = append(out, CustomerTool{
			Name:        "admin_list_tickets",
			Description: "查询客服工单列表，返回表格数据。",
			Fn: adminToolWrapper(func(ctx context.Context, _ string) (string, *AdminRender, error) {
				rows, err := env.ListTickets(ctx, 50)
				if err != nil {
					return "", nil, err
				}
				return tableReply("客服工单", rows)
			}),
		})
	}
	return out
}

// adminToolWrapper 将带渲染的工具函数适配为 CustomerTool.Fn（写入会话 state.render）。
func adminToolWrapper(fn func(ctx context.Context, input string) (string, *AdminRender, error)) func(ctx context.Context, input string) (string, error) {
	return func(ctx context.Context, input string) (string, error) {
		reply, render, err := fn(ctx, input)
		if err != nil {
			return "", err
		}
		if state := adminStateFrom(ctx); state != nil && render != nil {
			state.setRender(render)
		}
		return reply, nil
	}
}

// tableReply 把行数据包装成表格渲染。
func tableReply(title string, rows []map[string]any) (string, *AdminRender, error) {
	if len(rows) == 0 {
		return "查询结果为空", &AdminRender{ChartType: "table", Title: title, Data: map[string]any{"columns": []string{}, "rows": [][]any{}}}, nil
	}
	cols := make([]string, 0)
	for k := range rows[0] {
		cols = append(cols, k)
	}
	tbl := make([][]any, 0, len(rows))
	for _, r := range rows {
		row := make([]any, 0, len(cols))
		for _, c := range cols {
			row = append(row, r[c])
		}
		tbl = append(tbl, row)
	}
	render := &AdminRender{ChartType: "table", Title: title, Data: map[string]any{"columns": cols, "rows": tbl}}
	b, _ := json.Marshal(rows)
	return fmt.Sprintf("查询到 %d 条：%s", len(rows), string(b)), render, nil
}

// parseAdminQuery 解析工具输入的 {query, limit}。
func parseAdminQuery(input string) (string, int) {
	q := ""
	limit := 20
	var m map[string]any
	if err := json.Unmarshal([]byte(input), &m); err == nil {
		if v, ok := m["query"].(string); ok {
			q = v
		}
		if v, ok := m["limit"].(float64); ok && int(v) > 0 {
			limit = int(v)
		}
	}
	return q, limit
}
