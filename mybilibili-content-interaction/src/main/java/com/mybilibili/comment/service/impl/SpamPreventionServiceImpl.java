package com.mybilibili.comment.service.impl;

import com.mybilibili.comment.service.SpamPreventionService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.concurrent.TimeUnit;

@Service
public class SpamPreventionServiceImpl implements SpamPreventionService {

    private static final String RATE_PREFIX = "security:rate:";

    @Autowired
    private StringRedisTemplate redisTemplate;

    private String key(Integer userId, String action) {
        return "spam:" + action + ":" + userId;
    }

    private int getMaxCount(String action) {
        String key = RATE_PREFIX + action + ":max_count";
        String value = redisTemplate.opsForValue().get(key);
        if (value == null) {
            return "comment".equals(action) ? 10 : 20;
        }
        return Integer.parseInt(value);
    }

    private int getWindowSeconds(String action) {
        String key = RATE_PREFIX + action + ":window_seconds";
        String value = redisTemplate.opsForValue().get(key);
        if (value == null) {
            return 60;
        }
        return Integer.parseInt(value);
    }

    @Override
    public boolean isRateLimited(Integer userId, String action) {
        if (userId == null) return false;
        int maxCount = getMaxCount(action);
        String k = key(userId, action);
        String countStr = redisTemplate.opsForValue().get(k);
        if (countStr == null) return false;
        int count = Integer.parseInt(countStr);
        return count >= maxCount;
    }

    @Override
    public void recordAction(Integer userId, String action) {
        if (userId == null) return;
        String k = key(userId, action);
        Long count = redisTemplate.opsForValue().increment(k);
        if (count != null && count == 1) {
            int window = getWindowSeconds(action);
            redisTemplate.expire(k, window, TimeUnit.SECONDS);
        }
    }

    @Override
    public long getRemainingCount(Integer userId, String action) {
        if (userId == null) return 0;
        int maxCount = getMaxCount(action);
        String k = key(userId, action);
        String countStr = redisTemplate.opsForValue().get(k);
        if (countStr == null) return maxCount;
        int count = Integer.parseInt(countStr);
        return Math.max(0, maxCount - count);
    }
}