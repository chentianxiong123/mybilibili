package interaction

import (
	"context"
	"database/sql"
	"time"

	pb "mybilibili/pkg/pb"
)

type InteractionRepository struct {
	db *sql.DB
}

func NewInteractionRepository(db *sql.DB) *InteractionRepository {
	return &InteractionRepository{db: db}
}

func (r *InteractionRepository) HasInteraction(ctx context.Context, userID int64, targetType, interactionType string, targetID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_interactions WHERE user_id = $1 AND target_type = $2 AND target_id = $3 AND interaction_type = $4`,
		userID, targetType, targetID, interactionType).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *InteractionRepository) AddInteraction(ctx context.Context, userID int64, targetType, interactionType string, targetID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_interactions (user_id, target_type, target_id, interaction_type) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, target_type, target_id, interaction_type) DO NOTHING`,
		userID, targetType, targetID, interactionType)
	return err
}

func (r *InteractionRepository) RemoveInteraction(ctx context.Context, userID int64, targetType, interactionType string, targetID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM user_interactions WHERE user_id = $1 AND target_type = $2 AND target_id = $3 AND interaction_type = $4`,
		userID, targetType, targetID, interactionType)
	return err
}

func (r *InteractionRepository) CountInteraction(ctx context.Context, targetType, interactionType string, targetID int64) (int32, error) {
	var count int32
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_interactions WHERE target_type = $1 AND target_id = $2 AND interaction_type = $3`,
		targetType, targetID, interactionType).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *InteractionRepository) GetInteractionIDs(ctx context.Context, userID int64, targetType, interactionType string) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT target_id FROM user_interactions WHERE user_id = $1 AND target_type = $2 AND interaction_type = $3 ORDER BY created_at DESC`,
		userID, targetType, interactionType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *InteractionRepository) Follow(ctx context.Context, followerID, followingID int64) (bool, error) {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO follows (follower_id, following_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		followerID, followingID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *InteractionRepository) Unfollow(ctx context.Context, followerID, followingID int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM follows WHERE follower_id = $1 AND following_id = $2`, followerID, followingID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *InteractionRepository) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM follows WHERE follower_id = $1 AND following_id = $2`, followerID, followingID).Scan(&count)
	return count > 0, err
}

func (r *InteractionRepository) CountFollows(ctx context.Context, userID int64) (following, followers int32, err error) {
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM follows WHERE follower_id = $1`, userID).Scan(&following)
	if err != nil {
		return
	}
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM follows WHERE following_id = $1`, userID).Scan(&followers)
	return
}

func (r *InteractionRepository) UpsertWatchHistory(ctx context.Context, userID, manuscriptID int64, progressSeconds int32) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO watch_history (user_id, manuscript_id, progress_seconds, watched_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (user_id, manuscript_id) DO UPDATE SET progress_seconds = $3, watched_at = NOW()`,
		userID, manuscriptID, progressSeconds)
	return err
}

func (r *InteractionRepository) GetWatchHistory(ctx context.Context, userID int64, page, pageSize int32) ([]*pb.WatchHistoryItem, error) {
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx,
		`SELECT m.id, m.title, m.cover_url, wh.progress_seconds, m.duration_seconds, wh.watched_at
		 FROM watch_history wh INNER JOIN manuscripts m ON m.id = wh.manuscript_id
		 WHERE wh.user_id = $1 ORDER BY wh.watched_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*pb.WatchHistoryItem
	for rows.Next() {
		item := &pb.WatchHistoryItem{}
		var watchedAt time.Time
		if err := rows.Scan(&item.ManuscriptId, &item.Title, &item.CoverUrl, &item.ProgressSeconds, &item.DurationSeconds, &watchedAt); err != nil {
			return nil, err
		}
		item.WatchedAt = watchedAt.Format("2006-01-02T15:04:05Z")
		items = append(items, item)
	}
	return items, nil
}

func (r *InteractionRepository) ClearWatchHistory(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM watch_history WHERE user_id = $1`, userID)
	return err
}

func (r *InteractionRepository) GetManuscriptTitle(ctx context.Context, id int64) (string, error) {
	var title string
	err := r.db.QueryRowContext(ctx, `SELECT title FROM manuscripts WHERE id = $1`, id).Scan(&title)
	return title, err
}

func (r *InteractionRepository) IncrementManuscriptCount(ctx context.Context, field string, manuscriptID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE manuscripts SET `+field+` = `+field+` + 1 WHERE id = $1`, manuscriptID)
	return err
}

func (r *InteractionRepository) DecrementManuscriptCount(ctx context.Context, field string, manuscriptID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE manuscripts SET `+field+` = GREATEST(`+field+` - 1, 0) WHERE id = $1`, manuscriptID)
	return err
}

// UpsertDailyMetric 当日指标累加（对齐旧版 analytics 每日聚合）。
// manuscript_daily_metrics 主键 (metric_date, manuscript_id)，存在则累加对应字段。
func (r *InteractionRepository) UpsertDailyMetric(ctx context.Context, manuscriptID, userID int64, field string, delta int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO manuscript_daily_metrics (metric_date, manuscript_id, user_id, `+field+`)
		 VALUES (CURRENT_DATE, $1, $2, $3)
		 ON CONFLICT (metric_date, manuscript_id)
		 DO UPDATE SET `+field+` = manuscript_daily_metrics.`+field+` + $3, updated_at = NOW()`,
		manuscriptID, userID, delta)
	return err
}

func (r *InteractionRepository) GetUserCoinCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT coin_count FROM users WHERE id = $1`, userID).Scan(&count)
	return count, err
}

func (r *InteractionRepository) DeductCoin(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET coin_count = coin_count - 1 WHERE id = $1 AND coin_count > 0`, userID)
	return err
}
