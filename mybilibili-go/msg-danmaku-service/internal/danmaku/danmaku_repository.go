package danmaku

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/lib/pq"
)

type Danmaku struct {
	ID           int64
	VideoID      int64
	ManuscriptID int64
	UserID       int64
	Content      string
	Time         float64
	Color        string
	Mode         int32
	CreatedAt    time.Time
}

type DanmakuRepository struct {
	db *sql.DB
}

func NewDanmakuRepository(db *sql.DB) *DanmakuRepository {
	return &DanmakuRepository{db: db}
}

func (r *DanmakuRepository) Create(ctx context.Context, d *Danmaku) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO danmaku (video_id, manuscript_id, user_id, content, time, color, mode)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		d.VideoID, d.ManuscriptID, d.UserID, d.Content, d.Time, d.Color, d.Mode).Scan(&id)
	return id, err
}

// UpsertDailyMetric 当日弹幕指标累加（对齐旧版 analytics 每日聚合）。
func (r *DanmakuRepository) UpsertDailyMetric(ctx context.Context, manuscriptID, userID int64, field string, delta int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO manuscript_daily_metrics (metric_date, manuscript_id, user_id, `+field+`)
		 VALUES (CURRENT_DATE, $1, $2, $3)
		 ON CONFLICT (metric_date, manuscript_id)
		 DO UPDATE SET `+field+` = manuscript_daily_metrics.`+field+` + $3, updated_at = NOW()`,
		manuscriptID, userID, delta)
	return err
}

func (r *DanmakuRepository) ListByVideo(ctx context.Context, videoID int64) ([]*Danmaku, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, video_id, manuscript_id, user_id, content, time, color, mode, created_at
		 FROM danmaku WHERE video_id = $1 ORDER BY time`, videoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Danmaku
	for rows.Next() {
		d := &Danmaku{}
		if err := rows.Scan(&d.ID, &d.VideoID, &d.ManuscriptID, &d.UserID, &d.Content, &d.Time, &d.Color, &d.Mode, &d.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, nil
}

func (r *DanmakuRepository) ListByTimeRange(ctx context.Context, videoID int64, startTime, endTime float64) ([]*Danmaku, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, video_id, manuscript_id, user_id, content, time, color, mode, created_at
		 FROM danmaku WHERE video_id = $1 AND time >= $2 AND time <= $3 ORDER BY time`,
		videoID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Danmaku
	for rows.Next() {
		d := &Danmaku{}
		if err := rows.Scan(&d.ID, &d.VideoID, &d.ManuscriptID, &d.UserID, &d.Content, &d.Time, &d.Color, &d.Mode, &d.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, nil
}

func (r *DanmakuRepository) Delete(ctx context.Context, id, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM danmaku WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *DanmakuRepository) CountByVideo(ctx context.Context, videoID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM danmaku WHERE video_id = $1`, videoID).Scan(&count)
	return count, err
}

func (r *DanmakuRepository) CountByManuscriptIDs(ctx context.Context, manuscriptIDs []int64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(manuscriptIDs))
	for _, mid := range manuscriptIDs {
		result[mid] = 0
	}
	if len(manuscriptIDs) == 0 {
		return result, nil
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT manuscript_id, COUNT(*) FROM danmaku WHERE manuscript_id = ANY($1) GROUP BY manuscript_id`,
		pq.Array(manuscriptIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var mid, cnt int64
		if err := rows.Scan(&mid, &cnt); err != nil {
			return nil, err
		}
		result[mid] = cnt
	}
	return result, nil
}

func (r *DanmakuRepository) TrendByDate(ctx context.Context, manuscriptIDs []int64, startDate, endDate string) (map[string]int, error) {
	result := make(map[string]int)
	if len(manuscriptIDs) == 0 {
		return result, nil
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT TO_CHAR(created_at, 'YYYY-MM-DD') AS day, COUNT(*)
		 FROM danmaku
		 WHERE manuscript_id = ANY($1)
		   AND created_at >= $2::date AND created_at < ($3::date + INTERVAL '1 day')
		 GROUP BY day ORDER BY day`,
		pq.Array(manuscriptIDs), startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var day string
		var cnt int
		if err := rows.Scan(&day, &cnt); err != nil {
			return nil, err
		}
		result[day] = cnt
	}
	return result, nil
}

func (r *DanmakuRepository) ListByCreator(ctx context.Context, userID, videoID int64, page, size int32) ([]*Danmaku, int64, error) {
	where := `JOIN videos v ON v.id = danmaku.video_id
	          JOIN manuscripts m ON m.id = v.manuscript_id
	          WHERE m.user_id = $1`
	args := []any{userID}
	if videoID > 0 {
		where += ` AND danmaku.video_id = $2`
		args = append(args, videoID)
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM danmaku `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := int64(page-1) * int64(size)
	query := `SELECT danmaku.id, danmaku.video_id, danmaku.manuscript_id, danmaku.user_id, danmaku.content, danmaku.time, danmaku.color, danmaku.mode, danmaku.created_at
	          FROM danmaku ` + where + ` ORDER BY danmaku.created_at DESC LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
	args = append(args, size, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*Danmaku
	for rows.Next() {
		d := &Danmaku{}
		if err := rows.Scan(&d.ID, &d.VideoID, &d.ManuscriptID, &d.UserID, &d.Content, &d.Time, &d.Color, &d.Mode, &d.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, d)
	}
	return list, total, nil
}

func (r *DanmakuRepository) DeleteByCreator(ctx context.Context, id, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM danmaku
		 WHERE id = $1 AND video_id IN (
		     SELECT v.id FROM videos v
		     JOIN manuscripts m ON m.id = v.manuscript_id
		     WHERE m.user_id = $2)`,
		id, userID)
	return err
}

type DanmakuBroadcaster struct {
	channels map[int64]chan *DanmakuEvent
}

type DanmakuEvent struct {
	ID        int64   `json:"id"`
	VideoID   int64   `json:"video_id"`
	UserID    int64   `json:"user_id"`
	Content   string  `json:"content"`
	Time      float64 `json:"time"`
	Color     string  `json:"color"`
	Mode      int32   `json:"mode"`
	CreatedAt string  `json:"created_at"`
}

func NewDanmakuBroadcaster() *DanmakuBroadcaster {
	return &DanmakuBroadcaster{channels: make(map[int64]chan *DanmakuEvent)}
}

func (b *DanmakuBroadcaster) Subscribe(videoID int64) <-chan *DanmakuEvent {
	if b.channels[videoID] == nil {
		b.channels[videoID] = make(chan *DanmakuEvent, 100)
	}
	return b.channels[videoID]
}

func (b *DanmakuBroadcaster) Unsubscribe(videoID int64, ch <-chan *DanmakuEvent) {
	// closed by caller
}

func (b *DanmakuBroadcaster) Broadcast(videoID int64, event *DanmakuEvent) {
	if b.channels[videoID] != nil {
		select {
		case b.channels[videoID] <- event:
		default:
		}
	}
}
