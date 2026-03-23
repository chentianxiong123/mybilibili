package com.mybilibili.web.service;

import com.mybilibili.common.entity.Conversation;
import com.mybilibili.common.vo.ConversationVO;

import java.util.List;

public interface ConversationService {

    Conversation createConversation(Integer userId, Integer targetUserId);

    ConversationVO getConversationById(Integer id);

    List<ConversationVO> getConversationsByUserId(Integer userId);

    void updateConversation(Conversation conversation);

    void updateUnreadCount(Integer id, Integer unreadCount);

    void deleteConversation(Integer id);

    Conversation getOrCreateConversation(Integer userId, Integer targetUserId);
}
