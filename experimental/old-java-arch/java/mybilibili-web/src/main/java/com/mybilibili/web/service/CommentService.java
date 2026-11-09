package com.mybilibili.web.service;

import com.mybilibili.common.enums.TargetType;
import com.mybilibili.common.vo.CommentVO;
import com.mybilibili.common.vo.ReplyVO;

import java.util.List;

public interface CommentService {
    // 新增：支持多种目标类型的评论
    CommentVO addComment(TargetType targetType, Integer targetId, Integer userId, String content);
    List<CommentVO> getCommentsByTarget(TargetType targetType, Integer targetId, Integer page, Integer size, Integer userId);

    // 向后兼容：保留原有接口
    CommentVO addComment(Integer manuscriptId, Integer userId, String content);
    List<CommentVO> getCommentsByManuscriptId(Integer manuscriptId, Integer page, Integer size, Integer userId);

    void deleteComment(Integer commentId, Integer userId);
    ReplyVO addReply(Integer commentId, Integer userId, String content, Integer replyToUserId);
    List<ReplyVO> getRepliesByCommentId(Integer commentId, Integer page, Integer size, Integer userId);
    void deleteReply(Integer replyId, Integer userId);
    void likeComment(Integer commentId, Integer userId);
    void unlikeComment(Integer commentId, Integer userId);
    void likeReply(Integer replyId, Integer userId);
    void unlikeReply(Integer replyId, Integer userId);
}
