package com.mybilibili.web.controller;

import com.mybilibili.common.entity.Video;
import com.mybilibili.common.enums.TargetType;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.CommentVO;
import com.mybilibili.common.vo.ReplyVO;
import com.mybilibili.web.mapper.VideoMapper;
import com.mybilibili.web.service.CommentService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.ArrayList;
import java.util.List;

@RestController
@RequestMapping("/comment")
@Tag(name = "评论相关接口", description = "评论的发表、查询、删除等操作")
public class CommentController {

    @Autowired
    private CommentService commentService;

    @Autowired
    private VideoMapper videoMapper;

    // 新增：统一的评论接口，支持多种目标类型
    @PostMapping("/add")
    @Operation(summary = "发表评论（新接口）", description = "支持对视频或动态发表评论")
    @SecurityRequirement(name = "JWT")
    public Result<CommentVO> addCommentV2(
            @RequestParam String targetType,
            @RequestParam Integer targetId,
            @RequestParam String content,
            HttpServletRequest request) {
        try {
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

            // 转换 targetType
            TargetType type = TargetType.fromCode(targetType);

            CommentVO commentVO = commentService.addComment(type, targetId, userId, content);
            return Result.success("评论成功", commentVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    // 新增：统一的获取评论列表接口
    @GetMapping("/list")
    @Operation(summary = "获取评论列表（新接口）", description = "支持获取视频或动态的评论列表")
    public Result<List<CommentVO>> getCommentsV2(
            @RequestParam String targetType,
            @RequestParam Integer targetId,
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            HttpServletRequest request) {
        try {
            // 尝试从JWT中获取用户ID，未登录则为null
            Integer userId = null;
            String token = request.getHeader("Authorization");
            if (token != null) {
                try {
                    userId = JwtUtils.getUserIdFromToken(token);
                } catch (Exception e) {
                    // token无效，继续处理
                }
            }

            // 转换 targetType
            TargetType type = TargetType.fromCode(targetType);

            List<CommentVO> comments = commentService.getCommentsByTarget(type, targetId, page, size, userId);
            return Result.success("获取成功", comments);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    // 向后兼容：保留原有接口
    @PostMapping
    @Operation(summary = "发表评论", description = "发表视频评论")
    @SecurityRequirement(name = "JWT")
    public Result<CommentVO> addComment(
            @RequestParam Integer videoId,
            @RequestParam String content,
            HttpServletRequest request) {
        try {
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

            // 根据videoId获取manuscriptId，评论关联到稿件
            Video video = videoMapper.selectById(videoId);
            if (video == null || video.getManuscriptId() == null) {
                return Result.error("视频不存在或未关联稿件");
            }
            Integer manuscriptId = video.getManuscriptId();

            CommentVO commentVO = commentService.addComment(manuscriptId, userId, content);
            return Result.success("评论成功", commentVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    // 向后兼容：保留原有接口
    @GetMapping("/video/{id}")
    @Operation(summary = "获取视频评论", description = "获取视频的评论列表")
    public Result<List<CommentVO>> getCommentsByVideoId(
            @PathVariable Integer id,
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            HttpServletRequest request) {
        try {
            // 尝试从JWT中获取用户ID，未登录则为null
            Integer userId = null;
            String token = request.getHeader("Authorization");
            if (token != null) {
                try {
                    userId = JwtUtils.getUserIdFromToken(token);
                } catch (Exception e) {
                    //  token无效，继续处理
                }
            }

            // 根据videoId获取manuscriptId，查询稿件的评论
            Video video = videoMapper.selectById(id);
            if (video == null || video.getManuscriptId() == null) {
                return Result.success("获取成功", new ArrayList<>());
            }
            Integer manuscriptId = video.getManuscriptId();

            List<CommentVO> comments = commentService.getCommentsByManuscriptId(manuscriptId, page, size, userId);
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
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

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
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

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
            // 尝试从JWT中获取用户ID，未登录则为null
            Integer userId = null;
            String token = request.getHeader("Authorization");
            if (token != null) {
                try {
                    userId = JwtUtils.getUserIdFromToken(token);
                } catch (Exception e) {
                    //  token无效，继续处理
                }
            }

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
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

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
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

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
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

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
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

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
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

            commentService.unlikeReply(id, userId);
            return Result.success("取消点赞成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
