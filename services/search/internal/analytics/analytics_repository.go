package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Overview 创作者数据概览，对齐老版 CreatorOverviewVO 字段（camelCase）
func (r *Repository) Overview(ctx context.Context, userID int64) (map[string]interface{}, error) {
	var totalViews, totalLikes, totalCoins, totalCollections, totalShares, totalComments, totalManuscripts int64
	r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(view_count),0), COALESCE(SUM(like_count),0),
		       COALESCE(SUM(coin_count),0), COALESCE(SUM(collect_count),0),
		       COALESCE(SUM(share_count),0), COALESCE(SUM(comment_count),0), COUNT(*)
		FROM manuscripts WHERE user_id = $1 AND status = 3`,
		userID).Scan(&totalViews, &totalLikes, &totalCoins, &totalCollections, &totalShares, &totalComments, &totalManuscripts)

	var totalDanmaku int64
	r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(danmaku_count),0) FROM manuscripts WHERE user_id = $1 AND status = 3`, userID).Scan(&totalDanmaku)

	var totalFollowers int64
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM follows WHERE following_id = $1`, userID).Scan(&totalFollowers)

	start7 := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	var viewsIncrease int64
	r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(view_count),0) FROM manuscript_daily_metrics WHERE user_id = $1 AND metric_date >= $2`, userID, start7).Scan(&viewsIncrease)

	var likesIncrease, coinsIncrease, collectionsIncrease, sharesIncrease int64
	r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(CASE WHEN i.interaction_type='LIKE' THEN 1 ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN i.interaction_type='COIN' THEN 1 ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN i.interaction_type='COLLECT' THEN 1 ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN i.interaction_type='SHARE' THEN 1 ELSE 0 END),0)
		FROM user_interactions i JOIN manuscripts m ON i.target_id = m.id
		WHERE m.user_id = $1 AND m.status = 3 AND i.target_type = 'MANUSCRIPT' AND DATE(i.created_at) >= $2`,
		userID, start7).Scan(&likesIncrease, &coinsIncrease, &collectionsIncrease, &sharesIncrease)

	var commentsIncrease int64
	r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM comments c JOIN manuscripts m ON c.manuscript_id = m.id
		WHERE m.user_id = $1 AND m.status = 3 AND DATE(c.created_at) >= $2`,
		userID, start7).Scan(&commentsIncrease)

	var followersIncrease int64
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM follows WHERE following_id = $1 AND created_at >= $2`, userID, start7).Scan(&followersIncrease)

	return map[string]interface{}{
		"totalViews":           totalViews,
		"totalLikes":           totalLikes,
		"totalCoins":           totalCoins,
		"totalCollections":     totalCollections,
		"totalShares":          totalShares,
		"totalComments":        totalComments,
		"totalManuscripts":     totalManuscripts,
		"totalDanmaku":         totalDanmaku,
		"totalFollowers":       totalFollowers,
		"viewsIncrease":        viewsIncrease,
		"likesIncrease":        likesIncrease,
		"coinsIncrease":        coinsIncrease,
		"collectionsIncrease":  collectionsIncrease,
		"sharesIncrease":       sharesIncrease,
		"commentsIncrease":     commentsIncrease,
		"followersIncrease":    followersIncrease,
		"danmakuIncrease":      int64(0),
		"updateTime":           time.Now().Format("2006-01-02"),
	}, nil
}

// Trend 播放/互动/评论/粉丝趋势，对齐老版 TrendDataVO
func (r *Repository) Trend(ctx context.Context, userID int64, days int) (map[string]interface{}, error) {
	if days < 1 {
		days = 7
	}
	start := time.Now().AddDate(0, 0, -(days - 1)).Format("2006-01-02")

	viewMap := map[string]int64{}
	rows, _ := r.db.QueryContext(ctx, `SELECT metric_date::text, COALESCE(SUM(view_count),0) FROM manuscript_daily_metrics WHERE user_id = $1 AND metric_date >= $2 GROUP BY metric_date`, userID, start)
	if rows != nil {
		for rows.Next() {
			var d string
			var v int64
			rows.Scan(&d, &v)
			viewMap[d] = v
		}
		rows.Close()
	}

	interactionMap := map[string]map[string]int64{}
	rows, _ = r.db.QueryContext(ctx, `
		SELECT DATE(i.created_at)::text,
		       SUM(CASE WHEN i.interaction_type='LIKE' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN i.interaction_type='COIN' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN i.interaction_type='COLLECT' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN i.interaction_type='SHARE' THEN 1 ELSE 0 END)
		FROM user_interactions i JOIN manuscripts m ON i.target_id = m.id
		WHERE m.user_id = $1 AND m.status = 3 AND i.target_type = 'MANUSCRIPT' AND DATE(i.created_at) >= $2
		GROUP BY DATE(i.created_at)`, userID, start)
	if rows != nil {
		for rows.Next() {
			var d string
			var lk, co, cl, sh int64
			rows.Scan(&d, &lk, &co, &cl, &sh)
			interactionMap[d] = map[string]int64{"likes": lk, "coins": co, "collects": cl, "shares": sh}
		}
		rows.Close()
	}

	commentMap := map[string]int64{}
	rows, _ = r.db.QueryContext(ctx, `
		SELECT DATE(c.created_at)::text, COUNT(*) FROM comments c JOIN manuscripts m ON c.manuscript_id = m.id
		WHERE m.user_id = $1 AND m.status = 3 AND DATE(c.created_at) >= $2 GROUP BY DATE(c.created_at)`, userID, start)
	if rows != nil {
		for rows.Next() {
			var d string
			var cnt int64
			rows.Scan(&d, &cnt)
			commentMap[d] = cnt
		}
		rows.Close()
	}

	fansMap := map[string]int64{}
	rows, _ = r.db.QueryContext(ctx, `SELECT DATE(created_at)::text, COUNT(*) FROM follows WHERE following_id = $1 AND created_at >= $2 GROUP BY DATE(created_at)`, userID, start)
	if rows != nil {
		for rows.Next() {
			var d string
			var cnt int64
			rows.Scan(&d, &cnt)
			fansMap[d] = cnt
		}
		rows.Close()
	}

	dates, views, likes, comments, followers, coins, collects, shares := []string{}, []int64{}, []int64{}, []int64{}, []int64{}, []int64{}, []int64{}, []int64{}
	end := time.Now()
	st := time.Now().AddDate(0, 0, -(days - 1))
	for d := st; !d.After(end); d = d.AddDate(0, 0, 1) {
		ds := d.Format("2006-01-02")
		dates = append(dates, ds)
		views = append(views, viewMap[ds])
		if day, ok := interactionMap[ds]; ok {
			likes = append(likes, day["likes"])
			coins = append(coins, day["coins"])
			collects = append(collects, day["collects"])
			shares = append(shares, day["shares"])
		} else {
			likes = append(likes, 0)
			coins = append(coins, 0)
			collects = append(collects, 0)
			shares = append(shares, 0)
		}
		comments = append(comments, commentMap[ds])
		followers = append(followers, fansMap[ds])
	}

	return map[string]interface{}{
		"dates": dates, "views": views, "likes": likes, "comments": comments,
		"followers": followers, "coins": coins, "collects": collects, "shares": shares,
		"danmaku": make([]int64, len(dates)),
	}, nil
}

// Ranking 创作者稿件排行，对齐老版 ManuscriptRankVO（仅本人稿件）
func (r *Repository) Ranking(ctx context.Context, userID int64, sortBy string, limit int) (map[string]interface{}, error) {
	orderBy := "view_count DESC"
	switch sortBy {
	case "likes":
		orderBy = "like_count DESC"
	case "comments":
		orderBy = "comment_count DESC"
	case "coins":
		orderBy = "coin_count DESC"
	case "time":
		orderBy = "upload_time DESC"
	}
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, title, cover_url, view_count, like_count, comment_count, danmaku_count, coin_count, collect_count, share_count, upload_time
		FROM manuscripts WHERE user_id = $1 AND status = 3 ORDER BY %s LIMIT $2`, orderBy), userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id, vc, lc, cc, dc, coC, clC, shC int64
		var title, cover sql.NullString
		var uploadTime sql.NullTime
		rows.Scan(&id, &title, &cover, &vc, &lc, &cc, &dc, &coC, &clC, &shC, &uploadTime)
		totalInter := lc + cc + dc + coC + clC + shC
		var rate float64
		if vc > 0 {
			rate = float64(totalInter) / float64(vc) * 100
		}
		list = append(list, map[string]interface{}{
			"id": id, "title": title.String, "coverUrl": cover.String,
			"viewCount": vc, "likeCount": lc, "commentCount": cc, "danmakuCount": dc,
			"coinCount": coC, "collectCount": clC, "shareCount": shC,
			"uploadTime": uploadTime.Time.Format("2006-01-02 15:04:05"),
			"interactionRate": rate,
		})
	}
	return map[string]interface{}{"list": list, "total": len(list)}, nil
}

// LatestComments 最新评论，对齐老版 LatestCommentVO
func (r *Repository) LatestComments(ctx context.Context, userID int64, limit int) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.content, c.manuscript_id, m.title, u.username, u.avatar, c.created_at
		FROM comments c JOIN manuscripts m ON c.manuscript_id = m.id
		JOIN users u ON c.user_id = u.id
		WHERE m.user_id = $1 AND m.status = 3 AND c.status = 0
		ORDER BY c.created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id, msID int64
		var content, title, username string
		var avatar sql.NullString
		var ct time.Time
		rows.Scan(&id, &content, &msID, &title, &username, &avatar, &ct)
		list = append(list, map[string]interface{}{
			"id": id, "content": content, "manuscriptId": msID,
			"manuscriptTitle": title, "username": username, "avatar": avatar.String,
			"time": formatTimeAgo(ct),
			"createTime": ct.Format("2006-01-02 15:04:05"),
		})
	}
	return list, nil
}

// FansTrend 粉丝增长趋势，对齐老版 FansTrendVO
func (r *Repository) FansTrend(ctx context.Context, userID int64, days int) (map[string]interface{}, error) {
	if days < 1 {
		days = 30
	}
	start := time.Now().AddDate(0, 0, -(days - 1)).Format("2006-01-02")
	dataMap := map[string]map[string]int64{}
	rows, err := r.db.QueryContext(ctx, `
		SELECT DATE(created_at)::text,
		       SUM(CASE WHEN interaction_type='FOLLOW' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN interaction_type='UNFOLLOW' THEN 1 ELSE 0 END)
		FROM user_interactions WHERE target_id = $1 AND target_type = 'USER' AND created_at >= $2
		GROUP BY DATE(created_at)`, userID, start)
	if err == nil && rows != nil {
		for rows.Next() {
			var d string
			var nf, uf int64
			rows.Scan(&d, &nf, &uf)
			dataMap[d] = map[string]int64{"newFollowers": nf, "unfollows": uf}
		}
		rows.Close()
	}
	if rows == nil {
		rows2, _ := r.db.QueryContext(ctx, `SELECT DATE(created_at)::text, COUNT(*) FROM follows WHERE following_id = $1 AND created_at >= $2 GROUP BY DATE(created_at)`, userID, start)
		if rows2 != nil {
			for rows2.Next() {
				var d string
				var nf int64
				rows2.Scan(&d, &nf)
				dataMap[d] = map[string]int64{"newFollowers": nf, "unfollows": 0}
			}
			rows2.Close()
		}
	}

	var currentFollowers int64
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM follows WHERE following_id = $1`, userID).Scan(&currentFollowers)

	totalNet := int64(0)
	for _, v := range dataMap {
		totalNet += v["newFollowers"] - v["unfollows"]
	}
	running := currentFollowers - totalNet

	dates, newF, unfollows, totalF := []string{}, []int64{}, []int64{}, []int64{}
	end := time.Now()
	st := time.Now().AddDate(0, 0, -(days - 1))
	for d := st; !d.After(end); d = d.AddDate(0, 0, 1) {
		ds := d.Format("2006-01-02")
		dates = append(dates, ds)
		if day, ok := dataMap[ds]; ok {
			newF = append(newF, day["newFollowers"])
			unfollows = append(unfollows, day["unfollows"])
			running += day["newFollowers"] - day["unfollows"]
		} else {
			newF = append(newF, 0)
			unfollows = append(unfollows, 0)
		}
		if running < 0 {
			running = 0
		}
		totalF = append(totalF, running)
	}

	var growthRate float64
	if currentFollowers > 0 && len(totalF) >= 7 {
		weekAgo := totalF[len(totalF)-7]
		if weekAgo > 0 {
			growthRate = float64(currentFollowers-weekAgo) / float64(weekAgo) * 100
		}
	}

	var newToday, unfollowToday int64
	if len(newF) > 0 {
		newToday = newF[len(newF)-1]
		unfollowToday = unfollows[len(unfollows)-1]
	}

	return map[string]interface{}{
		"dates": dates, "newFollowers": newF, "unfollows": unfollows, "totalFollowers": totalF,
		"currentFollowers": currentFollowers, "newFollowersToday": newToday,
		"unfollowsToday": unfollowToday, "growthRate": growthRate,
	}, nil
}

// FansRanking 粉丝互动/观看排行，对齐老版 FansRankingVO
func (r *Repository) FansRanking(ctx context.Context, userID int64, typ string, limit int) ([]map[string]interface{}, error) {
	if typ == "interaction" {
		rows, err := r.db.QueryContext(ctx, `
			SELECT u.id, u.username, u.avatar, COUNT(*) AS interactionCount
			FROM user_interactions i JOIN users u ON i.user_id = u.id
			JOIN manuscripts m ON i.target_id = m.id
			WHERE m.user_id = $1 AND m.status = 3 AND i.target_type = 'MANUSCRIPT'
			GROUP BY u.id ORDER BY interactionCount DESC LIMIT $2`, userID, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		out := []map[string]interface{}{}
		for rows.Next() {
			var id, cnt int64
			var username string
			var avatar sql.NullString
			rows.Scan(&id, &username, &avatar, &cnt)
			out = append(out, map[string]interface{}{
				"id": id, "username": username, "avatar": avatar.String, "interactionCount": cnt,
			})
		}
		return out, nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.username, u.avatar, COUNT(DISTINCT c.id) AS commentCount
		FROM comments c JOIN users u ON c.user_id = u.id
		JOIN manuscripts m ON c.manuscript_id = m.id
		WHERE m.user_id = $1 AND m.status = 3
		GROUP BY u.id ORDER BY commentCount DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]interface{}{}
	for rows.Next() {
		var id, cnt int64
		var username string
		var avatar sql.NullString
		rows.Scan(&id, &username, &avatar, &cnt)
		out = append(out, map[string]interface{}{
			"id": id, "username": username, "avatar": avatar.String, "interactionCount": cnt,
		})
	}
	return out, nil
}

// ManuscriptTrend 稿件播放趋势（累计），对齐老版 ManuscriptTrendVO
func (r *Repository) ManuscriptTrend(ctx context.Context, userID int64) (map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, COALESCE(view_count,0) AS views, upload_time
		FROM manuscripts WHERE user_id = $1 AND status = 3 ORDER BY upload_time ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dates, views, titles, danmaku := []string{}, []int64{}, []string{}, []int64{}
	cumulative := int64(0)
	for rows.Next() {
		var id, v int64
		var title string
		var ut sql.NullTime
		rows.Scan(&id, &title, &v, &ut)
		ds := ""
		if ut.Valid {
			ds = ut.Time.Format("2006-01-02")
		}
		dates = append(dates, ds)
		cumulative += v
		views = append(views, cumulative)
		titles = append(titles, title)
		danmaku = append(danmaku, 0)
	}
	return map[string]interface{}{
		"dates": dates, "views": views, "danmaku": danmaku, "titles": titles,
	}, nil
}

func formatTimeAgo(t time.Time) string {
	d := time.Since(t)
	if d.Minutes() < 1 {
		return "刚刚"
	}
	if d.Hours() < 1 {
		return fmt.Sprintf("%d分钟前", int(d.Minutes()))
	}
	if d.Hours() < 24 {
		return fmt.Sprintf("%d小时前", int(d.Hours()))
	}
	if d.Hours() < 24*30 {
		return fmt.Sprintf("%d天前", int(d.Hours()/24))
	}
	return t.Format("01-02 15:04")
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

func (s *Service) Trend(ctx context.Context, userID int64, days int) (map[string]interface{}, error) {
	if days < 1 || days > 365 {
		days = 7
	}
	return s.repo.Trend(ctx, userID, days)
}

func (s *Service) Ranking(ctx context.Context, userID int64, sortBy string, limit int) (map[string]interface{}, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.repo.Ranking(ctx, userID, sortBy, limit)
}

func (s *Service) LatestComments(ctx context.Context, userID int64, limit int) ([]map[string]interface{}, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.repo.LatestComments(ctx, userID, limit)
}

func (s *Service) FansTrend(ctx context.Context, userID int64, days int) (map[string]interface{}, error) {
	if days < 1 || days > 365 {
		days = 30
	}
	return s.repo.FansTrend(ctx, userID, days)
}

func (s *Service) FansRanking(ctx context.Context, userID int64, typ string, limit int) ([]map[string]interface{}, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.repo.FansRanking(ctx, userID, typ, limit)
}

func (s *Service) ManuscriptTrend(ctx context.Context, userID int64, days int) (map[string]interface{}, error) {
	return s.repo.ManuscriptTrend(ctx, userID)
}
