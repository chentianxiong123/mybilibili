package com.mybilibili.danmaku.websocket;

import org.springframework.stereotype.Component;
import org.springframework.web.socket.WebSocketSession;

import java.util.Map;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;

@Component
public class DanmakuSessionRegistry {

    private final Map<Integer, Set<WebSocketSession>> roomSessions = new ConcurrentHashMap<>();
    private final Map<String, Integer> sessionRooms = new ConcurrentHashMap<>();

    public void join(Integer videoId, WebSocketSession session) {
        leave(session);
        roomSessions.computeIfAbsent(videoId, k -> ConcurrentHashMap.newKeySet()).add(session);
        sessionRooms.put(session.getId(), videoId);
    }

    public void leave(WebSocketSession session) {
        Integer videoId = sessionRooms.remove(session.getId());
        if (videoId != null) {
            Set<WebSocketSession> sessions = roomSessions.get(videoId);
            if (sessions != null) {
                sessions.remove(session);
                if (sessions.isEmpty()) {
                    roomSessions.remove(videoId);
                }
            }
        }
    }

    public Set<WebSocketSession> getRoomSessions(Integer videoId) {
        return roomSessions.getOrDefault(videoId, Set.of());
    }

    public Integer getVideoId(WebSocketSession session) {
        return sessionRooms.get(session.getId());
    }
}
