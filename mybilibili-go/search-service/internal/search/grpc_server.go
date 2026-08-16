package search

import (
	"context"
	"log"

	pb "mybilibili/pkg/pb"
)

type GrpcServer struct {
	pb.UnimplementedSearchServiceServer
	svc *Service
}

func NewGrpcServer(svc *Service) *GrpcServer {
	return &GrpcServer{svc: svc}
}

func (s *GrpcServer) SearchVideos(ctx context.Context, req *pb.SearchVideosRequest) (*pb.SearchVideosResponse, error) {
	log.Printf("gRPC SearchVideos: keyword=%s page=%d", req.Keyword, req.Page)
	list, err := s.svc.Search(ctx, req.Keyword, req.CategoryId, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	resp := &pb.SearchVideosResponse{}
	for _, item := range list {
		pbItem := &pb.SearchVideoItem{}
		if id, ok := item["id"].(int64); ok {
			pbItem.Id = id
		} else if id, ok := item["manuscript_id"].(int64); ok {
			pbItem.Id = id
		}
		if title, ok := item["title"].(string); ok {
			pbItem.Title = title
		}
		if cover, ok := item["cover_url"].(string); ok {
			pbItem.Cover = cover
		}
		if cover, ok := item["cover"].(string); ok {
			pbItem.Cover = cover
		}
		resp.List = append(resp.List, pbItem)
	}
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	return resp, nil
}

func (s *GrpcServer) GetHotSearch(ctx context.Context, req *pb.GetHotSearchRequest) (*pb.GetHotSearchResponse, error) {
	keywords, err := s.svc.Hot(ctx)
	if err != nil {
		return &pb.GetHotSearchResponse{}, nil
	}
	resp := &pb.GetHotSearchResponse{}
	for _, kw := range keywords {
		resp.List = append(resp.List, &pb.HotKeyword{Keyword: kw})
	}
	return resp, nil
}

func (s *GrpcServer) GetSuggestions(ctx context.Context, req *pb.GetSuggestionsRequest) (*pb.GetSuggestionsResponse, error) {
	suggestions, err := s.svc.Suggest(ctx, req.Keyword, req.Limit)
	if err != nil {
		return &pb.GetSuggestionsResponse{}, nil
	}
	return &pb.GetSuggestionsResponse{Suggestions: suggestions}, nil
}