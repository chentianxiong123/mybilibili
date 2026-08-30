package com.mybilibili.live.websocket;

import org.springframework.util.StringUtils;
import org.springframework.web.socket.WebSocketSession;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

final class MeetingSessionRegistry {

    private final Map<String, Map<String, WebSocketSession>> roomSessions = new ConcurrentHashMap<>();
    private final Map<String, Map<String, WebSocketSession>> waitingSessions = new ConcurrentHashMap<>();
    private final Map<String, String> sessionRooms = new ConcurrentHashMap<>();
    private final Map<String, String> waitingSessionRooms = new ConcurrentHashMap<>();
    private final Map<String, Long> sessionUsers = new ConcurrentHashMap<>();
    private final Map<String, String> sessionUserNames = new ConcurrentHashMap<>();

    SessionSnapshot snapshot(WebSocketSession session) {
        return new SessionSnapshot(
                sessionRooms.get(session.getId()),
                sessionUsers.get(session.getId()),
                sessionUserNames.get(session.getId())
        );
    }

    Map<String, WebSocketSession> getRoomSessions(String roomCode) {
        return StringUtils.hasText(roomCode)
                ? roomSessions.getOrDefault(roomCode, Map.of())
                : Map.of();
    }

    void registerRoomSession(String roomCode, WebSocketSession session, Long userId, String userName) {
        roomSessions.computeIfAbsent(roomCode, k -> new ConcurrentHashMap<>()).put(session.getId(), session);
        sessionRooms.put(session.getId(), roomCode);
        sessionUsers.put(session.getId(), userId);
        sessionUserNames.put(session.getId(), normalizeUserName(userName));
    }

    void registerWaitingSession(String roomCode, WebSocketSession session, Long userId, String userName) {
        removeWaitingSession(session);
        waitingSessions.computeIfAbsent(roomCode, k -> new ConcurrentHashMap<>()).put(session.getId(), session);
        waitingSessionRooms.put(session.getId(), roomCode);
        sessionUsers.put(session.getId(), userId);
        sessionUserNames.put(session.getId(), normalizeUserName(userName));
    }

    void removeSession(WebSocketSession session) {
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
        removeWaitingSession(session);
    }

    void removeWaitingSession(WebSocketSession session) {
        String sessionId = session.getId();
        String waitingRoomCode = waitingSessionRooms.remove(sessionId);
        if (waitingRoomCode == null) return;
        Map<String, WebSocketSession> waiting = waitingSessions.get(waitingRoomCode);
        if (waiting != null) {
            waiting.remove(sessionId);
            if (waiting.isEmpty()) {
                waitingSessions.remove(waitingRoomCode);
            }
        }
        if (!sessionRooms.containsKey(sessionId)) {
            sessionUsers.remove(sessionId);
            sessionUserNames.remove(sessionId);
        }
    }

    String resolveRoomCode(WebSocketSession session, MeetingWebSocketHandler.SignalingMessage msg) {
        String roomCode = sessionRooms.get(session.getId());
        if (StringUtils.hasText(roomCode)) {
            return roomCode;
        }
        return StringUtils.hasText(msg.roomCode) ? msg.roomCode : null;
    }

    List<Map<String, Object>> collectExistingUsers(String roomCode) {
        List<Map<String, Object>> existingUsers = new ArrayList<>();
        for (WebSocketSession s : getRoomSessions(roomCode).values()) {
            Long uid = sessionUsers.get(s.getId());
            if (uid == null) {
                continue;
            }
            Map<String, Object> user = new LinkedHashMap<>();
            user.put("userId", uid);
            user.put("userName", normalizeUserName(sessionUserNames.get(s.getId())));
            existingUsers.add(user);
        }
        return existingUsers;
    }

    WebSocketSession findWaitingSessionByUser(String roomCode, Long userId) {
        Map<String, WebSocketSession> waiting = waitingSessions.get(roomCode);
        if (waiting == null) return null;
        for (WebSocketSession session : waiting.values()) {
            Long uid = sessionUsers.get(session.getId());
            if (userId.equals(uid)) {
                return session;
            }
        }
        return null;
    }

    Long getUserId(WebSocketSession session) {
        return sessionUsers.get(session.getId());
    }

    String getUserName(WebSocketSession session) {
        return sessionUserNames.get(session.getId());
    }

    private String normalizeUserName(String userName) {
        return userName == null ? "" : userName;
    }

    record SessionSnapshot(String roomCode, Long userId, String userName) {
    }
}
