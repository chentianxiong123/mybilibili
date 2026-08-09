package core

import (
	"context"

	pb "mybilibili/internal/core/pb"
)

type InteractionHandler struct {
	pb.UnimplementedInteractionServiceServer
	svc *InteractionService
}

func NewInteractionHandler(svc *InteractionService) *InteractionHandler {
	return &InteractionHandler{svc: svc}
}

func (h *InteractionHandler) LikeManuscript(ctx context.Context, req *pb.LikeManuscriptRequest) (*pb.LikeManuscriptResponse, error) {
	return h.svc.LikeManuscript(ctx, req)
}

func (h *InteractionHandler) UnlikeManuscript(ctx context.Context, req *pb.UnlikeManuscriptRequest) (*pb.UnlikeManuscriptResponse, error) {
	return h.svc.UnlikeManuscript(ctx, req)
}

func (h *InteractionHandler) CoinManuscript(ctx context.Context, req *pb.CoinManuscriptRequest) (*pb.CoinManuscriptResponse, error) {
	return h.svc.CoinManuscript(ctx, req)
}

func (h *InteractionHandler) CollectManuscript(ctx context.Context, req *pb.CollectManuscriptRequest) (*pb.CollectManuscriptResponse, error) {
	return h.svc.CollectManuscript(ctx, req)
}

func (h *InteractionHandler) UncollectManuscript(ctx context.Context, req *pb.UncollectManuscriptRequest) (*pb.UncollectManuscriptResponse, error) {
	return h.svc.UncollectManuscript(ctx, req)
}

func (h *InteractionHandler) ShareManuscript(ctx context.Context, req *pb.ShareManuscriptRequest) (*pb.ShareManuscriptResponse, error) {
	return h.svc.ShareManuscript(ctx, req)
}

func (h *InteractionHandler) GetInteractionStatus(ctx context.Context, req *pb.GetInteractionStatusRequest) (*pb.GetInteractionStatusResponse, error) {
	return h.svc.GetInteractionStatus(ctx, req)
}

func (h *InteractionHandler) FollowUser(ctx context.Context, req *pb.FollowUserRequest) (*pb.FollowUserResponse, error) {
	return h.svc.FollowUser(ctx, req)
}

func (h *InteractionHandler) UnfollowUser(ctx context.Context, req *pb.UnfollowUserRequest) (*pb.UnfollowUserResponse, error) {
	return h.svc.UnfollowUser(ctx, req)
}

func (h *InteractionHandler) CheckFollow(ctx context.Context, req *pb.CheckFollowRequest) (*pb.CheckFollowResponse, error) {
	return h.svc.CheckFollow(ctx, req)
}

func (h *InteractionHandler) GetFollowCount(ctx context.Context, req *pb.GetFollowCountRequest) (*pb.GetFollowCountResponse, error) {
	return h.svc.GetFollowCount(ctx, req)
}

func (h *InteractionHandler) GetLikedManuscripts(ctx context.Context, req *pb.GetLikedManuscriptsRequest) (*pb.GetLikedManuscriptsResponse, error) {
	return h.svc.GetLikedManuscripts(ctx, req)
}

func (h *InteractionHandler) GetCollectedManuscripts(ctx context.Context, req *pb.GetCollectedManuscriptsRequest) (*pb.GetCollectedManuscriptsResponse, error) {
	return h.svc.GetCollectedManuscripts(ctx, req)
}

func (h *InteractionHandler) AddWatchHistory(ctx context.Context, req *pb.AddWatchHistoryRequest) (*pb.AddWatchHistoryResponse, error) {
	return h.svc.AddWatchHistory(ctx, req)
}

func (h *InteractionHandler) GetWatchHistory(ctx context.Context, req *pb.GetWatchHistoryRequest) (*pb.GetWatchHistoryResponse, error) {
	return h.svc.GetWatchHistory(ctx, req)
}

func (h *InteractionHandler) ClearWatchHistory(ctx context.Context, req *pb.ClearWatchHistoryRequest) (*pb.ClearWatchHistoryResponse, error) {
	return h.svc.ClearWatchHistory(ctx, req)
}
