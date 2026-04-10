package com.mybilibili.comment.service.impl;

import com.mybilibili.comment.feign.UserClient;
import com.mybilibili.comment.mapper.CommentMapper;
import com.mybilibili.comment.service.CommentService;
import com.mybilibili.comment.service.CreatorCommentService;
import com.mybilibili.common.entity.Comment;
import com.mybilibili.common.vo.CreatorCommentVO;
import com.mybilibili.common.vo.ReplyVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.stream.Collectors;

@Service
public class CreatorCommentServiceImpl implements CreatorCommentService {

    @Autowired
    private CommentMapper commentMapper;

    @Autowired
    private CommentService commentService;

    @Autowired
    private UserClient userClient;

    @Override
    public List<CreatorCommentVO> getCreatorComments(Integer userId, Integer manuscriptId, Integer page, Integer size) {
        int offset = (page - 1) * size;
        List<Comment> comments = commentMapper.selectByCreatorId(userId, manuscriptId, offset, size);

        return comments.stream().map(comment -> {
            CreatorCommentVO vo = new CreatorCommentVO();
            vo.setId(comment.getId());
            vo.setManuscriptId(comment.getManuscriptId());
            vo.setUserId(comment.getUserId());
            vo.setContent(comment.getContent());
            vo.setLikeCount(comment.getLikeCount());
            vo.setReplyCount(comment.getReplyCount());
            vo.setCreateTime(comment.getCreatedAt());

            UserVO user = getUserById(comment.getUserId());
            if (user != null) {
                vo.setUserName(user.getNickname());
                vo.setUserAvatar(user.getAvatar());
            }

            return vo;
        }).collect(Collectors.toList());
    }

    @Override
    public int countCreatorComments(Integer userId, Integer manuscriptId) {
        return commentMapper.countByCreatorId(userId, manuscriptId);
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
    public ReplyVO replyComment(Integer commentId, Integer userId, String content, Integer replyToUserId) {
        return commentService.addReply(commentId, userId, content, replyToUserId);
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