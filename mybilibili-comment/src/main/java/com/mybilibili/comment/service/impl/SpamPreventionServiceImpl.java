package com.mybilibili.comment.service.impl;

import com.mybilibili.comment.service.SpamPreventionService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.concurrent.TimeUnit;

@Service
public class SpamPreventionServiceImpl implements SpamPreventionService {

    /** 评论：60秒内最多10条 */
    private static final int COMMENT_WINDOW_SECONDS = 60;
    private static final int COMMENT_MAX_COUNT = 10;

    /** 回复：60秒内最多20条 */
    private static final int REPLY_WINDOW_SECONDS = 60;
    private static final int REPLY_MAX_COUNT = 20;

    @Autowired
    private StringRedisTemplate redisTemplate;

    private String key(Integer userId, String action) {
        return "spam:" + action + ":" + userId;
    }

    @Override
    public boolean isRateLimited(Integer userId, String action) {
        if (userId == null) return false;
        int maxCount = "comment".equals(action) ? COMMENT_MAX_COUNT : REPLY_MAX_COUNT;
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
            int window = "comment".equals(action) ? COMMENT_WINDOW_SECONDS : REPLY_WINDOW_SECONDS;
            redisTemplate.expire(k, window, TimeUnit.SECONDS);
        }
    }

    @Override
    public long getRemainingCount(Integer userId, String action) {
        if (userId == null) return 0;
        int maxCount = "comment".equals(action) ? COMMENT_MAX_COUNT : REPLY_MAX_COUNT;
        String k = key(userId, action);
        String countStr = redisTemplate.opsForValue().get(k);
        if (countStr == null) return maxCount;
        int count = Integer.parseInt(countStr);
        return Math.max(0, maxCount - count);
    }
}