package core

import (
	"context"

	pb "mybilibili/internal/core/pb"
)

type ManuscriptHandler struct {
	pb.UnimplementedManuscriptServiceServer
	svc *ManuscriptService
}

func NewManuscriptHandler(svc *ManuscriptService) *ManuscriptHandler {
	return &ManuscriptHandler{svc: svc}
}

func (h *ManuscriptHandler) GetManuscript(ctx context.Context, req *pb.GetManuscriptRequest) (*pb.GetManuscriptResponse, error) {
	return h.svc.GetManuscript(ctx, req)
}

func (h *ManuscriptHandler) GetManuscriptWithVideos(ctx context.Context, req *pb.GetManuscriptWithVideosRequest) (*pb.GetManuscriptResponse, error) {
	return h.svc.GetManuscriptWithVideos(ctx, req)
}

func (h *ManuscriptHandler) ListUserManuscripts(ctx context.Context, req *pb.ListUserManuscriptsRequest) (*pb.ListUserManuscriptsResponse, error) {
	return h.svc.ListUserManuscripts(ctx, req)
}

func (h *ManuscriptHandler) ListRecommended(ctx context.Context, req *pb.ListRecommendedRequest) (*pb.ListRecommendedResponse, error) {
	return h.svc.ListRecommended(ctx, req)
}

func (h *ManuscriptHandler) ListHot(ctx context.Context, req *pb.ListHotRequest) (*pb.ListHotResponse, error) {
	return h.svc.ListHot(ctx, req)
}

func (h *ManuscriptHandler) ListByCategory(ctx context.Context, req *pb.ListByCategoryRequest) (*pb.ListByCategoryResponse, error) {
	return h.svc.ListByCategory(ctx, req)
}

func (h *ManuscriptHandler) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	return h.svc.ListCategories(ctx, req)
}

func (h *ManuscriptHandler) DeleteManuscript(ctx context.Context, req *pb.DeleteManuscriptRequest) (*pb.DeleteManuscriptResponse, error) {
	return h.svc.DeleteManuscript(ctx, req)
}

func (h *ManuscriptHandler) PublishManuscript(ctx context.Context, req *pb.PublishManuscriptRequest) (*pb.PublishManuscriptResponse, error) {
	return h.svc.PublishManuscript(ctx, req)
}

func (h *ManuscriptHandler) UnpublishManuscript(ctx context.Context, req *pb.UnpublishManuscriptRequest) (*pb.UnpublishManuscriptResponse, error) {
	return h.svc.UnpublishManuscript(ctx, req)
}

func (h *ManuscriptHandler) SearchUserManuscripts(ctx context.Context, req *pb.SearchUserManuscriptsRequest) (*pb.SearchUserManuscriptsResponse, error) {
	return h.svc.SearchUserManuscripts(ctx, req)
}