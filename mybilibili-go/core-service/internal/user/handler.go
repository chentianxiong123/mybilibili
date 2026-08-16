package user

import (
	"context"

	pb "mybilibili/pkg/pb"
)

type Handler struct {
	pb.UnimplementedUserServiceServer
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return h.svc.Register(ctx, req)
}

func (h *Handler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return h.svc.Login(ctx, req)
}

func (h *Handler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	return h.svc.GetUser(ctx, req)
}
