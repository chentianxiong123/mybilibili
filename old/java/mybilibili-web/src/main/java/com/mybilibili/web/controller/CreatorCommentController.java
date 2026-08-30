package com.mybilibili.web.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.CreatorCommentVO;
import com.mybilibili.common.vo.ReplyVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.web.service.CreatorCommentService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 创作者评论管理控制器
 */
@RestController
@RequestMapping("/creator/comments")
@Tag(name = "创作者评论管理", description = "创作者管理自己稿件下的评论")
public class CreatorCommentController {

    @Autowired
    private CreatorCommentService creatorCommentService;

    /**
     * 获取当前用户所有稿件的评论列表（分页，支持按稿件筛选）
     *
     * @param page         页码
     * @param size         每页数量
     * @param manuscriptId 稿件ID（可选）
     * @param request      HTTP请求
     * @return 评论列表和分页信息
     */
    @GetMapping
    @Operation(summary = "获取创作者评论列表", description = "获取当前用户所有稿件的评论列表，支持按稿件筛选")
    @SecurityRequirement(name = "JWT")
    public Result<Map<String, Object>> getCreatorComments(
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            @RequestParam(value = "manuscriptId", required = false) Integer manuscriptId,
            HttpServletRequest request) {
        try {
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

            // 获取评论列表
            List<CreatorCommentVO> comments = creatorCommentService.getCreatorComments(userId, manuscriptId, page, size);

            // 获取总数
            int total = creatorCommentService.countCreatorComments(userId, manuscriptId);

            // 构建返回结果
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

    /**
     * 删除评论（创作者权限，只能删除自己稿件下的评论）
     *
     * @param commentId 评论ID
     * @param request   HTTP请求
     * @return 操作结果
     */
    @DeleteMapping("/{commentId}")
    @Operation(summary = "删除评论", description = "创作者删除自己稿件下的评论")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteComment(
            @PathVariable Integer commentId,
            HttpServletRequest request) {
        try {
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

            creatorCommentService.deleteCommentByCreator(commentId, userId);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    /**
     * 回复评论
     *
     * @param commentId    评论ID
     * @param content      回复内容
     * @param replyToUserId 回复目标用户ID（可选）
     * @param request      HTTP请求
     * @return 回复信息
     */
    @PostMapping("/{commentId}/reply")
    @Operation(summary = "回复评论", description = "创作者回复自己稿件下的评论")
    @SecurityRequirement(name = "JWT")
    public Result<ReplyVO> replyComment(
            @PathVariable Integer commentId,
            @RequestParam String content,
            @RequestParam(required = false) Integer replyToUserId,
            HttpServletRequest request) {
        try {
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

            ReplyVO replyVO = creatorCommentService.replyComment(commentId, userId, content, replyToUserId);
            return Result.success("回复成功", replyVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
