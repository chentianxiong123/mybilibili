package moderation

import (
	"context"
	"database/sql"
	"time"
)

type ProhibitedWord struct {
	ID        int64
	Word      string
	MatchType string
	Category  string
	IsEnabled int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Report struct {
	ID           int64
	ReporterID   int64
	TargetType   string
	TargetID     int64
	ManuscriptID int64
	Reason       string
	Description  string
	Status       string
	AdminRemark  string
	ProcessedAt  *time.Time
	AIVerdict    string
	AIRiskLevel  string
	CreatedAt    time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListWords(ctx context.Context, page, size int32) ([]*ProhibitedWord, error) {
	offset := (page - 1) * size
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, word, match_type, COALESCE(category,''), is_enabled, created_at, updated_at
		 FROM prohibited_words ORDER BY id LIMIT $1 OFFSET $2`, size, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*ProhibitedWord
	for rows.Next() {
		p := &ProhibitedWord{}
		rows.Scan(&p.ID, &p.Word, &p.MatchType, &p.Category, &p.IsEnabled, &p.CreatedAt, &p.UpdatedAt)
		list = append(list, p)
	}
	return list, nil
}

func (r *Repository) CreateWord(ctx context.Context, word, matchType, category string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO prohibited_words (word, match_type, category) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
		word, matchType, category)
	return err
}

func (r *Repository) DeleteWord(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM prohibited_words WHERE id = $1`, id)
	return err
}

func (r *Repository) ContainsProhibited(ctx context.Context, content string) (bool, error) {
	var found bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM prohibited_words WHERE is_enabled = 1 AND $1 ILIKE '%' || word || '%')`,
		content).Scan(&found)
	return found, err
}

func (r *Repository) CreateReport(ctx context.Context, reporterID int64, targetType string, targetID, manuscriptID int64, reason, desc string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO reports (reporter_id, target_type, target_id, manuscript_id, reason, description)
		 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		reporterID, targetType, targetID, nullInt64(manuscriptID), reason, desc).Scan(&id)
	return id, err
}

func (r *Repository) ListReports(ctx context.Context, page, size int32, status string) ([]*Report, error) {
	offset := (page - 1) * size
	query := `SELECT id, reporter_id, target_type, target_id, COALESCE(manuscript_id,0), reason, description,
	          status, COALESCE(admin_remark,''), processed_at, COALESCE(ai_verdict,''), COALESCE(ai_risk_level,''), created_at
	          FROM reports`
	args := []interface{}{size, offset}
	if status != "" {
		query += ` WHERE status = $1`
		args = append([]interface{}{status}, args...)
	}
	query += ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Report
	for rows.Next() {
		rep := &Report{}
		var pt sql.NullTime
		rows.Scan(&rep.ID, &rep.ReporterID, &rep.TargetType, &rep.TargetID, &rep.ManuscriptID,
			&rep.Reason, &rep.Description, &rep.Status, &rep.AdminRemark, &pt, &rep.AIVerdict, &rep.AIRiskLevel, &rep.CreatedAt)
		if pt.Valid {
			rep.ProcessedAt = &pt.Time
		}
		list = append(list, rep)
	}
	return list, nil
}

func (r *Repository) ProcessReport(ctx context.Context, id int64, action, remark string) error {
	status := "RESOLVED"
	if action == "reject" {
		status = "REJECTED"
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE reports SET status=$1, admin_remark=$2, processed_at=NOW() WHERE id=$3`,
		status, remark, id)
	return err
}

func nullInt64(v int64) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListWords(ctx context.Context, page, size int32) ([]*ProhibitedWord, error) {
	return s.repo.ListWords(ctx, page, size)
}

func (s *Service) CreateWord(ctx context.Context, word, matchType, category string) error {
	return s.repo.CreateWord(ctx, word, matchType, category)
}

func (s *Service) DeleteWord(ctx context.Context, id int64) error {
	return s.repo.DeleteWord(ctx, id)
}

func (s *Service) ContainsProhibited(ctx context.Context, content string) (bool, error) {
	return s.repo.ContainsProhibited(ctx, content)
}

func (s *Service) SubmitReport(ctx context.Context, reporterID int64, targetType string, targetID, msID int64, reason, desc string) error {
	_, err := s.repo.CreateReport(ctx, reporterID, targetType, targetID, msID, reason, desc)
	return err
}

func (s *Service) ListReports(ctx context.Context, page, size int32, status string) ([]*Report, error) {
	return s.repo.ListReports(ctx, page, size, status)
}

func (s *Service) ProcessReport(ctx context.Context, id int64, action, remark string) error {
	return s.repo.ProcessReport(ctx, id, action, remark)
}
