package com.mybilibili.comment.controller;

import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/admin/security-settings")
@Tag(name = "安全设置接口", description = "频率限制和缓存配置管理")
public class SecuritySettingsController {

    private static final String RATE_PREFIX = "security:rate:";
    private static final String COMMENT_SUFFIX = ":comment";
    private static final String REPLY_SUFFIX = ":reply";
    private static final String CACHE_KEY = "security:cache:refresh_interval";

    @Autowired
    private StringRedisTemplate redisTemplate;

    @GetMapping
    @Operation(summary = "获取安全设置", description = "获取频率限制和缓存刷新间隔配置")
    public Result<?> getSettings() {
        Map<String, Object> settings = new HashMap<>();

        // 评论设置
        String commentMaxCount = redisTemplate.opsForValue().get(RATE_PREFIX + "comment:max_count");
        String commentWindow = redisTemplate.opsForValue().get(RATE_PREFIX + "comment:window_seconds");
        settings.put("commentMaxCount", commentMaxCount != null ? Integer.parseInt(commentMaxCount) : 10);
        settings.put("commentWindowSeconds", commentWindow != null ? Integer.parseInt(commentWindow) : 60);

        // 回复设置
        String replyMaxCount = redisTemplate.opsForValue().get(RATE_PREFIX + "reply:max_count");
        String replyWindow = redisTemplate.opsForValue().get(RATE_PREFIX + "reply:window_seconds");
        settings.put("replyMaxCount", replyMaxCount != null ? Integer.parseInt(replyMaxCount) : 20);
        settings.put("replyWindowSeconds", replyWindow != null ? Integer.parseInt(replyWindow) : 60);

        // 缓存刷新间隔
        String refreshInterval = redisTemplate.opsForValue().get(CACHE_KEY);
        settings.put("cacheRefreshIntervalSeconds", refreshInterval != null ? Integer.parseInt(refreshInterval) : 300);

        return Result.success(settings);
    }

    @PutMapping
    @Operation(summary = "更新安全设置", description = "更新频率限制和缓存刷新间隔配置")
    public Result<?> updateSettings(@RequestBody Map<String, Object> settings) {
        // 更新评论设置
        if (settings.containsKey("commentMaxCount")) {
            redisTemplate.opsForValue().set(
                RATE_PREFIX + "comment:max_count",
                String.valueOf(settings.get("commentMaxCount"))
            );
        }
        if (settings.containsKey("commentWindowSeconds")) {
            redisTemplate.opsForValue().set(
                RATE_PREFIX + "comment:window_seconds",
                String.valueOf(settings.get("commentWindowSeconds"))
            );
        }

        // 更新回复设置
        if (settings.containsKey("replyMaxCount")) {
            redisTemplate.opsForValue().set(
                RATE_PREFIX + "reply:max_count",
                String.valueOf(settings.get("replyMaxCount"))
            );
        }
        if (settings.containsKey("replyWindowSeconds")) {
            redisTemplate.opsForValue().set(
                RATE_PREFIX + "reply:window_seconds",
                String.valueOf(settings.get("replyWindowSeconds"))
            );
        }

        // 更新缓存刷新间隔
        if (settings.containsKey("cacheRefreshIntervalSeconds")) {
            redisTemplate.opsForValue().set(
                CACHE_KEY,
                String.valueOf(settings.get("cacheRefreshIntervalSeconds"))
            );
        }

        return Result.success("设置更新成功");
    }
}