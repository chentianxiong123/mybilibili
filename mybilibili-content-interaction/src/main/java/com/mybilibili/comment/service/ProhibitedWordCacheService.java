package com.mybilibili.comment.service;

import java.util.List;

public interface ProhibitedWordCacheService {
    List<String> check(String content);
    void refreshCache();
    long getCacheSize();
}
