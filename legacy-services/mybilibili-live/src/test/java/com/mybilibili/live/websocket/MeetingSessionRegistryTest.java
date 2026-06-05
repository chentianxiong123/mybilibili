package com.mybilibili.live.websocket;

import org.junit.jupiter.api.Test;
import org.springframework.web.socket.WebSocketSession;

import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertSame;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

class MeetingSessionRegistryTest {

    private final MeetingSessionRegistry registry = new MeetingSessionRegistry();

    @Test
    void registerRoomSessionStoresSessionAndUserSnapshot() {
        WebSocketSession session = session("host");

        registry.registerRoomSession("room-1", session, 1L, "主持人");

        MeetingSessionRegistry.SessionSnapshot snapshot = registry.snapshot(session);
        assertEquals("room-1", snapshot.roomCode());
        assertEquals(1L, snapshot.userId());
        assertEquals("主持人", snapshot.userName());
        assertSame(session, registry.getRoomSessions("room-1").get("host"));

        List<Map<String, Object>> existingUsers = registry.collectExistingUsers("room-1");
        assertEquals(1, existingUsers.size());
        assertEquals(1L, existingUsers.get(0).get("userId"));
        assertEquals("主持人", existingUsers.get(0).get("userName"));
    }

    @Test
    void registerWaitingSessionMovesSameSessionBetweenWaitingRooms() {
        WebSocketSession session = session("waiting");

        registry.registerWaitingSession("room-1", session, 2L, "观众");
        assertSame(session, registry.findWaitingSessionByUser("room-1", 2L));

        registry.registerWaitingSession("room-2", session, 2L, "观众");

        assertNull(registry.findWaitingSessionByUser("room-1", 2L));
        assertSame(session, registry.findWaitingSessionByUser("room-2", 2L));
    }

    @Test
    void removeSessionClearsRoomAndWaitingState() {
        WebSocketSession session = session("guest");
        registry.registerRoomSession("room-1", session, 3L, "参会者");
        registry.registerWaitingSession("room-2", session, 3L, "参会者");

        registry.removeSession(session);

        MeetingSessionRegistry.SessionSnapshot snapshot = registry.snapshot(session);
        assertNull(snapshot.roomCode());
        assertNull(snapshot.userId());
        assertNull(snapshot.userName());
        assertTrue(registry.getRoomSessions("room-1").isEmpty());
        assertNull(registry.findWaitingSessionByUser("room-2", 3L));
    }

    @Test
    void removeWaitingSessionClearsWaitingOnlyUserMetadata() {
        WebSocketSession session = session("waiting");
        registry.registerWaitingSession("room-1", session, 3L, "参会者");

        registry.removeWaitingSession(session);

        MeetingSessionRegistry.SessionSnapshot snapshot = registry.snapshot(session);
        assertNull(snapshot.roomCode());
        assertNull(snapshot.userId());
        assertNull(snapshot.userName());
        assertNull(registry.getUserId(session));
        assertNull(registry.getUserName(session));
        assertNull(registry.findWaitingSessionByUser("room-1", 3L));
    }

    @Test
    void nullUserNameIsNormalizedToEmptyString() {
        WebSocketSession session = session("anonymous");

        registry.registerRoomSession("room-1", session, 4L, null);

        assertEquals("", registry.getUserName(session));
        assertEquals("", registry.snapshot(session).userName());
        List<Map<String, Object>> existingUsers = registry.collectExistingUsers("room-1");
        assertEquals("", existingUsers.get(0).get("userName"));
    }

    private WebSocketSession session(String id) {
        WebSocketSession session = mock(WebSocketSession.class);
        when(session.getId()).thenReturn(id);
        return session;
    }
}
