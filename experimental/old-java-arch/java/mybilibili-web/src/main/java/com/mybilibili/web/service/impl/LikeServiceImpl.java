package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.UserInteraction;
import com.mybilibili.common.enums.InteractionType;
import com.mybilibili.web.mapper.UserInteractionMapper;
import com.mybilibili.web.service.LikeService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class LikeServiceImpl implements LikeService {

    @Autowired
    private UserInteractionMapper userInteractionMapper;

    @Override
    public boolean like(Integer userId, String targetType, Integer targetId) {
        // 检查是否已经点赞
        UserInteraction existing = userInteractionMapper.findByUserAndTarget(
                userId, targetType, targetId, InteractionType.LIKE.getCode());
        if (existing != null) {
            return false; // 已经点赞过
        }

        // 添加点赞记录
        UserInteraction interaction = new UserInteraction();
        interaction.setUserId(userId);
        interaction.setTargetType(targetType);
        interaction.setTargetId(targetId);
        interaction.setInteractionType(InteractionType.LIKE.getCode());
        userInteractionMapper.insert(interaction);

        return true;
    }

    @Override
    public boolean unlike(Integer userId, String targetType, Integer targetId) {
        // 检查是否已经点赞
        UserInteraction existing = userInteractionMapper.findByUserAndTarget(
                userId, targetType, targetId, InteractionType.LIKE.getCode());
        if (existing == null) {
            return false; // 还没有点赞
        }

        // 删除点赞记录
        userInteractionMapper.delete(userId, targetType, targetId, InteractionType.LIKE.getCode());
        return true;
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
    public Map<Integer, Boolean> batchIsLiked(Integer userId, String targetType, List<Integer> targetIds) {
        Map<Integer, Boolean> result = new HashMap<>();
        if (userId == null || targetIds == null || targetIds.isEmpty()) {
            for (Integer targetId : targetIds) {
                result.put(targetId, false);
            }
            return result;
        }

        List<UserInteraction> interactions = userInteractionMapper.findByUserAndTargets(
                userId, targetType, targetIds, InteractionType.LIKE.getCode());

        // 初始化所有为false
        for (Integer targetId : targetIds) {
            result.put(targetId, false);
        }

        // 设置已点赞的为true
        for (UserInteraction interaction : interactions) {
            result.put(interaction.getTargetId(), true);
        }

        return result;
    }

    @Override
    public int getLikeCount(String targetType, Integer targetId) {
        return userInteractionMapper.countByTarget(targetType, targetId, InteractionType.LIKE.getCode());
    }

    @Override
    public Map<Integer, Integer> batchGetLikeCount(String targetType, List<Integer> targetIds) {
        Map<Integer, Integer> result = new HashMap<>();
        if (targetIds == null || targetIds.isEmpty()) {
            return result;
        }

        List<UserInteraction> counts = userInteractionMapper.countByTargets(
                targetType, targetIds, InteractionType.LIKE.getCode());

        // 初始化所有为0
        for (Integer targetId : targetIds) {
            result.put(targetId, 0);
        }

        // 设置实际的点赞数
        for (UserInteraction count : counts) {
            result.put(count.getTargetId(), count.getId() != null ? count.getId().intValue() : 0); // 使用id字段存储count值
        }

        return result;
    }
}
