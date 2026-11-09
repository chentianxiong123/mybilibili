package com.mybilibili.comment.controller;

import com.mybilibili.comment.service.CommentService;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/admin/content-review")
@Tag(name = "内容审核接口", description = "评论和回复审核相关操作")
public class ContentReviewController {

    @Autowired
    private CommentService commentService;

    @GetMapping("/pending")
    @Operation(summary = "获取待审核列表", description = "获取被下架的评论和回复列表")
    public Result<?> getPendingList(
            @Parameter(description = "内容类型：COMMENT/REPLY") @RequestParam(required = false) String contentType,
            @Parameter(description = "页码") @RequestParam(defaultValue = "1") Integer page,
            @Parameter(description = "每页数量") @RequestParam(defaultValue = "10") Integer size) {
        return commentService.getPendingList(contentType, page, size);
    }

    @GetMapping("/all")
    @Operation(summary = "获取所有内容", description = "获取所有评论和回复")
    public Result<?> getAllContent(
            @Parameter(description = "内容类型：COMMENT/REPLY") @RequestParam(required = false) String contentType,
            @Parameter(description = "状态") @RequestParam(required = false) String status,
            @Parameter(description = "页码") @RequestParam(defaultValue = "1") Integer page,
            @Parameter(description = "每页数量") @RequestParam(defaultValue = "10") Integer size) {
        return commentService.getAllContent(contentType, status, page, size);
    }

    @PutMapping("/restore/{type}/{id}")
    @Operation(summary = "恢复内容", description = "恢复被下架的评论或回复")
    public Result<?> restoreContent(
            @PathVariable String type,
            @PathVariable Integer id) {
        return commentService.restoreContent(type, id);
    }

    @DeleteMapping("/{type}/{id}")
    @Operation(summary = "删除内容", description = "删除评论或回复")
    public Result<?> deleteContent(
            @PathVariable String type,
            @PathVariable Integer id) {
        return commentService.deleteContent(type, id);
    }

    @PostMapping("/batch")
    @Operation(summary = "批量处理", description = "批量恢复或删除内容")
    public Result<?> batchProcess(@RequestBody Map<String, Object> request) {
        String action = (String) request.get("action");
        @SuppressWarnings("unchecked")
        List<Map<String, Object>> items = (List<Map<String, Object>>) request.get("items");
        return commentService.batchProcess(action, items);
    }
}
