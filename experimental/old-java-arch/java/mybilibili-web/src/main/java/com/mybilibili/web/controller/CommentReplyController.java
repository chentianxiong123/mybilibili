package com.mybilibili.web.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.entity.CommentReply;
import com.mybilibili.web.service.CommentReplyService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

@RestController
@RequestMapping("/comment/reply")
@Tag(name = "评论回复相关接口", description = "回复评论、获取回复列表、点赞回复")
public class CommentReplyController {

    @Autowired
    private CommentReplyService commentReplyService;

    @PostMapping("/add")
    @Operation(summary = "回复评论", description = "回复指定的评论")
    @SecurityRequirement(name = "JWT")
    public Result<?> replyComment(@RequestParam Integer commentId, 
                                @RequestParam Integer replyUserId, 
                                @RequestParam String content, 
                                HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            return commentReplyService.replyComment(currentUserId, commentId, replyUserId, content);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/list/{commentId}")
    @Operation(summary = "获取评论回复列表", description = "获取指定评论的回复列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<CommentReply>> getCommentReplies(@PathVariable Integer commentId, 
                                                     @RequestParam(defaultValue = "1") Integer page, 
                                                     @RequestParam(defaultValue = "10") Integer limit) {
        return commentReplyService.getCommentReplies(commentId, page, limit);
    }

    @PostMapping("/like/{id}")
    @Operation(summary = "点赞回复", description = "点赞指定的回复")
    @SecurityRequirement(name = "JWT")
    public Result<?> likeReply(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            return commentReplyService.likeReply(currentUserId, id);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/like/{id}")
    @Operation(summary = "取消点赞回复", description = "取消点赞指定的回复")
    @SecurityRequirement(name = "JWT")
    public Result<?> unlikeReply(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            return commentReplyService.unlikeReply(currentUserId, id);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
