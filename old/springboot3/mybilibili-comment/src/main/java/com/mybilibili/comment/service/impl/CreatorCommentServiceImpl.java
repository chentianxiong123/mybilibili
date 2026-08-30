package com.mybilibili.comment.service.impl;

import com.mybilibili.comment.feign.UserClient;
import com.mybilibili.comment.mapper.CommentMapper;
import com.mybilibili.comment.mapper.ReplyMapper;
import com.mybilibili.comment.service.CommentService;
import com.mybilibili.comment.service.CreatorCommentService;
import com.mybilibili.common.entity.Comment;
import com.mybilibili.common.entity.Reply;
import com.mybilibili.common.vo.CreatorCommentVO;
import com.mybilibili.common.vo.ReplyVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;

@Service
public class CreatorCommentServiceImpl implements CreatorCommentService {

    @Autowired
    private CommentMapper commentMapper;

    @Autowired
    private ReplyMapper replyMapper;

    @Autowired
    private CommentService commentService;

    @Autowired
    private UserClient userClient;

    @Override
    public List<CreatorCommentVO> getCreatorComments(Integer userId, Integer manuscriptId, Integer page, Integer size, String sort, String commentType) {
        int offset = (page - 1) * size;
        List<CreatorCommentVO> result = new ArrayList<>();

        if ("reply".equals(commentType)) {
            // Only replies
            List<Reply> replies = replyMapper.selectByCreatorId(userId, manuscriptId, offset, size, sort);
            for (Reply reply : replies) {
                result.add(buildReplyCommentVO(reply));
            }
        } else if ("comment".equals(commentType)) {
            // Only comments
            List<Comment> comments = commentMapper.selectByCreatorId(userId, manuscriptId, offset, size, sort);
            for (Comment comment : comments) {
                result.add(buildCommentVO(comment));
            }
        } else {
            // "all" - merge both, sorted by time
            List<Comment> comments = commentMapper.selectByCreatorId(userId, manuscriptId, offset, size, sort);
            List<Reply> replies = replyMapper.selectByCreatorId(userId, manuscriptId, offset, size, sort);
            for (Comment comment : comments) {
                result.add(buildCommentVO(comment));
            }
            for (Reply reply : replies) {
                result.add(buildReplyCommentVO(reply));
            }
            // Sort combined list
            result.sort((a, b) -> {
                if ("likes".equals(sort)) {
                    return Integer.compare(b.getLikeCount() != null ? b.getLikeCount() : 0, a.getLikeCount() != null ? a.getLikeCount() : 0);
                } else {
                    if (a.getCreateTime() == null || b.getCreateTime() == null) return 0;
                    return b.getCreateTime().compareTo(a.getCreateTime());
                }
            });
        }
        return result;
    }

    @Override
    public int countCreatorComments(Integer userId, Integer manuscriptId, String commentType) {
        if ("reply".equals(commentType)) {
            return replyMapper.countByCreatorId(userId, manuscriptId);
        } else if ("comment".equals(commentType)) {
            return commentMapper.countByCreatorId(userId, manuscriptId);
        } else {
            return commentMapper.countByCreatorId(userId, manuscriptId) + replyMapper.countByCreatorId(userId, manuscriptId);
        }
    }

    @Override
    public void deleteCommentByCreator(Integer commentId, Integer userId) {
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }
        commentMapper.deleteById(commentId);
    }

    @Override
    public void deleteReplyByCreator(Integer replyId, Integer userId) {
        Reply reply = replyMapper.selectById(replyId);
        if (reply == null) {
            throw new RuntimeException("回复不存在");
        }
        replyMapper.deleteById(replyId);
    }

    @Override
    public ReplyVO replyComment(Integer commentId, Integer userId, String content, Integer replyToUserId) {
        return commentService.addReply(commentId, userId, content, replyToUserId);
    }

    private CreatorCommentVO buildCommentVO(Comment comment) {
        CreatorCommentVO vo = new CreatorCommentVO();
        vo.setId(comment.getId());
        vo.setManuscriptId(comment.getManuscriptId());
        vo.setUserId(comment.getUserId());
        vo.setContent(comment.getContent());
        vo.setLikeCount(comment.getLikeCount());
        vo.setReplyCount(comment.getReplyCount());
        vo.setCreateTime(comment.getCreatedAt());
        vo.setCommentType("comment");

        UserVO user = getUserById(comment.getUserId());
        if (user != null) {
            vo.setUserName(user.getNickname());
            vo.setUserAvatar(user.getAvatar());
        }
        return vo;
    }

    private CreatorCommentVO buildReplyCommentVO(Reply reply) {
        CreatorCommentVO vo = new CreatorCommentVO();
        vo.setId(reply.getId());
        vo.setUserId(reply.getUserId());
        vo.setContent(reply.getContent());
        vo.setLikeCount(reply.getLikeCount());
        vo.setCreateTime(reply.getCreatedAt());
        vo.setCommentType("reply");
        vo.setParentCommentId(reply.getCommentId());

        // Get the manuscript id from parent comment
        Comment parentComment = commentMapper.selectById(reply.getCommentId());
        if (parentComment != null) {
            vo.setManuscriptId(parentComment.getManuscriptId());
        }

        UserVO user = getUserById(reply.getUserId());
        if (user != null) {
            vo.setUserName(user.getNickname());
            vo.setUserAvatar(user.getAvatar());
        }

        if (reply.getReplyToUserId() != null) {
            UserVO replyToUser = getUserById(reply.getReplyToUserId());
            if (replyToUser != null) {
                vo.setReplyToUserName(replyToUser.getNickname());
            }
        }
        return vo;
    }

    private UserVO getUserById(Integer userId) {
        if (userId == null) {
            return null;
        }
        try {
            Result<UserVO> result = userClient.getUserById(userId);
            if (result != null && result.getData() != null) {
                return result.getData();
            }
        } catch (Exception e) {
        }
        return null;
    }
}
