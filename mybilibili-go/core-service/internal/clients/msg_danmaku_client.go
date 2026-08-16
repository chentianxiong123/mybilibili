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

type MsgDanmakuClient struct {
	conn   *grpc.ClientConn
	client pb.MsgDanmakuServiceClient
}

func NewMsgDanmakuClient() (*MsgDanmakuClient, error) {
	addr := os.Getenv("MSG_DANMAKU_GRPC_ADDR")
	if addr == "" {
		addr = "127.0.0.1:9086"
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("msg-danmaku gRPC dial: %w", err)
	}
	return &MsgDanmakuClient{conn: conn, client: pb.NewMsgDanmakuServiceClient(conn)}, nil
}

func (c *MsgDanmakuClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *MsgDanmakuClient) SendMessage(ctx context.Context, senderID, receiverID int64, content string, msgType int32) {
	if c == nil || c.client == nil {
		return
	}
	_, err := c.client.SendMessage(ctx, &pb.SendMessageRequest{
		SenderId: senderID, ReceiverId: receiverID, Content: content, MessageType: msgType,
	})
	if err != nil {
		log.Printf("msg-danmaku gRPC SendMessage error: %v", err)
	}
}