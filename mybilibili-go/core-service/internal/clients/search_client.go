package clients

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "mybilibili/pkg/pb"
)

type SearchClient struct {
	conn   *grpc.ClientConn
	client pb.SearchServiceClient
}

func NewSearchClient() (*SearchClient, error) {
	addr := os.Getenv("SEARCH_GRPC_ADDR")
	if addr == "" {
		addr = "127.0.0.1:9084"
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("search gRPC dial: %w", err)
	}
	return &SearchClient{conn: conn, client: pb.NewSearchServiceClient(conn)}, nil
}

func (c *SearchClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *SearchClient) SearchVideos(ctx context.Context, keyword string, categoryID int64, page, pageSize int32) (*pb.SearchVideosResponse, error) {
	if c == nil || c.client == nil {
		return &pb.SearchVideosResponse{}, nil
	}
	resp, err := c.client.SearchVideos(ctx, &pb.SearchVideosRequest{
		Keyword: keyword, CategoryId: categoryID, Page: page, PageSize: pageSize,
	})
	if err != nil {
		log.Printf("search gRPC SearchVideos error: %v", err)
		return &pb.SearchVideosResponse{}, nil
	}
	return resp, nil
}

func (c *SearchClient) GetHotSearch(ctx context.Context) ([]string, error) {
	if c == nil || c.client == nil {
		return nil, nil
	}
	resp, err := c.client.GetHotSearch(ctx, &pb.GetHotSearchRequest{})
	if err != nil {
		log.Printf("search gRPC GetHotSearch error: %v", err)
		return nil, nil
	}
	var out []string
	for _, k := range resp.List {
		out = append(out, k.Keyword)
	}
	return out, nil
}

func (c *SearchClient) GetSuggestions(ctx context.Context, keyword string, limit int32) ([]string, error) {
	if c == nil || c.client == nil {
		return nil, nil
	}
	resp, err := c.client.GetSuggestions(ctx, &pb.GetSuggestionsRequest{Keyword: keyword, Limit: limit})
	if err != nil {
		log.Printf("search gRPC GetSuggestions error: %v", err)
		return nil, nil
	}
	return resp.Suggestions, nil
}