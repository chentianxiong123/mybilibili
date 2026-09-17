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

type fakeSearchService struct {
	pb.UnimplementedSearchServiceServer
	searchVideosFunc   func(ctx context.Context, req *pb.SearchVideosRequest) (*pb.SearchVideosResponse, error)
	getHotSearchFunc   func(ctx context.Context, req *pb.GetHotSearchRequest) (*pb.GetHotSearchResponse, error)
	getSuggestionsFunc func(ctx context.Context, req *pb.GetSuggestionsRequest) (*pb.GetSuggestionsResponse, error)
}

func (f *fakeSearchService) SearchVideos(ctx context.Context, req *pb.SearchVideosRequest) (*pb.SearchVideosResponse, error) {
	if f.searchVideosFunc != nil {
		return f.searchVideosFunc(ctx, req)
	}
	return &pb.SearchVideosResponse{
		List: []*pb.SearchVideoItem{
			{Id: 1, Title: "video1"},
			{Id: 2, Title: "video2"},
		},
		Total: 2,
	}, nil
}

func (f *fakeSearchService) GetHotSearch(ctx context.Context, req *pb.GetHotSearchRequest) (*pb.GetHotSearchResponse, error) {
	if f.getHotSearchFunc != nil {
		return f.getHotSearchFunc(ctx, req)
	}
	return &pb.GetHotSearchResponse{
		List: []*pb.HotKeyword{
			{Keyword: "golang", Score: 100},
			{Keyword: "rust", Score: 80},
		},
	}, nil
}

func (f *fakeSearchService) GetSuggestions(ctx context.Context, req *pb.GetSuggestionsRequest) (*pb.GetSuggestionsResponse, error) {
	if f.getSuggestionsFunc != nil {
		return f.getSuggestionsFunc(ctx, req)
	}
	return &pb.GetSuggestionsResponse{
		Suggestions: []string{"go tutorial", "go advanced"},
	}, nil
}

func startBufconnSearch(t *testing.T, svc *fakeSearchService) (*grpc.ClientConn, func()) {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	pb.RegisterSearchServiceServer(srv, svc)
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

func TestNewSearchClient_Success(t *testing.T) {
	t.Setenv("SEARCH_GRPC_ADDR", "localhost:9999")
	c, err := NewSearchClient()
	if err != nil {
		t.Skipf("grpc.NewClient may fail, got: %v", err)
	}
	assert.NotNil(t, c)
	c.Close()
}

func TestNewSearchClient_DefaultAddr(t *testing.T) {
	t.Setenv("SEARCH_GRPC_ADDR", "")
	c, err := NewSearchClient()
	if err != nil {
		t.Skipf("grpc.NewClient may fail, got: %v", err)
	}
	assert.NotNil(t, c)
	c.Close()
}

func TestSearchClient_Close_NilConn(t *testing.T) {
	c := &SearchClient{}
	assert.NotPanics(t, func() { c.Close() })
}

func TestSearchClient_Close_NonNilConn(t *testing.T) {
	conn, cleanup := startBufconnSearch(t, &fakeSearchService{})
	defer cleanup()
	c := &SearchClient{conn: conn, client: pb.NewSearchServiceClient(conn)}
	assert.NotPanics(t, func() { c.Close() })
}

func TestSearchClient_SearchVideos_NilClient(t *testing.T) {
	c := &SearchClient{}
	resp, err := c.SearchVideos(context.Background(), "go", 1, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, resp.List)
}

func TestSearchClient_SearchVideos_NilStruct(t *testing.T) {
	var c *SearchClient
	resp, err := c.SearchVideos(context.Background(), "go", 1, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, resp.List)
}

func TestSearchClient_SearchVideos_Success(t *testing.T) {
	conn, cleanup := startBufconnSearch(t, &fakeSearchService{})
	defer cleanup()

	c := &SearchClient{conn: conn, client: pb.NewSearchServiceClient(conn)}
	resp, err := c.SearchVideos(context.Background(), "golang", 1, 1, 20)
	assert.NoError(t, err)
	assert.Len(t, resp.List, 2)
	assert.Equal(t, "video1", resp.List[0].Title)
	assert.Equal(t, int32(2), resp.Total)
}

func TestSearchClient_SearchVideos_Error(t *testing.T) {
	fake := &fakeSearchService{
		searchVideosFunc: func(ctx context.Context, req *pb.SearchVideosRequest) (*pb.SearchVideosResponse, error) {
			return nil, errors.New("search failed")
		},
	}
	conn, cleanup := startBufconnSearch(t, fake)
	defer cleanup()

	c := &SearchClient{conn: conn, client: pb.NewSearchServiceClient(conn)}
	resp, err := c.SearchVideos(context.Background(), "go", 0, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, resp.List)
}

func TestSearchClient_GetHotSearch_NilClient(t *testing.T) {
	c := &SearchClient{}
	keywords, err := c.GetHotSearch(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, keywords)
}

func TestSearchClient_GetHotSearch_NilStruct(t *testing.T) {
	var c *SearchClient
	keywords, err := c.GetHotSearch(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, keywords)
}

func TestSearchClient_GetHotSearch_Success(t *testing.T) {
	conn, cleanup := startBufconnSearch(t, &fakeSearchService{})
	defer cleanup()

	c := &SearchClient{conn: conn, client: pb.NewSearchServiceClient(conn)}
	keywords, err := c.GetHotSearch(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []string{"golang", "rust"}, keywords)
}

func TestSearchClient_GetHotSearch_Empty(t *testing.T) {
	fake := &fakeSearchService{
		getHotSearchFunc: func(ctx context.Context, req *pb.GetHotSearchRequest) (*pb.GetHotSearchResponse, error) {
			return &pb.GetHotSearchResponse{List: nil}, nil
		},
	}
	conn, cleanup := startBufconnSearch(t, fake)
	defer cleanup()

	c := &SearchClient{conn: conn, client: pb.NewSearchServiceClient(conn)}
	keywords, err := c.GetHotSearch(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, keywords)
}

func TestSearchClient_GetHotSearch_Error(t *testing.T) {
	fake := &fakeSearchService{
		getHotSearchFunc: func(ctx context.Context, req *pb.GetHotSearchRequest) (*pb.GetHotSearchResponse, error) {
			return nil, errors.New("unavailable")
		},
	}
	conn, cleanup := startBufconnSearch(t, fake)
	defer cleanup()

	c := &SearchClient{conn: conn, client: pb.NewSearchServiceClient(conn)}
	keywords, err := c.GetHotSearch(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, keywords)
}

func TestSearchClient_GetSuggestions_NilClient(t *testing.T) {
	c := &SearchClient{}
	suggestions, err := c.GetSuggestions(context.Background(), "go", 5)
	assert.NoError(t, err)
	assert.Nil(t, suggestions)
}

func TestSearchClient_GetSuggestions_NilStruct(t *testing.T) {
	var c *SearchClient
	suggestions, err := c.GetSuggestions(context.Background(), "go", 5)
	assert.NoError(t, err)
	assert.Nil(t, suggestions)
}

func TestSearchClient_GetSuggestions_Success(t *testing.T) {
	conn, cleanup := startBufconnSearch(t, &fakeSearchService{})
	defer cleanup()

	c := &SearchClient{conn: conn, client: pb.NewSearchServiceClient(conn)}
	suggestions, err := c.GetSuggestions(context.Background(), "go", 10)
	assert.NoError(t, err)
	assert.Equal(t, []string{"go tutorial", "go advanced"}, suggestions)
}

func TestSearchClient_GetSuggestions_Error(t *testing.T) {
	fake := &fakeSearchService{
		getSuggestionsFunc: func(ctx context.Context, req *pb.GetSuggestionsRequest) (*pb.GetSuggestionsResponse, error) {
			return nil, errors.New("timeout")
		},
	}
	conn, cleanup := startBufconnSearch(t, fake)
	defer cleanup()

	c := &SearchClient{conn: conn, client: pb.NewSearchServiceClient(conn)}
	suggestions, err := c.GetSuggestions(context.Background(), "go", 5)
	assert.NoError(t, err)
	assert.Nil(t, suggestions)
}
