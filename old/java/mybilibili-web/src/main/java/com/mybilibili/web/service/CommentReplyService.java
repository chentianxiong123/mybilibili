package com.mybilibili.web.service;

import com.mybilibili.common.entity.CommentReply;
import com.mybilibili.common.vo.Result;

import java.util.List;

public interface CommentReplyService {
    // 回复评论
    Result<?> replyComment(Integer userId, Integer commentId, Integer replyUserId, String content);
    
    // 获取评论的回复列表
    Result<List<CommentReply>> getCommentReplies(Integer commentId, Integer page, Integer limit);
    
    // 点赞回复
    Result<?> likeReply(Integer userId, Integer replyId);
    
    // 取消点赞回复
    Result<?> unlikeReply(Integer userId, Integer replyId);
}
