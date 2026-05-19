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

@Slf4j
@Component
public class MeetingWebSocketHandler extends TextWebSocketHandler {

    @Autowired
    private ObjectMapper objectMapper;

    private final Map<String, Map<String, WebSocketSession>> roomSessions = new ConcurrentHashMap<>();
    private final Map<String, String> sessionRooms = new ConcurrentHashMap<>();
    private final Map<String, Long> sessionUsers = new ConcurrentHashMap<>();
    private final Map<String, String> sessionUserNames = new ConcurrentHashMap<>();

    public static class SignalingMessage {
        public String type;
        public String roomCode;
        public Long userId;
        public String userName;
        public Long targetUserId;
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
            log.info("[Meeting WS] type={}, roomCode={}, userId={}, target={}",
                    msg.type, msg.roomCode, msg.userId, msg.targetUserId);

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
                case "linkmic-apply":
                case "linkmic-accepted":
                case "linkmic-rejected":
                case "linkmic-disconnected":
                case "toggle-audio":
                case "toggle-video":
                case "toggle-screen":
                    handleSignaling(session, msg);
                    break;
                default:
                    log.warn("[Meeting WS] 未知消息类型: {}", msg.type);
            }
        } catch (Exception e) {
            log.error("[Meeting WS] 处理消息失败: {}", e.getMessage(), e);
        }
    }

    private void handleJoin(WebSocketSession session, SignalingMessage msg) throws IOException {
        String roomCode = msg.roomCode;
        roomSessions.computeIfAbsent(roomCode, k -> new ConcurrentHashMap<>());
        Map<String, WebSocketSession> sessions = roomSessions.get(roomCode);

        // 先收集已有的成员（在添加新成员之前）
        List<Map<String, Object>> existingUsers = new ArrayList<>();
        for (WebSocketSession s : sessions.values()) {
            Long uid = sessionUsers.get(s.getId());
            String uname = sessionUserNames.get(s.getId());
            if (uid != null) {
                Map<String, Object> u = new ConcurrentHashMap<>();
                u.put("userId", uid);
                u.put("userName", uname == null ? "" : uname);
                existingUsers.add(u);
            }
        }

        // 注册新成员
        sessions.put(session.getId(), session);
        sessionRooms.put(session.getId(), roomCode);
        sessionUsers.put(session.getId(), msg.userId);
        sessionUserNames.put(session.getId(), msg.userName == null ? "" : msg.userName);

        // 通知现有成员，有新人加入（让他们发起 offer）
        for (WebSocketSession s : sessions.values()) {
            if (!s.getId().equals(session.getId())) {
                sendMessage(s, "user-joined", msg.userId, msg.userName, null, null);
            }
        }

        // 把现有成员列表发给新人
        sendMessage(session, "room-users", null, null, null, existingUsers);
    }

    private void handleLeave(WebSocketSession session, SignalingMessage msg) throws IOException {
        Long userId = sessionUsers.get(session.getId());
        String userName = sessionUserNames.get(session.getId());
        String roomCode = sessionRooms.get(session.getId());
        removeSession(session);
        if (roomCode != null) {
            Map<String, WebSocketSession> sessions = roomSessions.get(roomCode);
            if (sessions != null) {
                for (WebSocketSession s : sessions.values()) {
                    sendMessage(s, "user-left", userId, userName, null, null);
                }
            }
        }
    }

    /**
     * 信令转发：
     * - 如果指定了 targetUserId，仅投递给该用户（精确路由 offer/answer/ice）
     * - 否则广播给房间内除自己外的所有人（房间级通知，比如 linkmic-apply）
     */
    private void handleSignaling(WebSocketSession session, SignalingMessage msg) throws IOException {
        String roomCode = sessionRooms.get(session.getId());
        if (roomCode == null) return;
        Map<String, WebSocketSession> sessions = roomSessions.get(roomCode);
        if (sessions == null) return;

        if (msg.targetUserId != null) {
            for (WebSocketSession s : sessions.values()) {
                Long uid = sessionUsers.get(s.getId());
                if (uid != null && uid.equals(msg.targetUserId)) {
                    sendMessage(s, msg.type, msg.userId, msg.userName, msg.targetUserId, msg.data);
                    return;
                }
            }
            log.warn("[Meeting WS] 找不到 targetUserId={} 在房间 {}", msg.targetUserId, roomCode);
            return;
        }

        for (WebSocketSession s : sessions.values()) {
            if (!s.getId().equals(session.getId())) {
                sendMessage(s, msg.type, msg.userId, msg.userName, null, msg.data);
            }
        }
    }

    private void sendMessage(WebSocketSession session, String type, Object userId, Object userName,
                              Object targetUserId, Object data) throws IOException {
        Map<String, Object> payload = new ConcurrentHashMap<>();
        payload.put("type", type);
        if (userId != null) payload.put("userId", userId);
        if (userName != null) payload.put("userName", userName);
        if (targetUserId != null) payload.put("targetUserId", targetUserId);
        if (data != null) payload.put("data", data);
        String json = objectMapper.writeValueAsString(payload);
        synchronized (session) {
            if (session.isOpen()) {
                session.sendMessage(new TextMessage(json));
            }
        }
    }

    private void removeSession(WebSocketSession session) {
        String roomCode = sessionRooms.remove(session.getId());
        sessionUsers.remove(session.getId());
        sessionUserNames.remove(session.getId());
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
        String userName = sessionUserNames.get(session.getId());
        removeSession(session);
        if (roomCode != null && userId != null) {
            Map<String, WebSocketSession> sessions = roomSessions.get(roomCode);
            if (sessions != null) {
                for (WebSocketSession s : sessions.values()) {
                    try {
                        sendMessage(s, "user-left", userId, userName, null, null);
                    } catch (IOException ignored) {}
                }
            }
        }
    }
}
