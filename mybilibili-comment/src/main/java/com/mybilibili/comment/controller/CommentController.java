package com.mybilibili.comment.controller;

import com.mybilibili.comment.service.CommentService;
import com.mybilibili.common.vo.CommentVO;
import com.mybilibili.common.vo.ReplyVO;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

@RestController
@RequestMapping("/comment")
@Tag(name = "评论相关接口", description = "评论的发表、查询、删除等操作")
public class CommentController {

    @Autowired
    private CommentService commentService;

    @PostMapping("/add")
    @Operation(summary = "发表评论", description = "对稿件发表评论")
    @SecurityRequirement(name = "JWT")
    public Result<CommentVO> addComment(
            @RequestParam Integer manuscriptId,
            @RequestParam String content,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            CommentVO commentVO = commentService.addCommentByManuscriptId(manuscriptId, userId, content);
            return Result.success("评论成功", commentVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/list")
    @Operation(summary = "获取评论列表", description = "获取稿件的评论列表")
    public Result<List<CommentVO>> getComments(
            @RequestParam Integer manuscriptId,
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            @RequestParam(value = "sort", defaultValue = "new") String sort,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            List<CommentVO> comments = commentService.getCommentsByManuscriptId(manuscriptId, page, size, userId, sort);
            return Result.success("获取成功", comments);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/video/{id}")
    @Operation(summary = "获取视频评论", description = "获取视频的评论列表")
    public Result<List<CommentVO>> getCommentsByVideoId(
            @PathVariable Integer id,
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            @RequestParam(value = "sort", defaultValue = "new") String sort,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            List<CommentVO> comments = commentService.getCommentsByManuscriptId(id, page, size, userId, sort);
            return Result.success("获取成功", comments);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除评论", description = "删除自己的评论")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteComment(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            commentService.deleteComment(id, userId);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/reply")
    @Operation(summary = "回复评论", description = "回复视频评论")
    @SecurityRequirement(name = "JWT")
    public Result<ReplyVO> addReply(
            @RequestParam Integer commentId,
            @RequestParam String content,
            @RequestParam(required = false) Integer replyToUserId,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            ReplyVO replyVO = commentService.addReply(commentId, userId, content, replyToUserId);
            return Result.success("回复成功", replyVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}/replies")
    @Operation(summary = "获取评论回复", description = "获取评论的回复列表")
    public Result<List<ReplyVO>> getRepliesByCommentId(
            @PathVariable Integer id,
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            List<ReplyVO> replies = commentService.getRepliesByCommentId(id, page, size, userId);
            return Result.success("获取成功", replies);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/reply/{id}")
    @Operation(summary = "删除回复", description = "删除自己的回复")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteReply(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            commentService.deleteReply(id, userId);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{id}/like")
    @Operation(summary = "点赞评论", description = "点赞视频评论")
    @SecurityRequirement(name = "JWT")
    public Result<?> likeComment(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            commentService.likeComment(id, userId);
            return Result.success("点赞成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}/like")
    @Operation(summary = "取消点赞评论", description = "取消点赞视频评论")
    @SecurityRequirement(name = "JWT")
    public Result<?> unlikeComment(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            commentService.unlikeComment(id, userId);
            return Result.success("取消点赞成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/reply/{id}/like")
    @Operation(summary = "点赞回复", description = "点赞评论回复")
    @SecurityRequirement(name = "JWT")
    public Result<?> likeReply(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            commentService.likeReply(id, userId);
            return Result.success("点赞成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/reply/{id}/like")
    @Operation(summary = "取消点赞回复", description = "取消点赞评论回复")
    @SecurityRequirement(name = "JWT")
    public Result<?> unlikeReply(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            commentService.unlikeReply(id, userId);
            return Result.success("取消点赞成功", null);
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