package com.mybilibili.comment.controller;

import com.mybilibili.comment.service.CreatorCommentService;
import com.mybilibili.common.vo.CreatorCommentVO;
import com.mybilibili.common.vo.ReplyVO;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/creator/comments")
@Tag(name = "创作者评论管理", description = "创作者管理自己稿件下的评论")
public class CreatorCommentController {

    @Autowired
    private CreatorCommentService creatorCommentService;

    @GetMapping
    @Operation(summary = "获取创作者评论列表", description = "获取当前用户所有稿件的评论列表，支持按稿件筛选")
    @SecurityRequirement(name = "JWT")
    public Result<Map<String, Object>> getCreatorComments(
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            @RequestParam(value = "manuscriptId", required = false) Integer manuscriptId,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }

            List<CreatorCommentVO> comments = creatorCommentService.getCreatorComments(userId, manuscriptId, page, size);
            int total = creatorCommentService.countCreatorComments(userId, manuscriptId);

            Map<String, Object> result = new HashMap<>();
            result.put("list", comments);
            result.put("total", total);
            result.put("page", page);
            result.put("size", size);

            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{commentId}")
    @Operation(summary = "删除评论", description = "创作者删除自己稿件下的评论")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteComment(
            @PathVariable Integer commentId,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            creatorCommentService.deleteCommentByCreator(commentId, userId);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{commentId}/reply")
    @Operation(summary = "回复评论", description = "创作者回复自己稿件下的评论")
    @SecurityRequirement(name = "JWT")
    public Result<ReplyVO> replyComment(
            @PathVariable Integer commentId,
            @RequestParam String content,
            @RequestParam(required = false) Integer replyToUserId,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            ReplyVO replyVO = creatorCommentService.replyComment(commentId, userId, content, replyToUserId);
            return Result.success("回复成功", replyVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    private Integer getUserIdFromRequest(HttpServletRequest request) {
        String userIdStr = request.getHeader("X-User-Id");
        if (userIdStr != null && !userIdStr.isEmpty()) {
            try {
                return Integer.parseInt(userIdStr);
            } catch (NumberFormatException e) {
                return null;
            }
        }
        return null;
    }
}