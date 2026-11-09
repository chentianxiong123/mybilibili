package com.mybilibili.danmaku.websocket;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;

import java.util.Map;
import java.util.Set;

@Service
public class DanmakuBroadcastService {

    private static final Logger log = LoggerFactory.getLogger(DanmakuBroadcastService.class);
    private static final ObjectMapper objectMapper = new ObjectMapper();

    private final DanmakuSessionRegistry registry;

    public DanmakuBroadcastService(DanmakuSessionRegistry registry) {
        this.registry = registry;
    }

    public void broadcast(Integer videoId, Map<String, Object> payload) {
        Set<WebSocketSession> sessions = registry.getRoomSessions(videoId);
        if (sessions.isEmpty()) return;

        try {
            TextMessage message = new TextMessage(objectMapper.writeValueAsString(payload));
            for (WebSocketSession session : sessions) {
                if (session.isOpen()) {
                    try {
                        session.sendMessage(message);
                    } catch (Exception e) {
                        log.debug("Broadcast to session {} failed: {}", session.getId(), e.getMessage());
                    }
                }
            }
        } catch (Exception e) {
            log.error("Broadcast serialization failed: {}", e.getMessage());
        }
    }
}
