package core

import (
	"context"

	pb "mybilibili/internal/core/pb"
)

type ProfileRecorder interface {
	RecordWatch(ctx context.Context, userID, categoryID int64, tags []string, duration int64) error
	RecordLike(ctx context.Context, userID, categoryID int64, tags []string) error
	RecordCollect(ctx context.Context, userID, categoryID int64, tags []string) error
}

type InteractionService struct {
	repo            *InteractionRepository
	publisher       *EventPublisher
	messageRepo     *MessageRepository
	profileRecorder ProfileRecorder
}

func NewInteractionService(repo *InteractionRepository) *InteractionService {
	return &InteractionService{repo: repo}
}

func (s *InteractionService) SetMessageRepo(mr *MessageRepository) {
	s.messageRepo = mr
}

func (s *InteractionService) SetProfileRecorder(pr ProfileRecorder) {
	s.profileRecorder = pr
}

func (s *InteractionService) SetEventPublisher(p *EventPublisher) {
	s.publisher = p
}

func (s *InteractionService) LikeManuscript(ctx context.Context, req *pb.LikeManuscriptRequest) (*pb.LikeManuscriptResponse, error) {
	liked, _ := s.repo.HasInteraction(ctx, req.UserId, "MANUSCRIPT", "LIKE", req.ManuscriptId)
	if !liked {
		s.repo.AddInteraction(ctx, req.UserId, "MANUSCRIPT", "LIKE", req.ManuscriptId)
		s.repo.IncrementManuscriptCount(ctx, "like_count", req.ManuscriptId)
		s.repo.UpsertDailyMetric(ctx, req.ManuscriptId, req.UserId, "like_count", 1)
		s.publishAnalytics(ctx, req.ManuscriptId, req.UserId, "MANUSCRIPT_LIKE", "like_count", 1)
		s.sendLikeNotification(ctx, req.UserId, req.ManuscriptId)
		if s.profileRecorder != nil {
			var catID int64
			s.messageRepo.DB().QueryRowContext(ctx, `SELECT category_id FROM manuscripts WHERE id = $1`, req.ManuscriptId).Scan(&catID)
			s.profileRecorder.RecordLike(ctx, req.UserId, catID, nil)
		}
	}
	count, _ := s.repo.CountInteraction(ctx, "MANUSCRIPT", "LIKE", req.ManuscriptId)
	return &pb.LikeManuscriptResponse{Liked: true, LikeCount: count}, nil
}

func (s *InteractionService) UnlikeManuscript(ctx context.Context, req *pb.UnlikeManuscriptRequest) (*pb.UnlikeManuscriptResponse, error) {
	s.repo.RemoveInteraction(ctx, req.UserId, "MANUSCRIPT", "LIKE", req.ManuscriptId)
	s.repo.DecrementManuscriptCount(ctx, "like_count", req.ManuscriptId)
	s.repo.UpsertDailyMetric(ctx, req.ManuscriptId, req.UserId, "like_count", -1)
	s.publishAnalytics(ctx, req.ManuscriptId, req.UserId, "MANUSCRIPT_UNLIKE", "like_count", -1)
	count, _ := s.repo.CountInteraction(ctx, "MANUSCRIPT", "LIKE", req.ManuscriptId)
	return &pb.UnlikeManuscriptResponse{Liked: false, LikeCount: count}, nil
}

func (s *InteractionService) CoinManuscript(ctx context.Context, req *pb.CoinManuscriptRequest) (*pb.CoinManuscriptResponse, error) {
	coinCount, err := s.repo.GetUserCoinCount(ctx, req.UserId)
	if err != nil || coinCount <= 0 {
		return nil, ErrResourceExhausted("insufficient coins")
	}
	s.repo.AddInteraction(ctx, req.UserId, "MANUSCRIPT", "COIN", req.ManuscriptId)
	s.repo.IncrementManuscriptCount(ctx, "coin_count", req.ManuscriptId)
	s.repo.UpsertDailyMetric(ctx, req.ManuscriptId, req.UserId, "coin_count", 1)
	s.publishAnalytics(ctx, req.ManuscriptId, req.UserId, "MANUSCRIPT_COIN", "coin_count", 1)
	s.repo.DeductCoin(ctx, req.UserId)
	return &pb.CoinManuscriptResponse{Success: true}, nil
}

func (s *InteractionService) CollectManuscript(ctx context.Context, req *pb.CollectManuscriptRequest) (*pb.CollectManuscriptResponse, error) {
	collected, _ := s.repo.HasInteraction(ctx, req.UserId, "MANUSCRIPT", "COLLECT", req.ManuscriptId)
	if !collected {
		s.repo.AddInteraction(ctx, req.UserId, "MANUSCRIPT", "COLLECT", req.ManuscriptId)
		s.repo.IncrementManuscriptCount(ctx, "collect_count", req.ManuscriptId)
		s.repo.UpsertDailyMetric(ctx, req.ManuscriptId, req.UserId, "collect_count", 1)
		s.publishAnalytics(ctx, req.ManuscriptId, req.UserId, "MANUSCRIPT_COLLECT", "collect_count", 1)
		if s.profileRecorder != nil {
			var catID int64
			s.messageRepo.DB().QueryRowContext(ctx, `SELECT category_id FROM manuscripts WHERE id = $1`, req.ManuscriptId).Scan(&catID)
			s.profileRecorder.RecordCollect(ctx, req.UserId, catID, nil)
		}
	}
	return &pb.CollectManuscriptResponse{Collected: true}, nil
}

func (s *InteractionService) UncollectManuscript(ctx context.Context, req *pb.UncollectManuscriptRequest) (*pb.UncollectManuscriptResponse, error) {
	s.repo.RemoveInteraction(ctx, req.UserId, "MANUSCRIPT", "COLLECT", req.ManuscriptId)
	s.repo.DecrementManuscriptCount(ctx, "collect_count", req.ManuscriptId)
	s.repo.UpsertDailyMetric(ctx, req.ManuscriptId, req.UserId, "collect_count", -1)
	s.publishAnalytics(ctx, req.ManuscriptId, req.UserId, "MANUSCRIPT_UNCOLLECT", "collect_count", -1)
	return &pb.UncollectManuscriptResponse{Collected: false}, nil
}

func (s *InteractionService) ShareManuscript(ctx context.Context, req *pb.ShareManuscriptRequest) (*pb.ShareManuscriptResponse, error) {
	s.repo.AddInteraction(ctx, req.UserId, "MANUSCRIPT", "SHARE", req.ManuscriptId)
	s.repo.IncrementManuscriptCount(ctx, "share_count", req.ManuscriptId)
	s.repo.UpsertDailyMetric(ctx, req.ManuscriptId, req.UserId, "share_count", 1)
	s.publishAnalytics(ctx, req.ManuscriptId, req.UserId, "MANUSCRIPT_SHARE", "share_count", 1)
	return &pb.ShareManuscriptResponse{Success: true}, nil
}

func (s *InteractionService) GetInteractionStatus(ctx context.Context, req *pb.GetInteractionStatusRequest) (*pb.GetInteractionStatusResponse, error) {
	liked, _ := s.repo.HasInteraction(ctx, req.UserId, "MANUSCRIPT", "LIKE", req.ManuscriptId)
	collected, _ := s.repo.HasInteraction(ctx, req.UserId, "MANUSCRIPT", "COLLECT", req.ManuscriptId)
	shared, _ := s.repo.HasInteraction(ctx, req.UserId, "MANUSCRIPT", "SHARE", req.ManuscriptId)
	coinCount, _ := s.repo.CountInteraction(ctx, "MANUSCRIPT", "COIN", req.ManuscriptId)
	return &pb.GetInteractionStatusResponse{
		Liked: liked, Collected: collected, Shared: shared, CoinCount: coinCount,
	}, nil
}

func (s *InteractionService) FollowUser(ctx context.Context, req *pb.FollowUserRequest) (*pb.FollowUserResponse, error) {
	s.repo.Follow(ctx, req.UserId, req.TargetUserId)
	return &pb.FollowUserResponse{Following: true}, nil
}

func (s *InteractionService) UnfollowUser(ctx context.Context, req *pb.UnfollowUserRequest) (*pb.UnfollowUserResponse, error) {
	s.repo.Unfollow(ctx, req.UserId, req.TargetUserId)
	return &pb.UnfollowUserResponse{Following: false}, nil
}

func (s *InteractionService) CheckFollow(ctx context.Context, req *pb.CheckFollowRequest) (*pb.CheckFollowResponse, error) {
	following, _ := s.repo.IsFollowing(ctx, req.UserId, req.TargetUserId)
	return &pb.CheckFollowResponse{Following: following}, nil
}

func (s *InteractionService) GetFollowCount(ctx context.Context, req *pb.GetFollowCountRequest) (*pb.GetFollowCountResponse, error) {
	following, followers, _ := s.repo.CountFollows(ctx, req.UserId)
	return &pb.GetFollowCountResponse{FollowingCount: following, FollowerCount: followers}, nil
}

func (s *InteractionService) GetLikedManuscripts(ctx context.Context, req *pb.GetLikedManuscriptsRequest) (*pb.GetLikedManuscriptsResponse, error) {
	ids, _ := s.repo.GetInteractionIDs(ctx, req.UserId, "MANUSCRIPT", "LIKE")
	return &pb.GetLikedManuscriptsResponse{ManuscriptIds: ids}, nil
}

func (s *InteractionService) GetCollectedManuscripts(ctx context.Context, req *pb.GetCollectedManuscriptsRequest) (*pb.GetCollectedManuscriptsResponse, error) {
	ids, _ := s.repo.GetInteractionIDs(ctx, req.UserId, "MANUSCRIPT", "COLLECT")
	return &pb.GetCollectedManuscriptsResponse{ManuscriptIds: ids}, nil
}

func (s *InteractionService) AddWatchHistory(ctx context.Context, req *pb.AddWatchHistoryRequest) (*pb.AddWatchHistoryResponse, error) {
	s.repo.UpsertWatchHistory(ctx, req.UserId, req.ManuscriptId, req.ProgressSeconds)
	return &pb.AddWatchHistoryResponse{}, nil
}

func (s *InteractionService) GetWatchHistory(ctx context.Context, req *pb.GetWatchHistoryRequest) (*pb.GetWatchHistoryResponse, error) {
	items, _ := s.repo.GetWatchHistory(ctx, req.UserId, req.Page, req.PageSize)
	return &pb.GetWatchHistoryResponse{Items: items}, nil
}

func (s *InteractionService) ClearWatchHistory(ctx context.Context, req *pb.ClearWatchHistoryRequest) (*pb.ClearWatchHistoryResponse, error) {
	s.repo.ClearWatchHistory(ctx, req.UserId)
	return &pb.ClearWatchHistoryResponse{}, nil
}

func (s *InteractionService) publishAnalytics(ctx context.Context, manuscriptID, userID int64, eventType, metricType string, delta int64) {
	if s.publisher != nil {
		_ = s.publisher.PublishAnalytics(ctx, manuscriptID, userID, eventType, metricType, delta)
	}
}

func (s *InteractionService) sendLikeNotification(ctx context.Context, senderID, manuscriptID int64) {
	if s.messageRepo == nil {
		return
	}
	var ownerID int64
	err := s.messageRepo.DB().QueryRowContext(ctx,
		`SELECT user_id FROM manuscripts WHERE id = $1`, manuscriptID).Scan(&ownerID)
	if err != nil || ownerID == 0 {
		return
	}
	_, _ = s.messageRepo.SendMessage(ctx, senderID, ownerID, "liked your manuscript", 4)
}
