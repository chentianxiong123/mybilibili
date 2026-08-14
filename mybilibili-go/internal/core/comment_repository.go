package core

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	pb "mybilibili/internal/core/pb"
)

type Comment struct {
	ID           int64
	ManuscriptID int64
	UserID       int64
	Content      string
	LikeCount    int32
	ReplyCount   int32
	Status       int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Reply struct {
	ID            int64
	CommentID     int64
	UserID        int64
	ReplyToUserID sql.NullInt64
	Content       string
	LikeCount     int32
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) CreateComment(ctx context.Context, c *Comment) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO comments (manuscript_id, user_id, content) VALUES ($1, $2, $3) RETURNING id`,
		c.ManuscriptID, c.UserID, c.Content).Scan(&id)
	return id, err
}

// UpsertDailyMetric 当日评论/弹幕指标累加（对齐旧版 analytics 每日聚合）。
func (r *CommentRepository) UpsertDailyMetric(ctx context.Context, manuscriptID, userID int64, field string, delta int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO manuscript_daily_metrics (metric_date, manuscript_id, user_id, `+field+`)
		 VALUES (CURRENT_DATE, $1, $2, $3)
		 ON CONFLICT (metric_date, manuscript_id)
		 DO UPDATE SET `+field+` = manuscript_daily_metrics.`+field+` + $3, updated_at = NOW()`,
		manuscriptID, userID, delta)
	return err
}

func (r *CommentRepository) FindByID(ctx context.Context, id int64) (*Comment, error) {
	c := &Comment{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, manuscript_id, user_id, content, like_count, reply_count, status, created_at, updated_at
		 FROM comments WHERE id = $1`, id,
	).Scan(&c.ID, &c.ManuscriptID, &c.UserID, &c.Content, &c.LikeCount, &c.ReplyCount, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CommentRepository) ListByManuscript(ctx context.Context, manuscriptID int64, page, pageSize int32, sort string) ([]*Comment, error) {
	order := "ORDER BY created_at DESC"
	if sort == "hot" {
		order = "ORDER BY like_count DESC, created_at DESC"
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, manuscript_id, user_id, content, like_count, reply_count, status, created_at, updated_at
		 FROM comments WHERE manuscript_id = $1 AND status = 0 `+order+` LIMIT $2 OFFSET $3`,
		manuscriptID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Comment
	for rows.Next() {
		c := &Comment{}
		if err := rows.Scan(&c.ID, &c.ManuscriptID, &c.UserID, &c.Content, &c.LikeCount, &c.ReplyCount, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *CommentRepository) Delete(ctx context.Context, id, userID int64) error {
	res, err := r.db.ExecContext(ctx, `UPDATE comments SET status = 1 WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *CommentRepository) CreateReply(ctx context.Context, rep *Reply) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO replies (comment_id, user_id, reply_to_user_id, content) VALUES ($1, $2, $3, $4) RETURNING id`,
		rep.CommentID, rep.UserID, nullInt64(rep.ReplyToUserID), rep.Content).Scan(&id)
	return id, err
}

func (r *CommentRepository) FindReplyByID(ctx context.Context, id int64) (*Reply, error) {
	rep := &Reply{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, comment_id, user_id, reply_to_user_id, content, like_count, status, created_at, updated_at
		 FROM replies WHERE id = $1`, id,
	).Scan(&rep.ID, &rep.CommentID, &rep.UserID, &rep.ReplyToUserID, &rep.Content, &rep.LikeCount, &rep.Status, &rep.CreatedAt, &rep.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return rep, nil
}

func (r *CommentRepository) ListRepliesByComment(ctx context.Context, commentID int64, page, pageSize int32) ([]*Reply, error) {
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, comment_id, user_id, reply_to_user_id, content, like_count, status, created_at, updated_at
		 FROM replies WHERE comment_id = $1 AND status = 'NORMAL' ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
		commentID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Reply
	for rows.Next() {
		rep := &Reply{}
		if err := rows.Scan(&rep.ID, &rep.CommentID, &rep.UserID, &rep.ReplyToUserID, &rep.Content, &rep.LikeCount, &rep.Status, &rep.CreatedAt, &rep.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, rep)
	}
	return list, nil
}

func (r *CommentRepository) DeleteReply(ctx context.Context, id, userID int64) error {
	res, err := r.db.ExecContext(ctx, `UPDATE replies SET status = 'REMOVED' WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *CommentRepository) IncrementReplyCount(ctx context.Context, commentID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE comments SET reply_count = reply_count + 1 WHERE id = $1`, commentID)
	return err
}

func (r *CommentRepository) FindUserByID(ctx context.Context, userID int64) (*User, error) {
	u := &User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, nickname, avatar, level FROM users WHERE id = $1`, userID,
	).Scan(&u.ID, &u.Username, &u.Nickname, &u.Avatar, &u.Level)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *CommentRepository) IsCommentLiked(ctx context.Context, commentID, userID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_interactions WHERE target_type = 'COMMENT' AND target_id = $1 AND user_id = $2 AND interaction_type = 'LIKE'`,
		commentID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CommentRepository) IsReplyLiked(ctx context.Context, replyID, userID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_interactions WHERE target_type = 'REPLY' AND target_id = $1 AND user_id = $2 AND interaction_type = 'LIKE'`,
		replyID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CommentRepository) LikeTarget(ctx context.Context, targetType string, targetID, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_interactions (user_id, target_type, target_id, interaction_type) VALUES ($1, $2, $3, 'like')
		 ON CONFLICT (user_id, target_type, target_id, interaction_type) DO NOTHING`,
		userID, targetType, targetID)
	if err != nil {
		return err
	}
	if targetType == "COMMENT" {
		_, err = r.db.ExecContext(ctx, `UPDATE comments SET like_count = like_count + 1 WHERE id = $1`, targetID)
	} else if targetType == "REPLY" {
		_, err = r.db.ExecContext(ctx, `UPDATE replies SET like_count = like_count + 1 WHERE id = $1`, targetID)
	}
	return err
}

func (r *CommentRepository) UnlikeTarget(ctx context.Context, targetType string, targetID, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM user_interactions WHERE user_id = $1 AND target_type = $2 AND target_id = $3 AND interaction_type = 'LIKE'`,
		userID, targetType, targetID)
	if err != nil {
		return err
	}
	if targetType == "COMMENT" {
		_, err = r.db.ExecContext(ctx, `UPDATE comments SET like_count = GREATEST(like_count - 1, 0) WHERE id = $1`, targetID)
	} else if targetType == "REPLY" {
		_, err = r.db.ExecContext(ctx, `UPDATE replies SET like_count = GREATEST(like_count - 1, 0) WHERE id = $1`, targetID)
	}
	return err
}

// BatchGetLikeCounts 批量返回评论/回复的点赞数（对齐旧版 batchGetLikeCount）。
func (r *CommentRepository) BatchGetLikeCounts(ctx context.Context, targetType string, ids []int64) (map[int64]int64, error) {
	out := make(map[int64]int64, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	table := "comments"
	if targetType == "REPLY" {
		table = "replies"
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, like_count FROM `+table+` WHERE id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, count int64
		rows.Scan(&id, &count)
		out[id] = count
	}
	return out, nil
}

// BatchIsLiked 批量判断用户对评论/回复的点赞态（对齐旧版 batchIsLiked）。
func (r *CommentRepository) BatchIsLiked(ctx context.Context, userID int64, targetType string, ids []int64) (map[int64]bool, error) {
	out := make(map[int64]bool, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, 0, len(ids)+2)
	args = append(args, userID, targetType)
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+3)
		args = append(args, id)
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT target_id FROM user_interactions
		 WHERE user_id = $1 AND target_type = $2 AND interaction_type = 'like'
		  AND target_id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		out[id] = true
	}
	return out, nil
}

func commentToPB(c *Comment, userName, userAvatar string, userLevel int32, liked bool, replies []*pb.ReplyInfo) *pb.CommentInfo {
	return &pb.CommentInfo{
		Id:           c.ID,
		ManuscriptId: c.ManuscriptID,
		UserId:       c.UserID,
		UserName:     userName,
		UserAvatar:   userAvatar,
		UserLevel:    userLevel,
		Content:      c.Content,
		LikeCount:    c.LikeCount,
		ReplyCount:   c.ReplyCount,
		CreatedAt:    c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Liked:        liked,
		Replies:      replies,
	}
}

func replyToPB(rep *Reply, userName, userAvatar string, userLevel int32, replyToUserName string, liked bool) *pb.ReplyInfo {
	return &pb.ReplyInfo{
		Id:              rep.ID,
		CommentId:       rep.CommentID,
		UserId:          rep.UserID,
		UserName:        userName,
		UserAvatar:      userAvatar,
		UserLevel:       userLevel,
		Content:         rep.Content,
		LikeCount:       rep.LikeCount,
		CreatedAt:       rep.CreatedAt.Format("2006-01-02T15:04:05Z"),
		ReplyToUserName: replyToUserName,
		Liked:           liked,
	}
}

func nullInt64(n sql.NullInt64) *int64 {
	if n.Valid {
		return &n.Int64
	}
	return nil
}

func (r *CommentRepository) ListCommentsByCreator(ctx context.Context, userID, manuscriptID int64, page, pageSize int32, sort, commentType string) ([]*Comment, error) {
	where := `JOIN manuscripts m ON m.id = comments.manuscript_id WHERE m.user_id = $1`
	args := []any{userID}
	if manuscriptID > 0 {
		where += ` AND comments.manuscript_id = $2`
		args = append(args, manuscriptID)
	}
	if commentType == "reply" {
		where += ` AND comments.reply_count > 0`
	}
	order := `comments.created_at DESC`
	if sort == "latest" {
		order = `comments.created_at DESC`
	} else if sort == "oldest" {
		order = `comments.created_at ASC`
	} else if sort == "likes" {
		order = `comments.like_count DESC`
	}
	offset := int64(page-1) * int64(pageSize)
	query := `SELECT comments.id, comments.manuscript_id, comments.user_id, comments.content, comments.like_count, comments.reply_count, comments.status, comments.created_at, comments.updated_at
	          FROM comments ` + where + ` ORDER BY ` + order + ` LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
	args = append(args, pageSize, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Comment
	for rows.Next() {
		c := &Comment{}
		if err := rows.Scan(&c.ID, &c.ManuscriptID, &c.UserID, &c.Content, &c.LikeCount, &c.ReplyCount, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *CommentRepository) CountCommentsByCreator(ctx context.Context, userID, manuscriptID int64, commentType string) (int64, error) {
	where := `JOIN manuscripts m ON m.id = comments.manuscript_id WHERE m.user_id = $1`
	args := []any{userID}
	if manuscriptID > 0 {
		where += ` AND comments.manuscript_id = $2`
		args = append(args, manuscriptID)
	}
	if commentType == "reply" {
		where += ` AND comments.reply_count > 0`
	}
	var total int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM comments `+where, args...).Scan(&total)
	return total, err
}

func (r *CommentRepository) DeleteCommentByCreator(ctx context.Context, commentID, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM comments
		 WHERE id = $1 AND manuscript_id IN (
		     SELECT id FROM manuscripts WHERE user_id = $2)`,
		commentID, userID)
	return err
}

func (r *CommentRepository) WriteContentReview(ctx context.Context, typ string, userID int64, content string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO content_reviews (type, user_id, content) VALUES ($1, $2, $3)`, typ, userID, content)
	return err
}

func (r *CommentRepository) UpdateCommentStatus(ctx context.Context, id int64, status int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE comments SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (r *CommentRepository) DeleteReplyByCreator(ctx context.Context, replyID, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM replies
		 WHERE id = $1 AND comment_id IN (
		     SELECT c.id FROM comments c
		     JOIN manuscripts m ON m.id = c.manuscript_id
		     WHERE m.user_id = $2)`,
		replyID, userID)
	return err
}
