package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.entity.AiChatMessage;
import com.mybilibili.ai.entity.AiSession;
import com.mybilibili.ai.mapper.AiChatMessageMapper;
import com.mybilibili.ai.mapper.AiSessionMapper;
import com.mybilibili.ai.service.CustomerSessionService;
import com.mybilibili.ai.service.SupportTicketService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.*;
import java.util.stream.Collectors;

@Service
public class CustomerSessionServiceImpl implements CustomerSessionService {

    public static final String TYPE_CUSTOMER_SERVICE = "CUSTOMER_SERVICE";
    public static final Integer STATUS_WAITING_HUMAN = 1;
    public static final Integer STATUS_PROCESSED = 2;
    public static final Integer STATUS_ACTIVE = 0;

    @Autowired
    private AiSessionMapper aiSessionMapper;

    @Autowired
    private AiChatMessageMapper aiChatMessageMapper;

    @Autowired
    private SupportTicketService supportTicketService;

    @Override
    public List<Map<String, Object>> listPendingSessions() {
        List<AiSession> sessions = aiSessionMapper.selectByTypeAndStatus(TYPE_CUSTOMER_SERVICE, STATUS_WAITING_HUMAN);

        return sessions.stream().map(session -> {
            Map<String, Object> map = new HashMap<>();
            map.put("sessionId", session.getId());
            map.put("userId", session.getUserId());
            map.put("lastMessageTime", session.getUpdatedAt());
            map.put("createdAt", session.getCreatedAt());

            // 获取该会话的最后一条消息
            List<AiChatMessage> messages = aiChatMessageMapper.selectBySessionId(session.getId());
            int messageCount = messages.size();
            map.put("messageCount", messageCount);

            String lastMessage = "";
            if (!messages.isEmpty()) {
                lastMessage = messages.get(messages.size() - 1).getContent();
                if (lastMessage.length() > 50) {
                    lastMessage = lastMessage.substring(0, 50) + "...";
                }
            }
            map.put("lastMessage", lastMessage);

            return map;
        }).collect(Collectors.toList());
    }

    @Override
    public List<Map<String, Object>> getSessionMessages(Long sessionId) {
        List<AiChatMessage> messages = aiChatMessageMapper.selectBySessionId(sessionId);

        return messages.stream().map(msg -> {
            Map<String, Object> map = new HashMap<>();
            map.put("id", msg.getId());
            map.put("sessionId", msg.getSessionId());
            map.put("role", msg.getRole());
            map.put("content", msg.getContent());
            map.put("createdAt", msg.getCreatedAt());
            return map;
        }).collect(Collectors.toList());
    }

    @Override
    @Transactional
    public void sendHumanReply(Long sessionId, Long adminId, String content) {
        AiChatMessage message = new AiChatMessage();
        message.setSessionId(sessionId);
        message.setRole("HUMAN");
        message.setContent(content);
        message.setCreatedAt(new Date());
        aiChatMessageMapper.insert(message);

        // 更新会话的更新时间
        AiSession session = aiSessionMapper.selectById(sessionId);
        if (session != null) {
            session.setUpdatedAt(new Date());
            aiSessionMapper.updateById(session);
        }
    }

    @Override
    @Transactional
    public void markProcessed(Long sessionId, Long adminId) {
        AiSession session = aiSessionMapper.selectById(sessionId);
        if (session != null) {
            session.setStatus(STATUS_PROCESSED);
            session.setUpdatedAt(new Date());
            aiSessionMapper.updateById(session);
            supportTicketService.processBySession(sessionId, adminId, "人工客服会话已处理");
        }
    }

    @Override
    public long countPending() {
        return aiSessionMapper.countByTypeAndStatus(TYPE_CUSTOMER_SERVICE, STATUS_WAITING_HUMAN);
    }
}
