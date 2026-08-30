package com.mybilibili.web.service;

import java.util.List;

public interface ContentReviewService {
    /**
     * 检测内容中的违禁词
     * @param content 待检测内容
     * @return 检测到的违禁词列表
     */
    List<String> detectProhibitedWords(String content);
}
