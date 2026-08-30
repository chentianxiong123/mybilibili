package com.mybilibili.ai.service;

import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

public interface AdminAiService {
    SseEmitter sendMessage(Integer adminId, String content);
}