package com.mybilibili.ai.service;

import com.mybilibili.ai.entity.AiChatMessage;
import com.mybilibili.ai.entity.AiConversation;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.util.List;

public interface AiChatService {

    SseEmitter sendMessage(Integer userId, Long conversationId, String content);

    AiConversation createConversation(Integer userId);

    List<AiConversation> getConversations(Integer userId);

    List<AiChatMessage> getMessages(Long conversationId, Integer userId);

    void deleteConversation(Long conversationId, Integer userId);
}