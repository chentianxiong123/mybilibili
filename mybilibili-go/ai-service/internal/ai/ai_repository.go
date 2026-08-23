package ai

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"time"
)

type ApiConfig struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	BaseURL     string    `json:"base_url"`
	APIKey      string    `json:"api_key"`
	Model       string    `json:"model"`
	MaxTokens   int32     `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	Enabled     int32     `json:"enabled"`
	ExtraConfig string    `json:"extra_config"`
	CreatedAt   time.Time `json:"created_at"`
}

type Skill struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	SystemPrompt    string    `json:"system_prompt"`
	FewShotExamples string    `json:"few_shot_examples"`
	Type            string    `json:"type"`
	Enabled         int32     `json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListConfigs(ctx context.Context) ([]*ApiConfig, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, COALESCE(type,'LLM'), base_url, api_key, model, max_tokens, temperature, enabled, COALESCE(extra_config,''), created_at
		 FROM ai_api_configs ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*ApiConfig, 0)
	for rows.Next() {
		c := &ApiConfig{}
		rows.Scan(&c.ID, &c.Name, &c.Type, &c.BaseURL, &c.APIKey, &c.Model, &c.MaxTokens, &c.Temperature, &c.Enabled, &c.ExtraConfig, &c.CreatedAt)
		list = append(list, c)
	}
	return list, nil
}

func (r *Repository) ListConfigsByType(ctx context.Context, typ string) ([]*ApiConfig, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, COALESCE(type,'LLM'), base_url, api_key, model, max_tokens, temperature, enabled, COALESCE(extra_config,''), created_at
		 FROM ai_api_configs WHERE COALESCE(type,'LLM') = $1 ORDER BY id`, typ)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*ApiConfig, 0)
	for rows.Next() {
		c := &ApiConfig{}
		rows.Scan(&c.ID, &c.Name, &c.Type, &c.BaseURL, &c.APIKey, &c.Model, &c.MaxTokens, &c.Temperature, &c.Enabled, &c.ExtraConfig, &c.CreatedAt)
		list = append(list, c)
	}
	return list, nil
}

func (r *Repository) GetConfig(ctx context.Context, id int64) (*ApiConfig, error) {
	c := &ApiConfig{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, COALESCE(type,'LLM'), base_url, api_key, model, max_tokens, temperature, enabled, COALESCE(extra_config,''), created_at
		 FROM ai_api_configs WHERE id = $1`, id).
		Scan(&c.ID, &c.Name, &c.Type, &c.BaseURL, &c.APIKey, &c.Model, &c.MaxTokens, &c.Temperature, &c.Enabled, &c.ExtraConfig, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *Repository) CreateConfig(ctx context.Context, c *ApiConfig) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO ai_api_configs (name, type, base_url, api_key, model, max_tokens, temperature) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		c.Name, c.Type, c.BaseURL, c.APIKey, c.Model, c.MaxTokens, c.Temperature).Scan(&id)
	return id, err
}

func (r *Repository) UpdateConfig(ctx context.Context, c *ApiConfig) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE ai_api_configs SET name=$1, type=$2, base_url=$3, api_key=$4, model=$5, max_tokens=$6, temperature=$7 WHERE id=$8`,
		c.Name, c.Type, c.BaseURL, c.APIKey, c.Model, c.MaxTokens, c.Temperature, c.ID)
	return err
}

func (r *Repository) DeleteConfig(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM ai_api_configs WHERE id = $1`, id)
	return err
}

func (r *Repository) ToggleConfig(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE ai_api_configs SET enabled = CASE WHEN enabled = 1 THEN 0 ELSE 1 END WHERE id = $1`, id)
	return err
}

func (r *Repository) GetBinding(ctx context.Context, feature string) (int64, error) {
	var configID int64
	err := r.db.QueryRowContext(ctx,
		`SELECT api_config_id FROM ai_bindings WHERE feature = $1`, feature).Scan(&configID)
	return configID, err
}

func (r *Repository) SetBinding(ctx context.Context, feature string, configID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO ai_bindings (feature, api_config_id) VALUES ($1,$2) ON CONFLICT (feature) DO UPDATE SET api_config_id = $2`,
		feature, configID)
	return err
}

func (r *Repository) ListAllBindings(ctx context.Context) (map[string]int64, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT feature, api_config_id FROM ai_bindings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]int64)
	for rows.Next() {
		var f string
		var id int64
		if err := rows.Scan(&f, &id); err != nil {
			return nil, err
		}
		result[f] = id
	}
	return result, nil
}

func (r *Repository) ListSkills(ctx context.Context) ([]*Skill, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, description, system_prompt, few_shot_examples, type, enabled, created_at FROM ai_skills ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*Skill, 0)
	for rows.Next() {
		s := &Skill{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.SystemPrompt, &s.FewShotExamples, &s.Type, &s.Enabled, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *Repository) ListSkillsByType(ctx context.Context, typ string) ([]*Skill, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, description, system_prompt, few_shot_examples, type, enabled, created_at FROM ai_skills WHERE type = $1 ORDER BY id`, typ)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*Skill, 0)
	for rows.Next() {
		s := &Skill{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.SystemPrompt, &s.FewShotExamples, &s.Type, &s.Enabled, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *Repository) GetSkill(ctx context.Context, id int64) (*Skill, error) {
	s := &Skill{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, system_prompt, few_shot_examples, type, enabled, created_at FROM ai_skills WHERE id = $1`, id).
		Scan(&s.ID, &s.Name, &s.Description, &s.SystemPrompt, &s.FewShotExamples, &s.Type, &s.Enabled, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) CreateSkill(ctx context.Context, s *Skill) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO ai_skills (name, description, system_prompt, few_shot_examples, type) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		s.Name, s.Description, s.SystemPrompt, s.FewShotExamples, s.Type).Scan(&id)
	return id, err
}

func (r *Repository) UpdateSkill(ctx context.Context, id int64, s *Skill) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE ai_skills SET name=$1, description=$2, system_prompt=$3, few_shot_examples=$4, type=$5 WHERE id=$6`,
		s.Name, s.Description, s.SystemPrompt, s.FewShotExamples, s.Type, id)
	return err
}

func (r *Repository) DeleteSkill(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM ai_skills WHERE id = $1`, id)
	return err
}

func (r *Repository) ToggleSkill(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE ai_skills SET enabled = CASE WHEN enabled = 1 THEN 0 ELSE 1 END WHERE id = $1`, id)
	return err
}

func (r *Repository) CountSkillByName(ctx context.Context, name string) (int64, error) {
	var cnt int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai_skills WHERE name = $1`, name).Scan(&cnt)
	return cnt, err
}

func (r *Repository) UsageOverview(ctx context.Context) (map[string]any, error) {
	var totalCalls, totalTokens int64
	var totalDuration float64
	var featureCount, userCount int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai_usage_logs`).Scan(&totalCalls); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(token_count),0) FROM ai_usage_logs`).Scan(&totalTokens); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(duration_ms),0) FROM ai_usage_logs`).Scan(&totalDuration); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT feature) FROM ai_usage_logs`).Scan(&featureCount); err != nil {
		return nil, err
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT user_id) FROM ai_usage_logs WHERE user_id IS NOT NULL`).Scan(&userCount); err != nil {
		return nil, err
	}
	return map[string]any{
		"total_calls":     totalCalls,
		"total_tokens":    totalTokens,
		"total_duration":  totalDuration,
		"feature_count":   featureCount,
		"user_count":      userCount,
		"avg_duration_ms": avgDuration(totalDuration, totalCalls),
	}, nil
}

func avgDuration(total float64, calls int64) float64 {
	if calls == 0 {
		return 0
	}
	return total / float64(calls)
}

func (r *Repository) UsageByFeature(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT feature, COUNT(*), COALESCE(SUM(token_count),0), COALESCE(AVG(duration_ms),0)
		 FROM ai_usage_logs GROUP BY feature ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]map[string]any, 0)
	for rows.Next() {
		var f string
		var calls int64
		var tokens int64
		var avg float64
		if err := rows.Scan(&f, &calls, &tokens, &avg); err != nil {
			return nil, err
		}
		result = append(result, map[string]any{"feature": f, "calls": calls, "tokens": tokens, "avg_duration_ms": avg})
	}
	return result, nil
}

func (r *Repository) UsageDaily(ctx context.Context, startDate string) ([]map[string]any, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT TO_CHAR(created_at, 'YYYY-MM-DD') AS day, COUNT(*), COALESCE(SUM(token_count),0)
		 FROM ai_usage_logs WHERE created_at >= $1::date
		 GROUP BY day ORDER BY day`, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]map[string]any, 0)
	for rows.Next() {
		var day string
		var calls, tokens int64
		if err := rows.Scan(&day, &calls, &tokens); err != nil {
			return nil, err
		}
		result = append(result, map[string]any{"date": day, "calls": calls, "tokens": tokens})
	}
	return result, nil
}

func (r *Repository) ListPendingSessions(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, title, status, created_at
		 FROM ai_conversations WHERE status = 0 ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]map[string]any, 0)
	for rows.Next() {
		var id, userID int64
		var title string
		var status int32
		var createdAt time.Time
		if err := rows.Scan(&id, &userID, &title, &status, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, map[string]any{
			"id": id, "user_id": userID, "title": title,
			"status": status, "created_at": createdAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return result, nil
}

func (r *Repository) GetSessionMessages(ctx context.Context, sessionID int64) ([]map[string]any, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, role, content, token_count, created_at FROM ai_chat_messages WHERE conversation_id = $1 ORDER BY id`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]map[string]any, 0)
	for rows.Next() {
		var id int64
		var role, content string
		var tokens int32
		var createdAt time.Time
		if err := rows.Scan(&id, &role, &content, &tokens, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, map[string]any{
			"id": id, "role": role, "content": content, "token_count": tokens,
			"created_at": createdAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return result, nil
}

func (r *Repository) SendSessionReply(ctx context.Context, sessionID, adminID int64, content string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO ai_chat_messages (conversation_id, role, content) VALUES ($1, 'assistant', $2)`,
		sessionID, content)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `UPDATE ai_conversations SET updated_at = NOW() WHERE id = $1`, sessionID)
	return err
}

func (r *Repository) MarkSessionProcessed(ctx context.Context, sessionID, adminID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE ai_conversations SET status = 1, updated_at = NOW() WHERE id = $1`, sessionID)
	return err
}

func (r *Repository) CountPendingSessions(ctx context.Context) (int64, error) {
	var cnt int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai_conversations WHERE status = 0`).Scan(&cnt)
	return cnt, err
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListConfigs(ctx context.Context) ([]*ApiConfig, error) {
	return s.repo.ListConfigs(ctx)
}

func (s *Service) CreateConfig(ctx context.Context, c *ApiConfig) error {
	_, err := s.repo.CreateConfig(ctx, c)
	return err
}

func (s *Service) DeleteConfig(ctx context.Context, id int64) error {
	return s.repo.DeleteConfig(ctx, id)
}

func (s *Service) ToggleConfig(ctx context.Context, id int64) error {
	return s.repo.ToggleConfig(ctx, id)
}

func (s *Service) GetBinding(ctx context.Context, feature string) (int64, error) {
	return s.repo.GetBinding(ctx, feature)
}

func (s *Service) SetBinding(ctx context.Context, feature string, configID int64) error {
	return s.repo.SetBinding(ctx, feature, configID)
}

func (s *Service) ListAllBindings(ctx context.Context) (map[string]int64, error) {
	return s.repo.ListAllBindings(ctx)
}

func (s *Service) ListConfigsByType(ctx context.Context, typ string) ([]*ApiConfig, error) {
	return s.repo.ListConfigsByType(ctx, typ)
}

func (s *Service) GetConfig(ctx context.Context, id int64) (*ApiConfig, error) {
	return s.repo.GetConfig(ctx, id)
}

func (s *Service) UpdateConfig(ctx context.Context, c *ApiConfig) error {
	return s.repo.UpdateConfig(ctx, c)
}

func (s *Service) ListSkills(ctx context.Context) ([]*Skill, error) {
	return s.repo.ListSkills(ctx)
}

func (s *Service) ListSkillsByType(ctx context.Context, typ string) ([]*Skill, error) {
	return s.repo.ListSkillsByType(ctx, typ)
}

func (s *Service) GetSkill(ctx context.Context, id int64) (*Skill, error) {
	return s.repo.GetSkill(ctx, id)
}

func (s *Service) CreateSkill(ctx context.Context, sk *Skill) error {
	_, err := s.repo.CreateSkill(ctx, sk)
	return err
}

func (s *Service) UpdateSkill(ctx context.Context, id int64, sk *Skill) error {
	return s.repo.UpdateSkill(ctx, id, sk)
}

func (s *Service) DeleteSkill(ctx context.Context, id int64) error {
	return s.repo.DeleteSkill(ctx, id)
}

func (s *Service) ToggleSkill(ctx context.Context, id int64) error {
	return s.repo.ToggleSkill(ctx, id)
}

func (s *Service) CreateMissingCustomerServiceDefaults(ctx context.Context) (int, error) {
	defaults := []struct{ name, desc, typ string }{
		{"账号与登录问题", "账号注册、登录、密码找回、第三方绑定等问题的处理", "CUSTOMER_SERVICE"},
		{"视频上传与发布", "视频上传、转码、稿件审核、发布状态等问题", "CUSTOMER_SERVICE"},
		{"直播与互动", "开播、直播间设置、弹幕、连麦等问题", "CUSTOMER_SERVICE"},
		{"会员与消费", "会员购买、充值、订单、退款等问题", "CUSTOMER_SERVICE"},
		{"举报与违规", "内容举报、违规申诉、封禁处理等问题", "CUSTOMER_SERVICE"},
		{"其他问题", "无法归类的其他平台问题", "CUSTOMER_SERVICE"},
	}
	created := 0
	for _, d := range defaults {
		cnt, err := s.repo.CountSkillByName(ctx, d.name)
		if err != nil {
			continue
		}
		if cnt > 0 {
			continue
		}
		_, err = s.repo.CreateSkill(ctx, &Skill{Name: d.name, Description: d.desc, Type: d.typ})
		if err != nil {
			continue
		}
		created++
	}
	return created, nil
}

func (s *Service) MatchCustomerServiceSkill(ctx context.Context, content string) (map[string]any, error) {
	skills, err := s.repo.ListSkillsByType(ctx, "CUSTOMER_SERVICE")
	if err != nil {
		return nil, err
	}
	for _, sk := range skills {
		lower := strings.ToLower(content)
		if strings.Contains(strings.ToLower(sk.Name), lower) || strings.Contains(strings.ToLower(sk.Description), lower) {
			return map[string]any{"name": sk.Name, "id": sk.ID, "matched": true}, nil
		}
	}
	return map[string]any{"name": "default", "matched": false}, nil
}

func (s *Service) RouteSkills(ctx context.Context, content string, limit int) ([]map[string]any, error) {
	skills, err := s.repo.ListSkillsByType(ctx, "CUSTOMER_SERVICE")
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 3
	}
	type scored struct {
		skill *Skill
		score int
	}
	var scoredSkills []scored
	contentLower := strings.ToLower(content)
	words := strings.Fields(contentLower)
	for _, sk := range skills {
		if sk.Enabled == 0 {
			continue
		}
		score := 0
		nameLower := strings.ToLower(sk.Name)
		descLower := strings.ToLower(sk.Description)
		if nameLower == contentLower {
			score += 100
		} else if strings.Contains(nameLower, contentLower) {
			score += 50
		} else if strings.Contains(contentLower, nameLower) {
			score += 30
		}
		for _, word := range words {
			if len(word) < 2 {
				continue
			}
			if strings.Contains(nameLower, word) {
				score += 10
			}
			if strings.Contains(descLower, word) {
				score += 5
			}
		}
		if score > 0 {
			scoredSkills = append(scoredSkills, scored{skill: sk, score: score})
		}
	}
	sort.Slice(scoredSkills, func(i, j int) bool {
		return scoredSkills[i].score > scoredSkills[j].score
	})
	if len(scoredSkills) > limit {
		scoredSkills = scoredSkills[:limit]
	}
	result := make([]map[string]any, 0, len(scoredSkills))
	for _, ss := range scoredSkills {
		result = append(result, map[string]any{
			"id": ss.skill.ID, "name": ss.skill.Name, "score": ss.score, "matched": true,
		})
	}
	return result, nil
}

func (s *Service) UsageOverview(ctx context.Context) (map[string]any, error) {
	return s.repo.UsageOverview(ctx)
}

func (s *Service) UsageByFeature(ctx context.Context) ([]map[string]any, error) {
	return s.repo.UsageByFeature(ctx)
}

func (s *Service) UsageDaily(ctx context.Context, days int) ([]map[string]any, error) {
	start := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	return s.repo.UsageDaily(ctx, start)
}

func (s *Service) ListPendingSessions(ctx context.Context) ([]map[string]any, error) {
	return s.repo.ListPendingSessions(ctx)
}

func (s *Service) GetSessionMessages(ctx context.Context, sessionID int64) ([]map[string]any, error) {
	return s.repo.GetSessionMessages(ctx, sessionID)
}

func (s *Service) SendSessionReply(ctx context.Context, sessionID, adminID int64, content string) error {
	return s.repo.SendSessionReply(ctx, sessionID, adminID, content)
}

func (s *Service) MarkSessionProcessed(ctx context.Context, sessionID, adminID int64) error {
	return s.repo.MarkSessionProcessed(ctx, sessionID, adminID)
}

func (s *Service) CountPendingSessions(ctx context.Context) (int64, error) {
	return s.repo.CountPendingSessions(ctx)
}
