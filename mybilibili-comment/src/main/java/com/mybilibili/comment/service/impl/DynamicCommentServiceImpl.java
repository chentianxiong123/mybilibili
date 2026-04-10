package com.mybilibili.comment.service.impl;

import com.mybilibili.comment.feign.DynamicClient;
import com.mybilibili.comment.feign.LikeClient;
import com.mybilibili.comment.feign.UserClient;
import com.mybilibili.comment.mapper.DynamicCommentMapper;
import com.mybilibili.comment.service.DynamicCommentService;
import com.mybilibili.common.entity.DynamicComment;
import com.mybilibili.common.vo.DynamicCommentVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;

@Service
public class DynamicCommentServiceImpl implements DynamicCommentService {

    @Autowired
    private DynamicCommentMapper dynamicCommentMapper;

    @Autowired
    private LikeClient likeClient;

    @Autowired
    private UserClient userClient;

    @Autowired
    private DynamicClient dynamicClient;

    private static final String TARGET_TYPE_DYNAMIC_COMMENT = "DYNAMIC_COMMENT";
    private static final int COMMENT_EXPERIENCE = 5;
    private static final int REPLY_EXPERIENCE = 2;

    private UserVO getUserById(Integer userId) {
        if (userId == null) {
            return null;
        }
        try {
            Result<UserVO> result = userClient.getUserById(userId);
            if (result != null && result.getData() != null) {
                return result.getData();
            }
        } catch (Exception e) {
        }
        return null;
    }

    private DynamicCommentVO buildDynamicCommentVO(DynamicComment comment, Integer currentUserId) {
        DynamicCommentVO vo = new DynamicCommentVO();
        vo.setId(comment.getId());
        vo.setDynamicId(comment.getDynamicId());
        vo.setUserId(comment.getUserId());
        vo.setContent(comment.getContent());
        vo.setParentId(comment.getParentId());
        vo.setReplyUserId(comment.getReplyUserId());
        vo.setLikeCount(comment.getLikeCount());
        vo.setCreatedAt(comment.getCreatedAt());
        vo.setStatus(comment.getStatus());

        UserVO user = getUserById(comment.getUserId());
        if (user != null) {
            vo.setUserName(user.getNickname());
            vo.setUserAvatar(user.getAvatar());
        }

        return vo;
    }

    @Override
    public Result<DynamicCommentVO> addComment(Integer userId, Integer dynamicId, String content, Integer parentId, Integer replyUserId) {
        if (content == null || content.trim().isEmpty()) {
            return Result.error("评论内容不能为空");
        }

        DynamicComment comment = new DynamicComment();
        comment.setDynamicId(dynamicId);
        comment.setUserId(userId);
        comment.setContent(content.trim());
        comment.setParentId(parentId);
        comment.setReplyUserId(replyUserId);
        comment.setLikeCount(0);
        comment.setStatus(0);  // 0-正常 1-已删除

        dynamicCommentMapper.insert(comment);

        try {
            userClient.addExperience(userId, COMMENT_EXPERIENCE);
        } catch (Exception e) {
            // 忽略经验值添加失败
        }

        if (parentId == null) {
            try {
                dynamicClient.incrementCommentCount(dynamicId, 1);
            } catch (Exception e) {
            }
        }

        return Result.success("评论成功", buildDynamicCommentVO(comment, userId));
    }

    @Override
    public Result<?> deleteComment(Integer userId, Integer commentId) {
        DynamicComment comment = dynamicCommentMapper.selectById(commentId);
        if (comment == null) {
            return Result.error("评论不存在");
        }
        if (!comment.getUserId().equals(userId)) {
            return Result.error("没有权限删除此评论");
        }

        int delta = (comment.getParentId() == null) ? -1 : 0;
        dynamicCommentMapper.deleteById(commentId);

        if (delta != 0) {
            try {
                dynamicClient.incrementCommentCount(comment.getDynamicId(), delta);
            } catch (Exception e) {
            }
        }

        return Result.success("删除成功", null);
    }

    @Override
    public Result<List<DynamicCommentVO>> getCommentsByDynamicId(Integer dynamicId, Integer page, Integer size, Integer currentUserId) {
        int offset = (page - 1) * size;
        List<DynamicComment> comments = dynamicCommentMapper.selectByDynamicId(dynamicId, offset, size);
        List<DynamicCommentVO> voList = new ArrayList<>();
        for (DynamicComment comment : comments) {
            voList.add(buildDynamicCommentVO(comment, currentUserId));
        }
        return Result.success("获取成功", voList);
    }

    @Override
    public Result<List<DynamicCommentVO>> getRepliesByCommentId(Integer commentId, Integer currentUserId) {
        List<DynamicComment> replies = dynamicCommentMapper.selectRepliesByParentId(commentId);
        List<DynamicCommentVO> voList = new ArrayList<>();
        for (DynamicComment reply : replies) {
            voList.add(buildDynamicCommentVO(reply, currentUserId));
        }
        return Result.success("获取成功", voList);
    }

    @Override
    public Result<?> likeComment(Integer userId, Integer commentId) {
        DynamicComment comment = dynamicCommentMapper.selectById(commentId);
        if (comment == null) {
            return Result.error("评论不存在");
        }
        likeClient.like(userId, TARGET_TYPE_DYNAMIC_COMMENT, commentId);
        return Result.success("点赞成功", null);
    }

    @Override
    public Result<?> unlikeComment(Integer userId, Integer commentId) {
        DynamicComment comment = dynamicCommentMapper.selectById(commentId);
        if (comment == null) {
            return Result.error("评论不存在");
        }
        likeClient.unlike(userId, TARGET_TYPE_DYNAMIC_COMMENT, commentId);
        return Result.success("取消点赞成功", null);
    }
}