package com.mybilibili.web.service;

import java.util.List;
import java.util.Map;

public interface InteractionQueryService {

    Map<String, Object> getStatus(Integer userId, String targetType, Integer targetId);

    Map<Integer, Map<String, Object>> batchGetStatus(Integer userId, String targetType, List<Integer> targetIds);

    boolean isLiked(Integer userId, String targetType, Integer targetId);

    boolean isCollected(Integer userId, String targetType, Integer targetId);

    boolean isFollowing(Integer userId, Integer targetUserId);

    int getLikeCount(String targetType, Integer targetId);

    int getCollectCount(String targetType, Integer targetId);

    int getCoinCount(String targetType, Integer targetId);
}
