package com.mybilibili.admin.controller;

import com.mybilibili.admin.service.CommentService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/comments")
@Tag(name = "评论管理接口", description = "评论管理相关操作")
public class CommentController {

    @Autowired
    private CommentService commentService;

    @GetMapping
    @Operation(summary = "获取评论列表", description = "获取评论列表，支持分页、搜索和视频ID筛选")
    public Object getCommentList(
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "10") Integer size,
            @RequestParam(required = false) String keyword,
            @RequestParam(required = false) Integer videoId) {
        return commentService.getCommentList(page, size, keyword, videoId);
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取评论详情", description = "根据ID获取评论详情")
    public Object getCommentById(@PathVariable Integer id) {
        return commentService.getCommentById(id);
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除评论", description = "删除评论")
    public Object deleteComment(@PathVariable Integer id) {
        return commentService.deleteComment(id);
    }

    @PutMapping("/{id}/status")
    @Operation(summary = "更新评论状态", description = "更新评论状态，例如审核通过/拒绝")
    public Object updateCommentStatus(@PathVariable Integer id, @RequestParam Integer status) {
        return commentService.updateCommentStatus(id, status);
    }
}