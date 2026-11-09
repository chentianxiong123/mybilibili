package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.CommentReply;
import com.mybilibili.common.vo.Result;
import com.mybilibili.web.mapper.CommentReplyMapper;
import com.mybilibili.web.service.CommentReplyService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.Date;
import java.util.List;

@Service
public class CommentReplyServiceImpl implements CommentReplyService {

    @Autowired
    private CommentReplyMapper commentReplyMapper;

    @Override
    public Result<?> replyComment(Integer userId, Integer commentId, Integer replyUserId, String content) {
        try {
            CommentReply reply = new CommentReply();
            reply.setCommentId(commentId);
            reply.setUserId(userId);
            reply.setReplyUserId(replyUserId);
            reply.setContent(content);
            reply.setLikeCount(0);
            reply.setCreatedAt(new Date());
            reply.setStatus(0); // 正常状态

            int result = commentReplyMapper.insert(reply);
            if (result > 0) {
                return Result.success("回复成功");
            } else {
                return Result.error("回复失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<List<CommentReply>> getCommentReplies(Integer commentId, Integer page, Integer limit) {
        try {
            int offset = (page - 1) * limit;
            List<CommentReply> replyList = commentReplyMapper.getByCommentId(commentId, offset, limit);
            return Result.success("获取成功", replyList);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> likeReply(Integer userId, Integer replyId) {
        try {
            CommentReply reply = commentReplyMapper.getById(replyId);
            if (reply == null) {
                return Result.error("回复不存在");
            }

            // 增加点赞数
            int newLikeCount = reply.getLikeCount() + 1;
            commentReplyMapper.updateLikeCount(replyId, newLikeCount);
            return Result.success("点赞成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> unlikeReply(Integer userId, Integer replyId) {
        try {
            CommentReply reply = commentReplyMapper.getById(replyId);
            if (reply == null) {
                return Result.error("回复不存在");
            }

            // 减少点赞数，确保不小于0
            int newLikeCount = Math.max(0, reply.getLikeCount() - 1);
            commentReplyMapper.updateLikeCount(replyId, newLikeCount);
            return Result.success("取消点赞成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
