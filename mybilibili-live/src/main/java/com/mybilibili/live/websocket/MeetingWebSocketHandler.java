package com.mybilibili.live.websocket;

import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.web.socket.CloseStatus;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;
import org.springframework.web.socket.handler.TextWebSocketHandler;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Slf4j
@Component
public class MeetingWebSocketHandler extends TextWebSocketHandler {

    @Autowired
    private ObjectMapper objectMapper;

    private final Map<String, Map<String, WebSocketSession>> roomSessions = new ConcurrentHashMap<>();
    private final Map<String, String> sessionRooms = new ConcurrentHashMap<>();
    private final Map<String, Long> sessionUsers = new ConcurrentHashMap<>();

    public static class SignalingMessage {
        public String type;
        public String roomCode;
        public Long userId;
        public String userName;
        public Object data;
    }

    @Override
    public void afterConnectionEstablished(WebSocketSession session) throws Exception {
        log.info("[Meeting WS] 连接建立: {}", session.getId());
    }

    @Override
    protected void handleTextMessage(WebSocketSession session, TextMessage message) throws Exception {
        try {
            SignalingMessage msg = objectMapper.readValue(message.getPayload(), SignalingMessage.class);
            log.info("[Meeting WS] type={}, roomCode={}, userId={}", msg.type, msg.roomCode, msg.userId);

            switch (msg.type) {
                case "join":
                    handleJoin(session, msg);
                    break;
                case "leave":
                    handleLeave(session, msg);
                    break;
                case "offer":
                case "answer":
                case "ice-candidate":
                    handleSignaling(session, msg);
                    break;
                case "toggle-audio":
                case "toggle-video":
                case "toggle-screen":
                    handleMediaToggle(session, msg);
                    break;
            }
        } catch (Exception e) {
            log.error("[Meeting WS] 处理消息失败: {}", e.getMessage());
        }
    }

    private void handleJoin(WebSocketSession session, SignalingMessage msg) throws IOException {
        String roomCode = msg.roomCode;
        roomSessions.computeIfAbsent(roomCode, k -> new ConcurrentHashMap<>());
        Map<String, WebSocketSession> sessions = roomSessions.get(roomCode);
        sessions.put(session.getId(), session);
        sessionRooms.put(session.getId(), roomCode);
        sessionUsers.put(session.getId(), msg.userId);

        for (WebSocketSession s : sessions.values()) {
            if (!s.getId().equals(session.getId())) {
                sendMessage(s, "user-joined", msg.userId, msg.userName, null);
            }
        }

        List<String> userIds = sessions.values().stream()
                .filter(s -> !s.getId().equals(session.getId()))
                .map(s -> s.getId())
                .collect(Collectors.toList());
        sendMessage(session, "room-users", null, null, userIds);
    }

    private void handleLeave(WebSocketSession session, SignalingMessage msg) throws IOException {
        removeSession(session);
        String roomCode = msg.roomCode;
        Map<String, WebSocketSession> sessions = roomSessions.get(roomCode);
        if (sessions != null) {
            for (WebSocketSession s : sessions.values()) {
                sendMessage(s, "user-left", msg.userId, msg.userName, null);
            }
        }
    }

    private void handleSignaling(WebSocketSession session, SignalingMessage msg) throws IOException {
        String roomCode = sessionRooms.get(session.getId());
        if (roomCode == null) return;
        Map<String, WebSocketSession> sessions = roomSessions.get(roomCode);
        if (sessions == null) return;

        for (WebSocketSession s : sessions.values()) {
            if (!s.getId().equals(session.getId())) {
                sendMessage(s, msg.type, msg.userId, msg.userName, msg.data);
            }
        }
    }

    private void handleMediaToggle(WebSocketSession session, SignalingMessage msg) throws IOException {
        String roomCode = sessionRooms.get(session.getId());
        if (roomCode == null) return;
        Map<String, WebSocketSession> sessions = roomSessions.get(roomCode);
        if (sessions == null) return;

        for (WebSocketSession s : sessions.values()) {
            if (!s.getId().equals(session.getId())) {
                sendMessage(s, msg.type, msg.userId, msg.userName, msg.data);
            }
        }
    }

    private void sendMessage(WebSocketSession session, String type, Object userId, Object userName, Object data) throws IOException {
        Map<String, Object> payload = new ConcurrentHashMap<>();
        payload.put("type", type);
        payload.put("userId", userId);
        payload.put("userName", userName);
        payload.put("data", data);
        String json = objectMapper.writeValueAsString(payload);
        session.sendMessage(new TextMessage(json));
    }

    private void removeSession(WebSocketSession session) {
        String roomCode = sessionRooms.remove(session.getId());
        sessionUsers.remove(session.getId());
        if (roomCode != null) {
            Map<String, WebSocketSession> sessions = roomSessions.get(roomCode);
            if (sessions != null) {
                sessions.remove(session.getId());
                if (sessions.isEmpty()) {
                    roomSessions.remove(roomCode);
                }
            }
        }
    }

    @Override
    public void afterConnectionClosed(WebSocketSession session, CloseStatus status) throws Exception {
        log.info("[Meeting WS] 连接关闭: {}", session.getId());
        String roomCode = sessionRooms.get(session.getId());
        Long userId = sessionUsers.get(session.getId());
        removeSession(session);
        if (roomCode != null) {
            Map<String, WebSocketSession> sessions = roomSessions.get(roomCode);
            if (sessions != null) {
                for (WebSocketSession s : sessions.values()) {
                    try {
                        sendMessage(s, "user-left", userId, null, null);
                    } catch (IOException ignored) {}
                }
            }
        }
    }
}