package core

import (
	"context"
	"database/sql"

	pb "mybilibili/internal/core/pb"
)

type CommentService struct {
	repo *CommentRepository
}

func NewCommentService(repo *CommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

func (s *CommentService) AddComment(ctx context.Context, req *pb.AddCommentRequest) (*pb.AddCommentResponse, error) {
	if req.Content == "" {
		return nil, ErrInvalidArgument("content required")
	}
	c := &Comment{
		ManuscriptID: req.ManuscriptId,
		UserID:       req.UserId,
		Content:      req.Content,
	}
	id, err := s.repo.CreateComment(ctx, c)
	if err != nil {
		return nil, ErrInternal("failed to create comment")
	}
	c.ID = id
	s.repo.UpsertDailyMetric(ctx, c.ManuscriptID, c.UserID, "comment_count", 1)

	info := s.buildComment(ctx, c, req.UserId, nil)
	return &pb.AddCommentResponse{Comment: info}, nil
}

func (s *CommentService) ListComments(ctx context.Context, req *pb.ListCommentsRequest) (*pb.ListCommentsResponse, error) {
	list, err := s.repo.ListByManuscript(ctx, req.ManuscriptId, req.Page, req.PageSize, req.Sort)
	if err != nil {
		return nil, ErrInternal("database error")
	}

	var infos []*pb.CommentInfo
	for _, c := range list {
		infos = append(infos, s.buildComment(ctx, c, req.UserId, nil))
	}
	return &pb.ListCommentsResponse{Comments: infos}, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.DeleteCommentResponse, error) {
	if err := s.repo.Delete(ctx, req.Id, req.UserId); err != nil {
		return nil, ErrPermissionDenied("cannot delete")
	}
	return &pb.DeleteCommentResponse{}, nil
}

func (s *CommentService) AddReply(ctx context.Context, req *pb.AddReplyRequest) (*pb.AddReplyResponse, error) {
	if req.Content == "" {
		return nil, ErrInvalidArgument("content required")
	}

	rep := &Reply{
		CommentID: req.CommentId,
		UserID:    req.UserId,
		Content:   req.Content,
	}
	if req.ReplyToUserId > 0 {
		rep.ReplyToUserID = sql.NullInt64{Int64: req.ReplyToUserId, Valid: true}
	}

	id, err := s.repo.CreateReply(ctx, rep)
	if err != nil {
		return nil, ErrInternal("failed to create reply")
	}
	rep.ID = id

	s.repo.IncrementReplyCount(ctx, req.CommentId)

	if parent, perr := s.repo.FindByID(ctx, req.CommentId); perr == nil {
		s.repo.UpsertDailyMetric(ctx, parent.ManuscriptID, req.UserId, "comment_count", 1)
	}

	info := s.buildReply(ctx, rep, req.UserId)
	return &pb.AddReplyResponse{Reply: info}, nil
}

func (s *CommentService) GetReplies(ctx context.Context, req *pb.GetRepliesRequest) (*pb.GetRepliesResponse, error) {
	list, err := s.repo.ListRepliesByComment(ctx, req.CommentId, req.Page, req.PageSize)
	if err != nil {
		return nil, ErrInternal("database error")
	}

	var infos []*pb.ReplyInfo
	for _, rep := range list {
		infos = append(infos, s.buildReply(ctx, rep, req.UserId))
	}
	return &pb.GetRepliesResponse{Replies: infos}, nil
}

func (s *CommentService) DeleteReply(ctx context.Context, req *pb.DeleteReplyRequest) (*pb.DeleteReplyResponse, error) {
	if err := s.repo.DeleteReply(ctx, req.Id, req.UserId); err != nil {
		return nil, ErrPermissionDenied("cannot delete")
	}
	return &pb.DeleteReplyResponse{}, nil
}

func (s *CommentService) LikeComment(ctx context.Context, req *pb.LikeCommentRequest) (*pb.LikeCommentResponse, error) {
	if err := s.repo.LikeTarget(ctx, "comment", req.CommentId, req.UserId); err != nil {
		return nil, ErrInternal("failed to like")
	}
	return &pb.LikeCommentResponse{}, nil
}

func (s *CommentService) UnlikeComment(ctx context.Context, req *pb.UnlikeCommentRequest) (*pb.UnlikeCommentResponse, error) {
	if err := s.repo.UnlikeTarget(ctx, "comment", req.CommentId, req.UserId); err != nil {
		return nil, ErrInternal("failed to unlike")
	}
	return &pb.UnlikeCommentResponse{}, nil
}

func (s *CommentService) LikeReply(ctx context.Context, req *pb.LikeReplyRequest) (*pb.LikeReplyResponse, error) {
	if err := s.repo.LikeTarget(ctx, "reply", req.ReplyId, req.UserId); err != nil {
		return nil, ErrInternal("failed to like")
	}
	return &pb.LikeReplyResponse{}, nil
}

func (s *CommentService) UnlikeReply(ctx context.Context, req *pb.UnlikeReplyRequest) (*pb.UnlikeReplyResponse, error) {
	if err := s.repo.UnlikeTarget(ctx, "reply", req.ReplyId, req.UserId); err != nil {
		return nil, ErrInternal("failed to unlike")
	}
	return &pb.UnlikeReplyResponse{}, nil
}

func (s *CommentService) buildComment(ctx context.Context, c *Comment, currentUserID int64, replies []*pb.ReplyInfo) *pb.CommentInfo {
	user, err := s.repo.FindUserByID(ctx, c.UserID)
	userName := ""
	userAvatar := ""
	userLevel := int32(0)
	if err == nil {
		userName = user.Nickname
		userAvatar = user.Avatar
		userLevel = user.Level
	}

	liked := false
	if currentUserID > 0 {
		liked, _ = s.repo.IsCommentLiked(ctx, c.ID, currentUserID)
	}

	return commentToPB(c, userName, userAvatar, userLevel, liked, replies)
}

func (s *CommentService) buildReply(ctx context.Context, rep *Reply, currentUserID int64) *pb.ReplyInfo {
	user, err := s.repo.FindUserByID(ctx, rep.UserID)
	userName := ""
	userAvatar := ""
	userLevel := int32(0)
	if err == nil {
		userName = user.Nickname
		userAvatar = user.Avatar
		userLevel = user.Level
	}

	replyToUserName := ""
	if rep.ReplyToUserID.Valid {
		replyUser, err := s.repo.FindUserByID(ctx, rep.ReplyToUserID.Int64)
		if err == nil {
			replyToUserName = replyUser.Nickname
		}
	}

	liked := false
	if currentUserID > 0 {
		liked, _ = s.repo.IsReplyLiked(ctx, rep.ID, currentUserID)
	}

	return replyToPB(rep, userName, userAvatar, userLevel, replyToUserName, liked)
}
