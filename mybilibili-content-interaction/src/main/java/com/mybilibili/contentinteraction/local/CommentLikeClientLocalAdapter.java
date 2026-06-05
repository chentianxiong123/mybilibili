package com.mybilibili.contentinteraction.local;

import com.mybilibili.comment.feign.LikeClient;
import com.mybilibili.common.vo.Result;
import com.mybilibili.interaction.service.InteractionService;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Component
public class CommentLikeClientLocalAdapter implements LikeClient {

    private final InteractionService interactionService;

    public CommentLikeClientLocalAdapter(InteractionService interactionService) {
        this.interactionService = interactionService;
    }

    @Override
    public Result<?> like(Integer userId, String targetType, Integer targetId) {
        return interactionService.like(userId, targetType, targetId);
    }

    @Override
    public Result<?> unlike(Integer userId, String targetType, Integer targetId) {
        return interactionService.unlike(userId, targetType, targetId);
    }

    @Override
    public Result<Boolean> isLiked(Integer userId, String targetType, Integer targetId) {
        return interactionService.checkLike(userId, targetType, targetId);
    }

    @Override
    public Result<Map<Integer, Boolean>> batchIsLiked(Integer userId, String targetType, List<Integer> targetIds) {
        Map<Integer, Boolean> result = new HashMap<>();
        for (Integer targetId : targetIds) {
            Result<Boolean> checkResult = interactionService.checkLike(userId, targetType, targetId);
            result.put(targetId, checkResult.getData());
        }
        return Result.success(result);
    }

    @Override
    public Result<Integer> getLikeCount(String targetType, Integer targetId) {
        return Result.success(interactionService.getLikeCount(targetType, targetId));
    }

    @Override
    public Result<Map<Integer, Integer>> batchGetLikeCount(String targetType, List<Integer> targetIds) {
        Map<Integer, Integer> result = new HashMap<>();
        for (Integer targetId : targetIds) {
            result.put(targetId, interactionService.getLikeCount(targetType, targetId));
        }
        return Result.success(result);
    }
}
