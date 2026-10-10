package clients

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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
	c.SendNotification(ctx, senderID, receiverID, content, msgType, 0, 0)
}

// SendNotification 发通知类消息（2=回复 3=@ 4=赞稿件 6=赞评论），可带稿件/评论 linkage。
// fire-and-forget：失败只打日志，绝不能影响点赞/评论主流程。
// 必须脱离请求 ctx：HTTP handler 一返回 ctx 就被取消，继承它会让通知
// 全部变成 "context canceled"（点赞通知实测丢失）。
func (c *MsgDanmakuClient) SendNotification(ctx context.Context, senderID, receiverID int64, content string, msgType int32, targetID, commentID int64) {
	if c == nil || c.client == nil {
		return
	}
	nctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer cancel()
	_, err := c.client.SendMessage(nctx, &pb.SendMessageRequest{
		SenderId: senderID, ReceiverId: receiverID, Content: content, MessageType: msgType,
		TargetId: targetID, CommentId: commentID,
	})
	if err != nil {
		log.Printf("msg-danmaku gRPC SendMessage error: %v", err)
	}
}