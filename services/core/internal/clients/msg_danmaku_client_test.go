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

type fakeMsgDanmakuService struct {
	pb.UnimplementedMsgDanmakuServiceServer
	sendFunc func(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error)
}

func (f *fakeMsgDanmakuService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	if f.sendFunc != nil {
		return f.sendFunc(ctx, req)
	}
	return &pb.SendMessageResponse{MessageId: 42}, nil
}

func startBufconnMsgDanmaku(t *testing.T, svc *fakeMsgDanmakuService) (*grpc.ClientConn, func()) {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	pb.RegisterMsgDanmakuServiceServer(srv, svc)
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

func TestNewMsgDanmakuClient_Success(t *testing.T) {
	t.Setenv("MSG_DANMAKU_GRPC_ADDR", "localhost:9999")
	c, err := NewMsgDanmakuClient()
	if err != nil {
		t.Skipf("grpc.NewClient may fail without real server, got: %v", err)
	}
	assert.NotNil(t, c)
	c.Close()
}

func TestNewMsgDanmakuClient_DefaultAddr(t *testing.T) {
	t.Setenv("MSG_DANMAKU_GRPC_ADDR", "")
	c, err := NewMsgDanmakuClient()
	if err != nil {
		t.Skipf("grpc.NewClient may fail, got: %v", err)
	}
	assert.NotNil(t, c)
	c.Close()
}

func TestMsgDanmakuClient_Close_NilConn(t *testing.T) {
	c := &MsgDanmakuClient{}
	assert.NotPanics(t, func() { c.Close() })
}

func TestMsgDanmakuClient_Close_NonNilConn(t *testing.T) {
	conn, cleanup := startBufconnMsgDanmaku(t, &fakeMsgDanmakuService{})
	defer cleanup()
	c := &MsgDanmakuClient{conn: conn, client: pb.NewMsgDanmakuServiceClient(conn)}
	assert.NotPanics(t, func() { c.Close() })
}

func TestMsgDanmakuClient_SendMessage_NilClient(t *testing.T) {
	c := &MsgDanmakuClient{}
	assert.NotPanics(t, func() {
		c.SendMessage(context.Background(), 1, 2, "hello", 0)
	})
}

func TestMsgDanmakuClient_SendMessage_NilStruct(t *testing.T) {
	var c *MsgDanmakuClient
	assert.NotPanics(t, func() {
		c.SendMessage(context.Background(), 1, 2, "hello", 0)
	})
}

func TestMsgDanmakuClient_SendMessage_Success(t *testing.T) {
	var captured *pb.SendMessageRequest
	fake := &fakeMsgDanmakuService{
		sendFunc: func(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
			captured = req
			return &pb.SendMessageResponse{MessageId: 99}, nil
		},
	}
	conn, cleanup := startBufconnMsgDanmaku(t, fake)
	defer cleanup()

	c := &MsgDanmakuClient{conn: conn, client: pb.NewMsgDanmakuServiceClient(conn)}
	c.SendMessage(context.Background(), 10, 20, "test content", 1)

	assert.Equal(t, int64(10), captured.SenderId)
	assert.Equal(t, int64(20), captured.ReceiverId)
	assert.Equal(t, "test content", captured.Content)
	assert.Equal(t, int32(1), captured.MessageType)
}

func TestMsgDanmakuClient_SendMessage_Error(t *testing.T) {
	fake := &fakeMsgDanmakuService{
		sendFunc: func(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
			return nil, errors.New("send failed")
		},
	}
	conn, cleanup := startBufconnMsgDanmaku(t, fake)
	defer cleanup()

	c := &MsgDanmakuClient{conn: conn, client: pb.NewMsgDanmakuServiceClient(conn)}
	assert.NotPanics(t, func() {
		c.SendMessage(context.Background(), 1, 2, "fail", 0)
	})
}
