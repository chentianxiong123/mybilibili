package search

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

func (r *Repository) IncrementHotSearch(ctx context.Context, keyword string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO hot_search (keyword, search_count, score, expires_at)
		 VALUES ($1, 1, 1, NOW() + INTERVAL '24 hours')
		 ON CONFLICT (keyword) DO UPDATE SET
		   search_count = hot_search.search_count + 1,
		   score = hot_search.score + 1,
		   updated_at = NOW()`, keyword)
	return err
}

func (r *Repository) SetKeyword(ctx context.Context, keyword string, score, rank int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO hot_search (keyword, score, rank, expires_at)
		 VALUES ($1, $2, $3, NOW() + INTERVAL '24 hours')
		 ON CONFLICT (keyword) DO UPDATE SET score = $2, rank = $3, updated_at = NOW()`,
		keyword, score, rank)
	return err
}

func (r *Repository) SetRank(ctx context.Context, keyword string, rank int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE hot_search SET rank = $1, updated_at = NOW() WHERE keyword = $2`, rank, keyword)
	return err
}

func (r *Repository) SetScore(ctx context.Context, keyword string, score int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE hot_search SET score = $1, updated_at = NOW() WHERE keyword = $2`, score, keyword)
	return err
}

func (r *Repository) GetKeyword(ctx context.Context, keyword string) (map[string]interface{}, error) {
	var id, score, rank, searchCount int64
	var expiresAt sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT id, keyword, score, rank, search_count, expires_at FROM hot_search WHERE keyword = $1`, keyword,
	).Scan(&id, &keyword, &score, &rank, &searchCount, &expiresAt)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{
		"id": id, "keyword": keyword, "score": score, "rank": rank, "search_count": searchCount,
	}
	if expiresAt.Valid {
		result["expires_at"] = expiresAt.Time.Format("2006-01-02T15:04:05Z")
	}
	return result, nil
}

func (r *Repository) GetScore(ctx context.Context, keyword string) (int64, error) {
	var score int64
	err := r.db.QueryRowContext(ctx,
		`SELECT score FROM hot_search WHERE keyword = $1`, keyword).Scan(&score)
	return score, err
}

func (r *Repository) CleanExpiredHotSearch(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM hot_search WHERE expires_at IS NOT NULL AND expires_at < NOW()`)
	return err
}

func (r *Repository) DeleteOne(ctx context.Context, keyword string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM hot_search WHERE keyword = $1`, keyword)
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

func (s *Service) Suggest(ctx context.Context, keyword string, size int32) ([]string, error) {
	if keyword == "" {
		return []string{}, nil
	}
	rows, err := s.repo.db.QueryContext(ctx,
		`SELECT title FROM manuscripts WHERE status = 3 AND title ILIKE '%'||$1||'%'
		 ORDER BY view_count DESC LIMIT $2`, keyword, size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		rows.Scan(&t)
		out = append(out, t)
	}
	return out, nil
}

func (s *Service) Related(ctx context.Context, manuscriptID int64, size int32) ([]map[string]interface{}, error) {
	return s.repo.RecommendRelated(ctx, manuscriptID, 0, size)
}

func (s *Service) HotRecommend(ctx context.Context, categoryID int64, size int32) ([]map[string]interface{}, error) {
	query := `SELECT id, user_id, title, cover_url, view_count, like_count, upload_time
	          FROM manuscripts WHERE status = 3`
	args := []interface{}{}
	paramIdx := 1
	if categoryID > 0 {
		query += ` AND category_id = $1`
		paramIdx = 2
		args = append(args, categoryID)
	}
	query += fmt.Sprintf(` ORDER BY view_count DESC, like_count DESC LIMIT $%d`, paramIdx)
	args = append(args, size)
	rows, err := s.repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var id, userID, views, likes int64
		var title, cover string
		var created time.Time
		rows.Scan(&id, &userID, &title, &cover, &views, &likes, &created)
		out = append(out, map[string]interface{}{
			"manuscript_id": id, "user_id": userID, "title": title, "cover_url": cover,
			"view_count": views, "like_count": likes, "created_at": created.Format("2006-01-02 15:04:05"),
		})
	}
	return out, nil
}

func (s *Service) GetRecommendConfig(ctx context.Context) (string, error) {
	return s.repo.GetRecommendConfig(ctx)
}

func (s *Service) UpdateRecommendConfig(ctx context.Context, configJSON, updatedBy string) error {
	return s.repo.UpdateRecommendConfig(ctx, configJSON, updatedBy)
}

func (s *Service) IncrementHotSearch(ctx context.Context, keyword string) error {
	return s.repo.IncrementHotSearch(ctx, keyword)
}

func (s *Service) SetKeyword(ctx context.Context, keyword string, score, rank int) error {
	return s.repo.SetKeyword(ctx, keyword, score, rank)
}

func (s *Service) SetRank(ctx context.Context, keyword string, rank int) error {
	return s.repo.SetRank(ctx, keyword, rank)
}

func (s *Service) SetScore(ctx context.Context, keyword string, score int) error {
	return s.repo.SetScore(ctx, keyword, score)
}

func (s *Service) GetKeyword(ctx context.Context, keyword string) (map[string]interface{}, error) {
	return s.repo.GetKeyword(ctx, keyword)
}

func (s *Service) GetScore(ctx context.Context, keyword string) (int64, error) {
	return s.repo.GetScore(ctx, keyword)
}

func (s *Service) CleanExpiredHotSearch(ctx context.Context) error {
	return s.repo.CleanExpiredHotSearch(ctx)
}

func (s *Service) DeleteOne(ctx context.Context, keyword string) error {
	return s.repo.DeleteOne(ctx, keyword)
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
