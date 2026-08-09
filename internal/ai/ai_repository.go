package ai

import (
	"context"
	"database/sql"
	"time"
)

type ApiConfig struct {
	ID          int64
	Name        string
	BaseURL     string
	APIKey      string
	Model       string
	MaxTokens   int32
	Temperature float64
	Enabled     int32
	ExtraConfig string
	CreatedAt   time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListConfigs(ctx context.Context) ([]*ApiConfig, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, base_url, api_key, model, max_tokens, temperature, enabled, COALESCE(extra_config,''), created_at
		 FROM ai_api_configs ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*ApiConfig
	for rows.Next() {
		c := &ApiConfig{}
		rows.Scan(&c.ID, &c.Name, &c.BaseURL, &c.APIKey, &c.Model, &c.MaxTokens, &c.Temperature, &c.Enabled, &c.ExtraConfig, &c.CreatedAt)
		list = append(list, c)
	}
	return list, nil
}

func (r *Repository) CreateConfig(ctx context.Context, c *ApiConfig) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO ai_api_configs (name, base_url, api_key, model, max_tokens, temperature) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		c.Name, c.BaseURL, c.APIKey, c.Model, c.MaxTokens, c.Temperature).Scan(&id)
	return id, err
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
