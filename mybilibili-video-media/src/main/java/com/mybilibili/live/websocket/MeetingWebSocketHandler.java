package com.mybilibili.live.websocket;

import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import org.springframework.web.socket.CloseStatus;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;
import org.springframework.web.socket.handler.TextWebSocketHandler;

import java.io.IOException;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

@Slf4j
@Component
public class MeetingWebSocketHandler extends TextWebSocketHandler {

    @Autowired
    private ObjectMapper objectMapper;

    private final MeetingSessionRegistry sessionRegistry = new MeetingSessionRegistry();

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
            if (msg == null || !StringUtils.hasText(msg.type)) {
                return;
            }
            log.info("[Meeting WS] type={}, roomCode={}, userId={}, target={}",
                    msg.type, msg.roomCode, msg.userId, msg.targetUserId);

            switch (msg.type) {
                case MeetingWsMessageType.PING:
                    // 心跳，回个 pong，不广播
                    sendMessage(session, MeetingWsMessageType.PONG, null, null, null, null);
                    break;
                case MeetingWsMessageType.JOIN:
                    handleJoin(session, msg);
                    break;
                case MeetingWsMessageType.LEAVE:
                    handleLeave(session);
                    break;
                case MeetingWsMessageType.JOIN_REQUEST:
                    handleJoinRequest(session, msg);
                    break;
                case MeetingWsMessageType.ADMIT_USER:
                    handleAdmitUser(session, msg);
                    break;
                case MeetingWsMessageType.REJECT_USER:
                    handleRejectUser(session, msg);
                    break;
                default:
                    if (!MeetingWsMessageType.isForwarded(msg.type)) {
                        log.warn("[Meeting WS] 未知消息类型: {}", msg.type);
                        return;
                    }
                    handleSignaling(session, msg);
                    break;
            }
        } catch (Exception e) {
            log.error("[Meeting WS] 处理消息失败: {}", e.getMessage(), e);
        }
    }

    private void handleJoin(WebSocketSession session, SignalingMessage msg) throws IOException {
        String roomCode = msg.roomCode;
        if (!StringUtils.hasText(roomCode) || msg.userId == null) {
            log.warn("[Meeting WS] join 参数无效: roomCode={}, userId={}", roomCode, msg.userId);
            return;
        }
        sessionRegistry.removeSession(session);
        // 先收集已有的成员（在添加新成员之前）
        List<Map<String, Object>> existingUsers = sessionRegistry.collectExistingUsers(roomCode);
        sessionRegistry.registerRoomSession(roomCode, session, msg.userId, msg.userName);

        // 通知现有成员，有新人加入（让他们发起 offer）
        broadcastToRoom(roomCode, MeetingWsMessageType.USER_JOINED,
                msg.userId, msg.userName, null, null, session.getId());

        // 把现有成员列表发给新人
        sendMessage(session, MeetingWsMessageType.ROOM_USERS, null, null, null, existingUsers);

        broadcastViewerCount(roomCode);
    }

    private void handleJoinRequest(WebSocketSession session, SignalingMessage msg) throws IOException {
        String roomCode = msg.roomCode;
        if (!StringUtils.hasText(roomCode) || msg.userId == null) {
            log.warn("[Meeting WS] join-request 参数无效: roomCode={}, userId={}", roomCode, msg.userId);
            return;
        }
        sessionRegistry.registerWaitingSession(roomCode, session, msg.userId, msg.userName);

        broadcastToRoom(roomCode, MeetingWsMessageType.JOIN_REQUEST, msg.userId, msg.userName, null, msg.data);
    }

    private void handleAdmitUser(WebSocketSession session, SignalingMessage msg) throws IOException {
        handleWaitingDecision(session, msg, MeetingWsMessageType.ADMIT_USER, false);
    }

    private void handleRejectUser(WebSocketSession session, SignalingMessage msg) throws IOException {
        handleWaitingDecision(session, msg, MeetingWsMessageType.REJECT_USER, true);
    }

    private void broadcastViewerCount(String roomCode) {
        Map<String, WebSocketSession> sessions = sessionRegistry.getRoomSessions(roomCode);
        if (sessions.isEmpty()) return;
        int count = sessions.size();
        Map<String, Object> data = new LinkedHashMap<>();
        data.put("count", count);
        for (WebSocketSession s : sessions.values()) {
            try {
                sendMessage(s, MeetingWsMessageType.VIEWER_COUNT, null, null, null, data);
            } catch (IOException ignored) {}
        }
    }

    private void handleLeave(WebSocketSession session) throws IOException {
        MeetingSessionRegistry.SessionSnapshot snapshot = sessionRegistry.snapshot(session);
        sessionRegistry.removeSession(session);
        notifyRoomDeparture(snapshot.roomCode(), snapshot.userId(), snapshot.userName());
    }

    /**
     * 信令转发：
     * - 如果指定了 targetUserId，仅投递给该用户（精确路由 offer/answer/ice）
     * - 否则广播给房间内除自己外的所有人（房间级通知，比如 linkmic-apply）
     */
    private void handleSignaling(WebSocketSession session, SignalingMessage msg) throws IOException {
        String roomCode = sessionRegistry.resolveRoomCode(session, msg);
        if (roomCode == null) return;
        Map<String, WebSocketSession> sessions = sessionRegistry.getRoomSessions(roomCode);
        if (sessions.isEmpty()) return;

        if (msg.targetUserId != null) {
            for (WebSocketSession s : sessions.values()) {
                Long uid = sessionRegistry.getUserId(s);
                if (uid != null && uid.equals(msg.targetUserId)) {
                    sendMessage(s, msg.type, msg.userId, msg.userName, msg.targetUserId, msg.data);
                    return;
                }
            }
            log.warn("[Meeting WS] 找不到 targetUserId={} 在房间 {}", msg.targetUserId, roomCode);
            return;
        }

        broadcastToRoom(roomCode, msg.type, msg.userId, msg.userName, null, msg.data, session.getId());
    }

    private void handleWaitingDecision(WebSocketSession session, SignalingMessage msg,
                                       String messageType, boolean removeAfterSend) throws IOException {
        String roomCode = sessionRegistry.resolveRoomCode(session, msg);
        if (!StringUtils.hasText(roomCode) || msg.targetUserId == null) {
            log.warn("[Meeting WS] {} 参数无效: roomCode={}, target={}",
                    messageType, roomCode, msg.targetUserId);
            return;
        }

        WebSocketSession targetSession = sessionRegistry.findWaitingSessionByUser(roomCode, msg.targetUserId);
        if (targetSession == null) {
            return;
        }

        Long operatorId = sessionRegistry.getUserId(session);
        if (operatorId == null) {
            operatorId = msg.userId;
        }
        String operatorName = sessionRegistry.getUserName(session);
        if (!StringUtils.hasText(operatorName)) {
            operatorName = msg.userName;
        }
        sendMessage(targetSession, messageType, operatorId, operatorName, msg.targetUserId, msg.data);
        if (removeAfterSend) {
            sessionRegistry.removeWaitingSession(targetSession);
        }
    }

    private void broadcastToRoom(String roomCode, String type, Object userId, Object userName,
                                 Object targetUserId, Object data) throws IOException {
        broadcastToRoom(roomCode, type, userId, userName, targetUserId, data, null);
    }

    private void broadcastToRoom(String roomCode, String type, Object userId, Object userName,
                                 Object targetUserId, Object data, String excludedSessionId) throws IOException {
        broadcastToRoom(roomCode, type, userId, userName, targetUserId, data, excludedSessionId, false);
    }

    private void broadcastToRoom(String roomCode, String type, Object userId, Object userName,
                                 Object targetUserId, Object data, String excludedSessionId,
                                 boolean ignoreSendFailures) throws IOException {
        Map<String, WebSocketSession> sessions = sessionRegistry.getRoomSessions(roomCode);
        if (sessions.isEmpty()) {
            return;
        }
        for (WebSocketSession s : sessions.values()) {
            if (excludedSessionId != null && excludedSessionId.equals(s.getId())) {
                continue;
            }
            try {
                sendMessage(s, type, userId, userName, targetUserId, data);
            } catch (IOException e) {
                if (!ignoreSendFailures) {
                    throw e;
                }
            }
        }
    }

    private void notifyRoomDeparture(String roomCode, Long userId, String userName) throws IOException {
        notifyRoomDeparture(roomCode, userId, userName, false);
    }

    private void notifyRoomDeparture(String roomCode, Long userId, String userName,
                                     boolean ignoreSendFailures) throws IOException {
        if (roomCode == null) {
            return;
        }
        broadcastToRoom(roomCode, MeetingWsMessageType.USER_LEFT,
                userId, userName, null, null, null, ignoreSendFailures);
        broadcastViewerCount(roomCode);
    }

    private void sendMessage(WebSocketSession session, String type, Object userId, Object userName,
                              Object targetUserId, Object data) throws IOException {
        Map<String, Object> payload = new LinkedHashMap<>();
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

    @Override
    public void afterConnectionClosed(WebSocketSession session, CloseStatus status) throws Exception {
        log.info("[Meeting WS] 连接关闭: {}", session.getId());
        MeetingSessionRegistry.SessionSnapshot snapshot = sessionRegistry.snapshot(session);
        sessionRegistry.removeSession(session);
        if (snapshot.userId() != null) {
            notifyRoomDeparture(snapshot.roomCode(), snapshot.userId(), snapshot.userName(), true);
        }
    }
}
