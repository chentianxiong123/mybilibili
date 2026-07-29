package com.mybilibili.comment.service;

import com.mybilibili.common.vo.DynamicCommentVO;
import com.mybilibili.common.vo.Result;

import java.util.List;

public interface DynamicCommentService {

    Result<DynamicCommentVO> addComment(Integer userId, Integer dynamicId, String content, Integer parentId, Integer replyUserId);

    Result<?> deleteComment(Integer userId, Integer commentId);

    Result<List<DynamicCommentVO>> getCommentsByDynamicId(Integer dynamicId, Integer page, Integer size, String sort, Integer currentUserId);

    Result<List<DynamicCommentVO>> getRepliesByCommentId(Integer commentId, Integer currentUserId);

    Result<?> likeComment(Integer userId, Integer commentId);

    Result<?> unlikeComment(Integer userId, Integer commentId);
}