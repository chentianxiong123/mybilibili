package core

import (
	"context"

	pb "mybilibili/internal/core/pb"
)

type CommentHandler struct {
	pb.UnimplementedCommentServiceServer
	svc *CommentService
}

func NewCommentHandler(svc *CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

func (h *CommentHandler) AddComment(ctx context.Context, req *pb.AddCommentRequest) (*pb.AddCommentResponse, error) {
	return h.svc.AddComment(ctx, req)
}

func (h *CommentHandler) ListComments(ctx context.Context, req *pb.ListCommentsRequest) (*pb.ListCommentsResponse, error) {
	return h.svc.ListComments(ctx, req)
}

func (h *CommentHandler) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.DeleteCommentResponse, error) {
	return h.svc.DeleteComment(ctx, req)
}

func (h *CommentHandler) AddReply(ctx context.Context, req *pb.AddReplyRequest) (*pb.AddReplyResponse, error) {
	return h.svc.AddReply(ctx, req)
}

func (h *CommentHandler) GetReplies(ctx context.Context, req *pb.GetRepliesRequest) (*pb.GetRepliesResponse, error) {
	return h.svc.GetReplies(ctx, req)
}

func (h *CommentHandler) DeleteReply(ctx context.Context, req *pb.DeleteReplyRequest) (*pb.DeleteReplyResponse, error) {
	return h.svc.DeleteReply(ctx, req)
}

func (h *CommentHandler) LikeComment(ctx context.Context, req *pb.LikeCommentRequest) (*pb.LikeCommentResponse, error) {
	return h.svc.LikeComment(ctx, req)
}

func (h *CommentHandler) UnlikeComment(ctx context.Context, req *pb.UnlikeCommentRequest) (*pb.UnlikeCommentResponse, error) {
	return h.svc.UnlikeComment(ctx, req)
}

func (h *CommentHandler) LikeReply(ctx context.Context, req *pb.LikeReplyRequest) (*pb.LikeReplyResponse, error) {
	return h.svc.LikeReply(ctx, req)
}

func (h *CommentHandler) UnlikeReply(ctx context.Context, req *pb.UnlikeReplyRequest) (*pb.UnlikeReplyResponse, error) {
	return h.svc.UnlikeReply(ctx, req)
}
