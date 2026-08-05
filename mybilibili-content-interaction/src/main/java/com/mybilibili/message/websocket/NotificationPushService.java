package com.mybilibili.message.websocket;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;

import java.util.LinkedHashMap;
import java.util.Map;

@Service
public class NotificationPushService {

    private static final Logger log = LoggerFactory.getLogger(NotificationPushService.class);
    private static final ObjectMapper objectMapper = new ObjectMapper()
            .registerModule(new JavaTimeModule());

    private final NotificationSessionRegistry registry;

    public NotificationPushService(NotificationSessionRegistry registry) {
        this.registry = registry;
    }

    public boolean pushToUser(Integer userId, String notificationType, String title, String content, Map<String, Object> extra) {
        WebSocketSession session = registry.getSession(userId);
        if (session == null || !session.isOpen()) {
            return false;
        }
        try {
            Map<String, Object> payload = new LinkedHashMap<>();
            payload.put("type", "notification");
            payload.put("notificationType", notificationType);
            payload.put("title", title);
            payload.put("content", content);
            payload.put("timestamp", System.currentTimeMillis());
            if (extra != null) {
                payload.putAll(extra);
            }
            session.sendMessage(new TextMessage(objectMapper.writeValueAsString(payload)));
            return true;
        } catch (Exception e) {
            log.error("Push notification failed for userId={}: {}", userId, e.getMessage());
            return false;
        }
    }

    public boolean pushToUser(Integer userId, String notificationType, String title, String content) {
        return pushToUser(userId, notificationType, title, content, null);
    }

    public boolean pushUnreadCount(Integer userId, Map<String, Integer> counts) {
        WebSocketSession session = registry.getSession(userId);
        if (session == null || !session.isOpen()) {
            return false;
        }
        try {
            Map<String, Object> payload = new LinkedHashMap<>();
            payload.put("type", "unread_count");
            payload.put("counts", counts);
            payload.put("timestamp", System.currentTimeMillis());
            session.sendMessage(new TextMessage(objectMapper.writeValueAsString(payload)));
            return true;
        } catch (Exception e) {
            log.error("Push unread count failed for userId={}: {}", userId, e.getMessage());
            return false;
        }
    }

    public boolean isUserOnline(Integer userId) {
        return registry.isOnline(userId);
    }
}
