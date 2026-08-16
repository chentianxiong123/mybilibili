package ai

import (
	"context"

	pb "mybilibili/pkg/pb"
)

type GrpcServer struct {
	pb.UnimplementedAiServiceServer
	reviewSvc  *ReviewService
	summarySvc *SummaryService
}

func NewGrpcServer(reviewSvc *ReviewService, summarySvc *SummaryService) *GrpcServer {
	return &GrpcServer{reviewSvc: reviewSvc, summarySvc: summarySvc}
}

func (s *GrpcServer) ReviewContent(ctx context.Context, req *pb.ReviewContentRequest) (*pb.ReviewContentResponse, error) {
	scene := req.Scene
	if scene == "" {
		scene = "COMMENT"
	}
	resp, err := s.reviewSvc.Moderate(ctx, req.Content, scene)
	if err != nil {
		return &pb.ReviewContentResponse{Passed: true, Reason: "", Score: 0}, nil
	}
	passed, _ := resp["passed"].(bool)
	reason, _ := resp["reason"].(string)
	return &pb.ReviewContentResponse{Passed: passed, Reason: reason, Score: 0}, nil
}

func (s *GrpcServer) GetSummary(ctx context.Context, req *pb.GetSummaryRequest) (*pb.GetSummaryResponse, error) {
	if s.summarySvc == nil {
		return &pb.GetSummaryResponse{HasSummary: false}, nil
	}
	summary, err := s.summarySvc.GetSummary(ctx, req.VideoId)
	if err != nil {
		return &pb.GetSummaryResponse{HasSummary: false}, nil
	}
	return &pb.GetSummaryResponse{Summary: summary, HasSummary: true}, nil
}