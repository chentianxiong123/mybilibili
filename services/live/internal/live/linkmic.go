package live

import (
	"context"
	"database/sql"
)

const (
	LinkmicStatusApplying     int32 = 0
	LinkmicStatusConnected    int32 = 1
	LinkmicStatusDisconnected int32 = 2
	LinkmicStatusRejected     int32 = 3
)

type LinkmicRepository struct {
	db *sql.DB
}

func NewLinkmicRepository(db *sql.DB) *LinkmicRepository {
	return &LinkmicRepository{db: db}
}

func (r *LinkmicRepository) Apply(ctx context.Context, roomID, streamerID, viewerID int64) (*Linkmic, error) {
	lm := &Linkmic{}
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO live_linkmic (room_id, streamer_id, viewer_id, status, apply_time)
		 VALUES ($1,$2,$3,0,NOW()) RETURNING id, room_id, streamer_id, viewer_id, status, created_at`,
		roomID, streamerID, viewerID,
	).Scan(&lm.ID, &lm.RoomID, &lm.StreamerID, &lm.ViewerID, &lm.Status, &lm.CreatedAt)
	if err != nil {
		return nil, err
	}
	return lm, nil
}

func (r *LinkmicRepository) UpdateStatus(ctx context.Context, id int64, status int32) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE live_linkmic SET status = $1,
		 connect_time = CASE WHEN $2 = 1 AND connect_time IS NULL THEN NOW() ELSE connect_time END,
		 end_time = CASE WHEN $2 IN (2,3) AND end_time IS NULL THEN NOW() ELSE end_time END
		 WHERE id = $3`, status, status, id)
	return err
}

func (r *LinkmicRepository) ActiveByRoom(ctx context.Context, roomID int64) ([]*Linkmic, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, room_id, streamer_id, viewer_id, status, created_at
		 FROM live_linkmic WHERE room_id = $1 AND status IN (0,1) ORDER BY created_at`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Linkmic
	for rows.Next() {
		lm := &Linkmic{}
		rows.Scan(&lm.ID, &lm.RoomID, &lm.StreamerID, &lm.ViewerID, &lm.Status, &lm.CreatedAt)
		list = append(list, lm)
	}
	return list, nil
}

func (r *LinkmicRepository) PendingByRoom(ctx context.Context, roomID int64) ([]*Linkmic, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, room_id, streamer_id, viewer_id, status, created_at
		 FROM live_linkmic WHERE room_id = $1 AND status = 0 ORDER BY created_at`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Linkmic
	for rows.Next() {
		lm := &Linkmic{}
		rows.Scan(&lm.ID, &lm.RoomID, &lm.StreamerID, &lm.ViewerID, &lm.Status, &lm.CreatedAt)
		list = append(list, lm)
	}
	return list, nil
}

func (r *LinkmicRepository) QueuePosition(ctx context.Context, roomID, viewerID int64) (int, error) {
	var pos int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM live_linkmic WHERE room_id = $1 AND status = 0 AND created_at <
		 (SELECT created_at FROM live_linkmic WHERE room_id = $1 AND viewer_id = $2 AND status = 0)`,
		roomID, viewerID).Scan(&pos)
	if err != nil {
		return 0, err
	}
	return pos + 1, nil
}

func (r *LinkmicRepository) ToggleAudio(ctx context.Context, id int64) (int32, error) {
	var enabled int32
	err := r.db.QueryRowContext(ctx,
		`UPDATE live_linkmic SET audio_enabled = CASE WHEN audio_enabled = 1 THEN 0 ELSE 1 END
		 WHERE id = $1 RETURNING audio_enabled`, id).Scan(&enabled)
	return enabled, err
}

func (r *LinkmicRepository) ToggleVideo(ctx context.Context, id int64) (int32, error) {
	var enabled int32
	err := r.db.QueryRowContext(ctx,
		`UPDATE live_linkmic SET video_enabled = CASE WHEN video_enabled = 1 THEN 0 ELSE 1 END
		 WHERE id = $1 RETURNING video_enabled`, id).Scan(&enabled)
	return enabled, err
}

type LinkmicService struct {
	repo  *LinkmicRepository
	hub   *Hub
	rooms *Repository
}

func NewLinkmicService(repo *LinkmicRepository, hub *Hub, rooms *Repository) *LinkmicService {
	return &LinkmicService{repo: repo, hub: hub, rooms: rooms}
}

func (s *LinkmicService) Apply(ctx context.Context, roomID, userID int64) (*Linkmic, error) {
	room, err := s.rooms.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	lm, err := s.repo.Apply(ctx, roomID, room.UserID, userID)
	if err != nil {
		return nil, err
	}
	s.hub.BroadcastRoom(roomID, map[string]interface{}{
		"type": "linkmic_apply", "linkmic_id": lm.ID, "room_id": roomID, "viewer_id": userID,
	})
	return lm, nil
}

func (s *LinkmicService) Accept(ctx context.Context, linkmicID int64) error {
	return s.repo.UpdateStatus(ctx, linkmicID, LinkmicStatusConnected)
}

func (s *LinkmicService) Reject(ctx context.Context, linkmicID int64) error {
	return s.repo.UpdateStatus(ctx, linkmicID, LinkmicStatusRejected)
}

func (s *LinkmicService) Disconnect(ctx context.Context, linkmicID int64) error {
	return s.repo.UpdateStatus(ctx, linkmicID, LinkmicStatusDisconnected)
}

func (s *LinkmicService) Active(ctx context.Context, roomID int64) ([]*Linkmic, error) {
	return s.repo.ActiveByRoom(ctx, roomID)
}

func (s *LinkmicService) Pending(ctx context.Context, roomID int64) ([]*Linkmic, error) {
	return s.repo.PendingByRoom(ctx, roomID)
}

func (s *LinkmicService) QueuePosition(ctx context.Context, roomID, userID int64) (int, error) {
	return s.repo.QueuePosition(ctx, roomID, userID)
}

func (s *LinkmicService) ToggleAudio(ctx context.Context, linkmicID int64) (int32, error) {
	return s.repo.ToggleAudio(ctx, linkmicID)
}

func (s *LinkmicService) ToggleVideo(ctx context.Context, linkmicID int64) (int32, error) {
	return s.repo.ToggleVideo(ctx, linkmicID)
}