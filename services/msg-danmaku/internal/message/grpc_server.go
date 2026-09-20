package message

import (
	"context"

	pb "mybilibili/pkg/pb"
)

type GrpcServer struct {
	pb.UnimplementedMsgDanmakuServiceServer
	repo  *MessageRepository
	notif *NotificationBroadcaster
	cache *UnreadCache
}

func NewGrpcServer(repo *MessageRepository, notif *NotificationBroadcaster, cache *UnreadCache) *GrpcServer {
	return &GrpcServer{repo: repo, notif: notif, cache: cache}
}

func (s *GrpcServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	msg, err := s.repo.SendMessage(ctx, req.SenderId, req.ReceiverId, req.Content, req.MessageType)
	if err != nil {
		return nil, err
	}
	if s.notif != nil {
		s.notif.Send(req.ReceiverId, &NotificationEvent{
			Type:    "message",
			Content: req.Content,
			FromUID: req.SenderId,
		})
		s.pushUnread(ctx, req.ReceiverId)
	}
	if s.cache != nil {
		s.cache.Invalidate(ctx, req.ReceiverId)
	}
	return &pb.SendMessageResponse{MessageId: msg.ID}, nil
}

func (s *GrpcServer) pushUnread(ctx context.Context, userID int64) {
	counts := s.repo.GetUnreadCountsByType(ctx, userID)
	if s.cache != nil {
		if cached, err := s.cache.Counts(ctx, userID); err == nil {
			counts = cached
		}
	}
	s.notif.Send(userID, &NotificationEvent{Type: "unread_counts", Data: counts})
}