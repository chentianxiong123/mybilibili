package search

import (
	"context"
	"database/sql"

	"mybilibili/internal/abstraction"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SearchManuscripts(ctx context.Context, keyword string, categoryID int64, page, size int32) ([]map[string]interface{}, error) {
	offset := (page - 1) * size
	query := `SELECT id, title, description, cover_url, user_id, category_id, view_count, like_count,
	          COALESCE(duration,''), status, upload_time FROM manuscripts WHERE status = 3`
	args := []interface{}{size, offset}
	if keyword != "" {
		query += ` AND (title ILIKE '%' || $3 || '%' OR description ILIKE '%' || $3 || '%')`
		args = append([]interface{}{keyword}, args...)
	}
	if categoryID > 0 {
		query += ` AND category_id = $0`
	}
	query += ` ORDER BY upload_time DESC LIMIT $1 OFFSET $2`

	// rebuild args safely
	args = []interface{}{size, offset}
	filterCount := 2
	if keyword != "" {
		filterCount++
		query = `SELECT id, title, description, cover_url, user_id, category_id, view_count, like_count,
		         COALESCE(duration,''), status, upload_time FROM manuscripts WHERE status = 3
		         AND (title ILIKE '%' || $3 || '%' OR description ILIKE '%' || $3 || '%')
		         ORDER BY upload_time DESC LIMIT $2 OFFSET $1`
		args = []interface{}{size, offset, keyword}
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, userID, catID, viewCount, likeCount, status int64
		var title, desc, cover, duration, uploadTime string
		rows.Scan(&id, &title, &desc, &cover, &userID, &catID, &viewCount, &likeCount, &duration, &status, &uploadTime)
		list = append(list, map[string]interface{}{
			"id": id, "title": title, "description": desc, "cover_url": cover,
			"user_id": userID, "category_id": catID, "view_count": viewCount,
			"like_count": likeCount, "duration": duration, "status": status, "upload_time": uploadTime,
		})
	}
	return list, nil
}

func (r *Repository) HotSearch(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT title FROM manuscripts WHERE status = 3 ORDER BY view_count DESC LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []string
	for rows.Next() {
		var t string
		rows.Scan(&t)
		list = append(list, t)
	}
	return list, nil
}

func (r *Repository) RecommendRelated(ctx context.Context, manuscriptID, categoryID int64, size int32) ([]map[string]interface{}, error) {
	return r.SearchManuscripts(ctx, "", categoryID, 1, size)
}

func (r *Repository) GetRecommendConfig(ctx context.Context) (string, error) {
	var config string
	err := r.db.QueryRowContext(ctx,
		`SELECT config_json FROM recommend_configs WHERE config_key = 'default'`).Scan(&config)
	return config, err
}

func (r *Repository) UpdateRecommendConfig(ctx context.Context, configJSON, updatedBy string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO recommend_configs (config_key, config_json, updated_by)
		 VALUES ('default', $1, $2)
		 ON CONFLICT (config_key) DO UPDATE SET config_json = $1, updated_by = $2, updated_at = NOW()`,
		configJSON, updatedBy)
	return err
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Search(ctx context.Context, keyword string, categoryID int64, page, size int32) ([]map[string]interface{}, error) {
	return s.repo.SearchManuscripts(ctx, keyword, categoryID, page, size)
}

func (s *Service) Hot(ctx context.Context) ([]string, error) {
	return s.repo.HotSearch(ctx)
}

func (s *Service) Related(ctx context.Context, manuscriptID int64, size int32) ([]map[string]interface{}, error) {
	return s.repo.RecommendRelated(ctx, manuscriptID, 0, size)
}

func (s *Service) GetRecommendConfig(ctx context.Context) (string, error) {
	return s.repo.GetRecommendConfig(ctx)
}

func (s *Service) UpdateRecommendConfig(ctx context.Context, configJSON, updatedBy string) error {
	return s.repo.UpdateRecommendConfig(ctx, configJSON, updatedBy)
}

func (s *Service) CountIndexed(ctx context.Context) (int64, error) {
	var count int64
	err := s.repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM manuscripts WHERE status = 3`).Scan(&count)
	return count, err
}

func (s *Service) BulkIndex(ctx context.Context, engine abstraction.SearchEngine) (int, error) {
	rows, err := s.repo.db.QueryContext(ctx,
		`SELECT id, title FROM manuscripts WHERE status = 3`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	docs := make(map[string]interface{})
	count := 0
	for rows.Next() {
		var id int64
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			continue
		}
		docs[formatID(id)] = map[string]interface{}{"id": id, "title": title, "status": 3}
		count++
	}
	if len(docs) > 0 {
		if err := engine.BulkIndex(ctx, "manuscripts", docs); err != nil {
			return 0, err
		}
	}
	return count, nil
}
