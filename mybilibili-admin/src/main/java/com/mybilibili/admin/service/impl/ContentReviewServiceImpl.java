package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.mapper.ProhibitedWordMapper;
import com.mybilibili.admin.service.ContentReviewService;
import com.mybilibili.common.entity.ProhibitedWord;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;

@Service
public class ContentReviewServiceImpl implements ContentReviewService {

    @Autowired
    private ProhibitedWordMapper prohibitedWordMapper;

    // 使用CopyOnWriteArrayList实现线程安全的缓存
    private volatile List<ProhibitedWord> wordCache = new CopyOnWriteArrayList<>();

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
        List<ProhibitedWord> words = getAllEnabledWords();

        for (ProhibitedWord word : words) {
            if (word.getWord() == null || word.getWord().trim().isEmpty()) {
                continue;
            }

            String wordStr = word.getWord().trim();
            boolean isDetected = false;

            if ("EXACT".equals(word.getMatchType())) {
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

    @Override
    public List<ProhibitedWord> getAllEnabledWords() {
        // 如果缓存为空，从数据库加载
        if (wordCache.isEmpty()) {
            refreshWordCache();
        }
        return new ArrayList<>(wordCache);
    }

    @Override
    public void refreshWordCache() {
        List<ProhibitedWord> words = prohibitedWordMapper.selectAllEnabled();
        wordCache = new CopyOnWriteArrayList<>(words);
    }
}
