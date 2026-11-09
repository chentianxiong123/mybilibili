package com.mybilibili.admin.service;

import com.mybilibili.common.vo.Result;

import java.util.Map;

public interface CommentService {
    Result<?> getCommentList(Integer page, Integer size, String keyword, Integer videoId);
    Result<?> getCommentById(Integer id);
    Result<?> deleteComment(Integer id);
    Result<?> updateCommentStatus(Integer id, Integer status);
}