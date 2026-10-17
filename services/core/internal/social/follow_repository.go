package social

import (
	"context"
	"database/sql"
	"strconv"
	"time"
)

type Follow struct {
	ID          int64
	FollowerID  int64
	FollowingID int64
	CreatedAt   time.Time
}

type FollowRepository struct {
	db *sql.DB
}

func NewFollowRepository(db *sql.DB) *FollowRepository {
	return &FollowRepository{db: db}
}

func (r *FollowRepository) Follow(ctx context.Context, followerID, followingID int64) error {
	return r.FollowWithSource(ctx, followerID, followingID, 0)
}

// FollowWithSource 带稿件来源的关注（数据中心涨粉归因），sourceManuscriptID<=0 表示未知来源。
func (r *FollowRepository) FollowWithSource(ctx context.Context, followerID, followingID, sourceManuscriptID int64) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		var res sql.Result
		var err error
		if sourceManuscriptID > 0 {
			res, err = tx.ExecContext(ctx,
				`INSERT INTO follows (follower_id, following_id, source_manuscript_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
				followerID, followingID, sourceManuscriptID)
		} else {
			res, err = tx.ExecContext(ctx,
				`INSERT INTO follows (follower_id, following_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				followerID, followingID)
		}
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			return incrCountsTx(ctx, tx, followerID, followingID, 1)
		}
		return nil
	})
}

func (r *FollowRepository) Unfollow(ctx context.Context, followerID, followingID int64) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx,
			`DELETE FROM follows WHERE follower_id = $1 AND following_id = $2`, followerID, followingID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			return incrCountsTx(ctx, tx, followerID, followingID, -1)
		}
		return nil
	})
}

func (r *FollowRepository) withTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *FollowRepository) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2)`,
		followerID, followingID).Scan(&exists)
	return exists, err
}

func (r *FollowRepository) ListFollowing(ctx context.Context, userID int64, page, pageSize int32) ([]int64, error) {
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx,
		`SELECT following_id FROM follows WHERE follower_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, offset)
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

func (r *FollowRepository) ListFollowers(ctx context.Context, userID int64, page, pageSize int32) ([]int64, error) {
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx,
		`SELECT follower_id FROM follows WHERE following_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, offset)
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

func (r *FollowRepository) FollowingCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM follows WHERE follower_id = $1`, userID).Scan(&count)
	return count, err
}

func (r *FollowRepository) FollowerCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM follows WHERE following_id = $1`, userID).Scan(&count)
	return count, err
}

// ListFans 返回粉丝条目（含关注时间），支持互关筛选与总数。
type FanEntry struct {
	FollowerID int64
	FollowedAt time.Time
	IsMutual   bool
}

func (r *FollowRepository) ListFans(ctx context.Context, userID int64, page, pageSize int32, mutual bool) ([]FanEntry, int64, error) {
	offset := (page - 1) * pageSize
	var total int64
	var rows *sql.Rows
	var err error
	if mutual {
		err = r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM follows f
			 WHERE f.following_id = $1
			   AND EXISTS(SELECT 1 FROM follows f2 WHERE f2.follower_id = f.following_id AND f2.following_id = f.follower_id)`,
			userID).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
		rows, err = r.db.QueryContext(ctx,
			`SELECT f.follower_id, f.created_at, true FROM follows f
			 WHERE f.following_id = $1
			   AND EXISTS(SELECT 1 FROM follows f2 WHERE f2.follower_id = f.following_id AND f2.following_id = f.follower_id)
			 ORDER BY f.created_at DESC LIMIT $2 OFFSET $3`,
			userID, pageSize, offset)
	} else {
		err = r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM follows WHERE following_id = $1`, userID).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
		rows, err = r.db.QueryContext(ctx,
			`SELECT f.follower_id, f.created_at,
			        EXISTS(SELECT 1 FROM follows f2 WHERE f2.follower_id = f.following_id AND f2.following_id = f.follower_id)
			 FROM follows f WHERE f.following_id = $1
			 ORDER BY f.created_at DESC LIMIT $2 OFFSET $3`,
			userID, pageSize, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []FanEntry
	for rows.Next() {
		var e FanEntry
		if err := rows.Scan(&e.FollowerID, &e.FollowedAt, &e.IsMutual); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, nil
}

type FansStats struct {
	TotalFans    int64
	MutualFans   int64
	NewFansToday int64
	NewFansWeek  int64
	NewFansMonth int64
}

func (r *FollowRepository) FansStats(ctx context.Context, userID int64) (FansStats, error) {
	var s FansStats
	err := r.db.QueryRowContext(ctx,
		`SELECT
		 COUNT(*) FILTER (WHERE f.following_id = $1) AS total,
		 COUNT(*) FILTER (WHERE f.following_id = $1
		   AND EXISTS(SELECT 1 FROM follows f2 WHERE f2.follower_id = f.following_id AND f2.following_id = f.follower_id)) AS mutual,
		 COUNT(*) FILTER (WHERE f.following_id = $1 AND f.created_at >= CURRENT_DATE) AS new_today,
		 COUNT(*) FILTER (WHERE f.following_id = $1 AND f.created_at >= CURRENT_DATE - INTERVAL '7 days') AS new_week,
		 COUNT(*) FILTER (WHERE f.following_id = $1 AND f.created_at >= CURRENT_DATE - INTERVAL '30 days') AS new_month
		 FROM follows f`,
		userID).Scan(&s.TotalFans, &s.MutualFans, &s.NewFansToday, &s.NewFansWeek, &s.NewFansMonth)
	return s, err
}

func incrCountsTx(ctx context.Context, tx *sql.Tx, followerID, followingID int64, delta int) error {
	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET following_count = GREATEST(following_count + $1, 0) WHERE id = $2`, delta, followerID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE users SET follower_count = GREATEST(follower_count + $1, 0) WHERE id = $2`, delta, followingID)
	return err
}

func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
