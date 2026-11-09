package com.mybilibili.comment.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.comment.service.CommentService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/comment/admin")
@Tag(name = "评论管理接口", description = "评论管理相关操作")
public class CommentAdminController {

    @Autowired
    private CommentService commentService;

    @GetMapping("/list")
    @Operation(summary = "获取评论列表", description = "获取评论列表，支持分页、搜索和视频ID筛选")
    public Result<?> getCommentList(
            @Parameter(description = "页码") @RequestParam(defaultValue = "1") Integer page,
            @Parameter(description = "每页数量") @RequestParam(defaultValue = "10") Integer size,
            @Parameter(description = "搜索关键词") @RequestParam(required = false) String keyword,
            @Parameter(description = "稿件ID") @RequestParam(required = false) Integer manuscriptId) {
        return commentService.getAdminCommentList(page, size, keyword, manuscriptId);
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取评论详情", description = "根据ID获取评论详情")
    public Result<?> getCommentById(@PathVariable Integer id) {
        return commentService.getAdminCommentById(id);
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除评论", description = "删除评论")
    public Result<?> deleteComment(@PathVariable Integer id) {
        commentService.deleteCommentByAdmin(id);
        return Result.success("删除成功", null);
    }

    @PutMapping("/{id}/status")
    @Operation(summary = "更新评论状态", description = "更新评论状态，例如审核通过/拒绝")
    public Result<?> updateCommentStatus(@PathVariable Integer id, @RequestParam Integer status) {
        commentService.updateCommentStatusByAdmin(id, status);
        return Result.success("状态更新成功", null);
    }
}