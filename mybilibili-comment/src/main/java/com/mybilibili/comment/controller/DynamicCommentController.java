package com.mybilibili.comment.controller;

import com.mybilibili.comment.service.DynamicCommentService;
import com.mybilibili.common.vo.DynamicCommentVO;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

@RestController
@RequestMapping("/dynamic/comment")
@Tag(name = "动态评论", description = "动态评论相关接口")
public class DynamicCommentController {

    @Autowired
    private DynamicCommentService dynamicCommentService;

    @GetMapping("/list")
    @Operation(summary = "获取动态评论列表", description = "获取指定动态的评论列表")
    public Result<List<DynamicCommentVO>> getComments(
            @RequestParam Integer dynamicId,
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "10") Integer size,
            @RequestParam(defaultValue = "time") String sort,
            HttpServletRequest request) {
        try {
            Integer currentUserId = getUserIdFromRequest(request);
            return dynamicCommentService.getCommentsByDynamicId(dynamicId, page, size, sort, currentUserId);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/add")
    @Operation(summary = "发表评论", description = "对动态发表评论")
    @SecurityRequirement(name = "JWT")
    public Result<?> addComment(
            @RequestParam Integer dynamicId,
            @RequestParam String content,
            @RequestParam(required = false) Integer parentId,
            @RequestParam(required = false) Integer replyUserId,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            return dynamicCommentService.addComment(userId, dynamicId, content, parentId, replyUserId);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/delete/{commentId}")
    @Operation(summary = "删除评论", description = "删除自己的评论")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteComment(@PathVariable Integer commentId, HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            return dynamicCommentService.deleteComment(userId, commentId);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/replies")
    @Operation(summary = "获取评论回复", description = "获取指定评论的回复列表")
    public Result<List<DynamicCommentVO>> getReplies(
            @RequestParam Integer commentId,
            HttpServletRequest request) {
        try {
            Integer currentUserId = getUserIdFromRequest(request);
            return dynamicCommentService.getRepliesByCommentId(commentId, currentUserId);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/like/{commentId}")
    @Operation(summary = "点赞评论", description = "点赞指定的评论")
    @SecurityRequirement(name = "JWT")
    public Result<?> likeComment(@PathVariable Integer commentId, HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            return dynamicCommentService.likeComment(userId, commentId);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/like/{commentId}")
    @Operation(summary = "取消点赞评论", description = "取消点赞指定的评论")
    @SecurityRequirement(name = "JWT")
    public Result<?> unlikeComment(@PathVariable Integer commentId, HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            return dynamicCommentService.unlikeComment(userId, commentId);
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