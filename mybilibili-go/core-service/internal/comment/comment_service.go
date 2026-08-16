package comment

import (
	"context"
	"database/sql"
	"time"

	"mybilibili/core-service/internal/message"
	"mybilibili/pkg/abstraction"
	"mybilibili/pkg/errors"
	pb "mybilibili/pkg/pb"
)

type MessageRepository = message.MessageRepository

type CommentService struct {
	repo        *CommentRepository
	limiter     *commentRateLimiter
	messageRepo *MessageRepository
	reviewSvc   interface {
		ReviewComment(ctx context.Context, content string) (bool, error)
	}
	cacheStore abstraction.CacheStore
}

func NewCommentService(repo *CommentRepository) *CommentService {
	return &CommentService{
		repo:    repo,
		limiter: newCommentRateLimiter(10*time.Minute, 20),
	}
}

func (s *CommentService) Repo() *CommentRepository {
	return s.repo
}

func (s *CommentService) SetCacheStore(cs abstraction.CacheStore) {
	s.cacheStore = cs
}

func (s *CommentService) SetMessageRepo(mr *MessageRepository) {
	s.messageRepo = mr
}

func (s *CommentService) SetReviewService(rs interface {
	ReviewComment(ctx context.Context, content string) (bool, error)
}) {
	s.reviewSvc = rs
}

func (s *CommentService) AddComment(ctx context.Context, req *pb.AddCommentRequest) (*pb.AddCommentResponse, error) {
	if req.Content == "" {
		return nil, errors.ErrInvalidArgument("content required")
	}
	if s.limiter.record(req.UserId, time.Now()) {
		return nil, errors.ErrResourceExhausted("too many comments, please slow down")
	}
	if s.messageRepo != nil {
		var cnt int
		s.messageRepo.DB().QueryRowContext(ctx,
			`SELECT COUNT(*) FROM prohibited_words WHERE $1 ILIKE '%' || word || '%'`, req.Content).Scan(&cnt)
		if cnt > 0 {
			c := &Comment{
				ManuscriptID: req.ManuscriptId,
				UserID:       req.UserId,
				Content:      req.Content,
				Status:       1,
			}
			id, _ := s.repo.CreateComment(ctx, c)
			c.ID = id
			info := s.buildComment(ctx, c, req.UserId, nil)
			return &pb.AddCommentResponse{Comment: info}, nil
		}
	}
	c := &Comment{
		ManuscriptID: req.ManuscriptId,
		UserID:       req.UserId,
		Content:      req.Content,
	}
	id, err := s.repo.CreateComment(ctx, c)
	if err != nil {
		return nil, errors.ErrInternal("failed to create comment")
	}
	c.ID = id
	s.repo.UpsertDailyMetric(ctx, c.ManuscriptID, c.UserID, "comment_count", 1)

	s.repo.WriteContentReview(ctx, "comment", req.UserId, req.Content)
	if s.reviewSvc != nil {
		passed, _ := s.reviewSvc.ReviewComment(ctx, req.Content)
		if !passed {
			s.repo.UpdateCommentStatus(ctx, id, 1)
		}
	}

	info := s.buildComment(ctx, c, req.UserId, nil)
	return &pb.AddCommentResponse{Comment: info}, nil
}

func (s *CommentService) ListComments(ctx context.Context, req *pb.ListCommentsRequest) (*pb.ListCommentsResponse, error) {
	list, err := s.repo.ListByManuscript(ctx, req.ManuscriptId, req.Page, req.PageSize, req.Sort)
	if err != nil {
		return nil, errors.ErrInternal("database error")
	}

	var infos []*pb.CommentInfo
	for _, c := range list {
		infos = append(infos, s.buildComment(ctx, c, req.UserId, nil))
	}
	return &pb.ListCommentsResponse{Comments: infos}, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.DeleteCommentResponse, error) {
	if err := s.repo.Delete(ctx, req.Id, req.UserId); err != nil {
		return nil, errors.ErrPermissionDenied("cannot delete")
	}
	return &pb.DeleteCommentResponse{}, nil
}

func (s *CommentService) AddReply(ctx context.Context, req *pb.AddReplyRequest) (*pb.AddReplyResponse, error) {
	if req.Content == "" {
		return nil, errors.ErrInvalidArgument("content required")
	}
	if s.limiter.record(req.UserId, time.Now()) {
		return nil, errors.ErrResourceExhausted("too many comments, please slow down")
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
		return nil, errors.ErrInternal("failed to create reply")
	}
	rep.ID = id

	s.repo.IncrementReplyCount(ctx, req.CommentId)

	if parent, perr := s.repo.FindByID(ctx, req.CommentId); perr == nil {
		s.repo.UpsertDailyMetric(ctx, parent.ManuscriptID, req.UserId, "comment_count", 1)
	}

	s.sendReplyNotification(ctx, req.CommentId, req.UserId)

	info := s.buildReply(ctx, rep, req.UserId)
	return &pb.AddReplyResponse{Reply: info}, nil
}

func (s *CommentService) GetReplies(ctx context.Context, req *pb.GetRepliesRequest) (*pb.GetRepliesResponse, error) {
	list, err := s.repo.ListRepliesByComment(ctx, req.CommentId, req.Page, req.PageSize)
	if err != nil {
		return nil, errors.ErrInternal("database error")
	}

	var infos []*pb.ReplyInfo
	for _, rep := range list {
		infos = append(infos, s.buildReply(ctx, rep, req.UserId))
	}
	return &pb.GetRepliesResponse{Replies: infos}, nil
}

func (s *CommentService) DeleteReply(ctx context.Context, req *pb.DeleteReplyRequest) (*pb.DeleteReplyResponse, error) {
	if err := s.repo.DeleteReply(ctx, req.Id, req.UserId); err != nil {
		return nil, errors.ErrPermissionDenied("cannot delete")
	}
	return &pb.DeleteReplyResponse{}, nil
}

func (s *CommentService) LikeComment(ctx context.Context, req *pb.LikeCommentRequest) (*pb.LikeCommentResponse, error) {
	if err := s.repo.LikeTarget(ctx, "comment", req.CommentId, req.UserId); err != nil {
		return nil, errors.ErrInternal("failed to like")
	}
	s.sendCommentLikeNotification(ctx, "comment", req.CommentId, req.UserId)
	return &pb.LikeCommentResponse{}, nil
}

func (s *CommentService) UnlikeComment(ctx context.Context, req *pb.UnlikeCommentRequest) (*pb.UnlikeCommentResponse, error) {
	if err := s.repo.UnlikeTarget(ctx, "comment", req.CommentId, req.UserId); err != nil {
		return nil, errors.ErrInternal("failed to unlike")
	}
	return &pb.UnlikeCommentResponse{}, nil
}

func (s *CommentService) LikeReply(ctx context.Context, req *pb.LikeReplyRequest) (*pb.LikeReplyResponse, error) {
	if err := s.repo.LikeTarget(ctx, "reply", req.ReplyId, req.UserId); err != nil {
		return nil, errors.ErrInternal("failed to like")
	}
	s.sendCommentLikeNotification(ctx, "reply", req.ReplyId, req.UserId)
	return &pb.LikeReplyResponse{}, nil
}

func (s *CommentService) UnlikeReply(ctx context.Context, req *pb.UnlikeReplyRequest) (*pb.UnlikeReplyResponse, error) {
	if err := s.repo.UnlikeTarget(ctx, "reply", req.ReplyId, req.UserId); err != nil {
		return nil, errors.ErrInternal("failed to unlike")
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

func (s *CommentService) sendCommentLikeNotification(ctx context.Context, targetType string, targetID, senderID int64) {
	if s.messageRepo == nil {
		return
	}
	var ownerID int64
	table := "comments"
	if targetType == "reply" {
		table = "replies"
	}
	_ = s.messageRepo.DB().QueryRowContext(ctx,
		`SELECT user_id FROM `+table+` WHERE id = $1`, targetID).Scan(&ownerID)
	if ownerID == 0 || ownerID == senderID {
		return
	}
	_, _ = s.messageRepo.SendMessage(ctx, senderID, ownerID, "liked your comment", 6)
}

func (s *CommentService) sendReplyNotification(ctx context.Context, commentID, senderID int64) {
	if s.messageRepo == nil {
		return
	}
	var ownerID int64
	_ = s.messageRepo.DB().QueryRowContext(ctx,
		`SELECT user_id FROM comments WHERE id = $1`, commentID).Scan(&ownerID)
	if ownerID == 0 || ownerID == senderID {
		return
	}
	_, _ = s.messageRepo.SendMessage(ctx, senderID, ownerID, "replied to your comment", 2)
}
