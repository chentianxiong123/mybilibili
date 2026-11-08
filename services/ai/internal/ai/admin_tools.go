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

// AdminToolEnv 注入 admin 助手工具所需的后台能力（统计/查询/写操作），由 main.go 提供闭包。
type AdminToolEnv struct {
	Statistics           func(ctx context.Context) (map[string]any, error)
	ListManuscripts      func(ctx context.Context, q string, limit int) ([]map[string]any, error)
	ListUsers            func(ctx context.Context, q string, limit int) ([]map[string]any, error)
	ListTickets          func(ctx context.Context, limit int) ([]map[string]any, error)
	GetManuscript        func(ctx context.Context, id int64) (map[string]any, error)
	GetUser              func(ctx context.Context, id int64) (map[string]any, error)
	ListComments         func(ctx context.Context, q string, limit int) ([]map[string]any, error)
	ListReviews          func(ctx context.Context, status string, limit int) ([]map[string]any, error)
	ListAuditLogs        func(ctx context.Context, limit int) ([]map[string]any, error)
	ListConfigs          func(ctx context.Context) ([]map[string]any, error)
	ListSkills           func(ctx context.Context) ([]map[string]any, error)
	ManuscriptMetrics    func(ctx context.Context, id int64, days int) (map[string]any, error)
	UserGrowth           func(ctx context.Context, days int) (map[string]any, error)
	UpdateManuscriptStatus func(ctx context.Context, id, status int64) error
	UpdateUserStatus     func(ctx context.Context, id, status int64) error
	ReplyTicket          func(ctx context.Context, ticketNo, reply string) error
	CloseTicket          func(ctx context.Context, ticketNo string) error
	DeleteComment        func(ctx context.Context, id int64) error
}

// adminToolFn 是带渲染能力的工具函数（写入会话 state.render）。
type adminToolFn func(ctx context.Context, input string) (string, error)

// AdminTools 返回 admin 助手工具集（统计/查询渲染为 number/table/chart，写操作返回结果文本）。
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
	if env != nil && env.GetManuscript != nil {
		out = append(out, CustomerTool{
			Name:        "admin_get_manuscript",
			Description: "按 ID 查询稿件详情，返回表格数据。",
			Fn: adminToolWrapper(func(ctx context.Context, input string) (string, *AdminRender, error) {
				row, err := env.GetManuscript(ctx, parseIDArg(input, "id"))
				if err != nil {
					return "", nil, err
				}
				return detailReply("稿件详情", row)
			}),
		})
	}
	if env != nil && env.GetUser != nil {
		out = append(out, CustomerTool{
			Name:        "admin_get_user",
			Description: "按 ID 查询用户详情，返回表格数据。",
			Fn: adminToolWrapper(func(ctx context.Context, input string) (string, *AdminRender, error) {
				row, err := env.GetUser(ctx, parseIDArg(input, "id"))
				if err != nil {
					return "", nil, err
				}
				return detailReply("用户详情", row)
			}),
		})
	}
	if env != nil && env.ListComments != nil {
		out = append(out, CustomerTool{
			Name:        "admin_list_comments",
			Description: "查询评论列表，返回表格数据。",
			Fn: adminToolWrapper(func(ctx context.Context, input string) (string, *AdminRender, error) {
				q, limit := parseAdminQuery(input)
				rows, err := env.ListComments(ctx, q, limit)
				if err != nil {
					return "", nil, err
				}
				return tableReply("评论列表", rows)
			}),
		})
	}
	if env != nil && env.ListReviews != nil {
		out = append(out, CustomerTool{
			Name:        "admin_list_reviews",
			Description: "查询内容审核记录，返回表格数据。",
			Fn: adminToolWrapper(func(ctx context.Context, input string) (string, *AdminRender, error) {
				status, _ := parseAdminStatus(input)
				rows, err := env.ListReviews(ctx, status, 30)
				if err != nil {
					return "", nil, err
				}
				return tableReply("内容审核", rows)
			}),
		})
	}
	if env != nil && env.ListAuditLogs != nil {
		out = append(out, CustomerTool{
			Name:        "admin_list_audit_logs",
			Description: "查询后台审计日志，返回表格数据。",
			Fn: adminToolWrapper(func(ctx context.Context, _ string) (string, *AdminRender, error) {
				rows, err := env.ListAuditLogs(ctx, 30)
				if err != nil {
					return "", nil, err
				}
				return tableReply("审计日志", rows)
			}),
		})
	}
	if env != nil && env.ListConfigs != nil {
		out = append(out, CustomerTool{
			Name:        "admin_list_configs",
			Description: "查询 AI API 配置列表，返回表格数据。",
			Fn: adminToolWrapper(func(ctx context.Context, _ string) (string, *AdminRender, error) {
				rows, err := env.ListConfigs(ctx)
				if err != nil {
					return "", nil, err
				}
				return tableReply("AI 配置", rows)
			}),
		})
	}
	if env != nil && env.ListSkills != nil {
		out = append(out, CustomerTool{
			Name:        "admin_list_skills",
			Description: "查询 AI 技能列表，返回表格数据。",
			Fn: adminToolWrapper(func(ctx context.Context, _ string) (string, *AdminRender, error) {
				rows, err := env.ListSkills(ctx)
				if err != nil {
					return "", nil, err
				}
				return tableReply("AI 技能", rows)
			}),
		})
	}
	if env != nil && env.ManuscriptMetrics != nil {
		out = append(out, CustomerTool{
			Name:        "admin_manuscript_metrics",
			Description: "查询稿件日指标趋势（播放/点赞/投币），返回折线/柱状图数据。",
			Fn: adminToolWrapper(func(ctx context.Context, input string) (string, *AdminRender, error) {
				id := parseIDArg(input, "id")
				days := int(parseIDArg(input, "days"))
				if days == 0 {
					days = 14
				}
				data, err := env.ManuscriptMetrics(ctx, id, days)
				if err != nil {
					return "", nil, err
				}
				b, _ := json.Marshal(data)
				return string(b), &AdminRender{ChartType: "line", Title: "稿件日指标", Data: data}, nil
			}),
		})
	}
	if env != nil && env.UserGrowth != nil {
		out = append(out, CustomerTool{
			Name:        "admin_user_growth",
			Description: "查询用户增长趋势，返回折线图数据。",
			Fn: adminToolWrapper(func(ctx context.Context, input string) (string, *AdminRender, error) {
				days := int(parseIDArg(input, "days"))
				if days == 0 {
					days = 14
				}
				data, err := env.UserGrowth(ctx, days)
				if err != nil {
					return "", nil, err
				}
				b, _ := json.Marshal(data)
				return string(b), &AdminRender{ChartType: "line", Title: "用户增长", Data: data}, nil
			}),
		})
	}
	if env != nil && env.UpdateManuscriptStatus != nil {
		out = append(out, CustomerTool{
			Name:        "admin_update_manuscript_status",
			Description: "更新稿件状态（id+status，如 0=正常,1=下架）。",
			Fn: adminToolWrapper(func(ctx context.Context, input string) (string, *AdminRender, error) {
				id := parseIDArg(input, "id")
				status := parseIDArg(input, "status")
				if err := env.UpdateManuscriptStatus(ctx, id, status); err != nil {
					return "", nil, err
				}
				return fmt.Sprintf("已更新稿件 %d 状态为 %d", id, status), nil, nil
			}),
		})
	}
	if env != nil && env.UpdateUserStatus != nil {
		out = append(out, CustomerTool{
			Name:        "admin_update_user_status",
			Description: "更新用户状态（id+status，如 1=正常,0=封禁）。",
			Fn: adminToolWrapper(func(ctx context.Context, input string) (string, *AdminRender, error) {
				id := parseIDArg(input, "id")
				status := parseIDArg(input, "status")
				if err := env.UpdateUserStatus(ctx, id, status); err != nil {
					return "", nil, err
				}
				return fmt.Sprintf("已更新用户 %d 状态为 %d", id, status), nil, nil
			}),
		})
	}
	if env != nil && env.ReplyTicket != nil {
		out = append(out, CustomerTool{
			Name:        "admin_reply_ticket",
			Description: "回复客服工单（ticket_no+reply），标记 RESOLVED。",
			Fn: adminToolWrapper(func(ctx context.Context, input string) (string, *AdminRender, error) {
				no, reply := parseTicketReply(input)
				if err := env.ReplyTicket(ctx, no, reply); err != nil {
					return "", nil, err
				}
				return fmt.Sprintf("已回复工单 %s", no), nil, nil
			}),
		})
	}
	if env != nil && env.CloseTicket != nil {
		out = append(out, CustomerTool{
			Name:        "admin_close_ticket",
			Description: "关闭客服工单（ticket_no）。",
			Fn: adminToolWrapper(func(ctx context.Context, input string) (string, *AdminRender, error) {
				no, _ := parseTicketReply(input)
				if err := env.CloseTicket(ctx, no); err != nil {
					return "", nil, err
				}
				return fmt.Sprintf("已关闭工单 %s", no), nil, nil
			}),
		})
	}
	if env != nil && env.DeleteComment != nil {
		out = append(out, CustomerTool{
			Name:        "admin_delete_comment",
			Description: "删除评论（id）。",
			Fn: adminToolWrapper(func(ctx context.Context, input string) (string, *AdminRender, error) {
				id := parseIDArg(input, "id")
				if err := env.DeleteComment(ctx, id); err != nil {
					return "", nil, err
				}
				return fmt.Sprintf("已删除评论 %d", id), nil, nil
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

// parseAdminStatus 解析 {status}。
func parseAdminStatus(input string) (string, int) {
	var m map[string]any
	if err := json.Unmarshal([]byte(input), &m); err == nil {
		if v, ok := m["status"].(string); ok {
			return v, 0
		}
	}
	return "", 0
}

// parseTicketReply 解析 {ticket_no, reply}。
func parseTicketReply(input string) (string, string) {
	var m map[string]any
	if err := json.Unmarshal([]byte(input), &m); err == nil {
		no, _ := m["ticket_no"].(string)
		reply, _ := m["reply"].(string)
		return no, reply
	}
	return "", ""
}

// detailReply 把单条记录包装成一行表格渲染。
func detailReply(title string, row map[string]any) (string, *AdminRender, error) {
	if row == nil {
		return "未找到", &AdminRender{ChartType: "table", Title: title, Data: map[string]any{"columns": []string{}, "rows": [][]any{}}}, nil
	}
	return tableReply(title, []map[string]any{row})
}
