package com.mybilibili.message.websocket;

import org.springframework.stereotype.Component;
import org.springframework.web.socket.WebSocketSession;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Component
public class NotificationSessionRegistry {

    private final Map<Integer, WebSocketSession> userSessions = new ConcurrentHashMap<>();
    private final Map<String, Integer> sessionUsers = new ConcurrentHashMap<>();

    public void register(Integer userId, WebSocketSession session) {
        WebSocketSession existing = userSessions.put(userId, session);
        if (existing != null && existing.isOpen() && !existing.getId().equals(session.getId())) {
            try {
                existing.close();
            } catch (Exception ignored) {
            }
        }
        sessionUsers.put(session.getId(), userId);
    }

    public void remove(WebSocketSession session) {
        Integer userId = sessionUsers.remove(session.getId());
        if (userId != null) {
            userSessions.remove(userId, session);
        }
    }

    public WebSocketSession getSession(Integer userId) {
        return userSessions.get(userId);
    }

    public boolean isOnline(Integer userId) {
        WebSocketSession session = userSessions.get(userId);
        return session != null && session.isOpen();
    }

    public int getOnlineCount() {
        return userSessions.size();
    }
}
