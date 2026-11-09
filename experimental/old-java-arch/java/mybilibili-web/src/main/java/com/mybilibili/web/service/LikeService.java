package com.mybilibili.web.service;

import java.util.List;
import java.util.Map;

public interface LikeService {

    boolean like(Integer userId, String targetType, Integer targetId);

    boolean unlike(Integer userId, String targetType, Integer targetId);

    boolean isLiked(Integer userId, String targetType, Integer targetId);

    Map<Integer, Boolean> batchIsLiked(Integer userId, String targetType, List<Integer> targetIds);

    int getLikeCount(String targetType, Integer targetId);

    Map<Integer, Integer> batchGetLikeCount(String targetType, List<Integer> targetIds);
}
