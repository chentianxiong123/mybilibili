package social

import (
	"context"
	"errors"
)

type FollowService struct {
	repo *FollowRepository
}

func NewFollowService(repo *FollowRepository) *FollowService {
	return &FollowService{repo: repo}
}

func (s *FollowService) Follow(ctx context.Context, followerID, followingID int64) error {
	if followerID == followingID {
		return errors.New("cannot follow yourself")
	}
	return s.repo.Follow(ctx, followerID, followingID)
}

func (s *FollowService) Unfollow(ctx context.Context, followerID, followingID int64) error {
	return s.repo.Unfollow(ctx, followerID, followingID)
}

func (s *FollowService) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	return s.repo.IsFollowing(ctx, followerID, followingID)
}

func (s *FollowService) ListFollowing(ctx context.Context, userID int64, page, pageSize int32) ([]int64, error) {
	return s.repo.ListFollowing(ctx, userID, page, pageSize)
}

func (s *FollowService) ListFollowers(ctx context.Context, userID int64, page, pageSize int32) ([]int64, error) {
	return s.repo.ListFollowers(ctx, userID, page, pageSize)
}

func (s *FollowService) FollowingCount(ctx context.Context, userID int64) (int64, error) {
	return s.repo.FollowingCount(ctx, userID)
}

func (s *FollowService) FollowerCount(ctx context.Context, userID int64) (int64, error) {
	return s.repo.FollowerCount(ctx, userID)
}
