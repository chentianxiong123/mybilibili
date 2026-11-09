package com.mybilibili.ai.service;

import java.util.Map;

public interface ContentReviewService {

    Map<String, Object> reviewContent(String content, String reason);
}