package search

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"mybilibili/search-service/internal/hot"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SearchManuscripts(ctx context.Context, keyword string, categoryID int64, page, size int32) ([]map[string]interface{}, error) {
	offset := (page - 1) * size
	query := `SELECT m.id, m.title, m.description, m.cover_url, m.user_id, m.category_id,
	                 m.view_count, m.like_count,
	                 COALESCE(m.comment_count,0), COALESCE(m.danmaku_count,0),
	                 COALESCE(m.duration,''), m.status, m.upload_time,
	                 COALESCE(u.id,0), COALESCE(u.username,''), COALESCE(u.nickname,''),
	                 COALESCE(u.avatar,''), COALESCE(u.level,0)
	          FROM manuscripts m LEFT JOIN users u ON m.user_id = u.id
	          WHERE m.status = 3`
	var args []interface{}
	kwIdx := 0
	if keyword != "" {
		args = append(args, keyword)
		kwIdx = len(args)
		query += ` AND m.search_vector @@ plainto_tsquery('zh_cn', $1)`
	}
	if categoryID > 0 {
		args = append(args, categoryID)
		query += fmt.Sprintf(` AND m.category_id = $%d`, len(args))
	}
	if kwIdx > 0 {
		query += fmt.Sprintf(` ORDER BY ts_rank_cd(m.search_vector, plainto_tsquery('zh_cn', $%d)) DESC, m.view_count DESC, m.upload_time DESC`, kwIdx)
	} else {
		query += ` ORDER BY m.upload_time DESC`
	}
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, size, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, userID, catID, viewCount, likeCount, commentCount, danmakuCount, status int64
		var title, desc, cover, duration, uploadTime string
		var uid, ulevel int64
		var uname, unick, uavatar string
		rows.Scan(&id, &title, &desc, &cover, &userID, &catID,
			&viewCount, &likeCount, &commentCount, &danmakuCount,
			&duration, &status, &uploadTime,
			&uid, &uname, &unick, &uavatar, &ulevel)
		uploader := map[string]interface{}{
			"id": uid, "name": unick, "username": uname, "nickname": unick,
			"avatar": uavatar, "level": ulevel,
		}
		list = append(list, map[string]interface{}{
			"id": id, "title": title, "description": desc, "cover_url": cover,
			"user_id": userID, "category_id": catID, "view_count": viewCount,
			"like_count": likeCount, "comment_count": commentCount, "danmaku_count": danmakuCount,
			"duration": duration, "status": status, "upload_time": uploadTime,
			"uploader": uploader,
		})
	}
	return list, nil
}

func (r *Repository) HotSearch(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT keyword, score FROM hot_search
		 WHERE expires_at IS NULL OR expires_at > NOW()
		 ORDER BY score DESC LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var keyword string
		var score int64
		if err := rows.Scan(&keyword, &score); err != nil {
			return nil, err
		}
		list = append(list, map[string]interface{}{
			"rank":    len(list) + 1,
			"keyword": keyword,
			"score":   score,
		})
	}
	if len(list) > 0 {
		return list, nil
	}
	// 兜底：按播放量取已发布稿件标题
	titles, err := r.TopTitles(ctx, 10)
	if err != nil {
		return nil, err
	}
	for i, t := range titles {
		list = append(list, map[string]interface{}{
			"rank":    i + 1,
			"keyword": t,
			"score":   0,
		})
	}
	return list, nil
}

func (r *Repository) TopTitles(ctx context.Context, limit int) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT title FROM manuscripts WHERE status = 3 ORDER BY view_count DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
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
	repo    *Repository
	hotRepo *hot.Repository
}

func NewService(repo *Repository, hotRepo *hot.Repository) *Service {
	return &Service{repo: repo, hotRepo: hotRepo}
}

func (s *Service) Search(ctx context.Context, keyword string, categoryID int64, page, size int32) ([]map[string]interface{}, error) {
	if keyword != "" && s.hotRepo != nil {
		_ = s.hotRepo.Increment(ctx, keyword)
	}
	return s.repo.SearchManuscripts(ctx, keyword, categoryID, page, size)
}

func (s *Service) Hot(ctx context.Context) ([]map[string]interface{}, error) {
	if s.hotRepo != nil {
		if list, err := s.hotRepo.Top(ctx, 10); err == nil && len(list) > 0 {
			return list, nil
		}
	}
	// 兜底：按播放量取已发布稿件标题
	return s.repo.HotSearch(ctx)
}

func (s *Service) Suggest(ctx context.Context, keyword string, size int32) ([]string, error) {
	if keyword == "" {
		return []string{}, nil
	}
	rows, err := s.repo.db.QueryContext(ctx,
		`SELECT title FROM manuscripts WHERE status = 3 AND search_vector @@ to_tsquery('zh_cn', $1 || ':*')
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
	query := `SELECT m.id, m.user_id, m.title, m.cover_url, m.view_count, m.like_count, m.upload_time, m.duration_seconds,
	                 u.id, u.username, u.nickname, u.avatar, u.level
	          FROM manuscripts m LEFT JOIN users u ON m.user_id = u.id
	          WHERE m.status = 3`
	args := []interface{}{}
	paramIdx := 1
	if categoryID > 0 {
		query += ` AND m.category_id = $1`
		paramIdx = 2
		args = append(args, categoryID)
	}
	query += fmt.Sprintf(` ORDER BY m.view_count DESC, m.like_count DESC LIMIT $%d`, paramIdx)
	args = append(args, size)
	rows, err := s.repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var id, userID, views, likes, durSec int64
		var title, cover string
		var created time.Time
		var uid, ulevel int64
		var uname, unick, uavatar string
		rows.Scan(&id, &userID, &title, &cover, &views, &likes, &created, &durSec,
			&uid, &uname, &unick, &uavatar, &ulevel)
		uploader := map[string]interface{}{
			"id": uid, "name": unick, "username": uname, "nickname": unick,
			"avatar": uavatar, "level": ulevel,
		}
		hours := durSec / 3600
		mins := (durSec % 3600) / 60
		secs := durSec % 60
		var duration string
		if hours > 0 {
			duration = fmt.Sprintf("%d:%02d:%02d", hours, mins, secs)
		} else {
			duration = fmt.Sprintf("%d:%02d", mins, secs)
		}
		out = append(out, map[string]interface{}{
			"manuscript_id": id, "user_id": userID, "title": title,
			"cover_url": cover, "view_count": views, "like_count": likes,
			"created_at": created.Format("2006-01-02 15:04:05"),
			"duration": duration, "duration_seconds": durSec,
			"uploader": uploader,
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
	if s.hotRepo == nil {
		return nil
	}
	return s.hotRepo.Increment(ctx, keyword)
}

func (s *Service) SetKeyword(ctx context.Context, keyword string, score, rank int) error {
	if s.hotRepo == nil {
		return nil
	}
	_ = s.hotRepo.UpdateScore(ctx, keyword, float64(score))
	return s.hotRepo.SetRank(ctx, keyword, int64(rank))
}

func (s *Service) SetRank(ctx context.Context, keyword string, rank int) error {
	if s.hotRepo == nil {
		return nil
	}
	return s.hotRepo.SetRank(ctx, keyword, int64(rank))
}

func (s *Service) SetScore(ctx context.Context, keyword string, score int) error {
	if s.hotRepo == nil {
		return nil
	}
	return s.hotRepo.UpdateScore(ctx, keyword, float64(score))
}

func (s *Service) GetKeyword(ctx context.Context, keyword string) (map[string]interface{}, error) {
	if s.hotRepo == nil {
		return nil, context.Canceled
	}
	item, err := s.hotRepo.Get(ctx, keyword)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) GetScore(ctx context.Context, keyword string) (int64, error) {
	if s.hotRepo == nil {
		return 0, context.Canceled
	}
	item, err := s.hotRepo.Get(ctx, keyword)
	if err != nil {
		return 0, err
	}
	score, _ := item["score"].(int64)
	return score, nil
}

func (s *Service) CleanExpiredHotSearch(ctx context.Context) error {
	if s.hotRepo == nil {
		return nil
	}
	return s.hotRepo.CleanExpired(ctx, 100)
}

func (s *Service) DeleteOne(ctx context.Context, keyword string) error {
	if s.hotRepo == nil {
		return nil
	}
	return s.hotRepo.Delete(ctx, keyword)
}

func (s *Service) CountIndexed(ctx context.Context) (int64, error) {
	var count int64
	err := s.repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM manuscripts WHERE search_vector IS NOT NULL`).Scan(&count)
	return count, err
}

type IndexStatus struct {
	Engine         string  `json:"engine"`
	Config         string  `json:"config"`
	IndexName      string  `json:"indexName"`
	TotalCount     int64   `json:"totalCount"`
	PublishedCount int64   `json:"publishedCount"`
	IndexedCount   int64   `json:"indexedCount"`
	NullCount      int64   `json:"nullCount"`
	Coverage       float64 `json:"coverage"`
	GINIndex       bool    `json:"ginIndex"`
	Trigger        bool    `json:"trigger"`
	Health         string  `json:"health"`
}

func (s *Service) GetIndexStatus(ctx context.Context) (*IndexStatus, error) {
	st := &IndexStatus{
		Engine:    "postgres-fts",
		Config:    "zh_cn (zhparser)",
		IndexName: "manuscripts",
	}
	var total, published, indexed, nullCount int64
	err := s.repo.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM manuscripts),
			(SELECT COUNT(*) FROM manuscripts WHERE status = 3),
			(SELECT COUNT(*) FROM manuscripts WHERE search_vector IS NOT NULL),
			(SELECT COUNT(*) FROM manuscripts WHERE search_vector IS NULL)`).Scan(&total, &published, &indexed, &nullCount)
	if err != nil {
		return nil, err
	}
	st.TotalCount = total
	st.PublishedCount = published
	st.IndexedCount = indexed
	st.NullCount = nullCount
	if total > 0 {
		st.Coverage = float64(indexed) / float64(total) * 100
	}

	var ginCount int
	_ = s.repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pg_indexes
		WHERE tablename = 'manuscripts' AND indexname = 'idx_manuscripts_search'`).Scan(&ginCount)
	st.GINIndex = ginCount > 0

	var trgCount int
	_ = s.repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pg_trigger
		WHERE tgname = 'trg_manuscript_search_vector'`).Scan(&trgCount)
	st.Trigger = trgCount > 0

	switch {
	case !st.GINIndex || !st.Trigger:
		st.Health = "warning"
	case nullCount > 0:
		st.Health = "degraded"
	default:
		st.Health = "active"
	}
	return st, nil
}

func (s *Service) RebuildSearchVector(ctx context.Context) (int64, error) {
	res, err := s.repo.db.ExecContext(ctx, `
		UPDATE manuscripts
		SET search_vector = to_tsvector('zh_cn', COALESCE(title, '') || ' ' || COALESCE(description, ''))`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
