package core

import (
	"context"
	"time"
)

type DanmakuService struct {
	repo       *DanmakuRepository
	broadcaster *DanmakuBroadcaster
}

func NewDanmakuService(repo *DanmakuRepository, broadcaster *DanmakuBroadcaster) *DanmakuService {
	return &DanmakuService{repo: repo, broadcaster: broadcaster}
}

func (s *DanmakuService) Send(ctx context.Context, videoID, manuscriptID, userID int64, content string, tm float64, color string, mode int32) (*DanmakuEvent, error) {
	d := &Danmaku{
		VideoID: videoID, ManuscriptID: manuscriptID, UserID: userID,
		Content: content, Time: tm, Color: color, Mode: mode,
	}
	id, err := s.repo.Create(ctx, d)
	if err != nil {
		return nil, err
	}

	event := &DanmakuEvent{
		ID: id, VideoID: videoID, UserID: userID,
		Content: content, Time: tm, Color: color, Mode: mode,
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z"),
	}
	s.broadcaster.Broadcast(videoID, event)
	return event, nil
}

func (s *DanmakuService) ListByVideo(ctx context.Context, videoID int64) ([]*DanmakuEvent, error) {
	list, err := s.repo.ListByVideo(ctx, videoID)
	if err != nil {
		return nil, err
	}
	var events []*DanmakuEvent
	for _, d := range list {
		events = append(events, &DanmakuEvent{
			ID: d.ID, VideoID: d.VideoID, UserID: d.UserID,
			Content: d.Content, Time: d.Time, Color: d.Color, Mode: d.Mode,
			CreatedAt: d.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return events, nil
}

func (s *DanmakuService) ListByTimeRange(ctx context.Context, videoID int64, startTime, endTime float64) ([]*DanmakuEvent, error) {
	list, err := s.repo.ListByTimeRange(ctx, videoID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	var events []*DanmakuEvent
	for _, d := range list {
		events = append(events, &DanmakuEvent{
			ID: d.ID, VideoID: d.VideoID, UserID: d.UserID,
			Content: d.Content, Time: d.Time, Color: d.Color, Mode: d.Mode,
			CreatedAt: d.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return events, nil
}