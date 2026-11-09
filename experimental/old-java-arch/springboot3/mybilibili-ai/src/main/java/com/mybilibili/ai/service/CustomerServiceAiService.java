package com.mybilibili.ai.service;

import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

public interface CustomerServiceAiService {
    SseEmitter chat(Long userId, String content);
    void transferToHuman(Long sessionId, Long userId, String reason);
}