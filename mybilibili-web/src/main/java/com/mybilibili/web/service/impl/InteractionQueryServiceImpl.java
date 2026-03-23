package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.UserInteraction;
import com.mybilibili.common.enums.InteractionType;
import com.mybilibili.web.mapper.UserInteractionMapper;
import com.mybilibili.web.service.InteractionQueryService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class InteractionQueryServiceImpl implements InteractionQueryService {

    @Autowired
    private UserInteractionMapper userInteractionMapper;

    @Override
    public Map<String, Object> getStatus(Integer userId, String targetType, Integer targetId) {
        Map<String, Object> status = new HashMap<>();

        if (userId == null) {
            status.put("isLiked", false);
            status.put("isCollected", false);
            status.put("likeCount", getLikeCount(targetType, targetId));
            status.put("collectCount", getCollectCount(targetType, targetId));
            return status;
        }

        status.put("isLiked", isLiked(userId, targetType, targetId));
        status.put("isCollected", isCollected(userId, targetType, targetId));
        status.put("likeCount", getLikeCount(targetType, targetId));
        status.put("collectCount", getCollectCount(targetType, targetId));

        return status;
    }

    @Override
    public Map<Integer, Map<String, Object>> batchGetStatus(Integer userId, String targetType, List<Integer> targetIds) {
        Map<Integer, Map<String, Object>> result = new HashMap<>();

        if (targetIds == null || targetIds.isEmpty()) {
            return result;
        }

        // 批量查询点赞状态
        Map<Integer, Boolean> likeStatusMap = new HashMap<>();
        Map<Integer, Boolean> collectStatusMap = new HashMap<>();

        if (userId != null) {
            List<UserInteraction> likeInteractions = userInteractionMapper.findByUserAndTargets(
                    userId, targetType, targetIds, InteractionType.LIKE.getCode());
            List<UserInteraction> collectInteractions = userInteractionMapper.findByUserAndTargets(
                    userId, targetType, targetIds, InteractionType.COLLECT.getCode());

            for (Integer targetId : targetIds) {
                likeStatusMap.put(targetId, false);
                collectStatusMap.put(targetId, false);
            }

            for (UserInteraction interaction : likeInteractions) {
                likeStatusMap.put(interaction.getTargetId(), true);
            }

            for (UserInteraction interaction : collectInteractions) {
                collectStatusMap.put(interaction.getTargetId(), true);
            }
        } else {
            for (Integer targetId : targetIds) {
                likeStatusMap.put(targetId, false);
                collectStatusMap.put(targetId, false);
            }
        }

        // 批量查询点赞数
        Map<Integer, Integer> likeCountMap = new HashMap<>();
        List<UserInteraction> likeCounts = userInteractionMapper.countByTargets(
                targetType, targetIds, InteractionType.LIKE.getCode());
        for (Integer targetId : targetIds) {
            likeCountMap.put(targetId, 0);
        }
        for (UserInteraction count : likeCounts) {
            likeCountMap.put(count.getTargetId(), (int) (long) count.getId());
        }

        // 批量查询收藏数
        Map<Integer, Integer> collectCountMap = new HashMap<>();
        List<UserInteraction> collectCounts = userInteractionMapper.countByTargets(
                targetType, targetIds, InteractionType.COLLECT.getCode());
        for (Integer targetId : targetIds) {
            collectCountMap.put(targetId, 0);
        }
        for (UserInteraction count : collectCounts) {
            collectCountMap.put(count.getTargetId(), (int) (long) count.getId());
        }

        // 组装结果
        for (Integer targetId : targetIds) {
            Map<String, Object> status = new HashMap<>();
            status.put("isLiked", likeStatusMap.get(targetId));
            status.put("isCollected", collectStatusMap.get(targetId));
            status.put("likeCount", likeCountMap.get(targetId));
            status.put("collectCount", collectCountMap.get(targetId));
            result.put(targetId, status);
        }

        return result;
    }

    @Override
    public boolean isLiked(Integer userId, String targetType, Integer targetId) {
        if (userId == null) {
            return false;
        }
        UserInteraction interaction = userInteractionMapper.findByUserAndTarget(
                userId, targetType, targetId, InteractionType.LIKE.getCode());
        return interaction != null;
    }

    @Override
    public boolean isCollected(Integer userId, String targetType, Integer targetId) {
        if (userId == null) {
            return false;
        }
        UserInteraction interaction = userInteractionMapper.findByUserAndTarget(
                userId, targetType, targetId, InteractionType.COLLECT.getCode());
        return interaction != null;
    }

    @Override
    public boolean isFollowing(Integer userId, Integer targetUserId) {
        if (userId == null || targetUserId == null) {
            return false;
        }
        UserInteraction interaction = userInteractionMapper.findByUserAndTarget(
                userId, "USER", targetUserId, InteractionType.FOLLOW.getCode());
        return interaction != null;
    }

    @Override
    public int getLikeCount(String targetType, Integer targetId) {
        return userInteractionMapper.countByTarget(targetType, targetId, InteractionType.LIKE.getCode());
    }

    @Override
    public int getCollectCount(String targetType, Integer targetId) {
        return userInteractionMapper.countByTarget(targetType, targetId, InteractionType.COLLECT.getCode());
    }

    @Override
    public int getCoinCount(String targetType, Integer targetId) {
        // 投币数需要从其他表获取，暂时返回0
        return 0;
    }
}
