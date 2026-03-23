package com.mybilibili.web.service.impl;

import com.mybilibili.web.mapper.ProhibitedWordMapper;
import com.mybilibili.web.service.ContentReviewService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class ContentReviewServiceImpl implements ContentReviewService {

    @Autowired
    private ProhibitedWordMapper prohibitedWordMapper;

    // 缓存违禁词，key为词，value为匹配类型
    private volatile Map<String, String> wordCache = new ConcurrentHashMap<>();

    @PostConstruct
    public void init() {
        refreshWordCache();
    }

    @Override
    public List<String> detectProhibitedWords(String content) {
        List<String> detectedWords = new ArrayList<>();
        if (content == null || content.trim().isEmpty()) {
            return detectedWords;
        }

        String trimmedContent = content.trim();
        Map<String, String> words = new ConcurrentHashMap<>(wordCache);

        for (Map.Entry<String, String> entry : words.entrySet()) {
            String wordStr = entry.getKey();
            String matchType = entry.getValue();

            if (wordStr == null || wordStr.isEmpty()) {
                continue;
            }

            boolean isDetected = false;

            if ("EXACT".equals(matchType)) {
                // 精确匹配
                isDetected = trimmedContent.equals(wordStr);
            } else {
                // 包含匹配（默认）
                isDetected = trimmedContent.contains(wordStr);
            }

            if (isDetected && !detectedWords.contains(wordStr)) {
                detectedWords.add(wordStr);
            }
        }

        return detectedWords;
    }

    public void refreshWordCache() {
        List<Map<String, Object>> words = prohibitedWordMapper.selectAllEnabled();
        Map<String, String> newCache = new ConcurrentHashMap<>();
        for (Map<String, Object> word : words) {
            String wordStr = (String) word.get("word");
            String matchType = (String) word.get("match_type");
            if (wordStr != null && !wordStr.isEmpty()) {
                newCache.put(wordStr, matchType != null ? matchType : "CONTAINS");
            }
        }
        wordCache = newCache;
    }
}
