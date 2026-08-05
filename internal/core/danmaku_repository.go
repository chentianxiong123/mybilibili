package core

import (
	"context"
	"database/sql"
	"time"
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

type DanmakuBroadcaster struct {
	channels map[int64]chan *DanmakuEvent
}

type DanmakuEvent struct {
	ID           int64   `json:"id"`
	VideoID      int64   `json:"video_id"`
	UserID       int64   `json:"user_id"`
	Content      string  `json:"content"`
	Time         float64 `json:"time"`
	Color        string  `json:"color"`
	Mode         int32   `json:"mode"`
	CreatedAt    string  `json:"created_at"`
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