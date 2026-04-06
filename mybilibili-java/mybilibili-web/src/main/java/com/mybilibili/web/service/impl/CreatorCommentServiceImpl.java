package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.Comment;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.Reply;
import com.mybilibili.common.vo.CreatorCommentVO;
import com.mybilibili.common.vo.ReplyVO;
import com.mybilibili.web.mapper.CommentMapper;
import com.mybilibili.web.mapper.ManuscriptMapper;
import com.mybilibili.web.mapper.ReplyMapper;
import com.mybilibili.web.mapper.UserMapper;
import com.mybilibili.web.service.CreatorCommentService;
import com.mybilibili.web.service.LikeService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * 创作者评论管理服务实现
 */
@Service
public class CreatorCommentServiceImpl implements CreatorCommentService {

    @Autowired
    private CommentMapper commentMapper;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private ReplyMapper replyMapper;

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private LikeService likeService;

    private static final String TARGET_TYPE_COMMENT = "COMMENT";
    private static final String TARGET_TYPE_REPLY = "REPLY";

    @Override
    public List<CreatorCommentVO> getCreatorComments(Integer userId, Integer manuscriptId, Integer page, Integer size) {
        int offset = (page - 1) * size;
        List<Comment> comments = commentMapper.selectByCreatorId(userId, manuscriptId, offset, size);

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

        List<CreatorCommentVO> commentVOs = new ArrayList<>();
        for (Comment comment : comments) {
            CreatorCommentVO commentVO = buildCreatorCommentVO(comment, userId, likeStatusMap, likeCountMap);
            commentVOs.add(commentVO);
        }

        return commentVOs;
    }

    @Override
    public int countCreatorComments(Integer userId, Integer manuscriptId) {
        return commentMapper.countByCreatorId(userId, manuscriptId);
    }

    @Override
    public void deleteCommentByCreator(Integer commentId, Integer userId) {
        // 检查评论是否存在
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }

        // 检查评论所属稿件是否属于该创作者
        Integer manuscriptId = comment.getManuscriptId();
        if (manuscriptId == null) {
            throw new RuntimeException("评论未关联稿件");
        }

        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null || !manuscript.getUserId().equals(userId)) {
            throw new RuntimeException("没有权限删除此评论");
        }

        // 删除评论
        commentMapper.delete(commentId);
    }

    @Override
    public ReplyVO replyComment(Integer commentId, Integer userId, String content, Integer replyToUserId) {
        // 检查评论是否存在
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }

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

        // 保存回复
        replyMapper.insert(reply);

        // 更新评论的回复数
        commentMapper.updateReplyCount(commentId, 1);

        // 构建返回对象
        return buildReplyVO(reply, userId);
    }

    private CreatorCommentVO buildCreatorCommentVO(Comment comment, Integer userId,
                                                    Map<Integer, Boolean> likeStatusMap,
                                                    Map<Integer, Integer> likeCountMap) {
        CreatorCommentVO commentVO = new CreatorCommentVO();
        commentVO.setId(comment.getId());
        commentVO.setManuscriptId(comment.getManuscriptId());
        commentVO.setUserId(comment.getUserId());
        commentVO.setContent(comment.getContent());
        commentVO.setLikeCount(likeCountMap.getOrDefault(comment.getId(), 0));
        commentVO.setReplyCount(comment.getReplyCount());
        commentVO.setCreateTime(comment.getCreatedAt());

        // 设置稿件信息
        Integer manuscriptId = comment.getManuscriptId();
        if (manuscriptId != null) {
            Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
            if (manuscript != null) {
                commentVO.setManuscriptTitle(manuscript.getTitle());
                commentVO.setManuscriptCover(manuscript.getCoverUrl());
            }
        }

        // 设置评论用户信息
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
        ReplyVO replyVO = new ReplyVO();
        replyVO.setId(reply.getId());
        replyVO.setCommentId(reply.getCommentId());
        replyVO.setUserId(reply.getUserId());
        replyVO.setContent(reply.getContent());
        replyVO.setLikeCount(likeService.getLikeCount(TARGET_TYPE_REPLY, reply.getId()));
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
        replyVO.setLiked(likeService.isLiked(userId, TARGET_TYPE_REPLY, reply.getId()));

        return replyVO;
    }
}
