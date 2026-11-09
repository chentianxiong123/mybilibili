package com.mybilibili.web.service;

import com.mybilibili.common.entity.DynamicComment;
import com.mybilibili.common.vo.Result;

import java.util.List;

public interface DynamicCommentService {
    
    /**
     * 发表评论
     */
    Result<?> addComment(Integer userId, Integer dynamicId, String content, Integer parentId, Integer replyUserId);
    
    /**
     * 删除评论
     */
    Result<?> deleteComment(Integer userId, Integer commentId);
    
    /**
     * 获取动态评论列表
     */
    Result<List<DynamicComment>> getCommentsByDynamicId(Integer dynamicId, Integer page, Integer size, Integer currentUserId);
    
    /**
     * 获取评论回复列表
     */
    Result<List<DynamicComment>> getRepliesByCommentId(Integer commentId, Integer currentUserId);
    
    /**
     * 点赞评论（使用复用的 LikeService）
     */
    Result<?> likeComment(Integer userId, Integer commentId);
    
    /**
     * 取消点赞评论（使用复用的 LikeService）
     */
    Result<?> unlikeComment(Integer userId, Integer commentId);
    
    /**
     * 获取评论点赞数
     */
    int getCommentLikeCount(Integer commentId);
    
    /**
     * 检查用户是否已点赞评论
     */
    boolean isCommentLiked(Integer userId, Integer commentId);
}
