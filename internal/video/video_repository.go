package video

import (
	"context"
	"database/sql"
	"time"
)

type Video struct {
	ID              int64
	ManuscriptID    int64
	VideoOrder      int32
	Title           string
	Description     string
	PlayURLHD       string
	PlayURLSD       string
	PlayURLLD       string
	UploadTime      time.Time
	UpdatedAt       time.Time
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

type BannerImage struct {
	ID         int64
	Title      string
	ImageURL   string
	LinkURL    string
	SortOrder  int32
	Status     int32
	Type       int32
	CategoryID int64
	StartTime  *time.Time
	EndTime    *time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetVideoByID(ctx context.Context, id int64) (*Video, error) {
	v := &Video{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, manuscript_id, video_order, title, description, play_url_hd, play_url_sd, play_url_ld,
		        upload_time, updated_at, process_progress, process_stage, has_subtitle, has_summary,
		        process_status, process_error, source_video_url, duration_seconds
		 FROM videos WHERE id = $1`, id,
	).Scan(&v.ID, &v.ManuscriptID, &v.VideoOrder, &v.Title, &v.Description,
		&v.PlayURLHD, &v.PlayURLSD, &v.PlayURLLD, &v.UploadTime, &v.UpdatedAt,
		&v.ProcessProgress, &v.ProcessStage, &v.HasSubtitle, &v.HasSummary,
		&v.ProcessStatus, &v.ProcessError, &v.SourceVideoURL, &v.DurationSeconds)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *Repository) ListByManuscript(ctx context.Context, manuscriptID int64) ([]*Video, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, manuscript_id, video_order, title, description, play_url_hd, play_url_sd, play_url_ld,
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
		rows.Scan(&v.ID, &v.ManuscriptID, &v.VideoOrder, &v.Title, &v.Description,
			&v.PlayURLHD, &v.PlayURLSD, &v.PlayURLLD, &v.UploadTime, &v.UpdatedAt,
			&v.ProcessProgress, &v.ProcessStage, &v.HasSubtitle, &v.HasSummary,
			&v.ProcessStatus, &v.ProcessError, &v.SourceVideoURL, &v.DurationSeconds)
		list = append(list, v)
	}
	return list, nil
}

func (r *Repository) ListCategories(ctx context.Context) ([]*Category, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, icon, sort_order FROM categories ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Category
	for rows.Next() {
		c := &Category{}
		rows.Scan(&c.ID, &c.Name, &c.Icon, &c.SortOrder)
		list = append(list, c)
	}
	return list, nil
}

func (r *Repository) GetCategoryByID(ctx context.Context, id int64) (*Category, error) {
	c := &Category{}
	err := r.db.QueryRowContext(ctx, `SELECT id, name, icon, sort_order FROM categories WHERE id = $1`, id).
		Scan(&c.ID, &c.Name, &c.Icon, &c.SortOrder)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *Repository) CreateCategory(ctx context.Context, name, icon string, sortOrder int32) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO categories (name, icon, sort_order) VALUES ($1, $2, $3) RETURNING id`,
		name, icon, sortOrder).Scan(&id)
	return id, err
}

func (r *Repository) UpdateCategory(ctx context.Context, id int64, name, icon string, sortOrder int32) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE categories SET name=$1, icon=$2, sort_order=$3, updated_at=NOW() WHERE id=$4`,
		name, icon, sortOrder, id)
	return err
}

func (r *Repository) DeleteCategory(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, id)
	return err
}

func (r *Repository) ListBanners(ctx context.Context, bannerType int32) ([]*BannerImage, error) {
	return r.listBanners(ctx, bannerType, 0)
}

func (r *Repository) ListBannersByCategory(ctx context.Context, bannerType int32, categoryID int64) ([]*BannerImage, error) {
	return r.listBanners(ctx, bannerType, categoryID)
}

func (r *Repository) listBanners(ctx context.Context, bannerType int32, categoryID int64) ([]*BannerImage, error) {
	query := `SELECT id, title, image_url, link_url, sort_order, status, type, COALESCE(category_id,0),
		        start_time, end_time FROM banner_images
		 WHERE type = $1 AND status = 1 AND (start_time IS NULL OR start_time <= NOW())
		   AND (end_time IS NULL OR end_time >= NOW())`
	args := []any{bannerType}
	if categoryID > 0 {
		query += ` AND category_id = $2`
		args = append(args, categoryID)
	}
	query += ` ORDER BY sort_order`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*BannerImage
	for rows.Next() {
		b := &BannerImage{}
		var st, et sql.NullTime
		rows.Scan(&b.ID, &b.Title, &b.ImageURL, &b.LinkURL, &b.SortOrder, &b.Status, &b.Type, &b.CategoryID, &st, &et)
		if st.Valid {
			b.StartTime = &st.Time
		}
		if et.Valid {
			b.EndTime = &et.Time
		}
		list = append(list, b)
	}
	return list, nil
}

func (r *Repository) CreateBanner(ctx context.Context, b *BannerImage) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO banner_images (title, image_url, link_url, sort_order, type, category_id, start_time, end_time)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		b.Title, b.ImageURL, b.LinkURL, b.SortOrder, b.Type, nullInt64(b.CategoryID), b.StartTime, b.EndTime).Scan(&id)
	return id, err
}

func (r *Repository) UpdateBanner(ctx context.Context, id int64, b *BannerImage) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE banner_images SET title=$1, image_url=$2, link_url=$3, sort_order=$4, status=$5, category_id=$6, start_time=$7, end_time=$8 WHERE id=$9`,
		b.Title, b.ImageURL, b.LinkURL, b.SortOrder, b.Status, nullInt64(b.CategoryID), b.StartTime, b.EndTime, id)
	return err
}

func (r *Repository) DeleteBanner(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM banner_images WHERE id = $1`, id)
	return err
}

func (r *Repository) GetStatistics(ctx context.Context) (map[string]interface{}, error) {
	var msCount, userCount, viewCount, pendingCount int64
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM manuscripts`).Scan(&msCount)
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&userCount)
	r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(view_count),0) FROM manuscripts`).Scan(&viewCount)
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM manuscripts WHERE status = 0`).Scan(&pendingCount)
	return map[string]interface{}{
		"manuscript_count": msCount, "user_count": userCount,
		"view_count": viewCount, "pending_count": pendingCount,
	}, nil
}

func nullInt64(v int64) interface{} {
	if v == 0 {
		return nil
	}
	return v
}
