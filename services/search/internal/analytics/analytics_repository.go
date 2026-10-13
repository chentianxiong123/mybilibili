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
func (r *Repository) Overview(ctx context.Context, userID int64, days int) (map[string]interface{}, error) {
	if days != 7 && days != 30 {
		days = 7
	}
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

	start := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	prevStart := time.Now().AddDate(0, 0, -days*2).Format("2006-01-02")
	var viewsIncrease, prevViews int64
	r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(view_count),0) FROM manuscript_daily_metrics WHERE user_id = $1 AND metric_date >= $2`, userID, start).Scan(&viewsIncrease)
	r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(view_count),0) FROM manuscript_daily_metrics WHERE user_id = $1 AND metric_date >= $2 AND metric_date < $3`, userID, prevStart, start).Scan(&prevViews)

	var likesIncrease, coinsIncrease, collectionsIncrease, sharesIncrease int64
	r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(CASE WHEN UPPER(i.interaction_type)='LIKE' THEN 1 ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN UPPER(i.interaction_type)='COIN' THEN 1 ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN UPPER(i.interaction_type)='COLLECT' THEN 1 ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN UPPER(i.interaction_type)='SHARE' THEN 1 ELSE 0 END),0)
		FROM user_interactions i JOIN manuscripts m ON i.target_id = m.id
		WHERE m.user_id = $1 AND m.status = 3 AND i.target_type = 'MANUSCRIPT' AND DATE(i.created_at) >= $2`,
		userID, start).Scan(&likesIncrease, &coinsIncrease, &collectionsIncrease, &sharesIncrease)

	var prevLikes, prevCoins, prevCollects, prevShares int64
	r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(CASE WHEN UPPER(i.interaction_type)='LIKE' THEN 1 ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN UPPER(i.interaction_type)='COIN' THEN 1 ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN UPPER(i.interaction_type)='COLLECT' THEN 1 ELSE 0 END),0),
		       COALESCE(SUM(CASE WHEN UPPER(i.interaction_type)='SHARE' THEN 1 ELSE 0 END),0)
		FROM user_interactions i JOIN manuscripts m ON i.target_id = m.id
		WHERE m.user_id = $1 AND m.status = 3 AND i.target_type = 'MANUSCRIPT' AND DATE(i.created_at) >= $2 AND DATE(i.created_at) < $3`,
		userID, prevStart, start).Scan(&prevLikes, &prevCoins, &prevCollects, &prevShares)

	var commentsIncrease, prevComments int64
	r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM comments c JOIN manuscripts m ON c.manuscript_id = m.id
		WHERE m.user_id = $1 AND m.status = 3 AND DATE(c.created_at) >= $2`,
		userID, start).Scan(&commentsIncrease)
	r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM comments c JOIN manuscripts m ON c.manuscript_id = m.id
		WHERE m.user_id = $1 AND m.status = 3 AND DATE(c.created_at) >= $2 AND DATE(c.created_at) < $3`,
		userID, prevStart, start).Scan(&prevComments)

	var danmakuIncrease, prevDanmaku int64
	r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM danmaku d JOIN manuscripts m ON d.manuscript_id = m.id
		WHERE m.user_id = $1 AND m.status = 3 AND DATE(d.created_at) >= $2`,
		userID, start).Scan(&danmakuIncrease)
	r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM danmaku d JOIN manuscripts m ON d.manuscript_id = m.id
		WHERE m.user_id = $1 AND m.status = 3 AND DATE(d.created_at) >= $2 AND DATE(d.created_at) < $3`,
		userID, prevStart, start).Scan(&prevDanmaku)

	var followersIncrease, prevFollowers int64
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM follows WHERE following_id = $1 AND created_at >= $2`, userID, start).Scan(&followersIncrease)
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM follows WHERE following_id = $1 AND created_at >= $2 AND created_at < $3`, userID, prevStart, start).Scan(&prevFollowers)

	var spaceVisits, prevSpaceVisits int64
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM space_visits WHERE host_user_id = $1 AND visit_date >= $2`, userID, start).Scan(&spaceVisits)
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM space_visits WHERE host_user_id = $1 AND visit_date >= $2 AND visit_date < $3`, userID, prevStart, start).Scan(&prevSpaceVisits)

	return map[string]interface{}{
		"totalViews":          totalViews,
		"totalLikes":          totalLikes,
		"totalCoins":          totalCoins,
		"totalCollections":    totalCollections,
		"totalShares":         totalShares,
		"totalComments":       totalComments,
		"totalManuscripts":    totalManuscripts,
		"totalDanmaku":        totalDanmaku,
		"totalFollowers":      totalFollowers,
		"viewsIncrease":       viewsIncrease,
		"likesIncrease":       likesIncrease,
		"coinsIncrease":       coinsIncrease,
		"collectionsIncrease": collectionsIncrease,
		"sharesIncrease":      sharesIncrease,
		"commentsIncrease":    commentsIncrease,
		"followersIncrease":   followersIncrease,
		"danmakuIncrease":     danmakuIncrease,
		"prevViews":           prevViews,
		"prevLikes":           prevLikes,
		"prevCoins":           prevCoins,
		"prevCollections":     prevCollects,
		"prevShares":          prevShares,
		"prevComments":        prevComments,
		"prevDanmaku":         prevDanmaku,
		"prevFollowers":       prevFollowers,
		"spaceVisits":         spaceVisits,
		"prevSpaceVisits":     prevSpaceVisits,
		"updateTime":          time.Now().Format("2006-01-02"),
		"periodDays":          days,
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
		       SUM(CASE WHEN UPPER(i.interaction_type)='LIKE' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN UPPER(i.interaction_type)='COIN' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN UPPER(i.interaction_type)='COLLECT' THEN 1 ELSE 0 END),
		       SUM(CASE WHEN UPPER(i.interaction_type)='SHARE' THEN 1 ELSE 0 END)
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
	st := time.Now().AddDate(0, 0, -(days - 1))
	end := st.AddDate(0, 0, days-1)
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
			"uploadTime":      uploadTime.Time.Format("2006-01-02 15:04:05"),
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
			"time":       formatTimeAgo(ct),
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
	st := time.Now().AddDate(0, 0, -(days - 1))
	end := st.AddDate(0, 0, days-1)
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

// FansByManuscript 各稿件带来的涨粉数（来源归因，历史累积）。
func (r *Repository) FansByManuscript(ctx context.Context, userID int64) (map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT f.source_manuscript_id, COUNT(*)
		  FROM follows f
		 WHERE f.following_id = $1 AND f.source_manuscript_id IS NOT NULL
		 GROUP BY f.source_manuscript_id`, userID)
	if err != nil {
		return map[string]interface{}{"list": []map[string]interface{}{}}, nil
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var mid, cnt int64
		if err := rows.Scan(&mid, &cnt); err != nil {
			continue
		}
		list = append(list, map[string]interface{}{"manuscriptId": mid, "fansGained": cnt})
	}
	return map[string]interface{}{"list": list}, nil
}

// ManuscriptInsights 各稿件洞察（粉丝观看率/游客播放占比/平均播放进度）+ 粉丝播放趋势。
// 口径：watch_history 按（用户，稿件）去重记最新进度；fan=观看者已关注作者；
// 游客播放≈总播放-登录观众数（钳制≥0）；平均进度=LEAST(进度，时长)/时长 的均值。
func (r *Repository) ManuscriptInsights(ctx context.Context, userID int64, days int) (map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT w.manuscript_id,
		       COUNT(*) AS logged_viewers,
		       COUNT(*) FILTER (WHERE f.follower_id IS NOT NULL) AS fan_viewers,
		       COALESCE(m.view_count,0) AS view_count,
		       AVG(CASE WHEN COALESCE(m.duration_seconds,0) > 0
		                THEN LEAST(w.progress_seconds, m.duration_seconds)::float / m.duration_seconds END) AS avg_progress
		  FROM watch_history w
		  JOIN manuscripts m ON m.id = w.manuscript_id
		  LEFT JOIN follows f ON f.follower_id = w.user_id AND f.following_id = m.user_id
		 WHERE m.user_id = $1 AND m.status = 3
		 GROUP BY w.manuscript_id, m.view_count`, userID)
	list := []map[string]interface{}{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var mid, logged, fans, views int64
			var avgProgress sql.NullFloat64
			if err := rows.Scan(&mid, &logged, &fans, &views, &avgProgress); err != nil {
				continue
			}
			var fanRate, guestRate float64
			if logged > 0 {
				fanRate = float64(fans) / float64(logged)
			}
			denom := views
			if denom < logged {
				denom = logged
			}
			if denom > 0 {
				guestViews := denom - logged
				if guestViews < 0 {
					guestViews = 0
				}
				guestRate = float64(guestViews) / float64(denom)
			}
			item := map[string]interface{}{
				"manuscriptId": mid, "fanRate": fanRate, "guestRate": guestRate,
			}
			if avgProgress.Valid {
				item["avgProgress"] = avgProgress.Float64
			}
			list = append(list, item)
		}
	}

	// 粉丝播放趋势：每日去重粉丝观众数
	start := time.Now().AddDate(0, 0, -(days - 1)).Format("2006-01-02")
	fanMap := map[string]int64{}
	frows, ferr := r.db.QueryContext(ctx, `
		SELECT DATE(w.watched_at)::text, COUNT(DISTINCT w.user_id)
		  FROM watch_history w
		  JOIN manuscripts m ON m.id = w.manuscript_id
		  JOIN follows f ON f.follower_id = w.user_id AND f.following_id = m.user_id
		 WHERE m.user_id = $1 AND m.status = 3 AND DATE(w.watched_at) >= $2
		 GROUP BY DATE(w.watched_at)`, userID, start)
	if ferr == nil && frows != nil {
		for frows.Next() {
			var d string
			var v int64
			if err := frows.Scan(&d, &v); err == nil {
				fanMap[d] = v
			}
		}
		frows.Close()
	}
	dates, fans := []string{}, []int64{}
	st := time.Now().AddDate(0, 0, -(days - 1))
	end := st.AddDate(0, 0, days-1)
	for d := st; !d.After(end); d = d.AddDate(0, 0, 1) {
		ds := d.Format("2006-01-02")
		dates = append(dates, ds)
		fans = append(fans, fanMap[ds])
	}
	return map[string]interface{}{
		"list":     list,
		"fanTrend": map[string]interface{}{"dates": dates, "fans": fans},
	}, nil
}

type terminalItem struct {
	Key     string  `json:"key"`
	Label   string  `json:"label"`
	Value   int64   `json:"value"`
	Percent float64 `json:"percent"`
}

// AccountDiagnosis 账号诊断：表现总结（我的指标 vs 同类UP主均值 雷达数据）
// + 播放分布（人群分布 游客/粉丝、观看途径 客户端类型、来源稿件占比昨日）。
// 数据口径：投稿/播放/点赞/新增关注 取近30天；同类UP主=其他UP主同期均值；
// 人群分布/观看途径 取近30天 watch_history；来源稿件占比 取最近有数据日。
func (r *Repository) AccountDiagnosis(ctx context.Context, userID int64) (map[string]interface{}, error) {
	days := 30
	start := time.Now().AddDate(0, 0, -(days - 1)).Format("2006-01-02")

	// —— 我的指标（近30天）——
	var myUploads, myViews, myLikes, myFollowers int64
	r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT m.id),
		       COALESCE(SUM(m.view_count),0),
		       COALESCE(SUM(m.like_count),0),
		       (SELECT COUNT(*) FROM follows f WHERE f.following_id = $1 AND DATE(f.created_at) >= $2)
		  FROM manuscripts m
		 WHERE m.user_id = $1 AND m.status = 3`,
		userID, start).Scan(&myUploads, &myViews, &myLikes, &myFollowers)

	// —— 同类UP主均值（其他UP主同期）——
	var peerUploads, peerViews, peerLikes, peerFollowers float64
	var peerCount float64
	peers, err := r.db.QueryContext(ctx, `
		SELECT COUNT(DISTINCT m.id),
		       COALESCE(SUM(m.view_count),0),
		       COALESCE(SUM(m.like_count),0),
		       (SELECT COUNT(*) FROM follows f WHERE f.following_id = m.user_id AND DATE(f.created_at) >= $2)
		  FROM manuscripts m
		 WHERE m.user_id <> $1 AND m.status = 3
		 GROUP BY m.user_id`, userID, start)
	if err == nil {
		defer peers.Close()
		for peers.Next() {
			var u, v, l, fo int64
			if err := peers.Scan(&u, &v, &l, &fo); err == nil {
				peerUploads += float64(u)
				peerViews += float64(v)
				peerLikes += float64(l)
				peerFollowers += float64(fo)
				peerCount++
			}
		}
	}
	if peerCount > 0 {
		peerUploads /= peerCount
		peerViews /= peerCount
		peerLikes /= peerCount
		peerFollowers /= peerCount
	}

	// —— 人群分布（近30天，游客=未关注观看者 / 粉丝=已关注观看者）——
	var guests, fans int64
	r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE f.follower_id IS NULL),
			COUNT(*) FILTER (WHERE f.follower_id IS NOT NULL)
		  FROM watch_history wh
		  JOIN manuscripts m ON m.id = wh.manuscript_id
		  LEFT JOIN follows f ON f.follower_id = wh.user_id AND f.following_id = m.user_id
		 WHERE m.user_id = $1 AND m.status = 3 AND DATE(wh.watched_at) >= $2`,
		userID, start).Scan(&guests, &fans)
	totalAudience := guests + fans
	if totalAudience == 0 {
		guests, fans = 1, 0
		totalAudience = 1
	}

	// —— 观看途径（近30天，按客户端类型）——
	termRows, err := r.db.QueryContext(ctx, `
		SELECT wh.client_type, COUNT(*)
		  FROM watch_history wh
		  JOIN manuscripts m ON m.id = wh.manuscript_id
		 WHERE m.user_id = $1 AND m.status = 3 AND DATE(wh.watched_at) >= $2
		 GROUP BY wh.client_type`, userID, start)
	termMap := map[string]int64{}
	if err == nil {
		defer termRows.Close()
		for termRows.Next() {
			var ct string
			var cnt int64
			if err := termRows.Scan(&ct, &cnt); err == nil {
				termMap[ct] += cnt
			}
		}
	}
	terminalOrder := []struct{ key, label string }{
		{"mobile", "移动端"}, {"pc", "PC"}, {"tv", "云视听小电视"}, {"h5", "h5及小程序"}, {"other", "其他"},
	}
	terminals := []terminalItem{}
	var termTotal int64
	for _, k := range terminalOrder {
		termTotal += termMap[k.key]
	}
	for _, k := range terminalOrder {
		v := termMap[k.key]
		var pct float64
		if termTotal > 0 {
			pct = float64(v) / float64(termTotal)
		}
		terminals = append(terminals, terminalItem{Key: k.key, Label: k.label, Value: v, Percent: pct})
	}

	// —— 来源稿件占比（最近有数据日，各稿件当日播放占比）——
	var lastDate string
	r.db.QueryRowContext(ctx, `
		SELECT MAX(metric_date)::text FROM manuscript_daily_metrics WHERE user_id = $1`, userID).Scan(&lastDate)
	srcRows, err := r.db.QueryContext(ctx, `
		SELECT dm.manuscript_id, m.title, dm.view_count
		  FROM manuscript_daily_metrics dm
		  JOIN manuscripts m ON m.id = dm.manuscript_id
		 WHERE dm.user_id = $1 AND dm.metric_date = $2 AND dm.metric_date IS NOT NULL
		 ORDER BY dm.view_count DESC LIMIT 15`, userID, lastDate)
	sources := []map[string]interface{}{}
	var srcTotal int64
	var srcIDs []int64
	var srcTitles []string
	var srcPlays []int64
	if err == nil {
		defer srcRows.Close()
		srcIDs = []int64{}
		srcTitles = []string{}
		srcPlays = []int64{}
		for srcRows.Next() {
			var mid int64
			var title sql.NullString
			var v int64
			if err := srcRows.Scan(&mid, &title, &v); err == nil {
				srcIDs = append(srcIDs, mid)
				srcTitles = append(srcTitles, title.String)
				srcPlays = append(srcPlays, v)
				srcTotal += v
			}
		}
	}
	if lastDate == "" {
		lastDate = time.Now().Format("2006-01-02")
	}
	for i := range srcIDs {
		var pct float64
		if srcTotal > 0 {
			pct = float64(srcPlays[i]) / float64(srcTotal)
		}
		sources = append(sources, map[string]interface{}{
			"manuscriptId": srcIDs[i],
			"title":        srcTitles[i],
			"plays":        srcPlays[i],
			"percent":      pct,
		})
	}

	return map[string]interface{}{
		"metrics": map[string]interface{}{
			"uploads": myUploads, "views": myViews, "likes": myLikes, "newFollowers": myFollowers,
			"peerUploads": peerUploads, "peerViews": peerViews, "peerLikes": peerLikes, "peerNewFollowers": peerFollowers,
		},
		"personDistribution": map[string]interface{}{
			"guestPercent": float64(guests) / float64(totalAudience),
			"fanPercent":   float64(fans) / float64(totalAudience),
		},
		"terminals":     terminals,
		"sources":       sources,
		"sourceDate":    lastDate,
		"periodDays":    days,
		"updateTime":    time.Now().Format("2006-01-02"),
		"lastDateLabel": lastDate,
	}, nil
}

// Income 收益数据：期间总收益 + 自定义充电收益（分）。无充值能力时真实为 0。
func (r *Repository) Income(ctx context.Context, userID int64, days int) (map[string]interface{}, error) {
	start := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	var totalCents, chargeCents int64
	r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents),0) FROM income_records WHERE user_id = $1 AND DATE(created_at) >= $2`,
		userID, start).Scan(&totalCents)
	r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents),0) FROM income_records WHERE user_id = $1 AND type = $2 AND DATE(created_at) >= $3`,
		userID, "custom_charge", start).Scan(&chargeCents)
	return map[string]interface{}{
		"totalAmount":  float64(totalCents) / 100,
		"chargeAmount": float64(chargeCents) / 100,
		"periodDays":   days,
		"updateTime":   time.Now().Format("2006-01-02 15:04:05"),
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

func (s *Service) Overview(ctx context.Context, userID int64, days int) (map[string]interface{}, error) {
	if days != 7 && days != 30 {
		days = 7
	}
	return s.repo.Overview(ctx, userID, days)
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

// FansByManuscript 各稿件带来的涨粉数（来源归因，历史累积），返回 manuscript_id -> 涨粉数。
func (s *Service) FansByManuscript(ctx context.Context, userID int64) (map[string]interface{}, error) {
	return s.repo.FansByManuscript(ctx, userID)
}

// ManuscriptInsights 各稿件洞察 + 粉丝播放趋势。
func (s *Service) ManuscriptInsights(ctx context.Context, userID int64, days int) (map[string]interface{}, error) {
	if days != 7 && days != 30 {
		days = 7
	}
	return s.repo.ManuscriptInsights(ctx, userID, days)
}

// Income 收益数据：总收益与自定义充电（期间内，分）。
func (s *Service) Income(ctx context.Context, userID int64, days int) (map[string]interface{}, error) {
	return s.repo.Income(ctx, userID, days)
}

// AccountDiagnosis 账号诊断：表现总结 + 播放分布。
func (s *Service) AccountDiagnosis(ctx context.Context, userID int64) (map[string]interface{}, error) {
	return s.repo.AccountDiagnosis(ctx, userID)
}
