package danmaku

import (
	"context"
	"time"
)

type DanmakuService struct {
	repo        *DanmakuRepository
	broadcaster *DanmakuBroadcaster
}

func NewDanmakuService(repo *DanmakuRepository, broadcaster *DanmakuBroadcaster) *DanmakuService {
	return &DanmakuService{repo: repo, broadcaster: broadcaster}
}

func (s *DanmakuService) Broadcaster() *DanmakuBroadcaster {
	return s.broadcaster
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
	if manuscriptID > 0 {
		s.repo.UpsertDailyMetric(ctx, manuscriptID, userID, "danmaku_count", 1)
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

func (s *DanmakuService) Delete(ctx context.Context, id, userID int64) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *DanmakuService) CountByVideo(ctx context.Context, videoID int64) (int64, error) {
	return s.repo.CountByVideo(ctx, videoID)
}

func (s *DanmakuService) CountByManuscriptIDs(ctx context.Context, manuscriptIDs []int64) (map[int64]int64, error) {
	return s.repo.CountByManuscriptIDs(ctx, manuscriptIDs)
}

func (s *DanmakuService) Trend(ctx context.Context, manuscriptIDs []int64, startDate, endDate string) (map[string]int, error) {
	return s.repo.TrendByDate(ctx, manuscriptIDs, startDate, endDate)
}

func (s *DanmakuService) CreatorList(ctx context.Context, userID, videoID int64, page, size int32) ([]*DanmakuEvent, int64, error) {
	list, total, err := s.repo.ListByCreator(ctx, userID, videoID, page, size)
	if err != nil {
		return nil, 0, err
	}
	var events []*DanmakuEvent
	for _, d := range list {
		events = append(events, &DanmakuEvent{
			ID: d.ID, VideoID: d.VideoID, UserID: d.UserID,
			Content: d.Content, Time: d.Time, Color: d.Color, Mode: d.Mode,
			CreatedAt: d.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return events, total, nil
}

func (s *DanmakuService) CreatorDelete(ctx context.Context, id, userID int64) error {
	return s.repo.DeleteByCreator(ctx, id, userID)
}
