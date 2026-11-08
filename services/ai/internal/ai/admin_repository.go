package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// ===== 管理后台 AI 助手：查询/渲染工具 =====

// GetManuscriptByID 稿件详情。
func (r *Repository) GetManuscriptByID(ctx context.Context, id int64) (map[string]any, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, title, description, user_id, category_id, view_count, like_count, coin_count,
		        collect_count, comment_count, status, review_status, upload_time
		 FROM manuscripts WHERE id=$1`, id)
	var (
		title, desc    string
		userID, catID  int64
		view, like     int64
		coin, collect  int64
		comments       int64
		status         int64
		reviewStatus   int64
		upload         time.Time
	)
	if err := row.Scan(&id, &title, &desc, &userID, &catID, &view, &like, &coin, &collect, &comments, &status, &reviewStatus, &upload); err != nil {
		return nil, err
	}
	return map[string]any{"ID": id, "标题": title, "描述": desc, "用户ID": userID, "分类ID": catID,
		"播放": view, "点赞": like, "投币": coin, "收藏": collect, "评论": comments,
		"状态": status, "审核状态": reviewStatus, "上传时间": upload.Format("2006-01-02 15:04")}, nil
}

// GetUserByID 用户详情。
func (r *Repository) GetUserByID(ctx context.Context, id int64) (map[string]any, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, username, nickname, email, level, status, created_at FROM users WHERE id=$1`, id)
	var (
		username, nickname, email string
		level, status              int64
		created                    time.Time
	)
	if err := row.Scan(&id, &username, &nickname, &email, &level, &status, &created); err != nil {
		return nil, err
	}
	return map[string]any{"ID": id, "用户名": username, "昵称": nickname, "邮箱": email,
		"等级": level, "状态": status, "注册时间": created.Format("2006-01-02")}, nil
}

// ListComments 评论列表。
func (r *Repository) ListComments(ctx context.Context, q string, limit int) ([]map[string]any, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	sqlStr := `SELECT id, article_id, user_id, content, like_count, status, created_at FROM article_comments`
	args := []any{}
	if q != "" {
		sqlStr += ` WHERE content ILIKE '%'||$1||'%'`
		args = append(args, q)
	}
	sqlStr += fmt.Sprintf(` ORDER BY id DESC LIMIT %d`, limit)
	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var (
			id, artID, userID, likeCount, status int64
			content                              string
			created                              time.Time
		)
		if err := rows.Scan(&id, &artID, &userID, &content, &likeCount, &status, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"ID": id, "稿件ID": artID, "用户ID": userID, "内容": content,
			"点赞": likeCount, "状态": status, "时间": created.Format("2006-01-02 15:04")})
	}
	return out, rows.Err()
}

// ListReviews 内容审核列表。
func (r *Repository) ListReviews(ctx context.Context, status string, limit int) ([]map[string]any, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	sqlStr := `SELECT id, type, user_id, content, status, reviewed_at, created_at FROM content_reviews`
	args := []any{}
	if status != "" {
		sqlStr += ` WHERE status=$1`
		args = append(args, status)
	}
	sqlStr += fmt.Sprintf(` ORDER BY id DESC LIMIT %d`, limit)
	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var (
			id, userID        int64
			typ, content, st  string
			reviewed, created *time.Time
		)
		if err := rows.Scan(&id, &typ, &userID, &content, &st, &reviewed, &created); err != nil {
			return nil, err
		}
		rt, ct := "", ""
		if reviewed != nil {
			rt = reviewed.Format("2006-01-02 15:04")
		}
		if created != nil {
			ct = created.Format("2006-01-02 15:04")
		}
		out = append(out, map[string]any{"ID": id, "类型": typ, "用户ID": userID, "内容": content,
			"状态": st, "审核时间": rt, "创建时间": ct})
	}
	return out, rows.Err()
}

// ListAuditLogs 审计日志列表。
func (r *Repository) ListAuditLogs(ctx context.Context, limit int) ([]map[string]any, error) {
	if limit <= 0 || limit > 200 {
		limit = 30
	}
	rows, err := r.db.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, operator_name, operator_role, module, action, target_type, target_id, created_at
		 FROM audit_logs ORDER BY id DESC LIMIT %d`, limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var (
			id                                      int64
			operatorName, role, module, action      string
			targetType, targetID                    string
			created                                 time.Time
		)
		if err := rows.Scan(&id, &operatorName, &role, &module, &action, &targetType, &targetID, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"ID": id, "操作人": operatorName, "角色": role, "模块": module,
			"动作": action, "目标类型": targetType, "目标ID": targetID, "时间": created.Format("2006-01-02 15:04")})
	}
	return out, rows.Err()
}

// ListAIConfigs AI 配置列表。
func (r *Repository) ListAIConfigs(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, base_url, model, max_tokens, temperature, enabled FROM ai_api_configs ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var (
			id, maxTokens, enabled int64
			name, baseURL, model   string
			temp                   float64
		)
		if err := rows.Scan(&id, &name, &baseURL, &model, &maxTokens, &temp, &enabled); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"ID": id, "名称": name, "地址": baseURL, "模型": model,
			"最大token": maxTokens, "温度": temp, "启用": enabled})
	}
	return out, rows.Err()
}

// ListAISkills AI 技能列表。
func (r *Repository) ListAISkills(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, description, type, enabled, created_at FROM ai_skills ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var (
			id, enabled          int64
			name, desc, typ      string
			created              time.Time
		)
		if err := rows.Scan(&id, &name, &desc, &typ, &enabled, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"ID": id, "名称": name, "描述": desc, "类型": typ,
			"启用": enabled, "创建": created.Format("2006-01-02")})
	}
	return out, rows.Err()
}

// GetManuscriptMetrics 稿件日指标（折线/柱状图）。
func (r *Repository) GetManuscriptMetrics(ctx context.Context, id int64, days int) (map[string]any, error) {
	if days <= 0 || days > 90 {
		days = 14
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT metric_date, view_count, like_count, coin_count FROM manuscript_daily_metrics
		 WHERE manuscript_id=$1 AND metric_date >= CURRENT_DATE - ($2 || ' days')::interval
		 ORDER BY metric_date LIMIT $2`,
		id, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var categories []string
	var views, likes, coins []int64
	for rows.Next() {
		var d string
		var v, l, c int64
		if err := rows.Scan(&d, &v, &l, &c); err != nil {
			return nil, err
		}
		categories = append(categories, d)
		views = append(views, v)
		likes = append(likes, l)
		coins = append(coins, c)
	}
	if len(categories) == 0 {
		return map[string]any{"categories": []string{}, "series": []map[string]any{}}, nil
	}
	return map[string]any{
		"categories": categories,
		"series": []map[string]any{
			{"name": "播放", "data": views},
			{"name": "点赞", "data": likes},
			{"name": "投币", "data": coins},
		},
	}, nil
}

// GetUserGrowth 用户增长（折线图）。
func (r *Repository) GetUserGrowth(ctx context.Context, days int) (map[string]any, error) {
	if days <= 0 || days > 90 {
		days = 14
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT to_char(d, 'MM-DD') AS day, COALESCE(cnt,0) FROM
		 generate_series(CURRENT_DATE - ($1-1), CURRENT_DATE, '1 day') d
		 LEFT JOIN (SELECT date_trunc('day', created_at)::date AS dd, COUNT(*) cnt
		   FROM users WHERE created_at >= CURRENT_DATE - ($1-1) GROUP BY dd) u
		 ON u.dd = d::date ORDER BY d`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var categories []string
	var counts []int64
	for rows.Next() {
		var d string
		var c int64
		if err := rows.Scan(&d, &c); err != nil {
			return nil, err
		}
		categories = append(categories, d)
		counts = append(counts, c)
	}
	return map[string]any{
		"categories": categories,
		"series":     []map[string]any{{"name": "新增用户", "data": counts}},
	}, nil
}

// ===== 管理后台 AI 助手：写操作工具 =====

// UpdateManuscriptStatus 稿件状态（0=正常,1=下架等）。
func (r *Repository) UpdateManuscriptStatus(ctx context.Context, id, status int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE manuscripts SET status=$2, updated_at=NOW() WHERE id=$1`, id, status)
	return err
}

// UpdateUserStatus 用户状态（1=正常,0=封禁）。
func (r *Repository) UpdateUserStatus(ctx context.Context, id, status int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET status=$2, updated_at=NOW() WHERE id=$1`, id, status)
	return err
}

// ReplyTicket 工单回复。
func (r *Repository) ReplyTicket(ctx context.Context, ticketNo, reply string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE support_tickets SET admin_reply=$2, status='RESOLVED' WHERE ticket_no=$1`, ticketNo, reply)
	return err
}

// CloseTicket 工单关闭。
func (r *Repository) CloseTicket(ctx context.Context, ticketNo string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE support_tickets SET status='CLOSED' WHERE ticket_no=$1`, ticketNo)
	return err
}

// DeleteComment 删除评论（逻辑/物理：物理删除）。
func (r *Repository) DeleteComment(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM article_comments WHERE id=$1`, id)
	return err
}

// parseIDArg 解析工具输入的整数参数。
func parseIDArg(input string, key string) int64 {
	var m map[string]any
	if err := json.Unmarshal([]byte(input), &m); err == nil {
		switch v := m[key].(type) {
		case float64:
			return int64(v)
		case string:
			n, _ := strconv.ParseInt(v, 10, 64)
			return n
		}
	}
	return 0
}
