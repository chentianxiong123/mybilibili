package meeting

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

type MeetingRoom struct {
	ID              int64
	RoomName        string
	RoomCode        string
	CreatorID       int64
	CreatorName     string
	MaxParticipants int32
	Status          int32
	StartTime       *time.Time
	EndTime         *time.Time
	ScheduledStart  *time.Time
	ScheduledEnd    *time.Time
	ScheduledReason string
	CreateTime      time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func genRoomCode() string {
	b := make([]byte, 6)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:6]
}

func (r *Repository) Create(ctx context.Context, roomName string, creatorID int64, creatorName string, maxParticipants int32) (int64, error) {
	code := genRoomCode()
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO meeting_room (room_name, room_code, creator_id, creator_name, max_participants)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		roomName, code, creatorID, creatorName, maxParticipants).Scan(&id)
	return id, err
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*MeetingRoom, error) {
	m := &MeetingRoom{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, room_name, room_code, creator_id, creator_name, max_participants, status,
		        start_time, end_time, scheduled_start, scheduled_end,
		        COALESCE(scheduled_reason,''), create_time
		 FROM meeting_room WHERE id = $1`, id,
	).Scan(&m.ID, &m.RoomName, &m.RoomCode, &m.CreatorID, &m.CreatorName,
		&m.MaxParticipants, &m.Status, &m.StartTime, &m.EndTime,
		&m.ScheduledStart, &m.ScheduledEnd, &m.ScheduledReason, &m.CreateTime)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (*MeetingRoom, error) {
	m := &MeetingRoom{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, room_name, room_code, creator_id, creator_name, max_participants, status,
		        start_time, end_time, scheduled_start, scheduled_end,
		        COALESCE(scheduled_reason,''), create_time
		 FROM meeting_room WHERE room_code = $1`, code,
	).Scan(&m.ID, &m.RoomName, &m.RoomCode, &m.CreatorID, &m.CreatorName,
		&m.MaxParticipants, &m.Status, &m.StartTime, &m.EndTime,
		&m.ScheduledStart, &m.ScheduledEnd, &m.ScheduledReason, &m.CreateTime)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *Repository) ListByCreator(ctx context.Context, creatorID int64) ([]*MeetingRoom, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id FROM meeting_room WHERE creator_id = $1 ORDER BY create_time DESC`, creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*MeetingRoom
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		m, _ := r.GetByID(ctx, id)
		list = append(list, m)
	}
	return list, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id int64, status int32) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE meeting_room SET status = $1, update_time = NOW() WHERE id = $2`, status, id)
	return err
}

func (r *Repository) Reserve(ctx context.Context, roomName string, creatorID int64, creatorName string, startT, endT time.Time, reason string) (int64, error) {
	code := genRoomCode()
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO meeting_room (room_name, room_code, creator_id, creator_name, status, scheduled_start, scheduled_end, scheduled_reason)
		 VALUES ($1,$2,$3,$4,3,$5,$6,$7) RETURNING id`,
		roomName, code, creatorID, creatorName, startT, endT, reason).Scan(&id)
	return id, err
}

func (r *Repository) AddParticipant(ctx context.Context, roomID, userID int64, userName, avatar string, role int32) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO meeting_participant (room_id, user_id, user_name, user_avatar, role, audio_enabled, video_enabled, join_time)
		 VALUES ($1,$2,$3,$4,$5,1,1,NOW()) ON CONFLICT DO NOTHING`,
		roomID, userID, userName, avatar, role)
	return err
}

func (r *Repository) RemoveParticipant(ctx context.Context, roomID, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE meeting_participant SET leave_time = NOW() WHERE room_id = $1 AND user_id = $2 AND leave_time IS NULL`,
		roomID, userID)
	return err
}

func (r *Repository) ListParticipants(ctx context.Context, roomID int64) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, user_name, COALESCE(user_avatar,''), role, audio_enabled, video_enabled, screen_share_enabled
		 FROM meeting_participant WHERE room_id = $1 AND leave_time IS NULL ORDER BY join_time`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, uid int64
		var name, avatar string
		var role, audio, video, screen int32
		rows.Scan(&id, &uid, &name, &avatar, &role, &audio, &video, &screen)
		list = append(list, map[string]interface{}{
			"id": id, "user_id": uid, "user_name": name, "avatar": avatar,
			"role": role, "audio_enabled": audio, "video_enabled": video, "screen_share_enabled": screen,
		})
	}
	return list, nil
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, roomName string, creatorID int64, creatorName string) (*MeetingRoom, error) {
	id, err := s.repo.Create(ctx, roomName, creatorID, creatorName, 5)
	if err != nil {
		return nil, err
	}
	s.repo.AddParticipant(ctx, id, creatorID, creatorName, "", 1)
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByCode(ctx context.Context, code string) (*MeetingRoom, error) {
	return s.repo.GetByCode(ctx, code)
}

func (s *Service) MyRooms(ctx context.Context, creatorID int64) ([]*MeetingRoom, error) {
	return s.repo.ListByCreator(ctx, creatorID)
}

func (s *Service) Join(ctx context.Context, roomCode string, userID int64, userName string) (*MeetingRoom, error) {
	m, err := s.repo.GetByCode(ctx, roomCode)
	if err != nil {
		return nil, err
	}
	s.repo.AddParticipant(ctx, m.ID, userID, userName, "", 0)
	return m, nil
}

func (s *Service) Leave(ctx context.Context, roomID, userID int64) error {
	return s.repo.RemoveParticipant(ctx, roomID, userID)
}

func (s *Service) Participants(ctx context.Context, roomID int64) ([]map[string]interface{}, error) {
	return s.repo.ListParticipants(ctx, roomID)
}

func (s *Service) End(ctx context.Context, roomID int64) error {
	return s.repo.UpdateStatus(ctx, roomID, 2)
}

func (s *Service) Reserve(ctx context.Context, roomName string, creatorID int64, creatorName string, startT, endT time.Time, reason string) (*MeetingRoom, error) {
	id, err := s.repo.Reserve(ctx, roomName, creatorID, creatorName, startT, endT, reason)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

var _ = strconv.Itoa
