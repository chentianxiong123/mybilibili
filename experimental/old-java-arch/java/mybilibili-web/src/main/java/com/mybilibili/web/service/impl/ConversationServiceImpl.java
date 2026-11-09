package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.Conversation;
import com.mybilibili.common.vo.ConversationVO;
import com.mybilibili.web.mapper.ConversationMapper;
import com.mybilibili.web.service.ConversationService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class ConversationServiceImpl implements ConversationService {

    @Autowired
    private ConversationMapper conversationMapper;

    @Override
    public Conversation createConversation(Integer userId, Integer targetUserId) {
        Conversation conversation = new Conversation();
        conversation.setUserId(userId);
        conversation.setTargetUserId(targetUserId);
        conversation.setUnreadCount(0);
        conversationMapper.insert(conversation);
        return conversation;
    }

    @Override
    public ConversationVO getConversationById(Integer id) {
        Conversation conversation = conversationMapper.selectById(id);
        if (conversation == null) {
            return null;
        }
        List<ConversationVO> list = conversationMapper.selectByUserId(conversation.getUserId());
        return list.stream()
                .filter(vo -> vo.getId().equals(id))
                .findFirst()
                .orElse(null);
    }

    @Override
    public List<ConversationVO> getConversationsByUserId(Integer userId) {
        return conversationMapper.selectByUserId(userId);
    }

    @Override
    public void updateConversation(Conversation conversation) {
        conversationMapper.update(conversation);
    }

    @Override
    public void updateUnreadCount(Integer id, Integer unreadCount) {
        conversationMapper.updateUnreadCount(id, unreadCount);
    }

    @Override
    public void deleteConversation(Integer id) {
        conversationMapper.deleteById(id);
    }

    @Override
    public Conversation getOrCreateConversation(Integer userId, Integer targetUserId) {
        Conversation conversation = conversationMapper.selectByUserAndTarget(userId, targetUserId);
        if (conversation == null) {
            conversation = createConversation(userId, targetUserId);
        }
        return conversation;
    }
}
