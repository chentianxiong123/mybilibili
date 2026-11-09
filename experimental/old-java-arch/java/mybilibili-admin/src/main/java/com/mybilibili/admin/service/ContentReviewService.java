package com.mybilibili.admin.service;

import com.mybilibili.common.entity.ProhibitedWord;

import java.util.List;

public interface ContentReviewService {
    /**
     * 检测内容中的违禁词
     * @param content 待检测内容
     * @return 检测到的违禁词列表
     */
    List<String> detectProhibitedWords(String content);
    
    /**
     * 获取所有启用的违禁词
     * @return 违禁词列表
     */
    List<ProhibitedWord> getAllEnabledWords();
    
    /**
     * 刷新违禁词缓存
     */
    void refreshWordCache();
}
