package live

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const (
	RoomStatusIdle  = "offline"
	RoomStatusLive  = "live"
	RoomStatusEnded = "ended"

	SeatStatusEmpty    int32 = 0
	SeatStatusPending  int32 = 1
	SeatStatusOccupied int32 = 2
	SeatStatusMuted    int32 = 3
	SeatStatusLocked   int32 = 4

	LinkmicStatusActive int32 = 0
	LinkmicStatusEnded  int32 = 1
)

type Service struct {
	repo *Repository
	hub  *Hub
}

func NewService(repo *Repository, hub *Hub) *Service {
	return &Service{repo: repo, hub: hub}
}

func (s *Service) CreateRoom(ctx context.Context, roomName, cover, category string, hostID int64, _ int32) (*Room, error) {
	if roomName == "" {
		return nil, ErrInvalidArgument("room name required")
	}

	room := &Room{
		RoomName: roomName, CoverURL: cover, Category: category,
		UserID: hostID, Status: RoomStatusIdle,
	}
	if err := s.repo.CreateRoom(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *Service) GetRoom(ctx context.Context, roomID int64) (*Room, error) {
	return s.repo.GetRoomByID(ctx, roomID)
}

func (s *Service) GetRoomByHost(ctx context.Context, hostID int64) (*Room, error) {
	return s.repo.GetRoomByUserID(ctx, hostID)
}

func (s *Service) ListLiveRooms(ctx context.Context, page, pageSize int32) ([]*Room, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	return s.repo.ListLiveRooms(ctx, page, pageSize)
}

func (s *Service) UpdateRoom(ctx context.Context, roomID, userID int64, roomName, cover, category string) error {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return ErrNotFound("room not found")
	}
	if room.UserID != userID {
		return ErrPermissionDenied("only host can update room")
	}
	if roomName != "" {
		room.RoomName = roomName
	}
	if cover != "" {
		room.CoverURL = cover
	}
	room.Category = category
	return s.repo.UpdateRoom(ctx, room)
}

func (s *Service) UpdateRoomStatus(ctx context.Context, roomID, userID int64, status string) error {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return ErrNotFound("room not found")
	}
	if room.UserID != userID {
		return ErrPermissionDenied("only host can update room status")
	}
	return s.repo.UpdateRoomStatus(ctx, roomID, status)
}

func (s *Service) ScheduleRoom(ctx context.Context, roomID, userID int64, scheduledAt *time.Time) error {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return ErrNotFound("room not found")
	}
	if room.UserID != userID {
		return ErrPermissionDenied("only host can schedule room")
	}
	ts := sql.NullTime{}
	if scheduledAt != nil {
		ts = sql.NullTime{Time: *scheduledAt, Valid: true}
	}
	return s.repo.ScheduleRoom(ctx, roomID, ts)
}

func (s *Service) HandleSRSCallback(ctx context.Context, action, streamKey string) error {
	if action != "on_publish" {
		return nil
	}
	room, err := s.repo.GetRoomByStreamKey(ctx, streamKey)
	if err != nil {
		return ErrNotFound("invalid stream key")
	}
	if err := s.repo.UpdateRoomStatus(ctx, room.ID, RoomStatusLive); err != nil {
		return err
	}
	s.hub.BroadcastRoom(room.ID, map[string]interface{}{
		"type": "room_live", "room_id": room.ID, "host_id": room.UserID,
	})
	return nil
}

func (s *Service) GetSeats(ctx context.Context, roomID int64) ([]*Seat, error) {
	return s.repo.GetSeats(ctx, roomID)
}

func (s *Service) ApplySeat(ctx context.Context, roomID, userID int64) error {
	seats, err := s.repo.GetSeats(ctx, roomID)
	if err != nil {
		return err
	}
	for _, seat := range seats {
		if seat.UserID == userID && seat.Status == SeatStatusOccupied {
			return nil
		}
	}

	for _, seat := range seats {
		if seat.Status == SeatStatusEmpty {
			seat.UserID = userID
			seat.Status = SeatStatusPending
			now := time.Now()
			seat.JoinedAt = sql.NullTime{Time: now, Valid: true}
			if err := s.repo.UpdateSeat(ctx, seat); err != nil {
				return err
			}
			s.hub.BroadcastRoom(roomID, map[string]interface{}{
				"type": "seat_updated", "room_id": roomID, "seats": seatsToMap(seats),
			})
			return nil
		}
	}
	return ErrInvalidArgument("no available seats")
}

func (s *Service) AcceptSeat(ctx context.Context, roomID, hostID, targetUserID int64) error {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return ErrNotFound("room not found")
	}
	if room.UserID != hostID {
		return ErrPermissionDenied("only host can accept seat")
	}

	seats, err := s.repo.GetSeats(ctx, roomID)
	if err != nil {
		return err
	}
	for _, seat := range seats {
		if seat.UserID == targetUserID && seat.Status == SeatStatusPending {
			seat.Status = SeatStatusOccupied
			seat.Muted = false
			if err := s.repo.UpdateSeat(ctx, seat); err != nil {
				return err
			}
			s.hub.BroadcastRoom(roomID, map[string]interface{}{
				"type": "seat_updated", "room_id": roomID, "seats": seatsToMap(seats),
			})
			return nil
		}
	}
	return ErrNotFound("pending seat request not found")
}

func (s *Service) RejectSeat(ctx context.Context, roomID, hostID, targetUserID int64) error {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return ErrNotFound("room not found")
	}
	if room.UserID != hostID {
		return ErrPermissionDenied("only host can reject seat")
	}

	seats, err := s.repo.GetSeats(ctx, roomID)
	if err != nil {
		return err
	}
	for _, seat := range seats {
		if seat.UserID == targetUserID && seat.Status == SeatStatusPending {
			seat.UserID = 0
			seat.Status = SeatStatusEmpty
			seat.JoinedAt = sql.NullTime{Valid: false}
			if err := s.repo.UpdateSeat(ctx, seat); err != nil {
				return err
			}
			s.hub.BroadcastRoom(roomID, map[string]interface{}{
				"type": "seat_updated", "room_id": roomID, "seats": seatsToMap(seats),
			})
			return nil
		}
	}
	return ErrNotFound("pending seat request not found")
}

func (s *Service) LeaveSeat(ctx context.Context, roomID, userID int64) error {
	seat, err := s.repo.GetSeatByUser(ctx, roomID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	seat.UserID = 0
	seat.Status = SeatStatusEmpty
	seat.Muted = false
	seat.JoinedAt = sql.NullTime{Valid: false}
	if err := s.repo.UpdateSeat(ctx, seat); err != nil {
		return err
	}

	seats, _ := s.repo.GetSeats(ctx, roomID)
	s.hub.BroadcastRoom(roomID, map[string]interface{}{
		"type": "seat_updated", "room_id": roomID, "seats": seatsToMap(seats),
	})
	return nil
}

func (s *Service) MuteSeat(ctx context.Context, roomID, operatorID int64, seatIndex int32, muted bool) error {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return ErrNotFound("room not found")
	}
	if room.UserID != operatorID {
		return ErrPermissionDenied("only host can mute")
	}

	seats, err := s.repo.GetSeats(ctx, roomID)
	if err != nil {
		return err
	}
	for _, seat := range seats {
		if seat.SeatIndex == seatIndex {
			seat.Muted = muted
			if muted {
				seat.Status = SeatStatusMuted
			} else {
				seat.Status = SeatStatusOccupied
			}
			if err := s.repo.UpdateSeat(ctx, seat); err != nil {
				return err
			}
			s.hub.BroadcastRoom(roomID, map[string]interface{}{
				"type": "seat_updated", "room_id": roomID, "seats": seatsToMap(seats),
			})
			return nil
		}
	}
	return ErrNotFound("seat not found")
}

func (s *Service) LockSeat(ctx context.Context, roomID, operatorID int64, seatIndex int32, locked bool) error {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return ErrNotFound("room not found")
	}
	if room.UserID != operatorID {
		return ErrPermissionDenied("only host can lock seat")
	}

	seats, err := s.repo.GetSeats(ctx, roomID)
	if err != nil {
		return err
	}
	for _, seat := range seats {
		if seat.SeatIndex == seatIndex {
			if locked {
				seat.Status = SeatStatusLocked
			} else {
				seat.Status = SeatStatusEmpty
			}
			if err := s.repo.UpdateSeat(ctx, seat); err != nil {
				return err
			}
			s.hub.BroadcastRoom(roomID, map[string]interface{}{
				"type": "seat_updated", "room_id": roomID, "seats": seatsToMap(seats),
			})
			return nil
		}
	}
	return ErrNotFound("seat not found")
}

func (s *Service) KickSeat(ctx context.Context, roomID, operatorID, targetUserID int64) error {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return ErrNotFound("room not found")
	}
	if room.UserID != operatorID {
		return ErrPermissionDenied("only host can kick")
	}

	seat, err := s.repo.GetSeatByUser(ctx, roomID, targetUserID)
	if err != nil {
		return ErrNotFound("user not on seat")
	}
	seat.UserID = 0
	seat.Status = SeatStatusEmpty
	seat.Muted = false
	seat.JoinedAt = sql.NullTime{Valid: false}
	if err := s.repo.UpdateSeat(ctx, seat); err != nil {
		return err
	}

	seats, _ := s.repo.GetSeats(ctx, roomID)
	s.hub.BroadcastRoom(roomID, map[string]interface{}{
		"type": "seat_updated", "room_id": roomID, "seats": seatsToMap(seats),
	})
	s.hub.SendToUser(targetUserID, map[string]interface{}{
		"type": "kicked", "room_id": roomID,
	})
	return nil
}

func (s *Service) CreateLinkmic(ctx context.Context, roomID, streamerID, viewerID int64) (*Linkmic, error) {
	return s.repo.CreateLinkmic(ctx, roomID, streamerID, viewerID)
}

func (s *Service) EndLinkmic(ctx context.Context, linkmicID int64) error {
	return s.repo.UpdateLinkmicStatus(ctx, linkmicID, LinkmicStatusEnded)
}

func seatsToMap(seats []*Seat) []map[string]interface{} {
	result := make([]map[string]interface{}, len(seats))
	for i, seat := range seats {
		m := map[string]interface{}{
			"index":  seat.SeatIndex,
			"status": seat.Status,
			"muted":  seat.Muted,
		}
		if seat.UserID != 0 {
			m["user_id"] = seat.UserID
		}
		result[i] = m
	}
	return result
}

func ErrInvalidArgument(msg string) error {
	return errors.New(msg)
}

func ErrNotFound(msg string) error {
	return errors.New(msg)
}

func ErrPermissionDenied(msg string) error {
	return errors.New(msg)
}