package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.DynamicComment;
import com.mybilibili.common.entity.User;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.web.mapper.DynamicCommentMapper;
import com.mybilibili.web.mapper.UserDynamicMapper;
import com.mybilibili.web.mapper.UserMapper;
import com.mybilibili.web.service.DynamicCommentService;
import com.mybilibili.web.service.LikeService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class DynamicCommentServiceImpl implements DynamicCommentService {

    @Autowired
    private DynamicCommentMapper dynamicCommentMapper;

    @Autowired
    private UserDynamicMapper userDynamicMapper;

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private LikeService likeService;

    private static final String TARGET_TYPE_COMMENT = "COMMENT";

    @Override
    @Transactional
    public Result<?> addComment(Integer userId, Integer dynamicId, String content, Integer parentId, Integer replyUserId) {
        try {
            // 检查动态是否存在
            if (userDynamicMapper.getById(dynamicId) == null) {
                return Result.error("动态不存在");
            }

            // 检查父评论是否存在（如果是回复）
            if (parentId != null) {
                DynamicComment parentComment = dynamicCommentMapper.getById(parentId);
                if (parentComment == null) {
                    return Result.error("父评论不存在");
                }
            }

            // 创建评论
            DynamicComment comment = new DynamicComment();
            comment.setDynamicId(dynamicId);
            comment.setUserId(userId);
            comment.setContent(content);
            comment.setParentId(parentId);
            comment.setReplyUserId(replyUserId);
            comment.setLikeCount(0);
            comment.setStatus(0);

            dynamicCommentMapper.insert(comment);

            // 更新动态评论数
            int commentCount = dynamicCommentMapper.countByDynamicId(dynamicId);
            userDynamicMapper.updateCommentCount(dynamicId, commentCount);

            return Result.success("评论成功", comment);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    @Transactional
    public Result<?> deleteComment(Integer userId, Integer commentId) {
        try {
            DynamicComment comment = dynamicCommentMapper.getById(commentId);
            if (comment == null) {
                return Result.error("评论不存在");
            }

            // 检查是否是评论作者
            if (!comment.getUserId().equals(userId)) {
                return Result.error("无权删除他人评论");
            }

            // 软删除评论
            dynamicCommentMapper.updateStatus(commentId, 1);

            // 更新动态评论数
            int commentCount = dynamicCommentMapper.countByDynamicId(comment.getDynamicId());
            userDynamicMapper.updateCommentCount(comment.getDynamicId(), commentCount);

            return Result.success("删除成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<List<DynamicComment>> getCommentsByDynamicId(Integer dynamicId, Integer page, Integer size, Integer currentUserId) {
        try {
            int offset = (page - 1) * size;
            List<DynamicComment> comments = dynamicCommentMapper.getByDynamicId(dynamicId, offset, size);

            // 批量查询点赞状态和点赞数
            List<Integer> commentIds = new ArrayList<>();
            for (DynamicComment comment : comments) {
                commentIds.add(comment.getId());
            }

            Map<Integer, Boolean> likeStatusMap = new HashMap<>();
            Map<Integer, Integer> likeCountMap = new HashMap<>();

            if (currentUserId != null && !commentIds.isEmpty()) {
                likeStatusMap = likeService.batchIsLiked(currentUserId, TARGET_TYPE_COMMENT, commentIds);
            } else {
                for (Integer id : commentIds) {
                    likeStatusMap.put(id, false);
                }
            }

            if (!commentIds.isEmpty()) {
                likeCountMap = likeService.batchGetLikeCount(TARGET_TYPE_COMMENT, commentIds);
            }

            // 设置点赞状态和点赞数
            for (DynamicComment comment : comments) {
                comment.setLikeCount(likeCountMap.getOrDefault(comment.getId(), 0));
            }

            return Result.success("获取成功", comments);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<List<DynamicComment>> getRepliesByCommentId(Integer commentId, Integer currentUserId) {
        try {
            List<DynamicComment> replies = dynamicCommentMapper.getRepliesByParentId(commentId);

            // 批量查询点赞状态和点赞数
            List<Integer> replyIds = new ArrayList<>();
            for (DynamicComment reply : replies) {
                replyIds.add(reply.getId());
            }

            Map<Integer, Boolean> likeStatusMap = new HashMap<>();
            Map<Integer, Integer> likeCountMap = new HashMap<>();

            if (currentUserId != null && !replyIds.isEmpty()) {
                likeStatusMap = likeService.batchIsLiked(currentUserId, TARGET_TYPE_COMMENT, replyIds);
            } else {
                for (Integer id : replyIds) {
                    likeStatusMap.put(id, false);
                }
            }

            if (!replyIds.isEmpty()) {
                likeCountMap = likeService.batchGetLikeCount(TARGET_TYPE_COMMENT, replyIds);
            }

            // 设置点赞状态和点赞数
            for (DynamicComment reply : replies) {
                reply.setLikeCount(likeCountMap.getOrDefault(reply.getId(), 0));
            }

            return Result.success("获取成功", replies);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    @Transactional
    public Result<?> likeComment(Integer userId, Integer commentId) {
        try {
            DynamicComment comment = dynamicCommentMapper.getById(commentId);
            if (comment == null) {
                return Result.error("评论不存在");
            }

            // 使用复用的 LikeService
            boolean result = likeService.like(userId, TARGET_TYPE_COMMENT, commentId);
            if (!result) {
                return Result.error("已经点赞过了");
            }

            // 更新评论点赞数
            int newLikeCount = likeService.getLikeCount(TARGET_TYPE_COMMENT, commentId);
            dynamicCommentMapper.updateLikeCount(commentId, newLikeCount);

            // 返回最新状态
            Map<String, Object> data = new HashMap<>();
            data.put("likeCount", newLikeCount);
            data.put("isLiked", true);
            return Result.success("点赞成功", data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    @Transactional
    public Result<?> unlikeComment(Integer userId, Integer commentId) {
        try {
            DynamicComment comment = dynamicCommentMapper.getById(commentId);
            if (comment == null) {
                return Result.error("评论不存在");
            }

            // 使用复用的 LikeService
            boolean result = likeService.unlike(userId, TARGET_TYPE_COMMENT, commentId);
            if (!result) {
                return Result.error("尚未点赞");
            }

            // 更新评论点赞数
            int newLikeCount = likeService.getLikeCount(TARGET_TYPE_COMMENT, commentId);
            dynamicCommentMapper.updateLikeCount(commentId, newLikeCount);

            // 返回最新状态
            Map<String, Object> data = new HashMap<>();
            data.put("likeCount", newLikeCount);
            data.put("isLiked", false);
            return Result.success("取消点赞成功", data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public int getCommentLikeCount(Integer commentId) {
        return likeService.getLikeCount(TARGET_TYPE_COMMENT, commentId);
    }

    @Override
    public boolean isCommentLiked(Integer userId, Integer commentId) {
        if (userId == null) {
            return false;
        }
        return likeService.isLiked(userId, TARGET_TYPE_COMMENT, commentId);
    }
}
