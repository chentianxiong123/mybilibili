package analytics

import (
	"context"
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Overview(ctx context.Context, userID int64) (map[string]interface{}, error) {
	var msCount, viewCount, likeCount, coinCount, collectCount, followerCount int64
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM manuscripts WHERE user_id = $1`, userID).Scan(&msCount)
	r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(view_count),0) FROM manuscripts WHERE user_id = $1`, userID).Scan(&viewCount)
	r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(like_count),0) FROM manuscripts WHERE user_id = $1`, userID).Scan(&likeCount)
	r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(coin_count),0) FROM manuscripts WHERE user_id = $1`, userID).Scan(&coinCount)
	r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(collect_count),0) FROM manuscripts WHERE user_id = $1`, userID).Scan(&collectCount)
	r.db.QueryRowContext(ctx, `SELECT follower_count FROM users WHERE id = $1`, userID).Scan(&followerCount)
	return map[string]interface{}{
		"manuscript_count": msCount, "view_count": viewCount, "like_count": likeCount,
		"coin_count": coinCount, "collect_count": collectCount, "follower_count": followerCount,
	}, nil
}

func (r *Repository) Trend(ctx context.Context, userID int64, days int) ([]map[string]interface{}, error) {
	start := time.Now().AddDate(0, 0, -days)
	rows, err := r.db.QueryContext(ctx,
		`SELECT metric_date, view_count, like_count, coin_count, collect_count, share_count, comment_count, danmaku_count
		 FROM manuscript_daily_metrics WHERE user_id = $1 AND metric_date >= $2 ORDER BY metric_date`,
		userID, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var date time.Time
		var v, l, co, cl, s, cm, d int64
		rows.Scan(&date, &v, &l, &co, &cl, &s, &cm, &d)
		list = append(list, map[string]interface{}{
			"date": date.Format("2006-01-02"), "view_count": v, "like_count": l,
			"coin_count": co, "collect_count": cl, "share_count": s,
			"comment_count": cm, "danmaku_count": d,
		})
	}
	return list, nil
}

func (r *Repository) Ranking(ctx context.Context, sortBy string, limit int) ([]map[string]interface{}, error) {
	orderBy := "view_count DESC"
	switch sortBy {
	case "likes":
		orderBy = "like_count DESC"
	case "coins":
		orderBy = "coin_count DESC"
	case "time":
		orderBy = "upload_time DESC"
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, view_count, like_count, coin_count, collect_count, upload_time
		 FROM manuscripts WHERE status = 3 ORDER BY `+orderBy+` LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, v, l, co, cl int64
		var title, t string
		rows.Scan(&id, &title, &v, &l, &co, &cl, &t)
		list = append(list, map[string]interface{}{
			"id": id, "title": title, "view_count": v, "like_count": l,
			"coin_count": co, "collect_count": cl, "upload_time": t,
		})
	}
	return list, nil
}

func (r *Repository) LatestComments(ctx context.Context, userID int64, limit int) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT c.id, c.content, c.manuscript_id, m.title, c.created_at
		 FROM comments c JOIN manuscripts m ON c.manuscript_id = m.id
		 WHERE m.user_id = $1 AND c.status = 0 ORDER BY c.created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, msID int64
		var content, title, t string
		rows.Scan(&id, &content, &msID, &title, &t)
		list = append(list, map[string]interface{}{
			"id": id, "content": content, "manuscript_id": msID, "title": title, "created_at": t,
		})
	}
	return list, nil
}

func (r *Repository) FansTrend(ctx context.Context, userID int64, days int) ([]map[string]interface{}, error) {
	start := time.Now().AddDate(0, 0, -days)
	rows, err := r.db.QueryContext(ctx,
		`SELECT DATE(created_at) as d, COUNT(*) as cnt FROM follows WHERE following_id = $1 AND created_at >= $2 GROUP BY d ORDER BY d`,
		userID, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var d time.Time
		var cnt int64
		rows.Scan(&d, &cnt)
		list = append(list, map[string]interface{}{"date": d.Format("2006-01-02"), "count": cnt})
	}
	return list, nil
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Overview(ctx context.Context, userID int64) (map[string]interface{}, error) {
	return s.repo.Overview(ctx, userID)
}

func (s *Service) Trend(ctx context.Context, userID int64, days int) ([]map[string]interface{}, error) {
	if days < 1 || days > 365 {
		days = 7
	}
	return s.repo.Trend(ctx, userID, days)
}

func (s *Service) Ranking(ctx context.Context, sortBy string, limit int) ([]map[string]interface{}, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.repo.Ranking(ctx, sortBy, limit)
}

func (s *Service) LatestComments(ctx context.Context, userID int64, limit int) ([]map[string]interface{}, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.repo.LatestComments(ctx, userID, limit)
}

func (s *Service) FansTrend(ctx context.Context, userID int64, days int) ([]map[string]interface{}, error) {
	if days < 1 || days > 365 {
		days = 7
	}
	return s.repo.FansTrend(ctx, userID, days)
}

func (s *Service) FansRanking(ctx context.Context, userID int64, limit int) ([]map[string]interface{}, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	rows, err := s.repo.db.QueryContext(ctx,
		`SELECT f.follower_id, u.nickname, u.avatar, f.created_at
		 FROM follows f JOIN users u ON f.follower_id = u.id
		 WHERE f.followee_id = $1
		 ORDER BY f.created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var id int64
		var nickname string
		var avatar sql.NullString
		var createdAt time.Time
		rows.Scan(&id, &nickname, &avatar, &createdAt)
		out = append(out, map[string]interface{}{
			"user_id": id, "nickname": nickname, "avatar": avatar.String,
			"follow_time": createdAt.Format("2006-01-02 15:04:05"),
		})
	}
	return out, nil
}

func (s *Service) ManuscriptTrend(ctx context.Context, userID int64, days int) ([]map[string]interface{}, error) {
	if days < 1 || days > 365 {
		days = 7
	}
	rows, err := s.repo.db.QueryContext(ctx,
		`SELECT date_trunc('day', created_at)::date AS day, COUNT(*), COALESCE(SUM(view_count), 0), COALESCE(SUM(like_count), 0)
		 FROM manuscripts WHERE user_id = $1 AND created_at >= NOW() - ($2 || ' days')::interval
		 GROUP BY day ORDER BY day`, userID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var day time.Time
		var cnt int64
		var views, likes int64
		rows.Scan(&day, &cnt, &views, &likes)
		out = append(out, map[string]interface{}{
			"date": day.Format("2006-01-02"), "count": cnt,
			"views": views, "likes": likes,
		})
	}
	return out, nil
}
