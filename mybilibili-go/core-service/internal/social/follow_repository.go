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
	return r.withTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx,
			`INSERT INTO follows (follower_id, following_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			followerID, followingID)
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
