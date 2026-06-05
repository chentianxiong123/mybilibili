package com.mybilibili.comment.service;

import com.mybilibili.common.vo.Result;

import java.util.List;
import java.util.Map;

public interface LikeInteractionPort {

    Result<?> like(Integer userId, String targetType, Integer targetId);

    Result<?> unlike(Integer userId, String targetType, Integer targetId);

    Result<Boolean> isLiked(Integer userId, String targetType, Integer targetId);

    Result<Map<Integer, Boolean>> batchIsLiked(Integer userId, String targetType, List<Integer> targetIds);

    Result<Integer> getLikeCount(String targetType, Integer targetId);

    Result<Map<Integer, Integer>> batchGetLikeCount(String targetType, List<Integer> targetIds);
}
