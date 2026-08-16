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

type AiClient struct {
	conn   *grpc.ClientConn
	client pb.AiServiceClient
}

func NewAiClient() (*AiClient, error) {
	addr := os.Getenv("AI_GRPC_ADDR")
	if addr == "" {
		addr = "127.0.0.1:9088"
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ai gRPC dial: %w", err)
	}
	return &AiClient{conn: conn, client: pb.NewAiServiceClient(conn)}, nil
}

func (c *AiClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *AiClient) ReviewContent(ctx context.Context, content, scene string) (bool, string) {
	if c == nil || c.client == nil {
		return true, ""
	}
	resp, err := c.client.ReviewContent(ctx, &pb.ReviewContentRequest{Content: content, Scene: scene})
	if err != nil {
		log.Printf("ai gRPC ReviewContent error: %v", err)
		return true, ""
	}
	return resp.Passed, resp.Reason
}

func (c *AiClient) ReviewComment(ctx context.Context, content string) (bool, error) {
	passed, _ := c.ReviewContent(ctx, content, "COMMENT")
	return passed, nil
}

func (c *AiClient) GetSummary(ctx context.Context, videoID, manuscriptID int64) (string, bool) {
	if c == nil || c.client == nil {
		return "", false
	}
	resp, err := c.client.GetSummary(ctx, &pb.GetSummaryRequest{VideoId: videoID, ManuscriptId: manuscriptID})
	if err != nil {
		log.Printf("ai gRPC GetSummary error: %v", err)
		return "", false
	}
	return resp.Summary, resp.HasSummary
}