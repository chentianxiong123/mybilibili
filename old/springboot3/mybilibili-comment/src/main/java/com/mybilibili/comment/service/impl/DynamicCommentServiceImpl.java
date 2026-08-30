package com.mybilibili.comment.service.impl;

import com.mybilibili.comment.feign.DynamicClient;
import com.mybilibili.comment.feign.LikeClient;
import com.mybilibili.comment.feign.MessageClient;
import com.mybilibili.comment.feign.UserClient;
import com.mybilibili.comment.mapper.DynamicCommentMapper;
import com.mybilibili.comment.service.DynamicCommentService;
import com.mybilibili.common.entity.DynamicComment;
import com.mybilibili.common.vo.DynamicCommentVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;

@Slf4j
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

    @Autowired
    private MessageClient messageClient;

    private static final String TARGET_TYPE_DYNAMIC_COMMENT = "DYNAMIC_COMMENT";
    private static final int COMMENT_EXPERIENCE = 5;
    private static final int REPLY_EXPERIENCE = 2;

    private UserVO getUserById(Integer userId) {
        if (userId == null) {
            return null;
        }
        try {
            Result<UserVO> result = userClient.getUserById(userId);
            log.info("获取用户信息, userId={}, result={}", userId, result);
            if (result != null && result.getData() != null) {
                return result.getData();
            }
        } catch (Exception e) {
            log.error("获取用户信息失败, userId={}, error={}", userId, e.getMessage());
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
        vo.setReplyToUserId(comment.getReplyUserId());
        vo.setLikeCount(comment.getLikeCount() != null ? comment.getLikeCount() : 0);
        vo.setCreatedAt(comment.getCreatedAt());
        vo.setCreateTime(comment.getCreatedAt());
        vo.setStatus(comment.getStatus());

        UserVO user = getUserById(comment.getUserId());
        String nickname = user != null ? user.getNickname() : null;
        vo.setUserName(isValidNickname(nickname) ? nickname : "用户" + comment.getUserId());
        if (user != null) {
            vo.setUserAvatar(user.getAvatar());
            vo.setUserLevel(user.getLevel());
        }

        if (comment.getReplyUserId() != null) {
            UserVO replyUser = getUserById(comment.getReplyUserId());
            String replyNickname = replyUser != null ? replyUser.getNickname() : null;
            vo.setReplyToUserName(isValidNickname(replyNickname) ? replyNickname : "用户" + comment.getReplyUserId());
        }

        if (currentUserId != null) {
            try {
                Result<Boolean> likeResult = likeClient.isLiked(currentUserId, TARGET_TYPE_DYNAMIC_COMMENT, comment.getId());
                vo.setLiked(likeResult != null && Boolean.TRUE.equals(likeResult.getData()));
            } catch (Exception e) {
                vo.setLiked(false);
            }
        } else {
            vo.setLiked(false);
        }

        return vo;
    }

    private boolean isValidNickname(String nickname) {
        return nickname != null && !nickname.trim().isEmpty() && !"string".equals(nickname.trim());
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

        System.out.println("========== 准备添加动态评论经验值，userId=" + userId + ", experience=" + COMMENT_EXPERIENCE + " ==========");
        try {
            Result<?> result = userClient.addExperience(userId, COMMENT_EXPERIENCE);
            System.out.println("========== 动态评论经验值调用返回: " + result + " ==========");
            if (result != null && result.getCode() == 200) {
                org.slf4j.LoggerFactory.getLogger(getClass()).info("用户 {} 动态评论成功，经验值 +{}", userId, COMMENT_EXPERIENCE);
            } else {
                org.slf4j.LoggerFactory.getLogger(getClass()).error("用户 {} 添加动态评论经验值失败，响应: {}", userId, result);
            }
        } catch (Exception e) {
            System.out.println("========== 动态评论经验值调用异常: " + e.getMessage() + " ==========");
            org.slf4j.LoggerFactory.getLogger(getClass()).error("用户 {} 添加动态评论经验值失败，原因: {}", userId, e.getMessage());
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
    public Result<List<DynamicCommentVO>> getCommentsByDynamicId(Integer dynamicId, Integer page, Integer size, String sort, Integer currentUserId) {
        int offset = (page - 1) * size;
        List<DynamicComment> comments = dynamicCommentMapper.selectByDynamicIdWithSort(dynamicId, offset, size, sort);
        List<DynamicCommentVO> voList = new ArrayList<>();
        for (DynamicComment comment : comments) {
            DynamicCommentVO vo = buildDynamicCommentVO(comment, currentUserId);
            List<DynamicComment> replies = dynamicCommentMapper.selectRepliesByParentId(comment.getId());
            List<DynamicCommentVO> replyVOs = new ArrayList<>();
            for (DynamicComment reply : replies) {
                replyVOs.add(buildDynamicCommentVO(reply, currentUserId));
            }
            vo.setReplies(replyVOs);
            vo.setReplyCount(replyVOs.size());
            voList.add(vo);
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
        Result<?> result = likeClient.like(userId, TARGET_TYPE_DYNAMIC_COMMENT, commentId);
        if (result == null || result.getCode() != 200) {
            return Result.error(result != null ? result.getMessage() : "点赞失败");
        }
        int newLikeCount = getLikeCount(TARGET_TYPE_DYNAMIC_COMMENT, commentId);
        dynamicCommentMapper.updateLikeCountDirect(commentId, newLikeCount);
        try {
            messageClient.sendCommentLikeNotification(userId, comment.getUserId(), commentId, comment.getContent());
        } catch (Exception e) {
            org.slf4j.LoggerFactory.getLogger(getClass()).warn("发送动态评论点赞通知失败: {}", e.getMessage());
        }
        return Result.success("点赞成功", null);
    }

    @Override
    public Result<?> unlikeComment(Integer userId, Integer commentId) {
        DynamicComment comment = dynamicCommentMapper.selectById(commentId);
        if (comment == null) {
            return Result.error("评论不存在");
        }
        Result<?> result = likeClient.unlike(userId, TARGET_TYPE_DYNAMIC_COMMENT, commentId);
        if (result == null || result.getCode() != 200) {
            return Result.error(result != null ? result.getMessage() : "取消点赞失败");
        }
        int newLikeCount = getLikeCount(TARGET_TYPE_DYNAMIC_COMMENT, commentId);
        dynamicCommentMapper.updateLikeCountDirect(commentId, newLikeCount);
        return Result.success("取消点赞成功", null);
    }

    private int getLikeCount(String targetType, Integer targetId) {
        try {
            Result<Integer> result = likeClient.getLikeCount(targetType, targetId);
            if (result != null && result.getData() != null) {
                return result.getData();
            }
        } catch (Exception e) {
            org.slf4j.LoggerFactory.getLogger(getClass()).error("获取点赞数失败: {}", e.getMessage());
        }
        throw new RuntimeException("获取点赞数失败");
    }
}