package live

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type Room struct {
	ID          int64
	RoomName    string
	UserID      int64
	StreamKey   string
	CoverURL    string
	Category    string
	Status      int32
	ViewerCount int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Seat struct {
	ID        int64
	RoomID    int64
	SeatIndex int32
	UserID    int64
	Status    int32
	Muted     bool
	JoinedAt  sql.NullTime
}

type Linkmic struct {
	ID         int64
	RoomID     int64
	StreamerID int64
	ViewerID   int64
	Status     int32
	CreatedAt  time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func genStreamKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func genRoomCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:8]
}

func (r *Repository) CreateRoom(ctx context.Context, room *Room) error {
	room.StreamKey = genStreamKey()
	return r.db.QueryRowContext(ctx,
		`INSERT INTO live_rooms (room_name, user_id, stream_key, cover_url, category)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`,
		room.RoomName, room.UserID, room.StreamKey,
		room.CoverURL, room.Category,
	).Scan(&room.ID, &room.CreatedAt, &room.UpdatedAt)
}

func (r *Repository) GetRoomByID(ctx context.Context, id int64) (*Room, error) {
	room := &Room{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, room_name, user_id, stream_key, cover_url, category, status, viewer_count, created_at, updated_at
		 FROM live_rooms WHERE id = $1`, id,
	).Scan(&room.ID, &room.RoomName, &room.UserID, &room.StreamKey,
		&room.CoverURL, &room.Category, &room.Status, &room.ViewerCount, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *Repository) GetRoomByStreamKey(ctx context.Context, key string) (*Room, error) {
	room := &Room{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, room_name, user_id, stream_key, cover_url, category, status, viewer_count, created_at, updated_at
		 FROM live_rooms WHERE stream_key = $1`, key,
	).Scan(&room.ID, &room.RoomName, &room.UserID, &room.StreamKey,
		&room.CoverURL, &room.Category, &room.Status, &room.ViewerCount, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *Repository) GetRoomByUserID(ctx context.Context, userID int64) (*Room, error) {
	room := &Room{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, room_name, user_id, stream_key, cover_url, category, status, viewer_count, created_at, updated_at
		 FROM live_rooms WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, userID,
	).Scan(&room.ID, &room.RoomName, &room.UserID, &room.StreamKey,
		&room.CoverURL, &room.Category, &room.Status, &room.ViewerCount, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *Repository) ListLiveRooms(ctx context.Context, page, pageSize int32) ([]*Room, error) {
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, room_name, user_id, stream_key, cover_url, category, status, viewer_count, created_at, updated_at
		 FROM live_rooms WHERE status = 1 ORDER BY viewer_count DESC LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Room
	for rows.Next() {
		room := &Room{}
		if err := rows.Scan(&room.ID, &room.RoomName, &room.UserID, &room.StreamKey,
			&room.CoverURL, &room.Category, &room.Status, &room.ViewerCount, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, room)
	}
	return list, nil
}

func (r *Repository) UpdateRoom(ctx context.Context, room *Room) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE live_rooms SET room_name=$1, cover_url=$2, category=$3, updated_at=NOW() WHERE id=$4`,
		room.RoomName, room.CoverURL, room.Category, room.ID)
	return err
}

func (r *Repository) UpdateRoomStatus(ctx context.Context, id int64, status int32) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE live_rooms SET status=$1, updated_at=NOW() WHERE id=$2`, status, id)
	return err
}

func (r *Repository) ScheduleRoom(ctx context.Context, id int64, scheduledAt sql.NullTime) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE live_rooms SET scheduled_at=$1, updated_at=NOW() WHERE id=$2`, scheduledAt, id)
	return err
}

func (r *Repository) IncrementViewerCount(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE live_rooms SET viewer_count = viewer_count + 1 WHERE id=$1`, id)
	return err
}

func (r *Repository) DecrementViewerCount(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE live_rooms SET viewer_count = GREATEST(viewer_count - 1, 0) WHERE id=$1`, id)
	return err
}

func (r *Repository) GetSeats(ctx context.Context, roomID int64) ([]*Seat, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, room_id, seat_index, COALESCE(user_id,0), status, muted, joined_at
		 FROM live_seats WHERE room_id = $1 ORDER BY seat_index`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Seat
	for rows.Next() {
		s := &Seat{}
		if err := rows.Scan(&s.ID, &s.RoomID, &s.SeatIndex, &s.UserID, &s.Status, &s.Muted, &s.JoinedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *Repository) InitSeats(ctx context.Context, roomID int64, maxSeats int32) error {
	vals := make([]string, maxSeats)
	for i := int32(0); i < maxSeats; i++ {
		vals[i] = fmt.Sprintf("(%d, %d, 0, false, NULL)", roomID, i)
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO live_seats (room_id, seat_index, status, muted, joined_at) VALUES `+
			strings.Join(vals, ", ")+
			` ON CONFLICT (room_id, seat_index) DO NOTHING`)
	return err
}

func (r *Repository) UpdateSeat(ctx context.Context, seat *Seat) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE live_seats SET user_id=$1, status=$2, muted=$3, joined_at=$4
		 WHERE room_id=$5 AND seat_index=$6`,
		nullInt64(seat.UserID), seat.Status, seat.Muted, seat.JoinedAt, seat.RoomID, seat.SeatIndex)
	return err
}

func (r *Repository) GetSeatByUser(ctx context.Context, roomID, userID int64) (*Seat, error) {
	s := &Seat{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, room_id, seat_index, COALESCE(user_id,0), status, muted, joined_at
		 FROM live_seats WHERE room_id = $1 AND user_id = $2`, roomID, userID,
	).Scan(&s.ID, &s.RoomID, &s.SeatIndex, &s.UserID, &s.Status, &s.Muted, &s.JoinedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) CreateLinkmic(ctx context.Context, roomID, streamerID, viewerID int64) (*Linkmic, error) {
	lm := &Linkmic{}
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO live_linkmic (room_id, streamer_id, viewer_id) VALUES ($1, $2, $3) RETURNING id, created_at`,
		roomID, streamerID, viewerID,
	).Scan(&lm.ID, &lm.CreatedAt)
	if err != nil {
		return nil, err
	}
	lm.RoomID = roomID
	lm.StreamerID = streamerID
	lm.ViewerID = viewerID
	return lm, nil
}

func (r *Repository) UpdateLinkmicStatus(ctx context.Context, id int64, status int32) error {
	_, err := r.db.ExecContext(ctx, `UPDATE live_linkmic SET status=$1 WHERE id=$2`, status, id)
	return err
}

func nullInt64(v int64) interface{} {
	if v == 0 {
		return nil
	}
	return v
}