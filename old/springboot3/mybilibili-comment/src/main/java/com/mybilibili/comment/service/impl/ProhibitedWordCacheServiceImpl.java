package com.mybilibili.comment.service.impl;

import com.mybilibili.comment.mapper.ProhibitedWordMapper;
import com.mybilibili.comment.service.ProhibitedWordCacheService;
import com.mybilibili.common.entity.ProhibitedWord;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import jakarta.annotation.PostConstruct;
import java.util.ArrayList;
import java.util.List;
import java.util.Set;

@Service
public class ProhibitedWordCacheServiceImpl implements ProhibitedWordCacheService {

    private static final String REDIS_SET_KEY = "prohibited_words:words";
    private static final String CACHE_REFRESH_KEY = "security:cache:refresh_interval";
    private static final int DEFAULT_REFRESH_INTERVAL = 300000; // 5分钟

    @Autowired
    private ProhibitedWordMapper prohibitedWordMapper;

    @Autowired
    private StringRedisTemplate redisTemplate;

    private volatile long refreshIntervalMs = DEFAULT_REFRESH_INTERVAL;

    @PostConstruct
    public void init() {
        refreshCache();
        loadRefreshInterval();
    }

    @Scheduled(fixedRateString = "#{@prohibitedWordCacheServiceImpl.refreshIntervalMs}")
    public void scheduledRefresh() {
        refreshCache();
    }

    private void loadRefreshInterval() {
        String value = redisTemplate.opsForValue().get(CACHE_REFRESH_KEY);
        if (value != null) {
            try {
                int seconds = Integer.parseInt(value);
                refreshIntervalMs = seconds * 1000L;
            } catch (NumberFormatException ignored) {}
        }
    }

    @Override
    public List<String> check(String content) {
        if (content == null || content.isEmpty()) {
            return new ArrayList<>();
        }
        List<String> found = new ArrayList<>();
        Set<String> cachedWords = redisTemplate.opsForSet().members(REDIS_SET_KEY);
        if (cachedWords != null) {
            for (String word : cachedWords) {
                if (content.contains(word)) {
                    found.add(word);
                }
            }
        }
        return found;
    }

    @Override
    public void refreshCache() {
        loadRefreshInterval();
        List<ProhibitedWord> words = prohibitedWordMapper.selectAllEnabled();
        if (words == null || words.isEmpty()) {
            redisTemplate.delete(REDIS_SET_KEY);
            return;
        }
        redisTemplate.delete(REDIS_SET_KEY);
        for (ProhibitedWord pw : words) {
            if (pw.getWord() != null && !pw.getWord().isEmpty()) {
                redisTemplate.opsForSet().add(REDIS_SET_KEY, pw.getWord());
            }
        }
    }

    @Override
    public long getCacheSize() {
        Long size = redisTemplate.opsForSet().size(REDIS_SET_KEY);
        return size != null ? size : 0;
    }
}