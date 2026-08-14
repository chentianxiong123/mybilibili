package social

import (
	"context"
	"database/sql"
	"time"
)

type Dynamic struct {
	ID              int64
	UserID          int64
	Content         string
	DynamicType     int32
	ImageURL        string
	RefManuscriptID int64
	LikeCount       int32
	CommentCount    int32
	ShareCount      int32
	Status          int32
	CreatedAt       time.Time
}

type DynamicComment struct {
	ID          int64
	DynamicID   int64
	UserID      int64
	Content     string
	ParentID    int64
	ReplyUserID int64
	LikeCount   int32
	Status      int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DynamicRepository struct {
	db *sql.DB
}

func NewDynamicRepository(db *sql.DB) *DynamicRepository {
	return &DynamicRepository{db: db}
}

func (r *DynamicRepository) Create(ctx context.Context, d *Dynamic) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO user_dynamics (user_id, content, dynamic_type, image_url, ref_manuscript_id)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		d.UserID, d.Content, d.DynamicType, d.ImageURL, nullInt64(d.RefManuscriptID)).Scan(&id)
	return id, err
}

func (r *DynamicRepository) GetByID(ctx context.Context, id int64) (*Dynamic, error) {
	d := &Dynamic{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, content, dynamic_type, COALESCE(image_url,''), COALESCE(ref_manuscript_id,0),
		        like_count, comment_count, share_count, status, created_at
		 FROM user_dynamics WHERE id = $1`, id,
	).Scan(&d.ID, &d.UserID, &d.Content, &d.DynamicType, &d.ImageURL, &d.RefManuscriptID,
		&d.LikeCount, &d.CommentCount, &d.ShareCount, &d.Status, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *DynamicRepository) ListByUser(ctx context.Context, userID int64, page, limit int32) ([]*Dynamic, error) {
	offset := (page - 1) * limit
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, content, dynamic_type, COALESCE(image_url,''), COALESCE(ref_manuscript_id,0),
		        like_count, comment_count, share_count, status, created_at
		 FROM user_dynamics WHERE user_id = $1 AND status = 0 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDynamics(rows)
}

func (r *DynamicRepository) ListFollowing(ctx context.Context, userID int64, page, limit int32) ([]*Dynamic, error) {
	offset := (page - 1) * limit
	rows, err := r.db.QueryContext(ctx,
		`SELECT d.id, d.user_id, d.content, d.dynamic_type, COALESCE(d.image_url,''), COALESCE(d.ref_manuscript_id,0),
		        d.like_count, d.comment_count, d.share_count, d.status, d.created_at
		 FROM user_dynamics d JOIN follows f ON d.user_id = f.following_id
		 WHERE f.follower_id = $1 AND d.status = 0 ORDER BY d.created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDynamics(rows)
}

func (r *DynamicRepository) ListAll(ctx context.Context, page, limit int32) ([]*Dynamic, error) {
	offset := (page - 1) * limit
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, content, dynamic_type, COALESCE(image_url,''), COALESCE(ref_manuscript_id,0),
		        like_count, comment_count, share_count, status, created_at
		 FROM user_dynamics WHERE status = 0 ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDynamics(rows)
}

func (r *DynamicRepository) Delete(ctx context.Context, id, userID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE user_dynamics SET status = 1 WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *DynamicRepository) IncrLikeCount(ctx context.Context, id int64, delta int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE user_dynamics SET like_count = GREATEST(like_count + $1, 0) WHERE id = $2`, delta, id)
	return err
}

func (r *DynamicRepository) IncrCommentCount(ctx context.Context, id int64, delta int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE user_dynamics SET comment_count = GREATEST(comment_count + $1, 0) WHERE id = $2`, delta, id)
	return err
}

func (r *DynamicRepository) IncrShareCount(ctx context.Context, id int64, delta int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE user_dynamics SET share_count = GREATEST(share_count + $1, 0) WHERE id = $2`, delta, id)
	return err
}

func (r *DynamicRepository) IsLiked(ctx context.Context, dynamicID, userID int64) (bool, error) {
	var cnt int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM dynamic_likes WHERE dynamic_id = $1 AND user_id = $2`,
		dynamicID, userID).Scan(&cnt)
	return cnt > 0, err
}

func (r *DynamicRepository) CreateComment(ctx context.Context, dc *DynamicComment) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO dynamic_comments (dynamic_id, user_id, content, parent_id, reply_user_id)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		dc.DynamicID, dc.UserID, dc.Content, nullInt64(dc.ParentID), nullInt64(dc.ReplyUserID)).Scan(&id)
	return id, err
}

func (r *DynamicRepository) ListComments(ctx context.Context, dynamicID int64, page, limit int32) ([]*DynamicComment, error) {
	offset := (page - 1) * limit
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, dynamic_id, user_id, content, COALESCE(parent_id,0), COALESCE(reply_user_id,0),
		        like_count, status, created_at, updated_at
		 FROM dynamic_comments WHERE dynamic_id = $1 AND status = 0 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		dynamicID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*DynamicComment
	for rows.Next() {
		dc := &DynamicComment{}
		if err := rows.Scan(&dc.ID, &dc.DynamicID, &dc.UserID, &dc.Content, &dc.ParentID, &dc.ReplyUserID,
			&dc.LikeCount, &dc.Status, &dc.CreatedAt, &dc.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, dc)
	}
	return list, nil
}

func (r *DynamicRepository) ListReplies(ctx context.Context, commentID int64, page, limit int32) ([]*DynamicComment, error) {
	offset := (page - 1) * limit
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, dynamic_id, user_id, content, COALESCE(parent_id,0), COALESCE(reply_user_id,0),
		        like_count, status, created_at, updated_at
		 FROM dynamic_comments WHERE parent_id = $1 AND status = 0 ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
		commentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*DynamicComment
	for rows.Next() {
		dc := &DynamicComment{}
		if err := rows.Scan(&dc.ID, &dc.DynamicID, &dc.UserID, &dc.Content, &dc.ParentID, &dc.ReplyUserID,
			&dc.LikeCount, &dc.Status, &dc.CreatedAt, &dc.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, dc)
	}
	return list, nil
}

func (r *DynamicRepository) DeleteComment(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE dynamic_comments SET status = 1 WHERE id = $1`, id)
	return err
}

func scanDynamics(rows *sql.Rows) ([]*Dynamic, error) {
	defer rows.Close()
	var list []*Dynamic
	for rows.Next() {
		d := &Dynamic{}
		if err := rows.Scan(&d.ID, &d.UserID, &d.Content, &d.DynamicType, &d.ImageURL, &d.RefManuscriptID,
			&d.LikeCount, &d.CommentCount, &d.ShareCount, &d.Status, &d.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, nil
}

func nullInt64(v int64) interface{} {
	if v == 0 {
		return nil
	}
	return v
}
