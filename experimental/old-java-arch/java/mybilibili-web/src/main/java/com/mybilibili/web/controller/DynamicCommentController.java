package com.mybilibili.web.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.web.service.DynamicCommentService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;

@RestController
@RequestMapping("/dynamic/comment")
@Tag(name = "动态评论", description = "动态评论相关接口")
public class DynamicCommentController {

    @Autowired
    private DynamicCommentService dynamicCommentService;

    @GetMapping("/list")
    @Operation(summary = "获取动态评论列表", description = "获取指定动态的评论列表")
    public Result<?> getComments(
            @RequestParam Integer dynamicId,
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "10") Integer size,
            HttpServletRequest request) {
        try {
            Integer currentUserId = null;
            try {
                currentUserId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            } catch (Exception ignored) {
            }
            return dynamicCommentService.getCommentsByDynamicId(dynamicId, page, size, currentUserId);
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
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            if (content == null || content.trim().isEmpty()) {
                return Result.error("评论内容不能为空");
            }
            return dynamicCommentService.addComment(userId, dynamicId, content.trim(), parentId, replyUserId);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/delete/{commentId}")
    @Operation(summary = "删除评论", description = "删除自己的评论")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteComment(@PathVariable Integer commentId, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            return dynamicCommentService.deleteComment(userId, commentId);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/replies")
    @Operation(summary = "获取评论回复", description = "获取指定评论的回复列表")
    public Result<?> getReplies(
            @RequestParam Integer commentId,
            HttpServletRequest request) {
        try {
            Integer currentUserId = null;
            try {
                currentUserId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            } catch (Exception ignored) {
            }
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
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
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
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            return dynamicCommentService.unlikeComment(userId, commentId);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
