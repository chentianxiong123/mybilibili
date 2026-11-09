package com.mybilibili.message.websocket;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.common.utils.JwtUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.web.socket.*;
import org.springframework.web.socket.handler.TextWebSocketHandler;
import org.springframework.web.util.UriComponentsBuilder;

import java.util.Map;

@Component
public class NotificationWebSocketHandler extends TextWebSocketHandler {

    private static final Logger log = LoggerFactory.getLogger(NotificationWebSocketHandler.class);
    private static final ObjectMapper objectMapper = new ObjectMapper();

    private final NotificationSessionRegistry registry;

    public NotificationWebSocketHandler(NotificationSessionRegistry registry) {
        this.registry = registry;
    }

    @Override
    public void afterConnectionEstablished(WebSocketSession session) throws Exception {
        Integer userId = extractUserId(session);
        if (userId == null) {
            session.close(CloseStatus.POLICY_VIOLATION);
            return;
        }
        registry.register(userId, session);
        log.debug("Notification WS connected: userId={}", userId);
        session.sendMessage(new TextMessage(objectMapper.writeValueAsString(
                Map.of("type", "connected", "message", "通知连接已建立")
        )));
    }

    @Override
    protected void handleTextMessage(WebSocketSession session, TextMessage message) throws Exception {
        String payload = message.getPayload();
        try {
            Map<?, ?> msg = objectMapper.readValue(payload, Map.class);
            String type = (String) msg.get("type");
            if ("ping".equals(type)) {
                session.sendMessage(new TextMessage(objectMapper.writeValueAsString(Map.of("type", "pong"))));
            }
        } catch (Exception e) {
            log.debug("Invalid WS message: {}", payload);
        }
    }

    @Override
    public void afterConnectionClosed(WebSocketSession session, CloseStatus status) {
        registry.remove(session);
        log.debug("Notification WS disconnected: sessionId={}", session.getId());
    }

    @Override
    public void handleTransportError(WebSocketSession session, Throwable exception) {
        log.debug("Notification WS transport error: {}", exception.getMessage());
        registry.remove(session);
    }

    private Integer extractUserId(WebSocketSession session) {
        try {
            String query = session.getUri().getQuery();
            if (query == null) return null;
            String token = UriComponentsBuilder.newInstance()
                    .query(query).build().getQueryParams().getFirst("token");
            if (token == null || token.isBlank()) return null;
            return JwtUtils.getUserIdFromToken(token);
        } catch (Exception e) {
            log.debug("WS auth failed: {}", e.getMessage());
            return null;
        }
    }
}
