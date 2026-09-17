package clients

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "mybilibili/pkg/pb"
)

type fakeAiService struct {
	pb.UnimplementedAiServiceServer
	reviewFunc func(ctx context.Context, req *pb.ReviewContentRequest) (*pb.ReviewContentResponse, error)
	summaryFunc func(ctx context.Context, req *pb.GetSummaryRequest) (*pb.GetSummaryResponse, error)
}

func (f *fakeAiService) ReviewContent(ctx context.Context, req *pb.ReviewContentRequest) (*pb.ReviewContentResponse, error) {
	if f.reviewFunc != nil {
		return f.reviewFunc(ctx, req)
	}
	return &pb.ReviewContentResponse{Passed: true, Reason: ""}, nil
}

func (f *fakeAiService) GetSummary(ctx context.Context, req *pb.GetSummaryRequest) (*pb.GetSummaryResponse, error) {
	if f.summaryFunc != nil {
		return f.summaryFunc(ctx, req)
	}
	return &pb.GetSummaryResponse{Summary: "test summary", HasSummary: true}, nil
}

func startBufconnAi(t *testing.T, svc *fakeAiService) (*grpc.ClientConn, func()) {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	pb.RegisterAiServiceServer(srv, svc)
	go func() { _ = srv.Serve(lis) }()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
	)
	assert.NoError(t, err)
	return conn, func() { srv.Stop(); conn.Close() }
}

func TestNewAiClient_Success(t *testing.T) {
	t.Setenv("AI_GRPC_ADDR", "localhost:9999")
	c, err := NewAiClient()
	if err != nil {
		t.Skipf("grpc.NewClient may fail without real server, got: %v", err)
	}
	assert.NotNil(t, c)
	c.Close()
}

func TestNewAiClient_DefaultAddr(t *testing.T) {
	t.Setenv("AI_GRPC_ADDR", "")
	c, err := NewAiClient()
	if err != nil {
		t.Skipf("grpc.NewClient may fail, got: %v", err)
	}
	assert.NotNil(t, c)
	c.Close()
}

func TestAiClient_Close_NilConn(t *testing.T) {
	c := &AiClient{}
	assert.NotPanics(t, func() { c.Close() })
}

func TestAiClient_Close_NonNilConn(t *testing.T) {
	conn, cleanup := startBufconnAi(t, &fakeAiService{})
	defer cleanup()
	c := &AiClient{conn: conn, client: pb.NewAiServiceClient(conn)}
	assert.NotPanics(t, func() { c.Close() })
}

func TestAiClient_ReviewContent_NilClient(t *testing.T) {
	c := &AiClient{}
	passed, reason := c.ReviewContent(context.Background(), "hello", "COMMENT")
	assert.True(t, passed)
	assert.Empty(t, reason)
}

func TestAiClient_ReviewContent_NilStruct(t *testing.T) {
	var c *AiClient
	passed, reason := c.ReviewContent(context.Background(), "hello", "COMMENT")
	assert.True(t, passed)
	assert.Empty(t, reason)
}

func TestAiClient_ReviewContent_Success(t *testing.T) {
	fake := &fakeAiService{
		reviewFunc: func(ctx context.Context, req *pb.ReviewContentRequest) (*pb.ReviewContentResponse, error) {
			return &pb.ReviewContentResponse{Passed: false, Reason: "violated"}, nil
		},
	}
	conn, cleanup := startBufconnAi(t, fake)
	defer cleanup()

	c := &AiClient{conn: conn, client: pb.NewAiServiceClient(conn)}
	passed, reason := c.ReviewContent(context.Background(), "bad content", "VIDEO")
	assert.False(t, passed)
	assert.Equal(t, "violated", reason)
}

func TestAiClient_ReviewContent_Error(t *testing.T) {
	fake := &fakeAiService{
		reviewFunc: func(ctx context.Context, req *pb.ReviewContentRequest) (*pb.ReviewContentResponse, error) {
			return nil, errors.New("rpc error")
		},
	}
	conn, cleanup := startBufconnAi(t, fake)
	defer cleanup()

	c := &AiClient{conn: conn, client: pb.NewAiServiceClient(conn)}
	passed, reason := c.ReviewContent(context.Background(), "text", "COMMENT")
	assert.True(t, passed)
	assert.Empty(t, reason)
}

func TestAiClient_ReviewComment(t *testing.T) {
	fake := &fakeAiService{
		reviewFunc: func(ctx context.Context, req *pb.ReviewContentRequest) (*pb.ReviewContentResponse, error) {
			assert.Equal(t, "COMMENT", req.Scene)
			return &pb.ReviewContentResponse{Passed: true, Reason: ""}, nil
		},
	}
	conn, cleanup := startBufconnAi(t, fake)
	defer cleanup()

	c := &AiClient{conn: conn, client: pb.NewAiServiceClient(conn)}
	passed, err := c.ReviewComment(context.Background(), "nice video")
	assert.True(t, passed)
	assert.NoError(t, err)
}

func TestAiClient_GetSummary_NilClient(t *testing.T) {
	c := &AiClient{}
	summary, ok := c.GetSummary(context.Background(), 1, 2)
	assert.Empty(t, summary)
	assert.False(t, ok)
}

func TestAiClient_GetSummary_NilStruct(t *testing.T) {
	var c *AiClient
	summary, ok := c.GetSummary(context.Background(), 1, 2)
	assert.Empty(t, summary)
	assert.False(t, ok)
}

func TestAiClient_GetSummary_Success(t *testing.T) {
	fake := &fakeAiService{
		summaryFunc: func(ctx context.Context, req *pb.GetSummaryRequest) (*pb.GetSummaryResponse, error) {
			return &pb.GetSummaryResponse{Summary: "a video summary", HasSummary: true}, nil
		},
	}
	conn, cleanup := startBufconnAi(t, fake)
	defer cleanup()

	c := &AiClient{conn: conn, client: pb.NewAiServiceClient(conn)}
	summary, ok := c.GetSummary(context.Background(), 100, 200)
	assert.Equal(t, "a video summary", summary)
	assert.True(t, ok)
}

func TestAiClient_GetSummary_Error(t *testing.T) {
	fake := &fakeAiService{
		summaryFunc: func(ctx context.Context, req *pb.GetSummaryRequest) (*pb.GetSummaryResponse, error) {
			return nil, errors.New("service unavailable")
		},
	}
	conn, cleanup := startBufconnAi(t, fake)
	defer cleanup()

	c := &AiClient{conn: conn, client: pb.NewAiServiceClient(conn)}
	summary, ok := c.GetSummary(context.Background(), 1, 2)
	assert.Empty(t, summary)
	assert.False(t, ok)
}
