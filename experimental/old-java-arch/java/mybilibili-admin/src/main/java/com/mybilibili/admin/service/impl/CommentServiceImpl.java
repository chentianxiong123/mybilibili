package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.mapper.CommentMapper;
import com.mybilibili.admin.service.CommentService;
import com.mybilibili.common.entity.Comment;
import com.mybilibili.common.vo.Result;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class CommentServiceImpl implements CommentService {

    @Autowired
    private CommentMapper commentMapper;

    @Override
    public Result<?> getCommentList(Integer page, Integer size, String keyword, Integer videoId) {
        try {
            if (page == null || page < 1) page = 1;
            if (size == null || size < 1) size = 10;
            if (keyword == null) keyword = "";

            int offset = (page - 1) * size;
            List<Comment> comments = commentMapper.selectCommentsByKeyword(offset, size, keyword, videoId);
            int total = commentMapper.countCommentsByKeyword(keyword, videoId);

            Map<String, Object> data = new HashMap<>();
            data.put("list", comments);
            data.put("total", total);
            data.put("page", page);
            data.put("size", size);

            return Result.success("获取评论列表成功", data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> getCommentById(Integer id) {
        try {
            Comment comment = commentMapper.selectById(id);
            if (comment == null) {
                return Result.error("评论不存在");
            }
            return Result.success("获取评论详情成功", comment);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> deleteComment(Integer id) {
        try {
            Comment comment = commentMapper.selectById(id);
            if (comment == null) {
                return Result.error("评论不存在");
            }

            int result = commentMapper.delete(id);
            if (result > 0) {
                return Result.success("删除评论成功", null);
            } else {
                return Result.error("删除评论失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> updateCommentStatus(Integer id, Integer status) {
        try {
            Comment comment = commentMapper.selectById(id);
            if (comment == null) {
                return Result.error("评论不存在");
            }

            int result = commentMapper.updateStatus(id, status);
            if (result > 0) {
                return Result.success("更新评论状态成功", null);
            } else {
                return Result.error("更新评论状态失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}