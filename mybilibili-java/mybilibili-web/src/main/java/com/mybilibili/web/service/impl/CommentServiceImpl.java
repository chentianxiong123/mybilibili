package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.Comment;
import com.mybilibili.common.entity.Reply;
import com.mybilibili.common.enums.TargetType;
import com.mybilibili.common.vo.CommentVO;
import com.mybilibili.common.vo.ReplyVO;
import com.mybilibili.web.mapper.CommentMapper;
import com.mybilibili.web.mapper.ReplyMapper;
import com.mybilibili.web.mapper.UserMapper;
import com.mybilibili.web.service.CommentService;
import com.mybilibili.web.service.ContentReviewService;
import com.mybilibili.web.service.LikeService;
import com.mybilibili.web.mapper.MessageMapper;
import com.mybilibili.common.entity.Message;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.Map;

@Service
public class CommentServiceImpl implements CommentService {

    @Autowired
    private CommentMapper commentMapper;

    @Autowired
    private ReplyMapper replyMapper;

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private LikeService likeService;

    @Autowired
    private ContentReviewService contentReviewService;

    @Autowired
    private MessageMapper messageMapper;

    private static final String TARGET_TYPE_COMMENT = "COMMENT";
    private static final String TARGET_TYPE_REPLY = "REPLY";

    // 消息类型常量
    private static final int MESSAGE_TYPE_REPLY = 2;  // 回复我的
    private static final int MESSAGE_TYPE_AT = 3;     // @我的

    // 新增：支持多种目标类型的评论（统一使用manuscript_id）
    @Override
    public CommentVO addComment(TargetType targetType, Integer targetId, Integer userId, String content) {
        // 违禁词检测
        List<String> prohibitedWords = contentReviewService.detectProhibitedWords(content);
        boolean hasProhibitedWords = !prohibitedWords.isEmpty();

        // 创建评论实体
        Comment comment = new Comment();
        comment.setManuscriptId(targetId);
        comment.setUserId(userId);
        comment.setContent(content);
        comment.setLikeCount(0);
        comment.setReplyCount(0);
        // 如果有违禁词，状态设为 2 (REMOVED)，否则 0 (NORMAL)
        comment.setStatus(hasProhibitedWords ? 2 : 0);

        // 保存评论
        commentMapper.insert(comment);

        // 构建返回对象
        CommentVO commentVO = buildCommentVO(comment, userId);
        // 如果有违禁词，在返回对象中标记
        commentVO.setHasProhibitedWords(hasProhibitedWords);
        if (hasProhibitedWords) {
            commentVO.setProhibitedWords(prohibitedWords);
        }
        return commentVO;
    }

    // 新增：根据目标类型和ID获取评论列表
    @Override
    public List<CommentVO> getCommentsByTarget(TargetType targetType, Integer targetId, Integer page, Integer size, Integer userId) {
        int offset = (page - 1) * size;
        List<Comment> comments = commentMapper.selectByTargetTypeAndTargetId(targetType, targetId, offset, size);

        if (comments.isEmpty()) {
            return new ArrayList<>();
        }

        // 批量获取评论ID列表
        List<Integer> commentIds = new ArrayList<>();
        for (Comment comment : comments) {
            commentIds.add(comment.getId());
        }

        // 批量查询点赞状态
        Map<Integer, Boolean> likeStatusMap = likeService.batchIsLiked(userId, TARGET_TYPE_COMMENT, commentIds);

        // 批量查询点赞数
        Map<Integer, Integer> likeCountMap = likeService.batchGetLikeCount(TARGET_TYPE_COMMENT, commentIds);

        List<CommentVO> commentVOs = new ArrayList<>();
        for (Comment comment : comments) {
            CommentVO commentVO = buildCommentVO(comment, userId, likeStatusMap, likeCountMap);
            // 获取前3条回复
            List<Reply> replies = replyMapper.selectByCommentId(comment.getId(), 0, 3);
            List<ReplyVO> replyVOs = new ArrayList<>();

            // 批量查询回复的点赞状态
            List<Integer> replyIds = new ArrayList<>();
            for (Reply reply : replies) {
                replyIds.add(reply.getId());
            }
            Map<Integer, Boolean> replyLikeStatusMap = likeService.batchIsLiked(userId, TARGET_TYPE_REPLY, replyIds);
            Map<Integer, Integer> replyLikeCountMap = likeService.batchGetLikeCount(TARGET_TYPE_REPLY, replyIds);

            for (Reply reply : replies) {
                replyVOs.add(buildReplyVO(reply, userId, replyLikeStatusMap, replyLikeCountMap));
            }
            commentVO.setReplies(replyVOs);
            commentVOs.add(commentVO);
        }

        return commentVOs;
    }

    // 向后兼容：保留原有接口实现
    @Override
    public CommentVO addComment(Integer manuscriptId, Integer userId, String content) {
        // 调用新方法，默认类型为 VIDEO
        return addComment(TargetType.VIDEO, manuscriptId, userId, content);
    }

    // 向后兼容：保留原有接口实现
    @Override
    public List<CommentVO> getCommentsByManuscriptId(Integer manuscriptId, Integer page, Integer size, Integer userId) {
        // 调用新方法，默认类型为 VIDEO
        return getCommentsByTarget(TargetType.VIDEO, manuscriptId, page, size, userId);
    }

    @Override
    public void deleteComment(Integer commentId, Integer userId) {
        // 检查评论是否存在
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }

        // 检查是否是评论的作者
        if (!comment.getUserId().equals(userId)) {
            throw new RuntimeException("没有权限删除此评论");
        }

        // 删除评论
        commentMapper.delete(commentId);
    }

    @Override
    public ReplyVO addReply(Integer commentId, Integer userId, String content, Integer replyToUserId) {
        // 检查评论是否存在
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }

        // 违禁词检测
        List<String> prohibitedWords = contentReviewService.detectProhibitedWords(content);
        boolean hasProhibitedWords = !prohibitedWords.isEmpty();

        // 如果有回复目标，自动添加@目标用户名
        if (replyToUserId != null) {
            com.mybilibili.common.entity.User targetUser = userMapper.findById(replyToUserId);
            if (targetUser != null) {
                content = "@" + targetUser.getNickname() + "：" + content;
            }
        }

        // 创建回复实体
        Reply reply = new Reply();
        reply.setCommentId(commentId);
        reply.setUserId(userId);
        reply.setReplyToUserId(replyToUserId);
        reply.setContent(content);
        reply.setLikeCount(0);
        // 如果有违禁词，状态设为 REMOVED，否则 NORMAL
        reply.setStatus(hasProhibitedWords ? "REMOVED" : "NORMAL");

        // 保存回复
        replyMapper.insert(reply);

        // 更新评论的回复数
        commentMapper.updateReplyCount(commentId, 1);

        // 发送消息通知（如果不是回复自己）
        if (replyToUserId != null && !replyToUserId.equals(userId)) {
            sendReplyMessage(userId, replyToUserId, content, commentId);
        }

        // 构建返回对象
        ReplyVO replyVO = buildReplyVO(reply, userId);
        // 如果有违禁词，在返回对象中标记
        replyVO.setHasProhibitedWords(hasProhibitedWords);
        if (hasProhibitedWords) {
            replyVO.setProhibitedWords(prohibitedWords);
        }
        return replyVO;
    }

    @Override
    public List<ReplyVO> getRepliesByCommentId(Integer commentId, Integer page, Integer size, Integer userId) {
        // 检查评论是否存在
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }

        int offset = (page - 1) * size;
        List<Reply> replies = replyMapper.selectByCommentId(commentId, offset, size);

        if (replies.isEmpty()) {
            return new ArrayList<>();
        }

        // 批量获取回复ID列表
        List<Integer> replyIds = new ArrayList<>();
        for (Reply reply : replies) {
            replyIds.add(reply.getId());
        }

        // 批量查询点赞状态
        Map<Integer, Boolean> likeStatusMap = likeService.batchIsLiked(userId, TARGET_TYPE_REPLY, replyIds);

        // 批量查询点赞数
        Map<Integer, Integer> likeCountMap = likeService.batchGetLikeCount(TARGET_TYPE_REPLY, replyIds);

        List<ReplyVO> replyVOs = new ArrayList<>();
        for (Reply reply : replies) {
            replyVOs.add(buildReplyVO(reply, userId, likeStatusMap, likeCountMap));
        }

        return replyVOs;
    }

    @Override
    public void deleteReply(Integer replyId, Integer userId) {
        // 检查回复是否存在
        Reply reply = replyMapper.selectById(replyId);
        if (reply == null) {
            throw new RuntimeException("回复不存在");
        }

        // 检查是否是回复的作者
        if (!reply.getUserId().equals(userId)) {
            throw new RuntimeException("没有权限删除此回复");
        }

        // 删除回复
        replyMapper.delete(replyId);

        // 更新评论的回复数
        commentMapper.updateReplyCount(reply.getCommentId(), -1);
    }

    @Override
    public void likeComment(Integer commentId, Integer userId) {
        // 检查评论是否存在
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }

        // 使用新的统一点赞服务
        boolean result = likeService.like(userId, TARGET_TYPE_COMMENT, commentId);
        if (result) {
            // 更新评论点赞数
            int newLikeCount = likeService.getLikeCount(TARGET_TYPE_COMMENT, commentId);
            commentMapper.updateLikeCountDirect(commentId, newLikeCount);
        }
    }

    @Override
    public void unlikeComment(Integer commentId, Integer userId) {
        // 检查评论是否存在
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }

        // 使用新的统一点赞服务
        boolean result = likeService.unlike(userId, TARGET_TYPE_COMMENT, commentId);
        if (result) {
            // 更新评论点赞数
            int newLikeCount = likeService.getLikeCount(TARGET_TYPE_COMMENT, commentId);
            commentMapper.updateLikeCountDirect(commentId, newLikeCount);
        }
    }

    @Override
    public void likeReply(Integer replyId, Integer userId) {
        // 检查回复是否存在
        Reply reply = replyMapper.selectById(replyId);
        if (reply == null) {
            throw new RuntimeException("回复不存在");
        }

        // 使用新的统一点赞服务
        boolean result = likeService.like(userId, TARGET_TYPE_REPLY, replyId);
        if (result) {
            // 更新回复点赞数
            int newLikeCount = likeService.getLikeCount(TARGET_TYPE_REPLY, replyId);
            replyMapper.updateLikeCountDirect(replyId, newLikeCount);
        }
    }

    @Override
    public void unlikeReply(Integer replyId, Integer userId) {
        // 检查回复是否存在
        Reply reply = replyMapper.selectById(replyId);
        if (reply == null) {
            throw new RuntimeException("回复不存在");
        }

        // 使用新的统一点赞服务
        boolean result = likeService.unlike(userId, TARGET_TYPE_REPLY, replyId);
        if (result) {
            // 更新回复点赞数
            int newLikeCount = likeService.getLikeCount(TARGET_TYPE_REPLY, replyId);
            replyMapper.updateLikeCountDirect(replyId, newLikeCount);
        }
    }

    private CommentVO buildCommentVO(Comment comment, Integer userId) {
        Map<Integer, Boolean> likeStatusMap = new java.util.HashMap<>();
        Map<Integer, Integer> likeCountMap = new java.util.HashMap<>();

        likeStatusMap.put(comment.getId(), likeService.isLiked(userId, TARGET_TYPE_COMMENT, comment.getId()));
        likeCountMap.put(comment.getId(), likeService.getLikeCount(TARGET_TYPE_COMMENT, comment.getId()));

        return buildCommentVO(comment, userId, likeStatusMap, likeCountMap);
    }

    private CommentVO buildCommentVO(Comment comment, Integer userId,
                                      Map<Integer, Boolean> likeStatusMap,
                                      Map<Integer, Integer> likeCountMap) {
        CommentVO commentVO = new CommentVO();
        commentVO.setId(comment.getId());
        commentVO.setVideoId(comment.getManuscriptId());
        commentVO.setUserId(comment.getUserId());
        commentVO.setContent(comment.getContent());
        commentVO.setLikeCount(likeCountMap.getOrDefault(comment.getId(), 0));
        commentVO.setReplyCount(comment.getReplyCount());
        commentVO.setCreateTime(comment.getCreatedAt());

        // 设置用户信息
        com.mybilibili.common.entity.User user = userMapper.findById(comment.getUserId());
        if (user != null) {
            commentVO.setUserName(user.getNickname());
            commentVO.setUserAvatar(user.getAvatar());
        }

        // 设置是否点赞
        commentVO.setLiked(likeStatusMap.getOrDefault(comment.getId(), false));

        return commentVO;
    }

    private ReplyVO buildReplyVO(Reply reply, Integer userId) {
        Map<Integer, Boolean> likeStatusMap = new java.util.HashMap<>();
        Map<Integer, Integer> likeCountMap = new java.util.HashMap<>();

        likeStatusMap.put(reply.getId(), likeService.isLiked(userId, TARGET_TYPE_REPLY, reply.getId()));
        likeCountMap.put(reply.getId(), likeService.getLikeCount(TARGET_TYPE_REPLY, reply.getId()));

        return buildReplyVO(reply, userId, likeStatusMap, likeCountMap);
    }

    private ReplyVO buildReplyVO(Reply reply, Integer userId,
                                  Map<Integer, Boolean> likeStatusMap,
                                  Map<Integer, Integer> likeCountMap) {
        ReplyVO replyVO = new ReplyVO();
        replyVO.setId(reply.getId());
        replyVO.setCommentId(reply.getCommentId());
        replyVO.setUserId(reply.getUserId());
        replyVO.setContent(reply.getContent());
        replyVO.setLikeCount(likeCountMap.getOrDefault(reply.getId(), 0));
        replyVO.setCreateTime(reply.getCreatedAt());

        // 设置用户信息
        com.mybilibili.common.entity.User user = userMapper.findById(reply.getUserId());
        if (user != null) {
            replyVO.setUserName(user.getNickname());
            replyVO.setUserAvatar(user.getAvatar());
        }

        // 设置回复目标用户信息
        if (reply.getReplyToUserId() != null) {
            com.mybilibili.common.entity.User targetUser = userMapper.findById(reply.getReplyToUserId());
            if (targetUser != null) {
                replyVO.setReplyToUserName(targetUser.getNickname());
            }
        }

        // 设置是否点赞
        replyVO.setLiked(likeStatusMap.getOrDefault(reply.getId(), false));

        return replyVO;
    }

    /**
     * 发送回复消息通知
     */
    private void sendReplyMessage(Integer senderId, Integer receiverId, String content, Integer commentId) {
        try {
            // 获取评论信息以获取视频ID
            Comment comment = commentMapper.selectById(commentId);
            Integer videoId = null;
            if (comment != null) {
                // 使用manuscriptId作为视频ID
                videoId = comment.getManuscriptId();
            }

            Message message = new Message();
            message.setSenderId(senderId);
            message.setReceiverId(receiverId);
            message.setContent("回复了你：" + content);
            message.setMessageType(MESSAGE_TYPE_REPLY);
            message.setTargetId(videoId);  // 设置视频ID
            message.setIsRead(false);
            message.setCreatedAt(new Date());
            messageMapper.insert(message);
        } catch (Exception e) {
            // 消息发送失败不影响回复功能，记录日志即可
            System.err.println("发送回复消息通知失败：" + e.getMessage());
        }
    }
}
