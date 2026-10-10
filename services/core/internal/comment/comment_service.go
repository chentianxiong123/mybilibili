package comment

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"mybilibili/pkg/abstraction"
	"mybilibili/pkg/errors"
	pb "mybilibili/pkg/pb"
	"mybilibili/pkg/repository"
)

type Notifier interface {
	SendMessage(ctx context.Context, senderID, receiverID int64, content string, msgType int32)
	SendNotification(ctx context.Context, senderID, receiverID int64, content string, msgType int32, targetID, commentID int64)
}

type CommentService struct {
	repo      *CommentRepository
	db        *sql.DB
	notifier  Notifier
	reviewSvc interface {
		ReviewComment(ctx context.Context, content string) (bool, error)
	}
	cacheStore abstraction.CacheStore

	// 评论与回复各自独立计数，阈值均可后台配置
	commentLimiter *commentRateLimiter
	replyLimiter   *commentRateLimiter

	prohibited  *prohibitedWordStore
	securityCfg *securityConfigProvider
}

func NewCommentService(repo *CommentRepository) *CommentService {
	cfg := defaultSecuritySettings()
	return &CommentService{
		repo:           repo,
		commentLimiter: newCommentRateLimiter(time.Duration(cfg.CommentWindowSeconds)*time.Second, cfg.CommentMaxCount),
		replyLimiter:   newCommentRateLimiter(time.Duration(cfg.ReplyWindowSeconds)*time.Second, cfg.ReplyMaxCount),
	}
}

func (s *CommentService) Repo() *CommentRepository {
	return s.repo
}

func (s *CommentService) SetDB(db *sql.DB) {
	s.db = db
}

// SetSecurityConfig 接上后台可配的限流阈值与违禁词缓存刷新间隔，
// 并拉起后台刷新协程。db 为 nil 时只保留默认阈值，不启协程。
func (s *CommentService) SetSecurityConfig(ctx context.Context, db *sql.DB) {
	if db == nil {
		return
	}
	s.securityCfg = newSecurityConfigProvider(db)
	cfg := s.securityCfg.get(ctx)
	s.apply(cfg)
	s.prohibited = newProhibitedWordStore(db, time.Duration(cfg.CacheRefreshIntervalSeconds)*time.Second)
	s.prohibited.StartRefresher(ctx, func() time.Duration {
		return time.Duration(s.securityCfg.get(ctx).CacheRefreshIntervalSeconds) * time.Second
	})
}

// InvalidateSecurityConfig 让下次读取强制回源，后台保存配置后调用以立即生效。
func (s *CommentService) InvalidateSecurityConfig(ctx context.Context) {
	if s.securityCfg != nil {
		s.securityCfg.invalidate()
	}
	s.refreshSecurityConfig(ctx)
}

// prohibitedWord 命中违禁词时返回该词，未命中或缓存不可用返回 ""。
func (s *CommentService) prohibitedWord(content string) string {
	if s.prohibited == nil {
		return ""
	}
	return s.prohibited.match(content)
}

func (s *CommentService) SetCacheStore(cs abstraction.CacheStore) {
	s.cacheStore = cs
}

func (s *CommentService) SetNotifier(n Notifier) {
	s.notifier = n
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
	if s.commentLimiter.record(req.UserId, time.Now()) {
		return nil, errors.ErrResourceExhausted("too many comments, please slow down")
	}
	// 命中违禁词：落库但标为待审核（status=1），跳过后续 AI 审核与计数
	if word := s.prohibitedWord(req.Content); word != "" {
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
	repository.UpsertDailyMetric(ctx, s.db, c.ManuscriptID, c.UserID, "comment_count", 1)
	_, _ = s.db.ExecContext(ctx, `UPDATE manuscripts SET comment_count = comment_count + 1 WHERE id = $1`, c.ManuscriptID)

	s.repo.WriteContentReview(ctx, "comment", req.UserId, req.Content)
	if s.reviewSvc != nil {
		passed, _ := s.reviewSvc.ReviewComment(ctx, req.Content)
		if !passed {
			s.repo.UpdateCommentStatus(ctx, id, 1)
		}
	}

	info := s.buildComment(ctx, c, req.UserId, nil)
	s.notifyMentions(ctx, req.Content, req.UserId, req.ManuscriptId, id)
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
	// 回复走独立的限流桶，与评论分开计数
	if s.replyLimiter.record(req.UserId, time.Now()) {
		return nil, errors.ErrResourceExhausted("too many replies, please slow down")
	}

	rep := &Reply{
		CommentID: req.CommentId,
		UserID:    req.UserId,
		Content:   req.Content,
	}
	// 命中违禁词：落库但标为待审核，公开列表只显示 NORMAL，管理员面板仍可见
	if word := s.prohibitedWord(req.Content); word != "" {
		rep.Status = statusPendingReview
	}
	if req.ReplyToUserId > 0 {
		rep.ReplyToUserID = sql.NullInt64{Int64: req.ReplyToUserId, Valid: true}
	}

	id, err := s.repo.CreateReply(ctx, rep)
	if err != nil {
		return nil, errors.ErrInternal("failed to create reply")
	}
	rep.ID = id

	// 待审核的回复不计入稿件评论数，也不推送通知
	if rep.Status == statusPendingReview {
		info := s.buildReply(ctx, rep, req.UserId)
		return &pb.AddReplyResponse{Reply: info}, nil
	}

	s.repo.IncrementReplyCount(ctx, req.CommentId)

	if parent, perr := s.repo.FindByID(ctx, req.CommentId); perr == nil {
		repository.UpsertDailyMetric(ctx, s.db, parent.ManuscriptID, req.UserId, "comment_count", 1)
		_, _ = s.db.ExecContext(ctx, `UPDATE manuscripts SET comment_count = comment_count + 1 WHERE id = $1`, parent.ManuscriptID)
	}

	s.sendReplyNotification(ctx, req.CommentId, req.UserId, req.Content, req.ReplyToUserId)
	if parent, perr := s.repo.FindByID(ctx, req.CommentId); perr == nil {
		s.notifyMentions(ctx, req.Content, req.UserId, parent.ManuscriptID, req.CommentId)
	}

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
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound("comment not found")
		}
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
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound("reply not found")
		}
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
	replyToUserID := int64(0)
	if rep.ReplyToUserID.Valid {
		replyToUserID = rep.ReplyToUserID.Int64
		replyUser, err := s.repo.FindUserByID(ctx, rep.ReplyToUserID.Int64)
		if err == nil {
			replyToUserName = replyUser.Nickname
		}
	}

	liked := false
	if currentUserID > 0 {
		liked, _ = s.repo.IsReplyLiked(ctx, rep.ID, currentUserID)
	}

	return replyToPB(rep, userName, userAvatar, userLevel, replyToUserName, replyToUserID, liked)
}

func (s *CommentService) sendCommentLikeNotification(ctx context.Context, targetType string, targetID, senderID int64) {
	if s.db == nil || s.notifier == nil {
		return
	}
	var ownerID, manuscriptID, commentID int64
	var text string
	if targetType == "reply" {
		rep, err := s.repo.FindReplyByID(ctx, targetID)
		if err != nil || rep.UserID == senderID {
			return
		}
		ownerID = rep.UserID
		text = rep.Content
		commentID = rep.CommentID
		if parent, err := s.repo.FindByID(ctx, rep.CommentID); err == nil {
			manuscriptID = parent.ManuscriptID
		}
	} else {
		c, err := s.repo.FindByID(ctx, targetID)
		if err != nil || c.UserID == senderID {
			return
		}
		ownerID = c.UserID
		text = c.Content
		commentID = c.ID
		manuscriptID = c.ManuscriptID
	}
	if ownerID == 0 {
		return
	}
	s.notifier.SendNotification(ctx, senderID, ownerID,
		"赞了你的评论\""+truncateRunes(text, 100)+"\"", 6, manuscriptID, commentID)
}

func (s *CommentService) sendReplyNotification(ctx context.Context, commentID, senderID int64, replyContent string, replyToUserID int64) {
	if s.db == nil || s.notifier == nil {
		return
	}
	parent, err := s.repo.FindByID(ctx, commentID)
	if err != nil {
		return
	}
	content := "回复了你的评论：" + truncateRunes(replyContent, 140)
	notified := map[int64]bool{}
	// 楼中楼被回复的人优先收到
	if replyToUserID > 0 && replyToUserID != senderID {
		s.notifier.SendNotification(ctx, senderID, replyToUserID, content, 2, parent.ManuscriptID, commentID)
		notified[replyToUserID] = true
	}
	if parent.UserID != 0 && parent.UserID != senderID && !notified[parent.UserID] {
		s.notifier.SendNotification(ctx, senderID, parent.UserID, content, 2, parent.ManuscriptID, commentID)
	}
}

// notifyMentions 解析内容里的 @昵称 并通知被@的人（type=3）。
// 评论和楼中楼回复都走这里；自己@自己不通知。
func (s *CommentService) notifyMentions(ctx context.Context, content string, senderID, manuscriptID, commentID int64) {
	if s.db == nil || s.notifier == nil {
		return
	}
	names := parseMentionNames(content)
	if len(names) == 0 {
		return
	}
	snippet := truncateRunes(content, 140)
	for _, name := range names {
		var uid int64
		err := s.db.QueryRowContext(ctx,
			`SELECT id FROM users WHERE nickname = $1 OR username = $1 LIMIT 1`, name).Scan(&uid)
		if err != nil || uid == 0 || uid == senderID {
			continue
		}
		s.notifier.SendNotification(ctx, senderID, uid,
			"在评论中@了你："+snippet, 3, manuscriptID, commentID)
	}
}

// parseMentionNames 提取 @xxx（@后连续非空字符，末尾标点剔除），去重返回。
func parseMentionNames(s string) []string {
	var out []string
	seen := map[string]bool{}
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] != '@' {
			continue
		}
		j := i + 1
		for j < len(runes) && runes[j] != '@' && runes[j] != ' ' && runes[j] != '\t' && runes[j] != '\n' {
			j++
		}
		name := strings.TrimRight(string(runes[i+1:j]), ",.;:!?，。；：！？、）】")
		if name != "" && !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
		i = j
	}
	return out
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
