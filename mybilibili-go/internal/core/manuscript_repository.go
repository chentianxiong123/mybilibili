package core

import (
	"context"
	"database/sql"

	pb "mybilibili/internal/core/pb"
)

type Manuscript struct {
	ID              int64
	Title           string
	Description     string
	CoverURL        string
	UserID          int64
	CategoryID      int64
	ViewCount       int64
	LikeCount       int64
	CoinCount       int64
	CollectCount    int64
	ShareCount      int64
	CommentCount    int64
	DanmakuCount    int64
	Status          int32
	ReviewStatus    int32
	ReviewReason    string
	ReviewTime      sql.NullTime
	ReviewerID      sql.NullInt64
	UploadTime      sql.NullTime
	UpdatedAt       sql.NullTime
	Duration        string
	DurationSeconds int32
}

type Video struct {
	ID              int64
	ManuscriptID    int64
	VideoOrder      int32
	Title           string
	Description     string
	PlayURLHd       string
	PlayURLSd       string
	PlayURLld       string
	UploadTime      sql.NullTime
	UpdatedAt       sql.NullTime
	ProcessProgress int32
	ProcessStage    string
	HasSubtitle     int32
	HasSummary      int32
	ProcessStatus   int32
	ProcessError    string
	SourceVideoURL  string
	DurationSeconds int32
}

type Category struct {
	ID        int64
	Name      string
	Icon      string
	SortOrder int32
}

type Tag struct {
	ID   int64
	Name string
}

type ManuscriptRepository struct {
	db *sql.DB
}

func NewManuscriptRepository(db *sql.DB) *ManuscriptRepository {
	return &ManuscriptRepository{db: db}
}

// IncrementViewCount 播放计数 +1（对齐旧版 ManuscriptServiceImpl.incrementViewCount）。
func (r *ManuscriptRepository) IncrementViewCount(ctx context.Context, manuscriptID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE manuscripts SET view_count = view_count + 1, updated_at = NOW() WHERE id = $1`, manuscriptID)
	return err
}

// UpsertDailyMetric 当日观看指标累加（对齐旧版 analytics 每日聚合）。
func (r *ManuscriptRepository) UpsertDailyMetric(ctx context.Context, manuscriptID, userID int64, field string, delta int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO manuscript_daily_metrics (metric_date, manuscript_id, user_id, `+field+`)
		 VALUES (CURRENT_DATE, $1, $2, $3)
		 ON CONFLICT (metric_date, manuscript_id)
		 DO UPDATE SET `+field+` = manuscript_daily_metrics.`+field+` + $3, updated_at = NOW()`,
		manuscriptID, userID, delta)
	return err
}

func (r *ManuscriptRepository) FindByID(ctx context.Context, id int64) (*Manuscript, error) {
	m := &Manuscript{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, title, description, cover_url, user_id, category_id,
		       view_count, like_count, coin_count, collect_count, share_count,
		       comment_count, danmaku_count, status, review_status, review_reason,
		       review_time, reviewer_id, upload_time, updated_at, duration, duration_seconds
		FROM manuscripts WHERE id = $1`, id,
	).Scan(&m.ID, &m.Title, &m.Description, &m.CoverURL, &m.UserID, &m.CategoryID,
		&m.ViewCount, &m.LikeCount, &m.CoinCount, &m.CollectCount, &m.ShareCount,
		&m.CommentCount, &m.DanmakuCount, &m.Status, &m.ReviewStatus, &m.ReviewReason,
		&m.ReviewTime, &m.ReviewerID, &m.UploadTime, &m.UpdatedAt, &m.Duration, &m.DurationSeconds)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *ManuscriptRepository) FindVideosByManuscriptID(ctx context.Context, manuscriptID int64) ([]*Video, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, manuscript_id, video_order, title, description, play_url_hd, play_url_sd, play_url_ld,
		       upload_time, updated_at, process_progress, process_stage, has_subtitle, has_summary,
		       process_status, process_error, source_video_url, duration_seconds
		FROM videos WHERE manuscript_id = $1 ORDER BY video_order`, manuscriptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Video
	for rows.Next() {
		v := &Video{}
		if err := rows.Scan(&v.ID, &v.ManuscriptID, &v.VideoOrder, &v.Title, &v.Description,
			&v.PlayURLHd, &v.PlayURLSd, &v.PlayURLld, &v.UploadTime, &v.UpdatedAt,
			&v.ProcessProgress, &v.ProcessStage, &v.HasSubtitle, &v.HasSummary,
			&v.ProcessStatus, &v.ProcessError, &v.SourceVideoURL, &v.DurationSeconds); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, nil
}

func (r *ManuscriptRepository) FindTagsByVideoID(ctx context.Context, videoID int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.name FROM tags t
		INNER JOIN video_tags vt ON vt.tag_id = t.id
		WHERE vt.video_id = $1`, videoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}
	return tags, nil
}

func (r *ManuscriptRepository) FindCategoryByID(ctx context.Context, id int64) (*Category, error) {
	c := &Category{}
	err := r.db.QueryRowContext(ctx, `SELECT id, name, icon, sort_order FROM categories WHERE id = $1`, id).
		Scan(&c.ID, &c.Name, &c.Icon, &c.SortOrder)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ManuscriptRepository) ListByUser(ctx context.Context, userID int64, status int32, page, pageSize int32) ([]*Manuscript, int64, error) {
	where := "WHERE user_id = $1"
	args := []interface{}{userID}
	argN := 2
	if status != 0 {
		where += " AND status = $" + itoa(argN)
		args = append(args, status)
		argN++
	}

	var total int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM manuscripts "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	order := " ORDER BY upload_time DESC LIMIT $" + itoa(argN) + " OFFSET $" + itoa(argN+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, cover_url, user_id, category_id,
		       view_count, like_count, coin_count, collect_count, share_count,
		       comment_count, danmaku_count, status, review_status, review_reason,
		       review_time, reviewer_id, upload_time, updated_at, duration, duration_seconds
		FROM manuscripts `+where+order, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*Manuscript
	for rows.Next() {
		m := &Manuscript{}
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.CoverURL, &m.UserID, &m.CategoryID,
			&m.ViewCount, &m.LikeCount, &m.CoinCount, &m.CollectCount, &m.ShareCount,
			&m.CommentCount, &m.DanmakuCount, &m.Status, &m.ReviewStatus, &m.ReviewReason,
			&m.ReviewTime, &m.ReviewerID, &m.UploadTime, &m.UpdatedAt, &m.Duration, &m.DurationSeconds); err != nil {
			return nil, 0, err
		}
		list = append(list, m)
	}
	return list, total, nil
}

func (r *ManuscriptRepository) ListRecommended(ctx context.Context) ([]*Manuscript, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, cover_url, user_id, category_id,
		       view_count, like_count, coin_count, collect_count, share_count,
		       comment_count, danmaku_count, status, review_status, review_reason,
		       review_time, reviewer_id, upload_time, updated_at, duration, duration_seconds
		FROM manuscripts WHERE status = 3 ORDER BY upload_time DESC LIMIT 20`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Manuscript
	for rows.Next() {
		m := &Manuscript{}
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.CoverURL, &m.UserID, &m.CategoryID,
			&m.ViewCount, &m.LikeCount, &m.CoinCount, &m.CollectCount, &m.ShareCount,
			&m.CommentCount, &m.DanmakuCount, &m.Status, &m.ReviewStatus, &m.ReviewReason,
			&m.ReviewTime, &m.ReviewerID, &m.UploadTime, &m.UpdatedAt, &m.Duration, &m.DurationSeconds); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

func (r *ManuscriptRepository) ListHot(ctx context.Context) ([]*Manuscript, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, cover_url, user_id, category_id,
		       view_count, like_count, coin_count, collect_count, share_count,
		       comment_count, danmaku_count, status, review_status, review_reason,
		       review_time, reviewer_id, upload_time, updated_at, duration, duration_seconds
		FROM manuscripts WHERE status = 3 ORDER BY view_count DESC LIMIT 20`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Manuscript
	for rows.Next() {
		m := &Manuscript{}
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.CoverURL, &m.UserID, &m.CategoryID,
			&m.ViewCount, &m.LikeCount, &m.CoinCount, &m.CollectCount, &m.ShareCount,
			&m.CommentCount, &m.DanmakuCount, &m.Status, &m.ReviewStatus, &m.ReviewReason,
			&m.ReviewTime, &m.ReviewerID, &m.UploadTime, &m.UpdatedAt, &m.Duration, &m.DurationSeconds); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

func (r *ManuscriptRepository) ListByCategory(ctx context.Context, categoryID int64, page, pageSize int32) ([]*Manuscript, int64, error) {
	var total int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM manuscripts WHERE category_id = $1 AND status = 3`, categoryID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, cover_url, user_id, category_id,
		       view_count, like_count, coin_count, collect_count, share_count,
		       comment_count, danmaku_count, status, review_status, review_reason,
		       review_time, reviewer_id, upload_time, updated_at, duration, duration_seconds
		FROM manuscripts WHERE category_id = $1 AND status = 3
		ORDER BY upload_time DESC LIMIT $2 OFFSET $3`, categoryID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*Manuscript
	for rows.Next() {
		m := &Manuscript{}
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.CoverURL, &m.UserID, &m.CategoryID,
			&m.ViewCount, &m.LikeCount, &m.CoinCount, &m.CollectCount, &m.ShareCount,
			&m.CommentCount, &m.DanmakuCount, &m.Status, &m.ReviewStatus, &m.ReviewReason,
			&m.ReviewTime, &m.ReviewerID, &m.UploadTime, &m.UpdatedAt, &m.Duration, &m.DurationSeconds); err != nil {
			return nil, 0, err
		}
		list = append(list, m)
	}
	return list, total, nil
}

func (r *ManuscriptRepository) ListCategories(ctx context.Context) ([]*Category, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, icon, sort_order FROM categories ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []*Category
	for rows.Next() {
		c := &Category{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &c.SortOrder); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

func (r *ManuscriptRepository) UpdateStatus(ctx context.Context, id, userID int64, status int32) error {
	res, err := r.db.ExecContext(ctx, `UPDATE manuscripts SET status = $1 WHERE id = $2 AND user_id = $3`, status, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ManuscriptRepository) Delete(ctx context.Context, id, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM manuscripts WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *ManuscriptRepository) SearchUser(ctx context.Context, userID int64, keyword string, sort string) ([]*Manuscript, error) {
	order := "ORDER BY upload_time DESC"
	if sort == "view" {
		order = "ORDER BY view_count DESC"
	} else if sort == "like" {
		order = "ORDER BY like_count DESC"
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, cover_url, user_id, category_id,
		       view_count, like_count, coin_count, collect_count, share_count,
		       comment_count, danmaku_count, status, review_status, review_reason,
		       review_time, reviewer_id, upload_time, updated_at, duration, duration_seconds
		FROM manuscripts WHERE user_id = $1 AND title ILIKE $2 `+order, userID, "%"+keyword+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Manuscript
	for rows.Next() {
		m := &Manuscript{}
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.CoverURL, &m.UserID, &m.CategoryID,
			&m.ViewCount, &m.LikeCount, &m.CoinCount, &m.CollectCount, &m.ShareCount,
			&m.CommentCount, &m.DanmakuCount, &m.Status, &m.ReviewStatus, &m.ReviewReason,
			&m.ReviewTime, &m.ReviewerID, &m.UploadTime, &m.UpdatedAt, &m.Duration, &m.DurationSeconds); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

func (r *ManuscriptRepository) ToPB(m *Manuscript, catName string, uploader *pb.UserInfo, videos []*pb.VideoItem, tags []string, firstVideoID int64, firstVideoPlayURL string) *pb.ManuscriptInfo {
	createdAt := ""
	updatedAt := ""
	if m.UploadTime.Valid {
		createdAt = m.UploadTime.Time.Format("2006-01-02T15:04:05Z")
	}
	if m.UpdatedAt.Valid {
		updatedAt = m.UpdatedAt.Time.Format("2006-01-02T15:04:05Z")
	}

	return &pb.ManuscriptInfo{
		Id:                m.ID,
		Title:             m.Title,
		Description:       m.Description,
		CoverUrl:          m.CoverURL,
		UserId:            m.UserID,
		CategoryId:        m.CategoryID,
		CategoryName:      catName,
		ViewCount:         m.ViewCount,
		LikeCount:         m.LikeCount,
		CoinCount:         m.CoinCount,
		CollectCount:      m.CollectCount,
		ShareCount:        m.ShareCount,
		CommentCount:      m.CommentCount,
		DanmakuCount:      m.DanmakuCount,
		Duration:          m.Duration,
		DurationSeconds:   m.DurationSeconds,
		Status:            m.Status,
		ReviewStatus:      m.ReviewStatus,
		ReviewReason:      m.ReviewReason,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		Uploader:          uploader,
		Tags:              tags,
		FirstVideoId:      firstVideoID,
		FirstVideoPlayUrl: firstVideoPlayURL,
		Videos:            videos,
	}
}

func videoToPB(v *Video) *pb.VideoItem {
	return &pb.VideoItem{
		Id:              v.ID,
		Title:           v.Title,
		Description:     v.Description,
		PlayUrl:         v.PlayURLHd,
		PlayUrlHd:       v.PlayURLHd,
		PlayUrlSd:       v.PlayURLSd,
		PlayUrlLd:       v.PlayURLld,
		DurationSeconds: v.DurationSeconds,
		VideoOrder:      v.VideoOrder,
		ProcessStatus:   v.ProcessStatus,
		ProcessProgress: v.ProcessProgress,
		ProcessStage:    v.ProcessStage,
		ProcessError:    v.ProcessError,
	}
}

func itoa(n int) string {
	return string(rune('0'+n%10)) + func() string {
		if n >= 10 {
			return string(rune('0' + n/10%10))
		}
		return ""
	}()
}
