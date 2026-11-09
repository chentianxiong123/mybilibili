package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.Video;
import com.mybilibili.web.mapper.VideoMapper;
import com.mybilibili.web.service.LikeHistoryService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.ZSetOperations;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Set;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

@Slf4j
@Service
public class LikeHistoryServiceImpl implements LikeHistoryService {

    @Autowired
    private StringRedisTemplate redisTemplate;

    @Autowired
    private VideoMapper videoMapper;

    private static final String KEY_USER_LIKES = "likes:user:%d:%s";
    private static final String KEY_TARGET_LIKES = "likes:target:%s:%d";
    private static final String KEY_VIDEO_MANUSCRIPT = "video:manuscript:%d";

    private static final long EXPIRE_DAYS = 365;

    @Override
    public boolean like(Integer userId, String targetType, Integer targetId) {
        if (userId == null || targetType == null || targetId == null) {
            return false;
        }

        String userLikesKey = String.format(KEY_USER_LIKES, userId, targetType);
        String targetLikesKey = String.format(KEY_TARGET_LIKES, targetType, targetId);
        long currentTimeMillis = System.currentTimeMillis();

        ZSetOperations<String, String> zSetOps = redisTemplate.opsForZSet();

        Double existingScore = zSetOps.score(userLikesKey, String.valueOf(targetId));
        if (existingScore != null) {
            log.debug("用户 {} 已经点赞过 {}/{}", userId, targetType, targetId);
            return false;
        }

        zSetOps.add(userLikesKey, String.valueOf(targetId), currentTimeMillis);
        redisTemplate.expire(userLikesKey, EXPIRE_DAYS, TimeUnit.DAYS);

        zSetOps.add(targetLikesKey, String.valueOf(userId), currentTimeMillis);
        redisTemplate.expire(targetLikesKey, EXPIRE_DAYS, TimeUnit.DAYS);

        log.info("用户 {} 点赞 {}/{}", userId, targetType, targetId);
        return true;
    }

    @Override
    public boolean unlike(Integer userId, String targetType, Integer targetId) {
        if (userId == null || targetType == null || targetId == null) {
            return false;
        }

        String userLikesKey = String.format(KEY_USER_LIKES, userId, targetType);
        String targetLikesKey = String.format(KEY_TARGET_LIKES, targetType, targetId);

        ZSetOperations<String, String> zSetOps = redisTemplate.opsForZSet();

        Long removed = zSetOps.remove(userLikesKey, String.valueOf(targetId));
        zSetOps.remove(targetLikesKey, String.valueOf(userId));

        log.info("用户 {} 取消点赞 {}/{}", userId, targetType, targetId);
        return removed != null && removed > 0;
    }

    @Override
    public boolean isLiked(Integer userId, String targetType, Integer targetId) {
        if (userId == null || targetType == null || targetId == null) {
            return false;
        }

        String userLikesKey = String.format(KEY_USER_LIKES, userId, targetType);
        ZSetOperations<String, String> zSetOps = redisTemplate.opsForZSet();

        Double score = zSetOps.score(userLikesKey, String.valueOf(targetId));
        return score != null;
    }

    @Override
    public List<Integer> getLikedTargetIds(Integer userId, String targetType, List<Integer> targetIds) {
        if (userId == null || targetIds == null || targetIds.isEmpty()) {
            return new ArrayList<>();
        }

        String userLikesKey = String.format(KEY_USER_LIKES, userId, targetType);
        ZSetOperations<String, String> zSetOps = redisTemplate.opsForZSet();

        Set<String> likedIds = zSetOps.range(userLikesKey, 0, -1);
        if (likedIds == null || likedIds.isEmpty()) {
            return new ArrayList<>();
        }

        Set<String> likedSet = likedIds.stream().collect(Collectors.toSet());
        return targetIds.stream()
                .filter(id -> likedSet.contains(String.valueOf(id)))
                .collect(Collectors.toList());
    }

    @Override
    public List<Integer> getUserLikeHistory(Integer userId, int days, int limit) {
        if (userId == null) {
            return new ArrayList<>();
        }

        String userLikesKey = String.format(KEY_USER_LIKES, userId, "VIDEO");
        ZSetOperations<String, String> zSetOps = redisTemplate.opsForZSet();

        long minScore = 0;
        if (days > 0) {
            minScore = System.currentTimeMillis() - (long) days * 24 * 60 * 60 * 1000;
        }

        Set<String> videoIds;
        if (days > 0) {
            videoIds = zSetOps.rangeByScore(userLikesKey, minScore, Double.MAX_VALUE);
        } else {
            videoIds = zSetOps.reverseRange(userLikesKey, 0, limit - 1);
        }

        if (videoIds == null || videoIds.isEmpty()) {
            return new ArrayList<>();
        }

        List<String> result = new ArrayList<>(videoIds);
        if (limit > 0 && result.size() > limit) {
            result = result.subList(0, limit);
        }

        return result.stream().map(Integer::parseInt).collect(Collectors.toList());
    }

    @Override
    public List<Integer> getUserLikedManuscriptIds(Integer userId, int days, int limit) {
        List<Integer> videoIds = getUserLikeHistory(userId, days, limit);
        if (videoIds.isEmpty()) {
            return new ArrayList<>();
        }

        List<Integer> manuscriptIds = new ArrayList<>();
        for (Integer videoId : videoIds) {
            Integer manuscriptId = getManuscriptIdByVideoId(videoId);
            if (manuscriptId != null && !manuscriptIds.contains(manuscriptId)) {
                manuscriptIds.add(manuscriptId);
            }
        }
        return manuscriptIds;
    }

    @Override
    public long getLikeCount(String targetType, Integer targetId) {
        if (targetType == null || targetId == null) {
            return 0;
        }

        String targetLikesKey = String.format(KEY_TARGET_LIKES, targetType, targetId);
        Long count = redisTemplate.opsForZSet().zCard(targetLikesKey);
        return count != null ? count : 0;
    }

    @Override
    public List<Object> getLikeCounts(String targetType, List<Integer> targetIds) {
        if (targetIds == null || targetIds.isEmpty()) {
            return new ArrayList<>();
        }

        List<Object> counts = new ArrayList<>();
        for (Integer targetId : targetIds) {
            counts.add(getLikeCount(targetType, targetId));
        }
        return counts;
    }

    private Integer getManuscriptIdByVideoId(Integer videoId) {
        try {
            String cacheKey = String.format(KEY_VIDEO_MANUSCRIPT, videoId);
            String cached = (String) redisTemplate.opsForValue().get(cacheKey);
            if (cached != null) {
                return Integer.parseInt(cached);
            }

            Video video = videoMapper.selectById(videoId);
            if (video != null && video.getManuscriptId() != null) {
                redisTemplate.opsForValue().set(cacheKey, String.valueOf(video.getManuscriptId()), 30, TimeUnit.DAYS);
                return video.getManuscriptId();
            }
        } catch (Exception e) {
            log.warn("获取视频对应稿件ID失败，videoId: {}", videoId, e);
        }
        return null;
    }
}
