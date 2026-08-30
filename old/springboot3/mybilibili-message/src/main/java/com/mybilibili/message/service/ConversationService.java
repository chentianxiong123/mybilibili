package com.mybilibili.message.service;

import com.mybilibili.common.entity.Conversation;
import com.mybilibili.common.vo.ConversationVO;

import java.util.List;

public interface ConversationService {

    Conversation createConversation(Integer userId, Integer targetUserId);

    ConversationVO getConversationById(Long id);

    List<ConversationVO> getConversationsByUserId(Integer userId);

    void updateConversation(Conversation conversation);

    void updateUnreadCount(Long id, Integer unreadCount);

    void deleteConversation(Long id);

    Conversation getOrCreateConversation(Integer userId, Integer targetUserId);
}