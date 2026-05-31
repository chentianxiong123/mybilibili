package com.mybilibili.live.websocket;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;
import org.springframework.test.util.ReflectionTestUtils;
import org.springframework.web.socket.CloseStatus;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;

import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.clearInvocations;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@SuppressWarnings("unchecked")
class MeetingWebSocketHandlerTest {

    private final ObjectMapper objectMapper = new ObjectMapper();
    private MeetingWebSocketHandler handler;

    @BeforeEach
    void setUp() {
        handler = new MeetingWebSocketHandler();
        ReflectionTestUtils.setField(handler, "objectMapper", objectMapper);
    }

    @Test
    void joinSendsExistingUsersToNewMemberAndNotifiesRoom() throws Exception {
        WebSocketSession host = session("host");
        WebSocketSession guest = session("guest");
        join(host, 1L, "主持人");
        clearInvocations(host);

        join(guest, 2L, "观众");

        List<Map<String, Object>> hostMessages = sentMessages(host);
        assertTrue(hasMessage(hostMessages, "user-joined", Map.of("userId", 2)));
        assertTrue(hasViewerCount(hostMessages, 2));

        List<Map<String, Object>> guestMessages = sentMessages(guest);
        Map<String, Object> roomUsers = firstMessage(guestMessages, "room-users");
        List<Map<String, Object>> existingUsers = (List<Map<String, Object>>) roomUsers.get("data");
        assertEquals(1, existingUsers.size());
        assertEquals(1, existingUsers.get(0).get("userId"));
        assertEquals("主持人", existingUsers.get(0).get("userName"));
        assertTrue(hasViewerCount(guestMessages, 2));
    }

    @Test
    void rejoinDoesNotIncludeCurrentSessionInExistingUsers() throws Exception {
        WebSocketSession host = session("host");
        join(host, 1L, "主持人");
        clearInvocations(host);

        join(host, 1L, "主持人");

        Map<String, Object> roomUsers = firstMessage(sentMessages(host), "room-users");
        List<Map<String, Object>> existingUsers = (List<Map<String, Object>>) roomUsers.get("data");
        assertTrue(existingUsers.isEmpty());
    }

    @Test
    void targetedSignalingOnlyRoutesToTargetUser() throws Exception {
        WebSocketSession host = session("host");
        WebSocketSession guest = session("guest");
        join(host, 1L, "主持人");
        join(guest, 2L, "观众");
        clearInvocations(host, guest);

        send(guest, Map.of(
                "type", "offer",
                "roomCode", "room-1",
                "userId", 2,
                "userName", "观众",
                "targetUserId", 1,
                "data", Map.of("sdp", "offer-sdp")
        ));

        List<Map<String, Object>> hostMessages = sentMessages(host);
        Map<String, Object> offer = firstMessage(hostMessages, "offer");
        assertEquals(2, offer.get("userId"));
        assertEquals(1, offer.get("targetUserId"));
        assertEquals(Map.of("sdp", "offer-sdp"), offer.get("data"));
        verify(guest, never()).sendMessage(any(TextMessage.class));
    }

    @Test
    void leaveNotifiesRemainingMembersAndUpdatesViewerCount() throws Exception {
        WebSocketSession host = session("host");
        WebSocketSession guest = session("guest");
        join(host, 1L, "主持人");
        join(guest, 2L, "观众");
        clearInvocations(host, guest);

        send(guest, Map.of(
                "type", "leave",
                "roomCode", "room-1",
                "userId", 2
        ));

        List<Map<String, Object>> hostMessages = sentMessages(host);
        assertTrue(hasMessage(hostMessages, "user-left", Map.of(
                "userId", 2,
                "userName", "观众"
        )));
        assertTrue(hasViewerCount(hostMessages, 1));
        verify(guest, never()).sendMessage(any(TextMessage.class));
    }

    @Test
    void connectionCloseNotifiesRemainingMembersAndUpdatesViewerCount() throws Exception {
        WebSocketSession host = session("host");
        WebSocketSession guest = session("guest");
        join(host, 1L, "主持人");
        join(guest, 2L, "观众");
        clearInvocations(host, guest);

        handler.afterConnectionClosed(guest, CloseStatus.NORMAL);

        List<Map<String, Object>> hostMessages = sentMessages(host);
        assertTrue(hasMessage(hostMessages, "user-left", Map.of(
                "userId", 2,
                "userName", "观众"
        )));
        assertTrue(hasViewerCount(hostMessages, 1));
        verify(guest, never()).sendMessage(any(TextMessage.class));
    }

    @Test
    void waitingRoomRequestIsForwardedToHostAndAdmitTargetsWaitingUser() throws Exception {
        WebSocketSession host = session("host");
        WebSocketSession waiting = session("waiting");
        join(host, 1L, "主持人");
        clearInvocations(host);

        send(waiting, Map.of(
                "type", "join-request",
                "roomCode", "room-1",
                "userId", 2,
                "userName", "观众",
                "data", Map.of("reason", "申请入会")
        ));

        Map<String, Object> request = firstMessage(sentMessages(host), "join-request");
        assertEquals(2, request.get("userId"));
        assertEquals("观众", request.get("userName"));
        assertEquals(Map.of("reason", "申请入会"), request.get("data"));

        send(host, Map.of(
                "type", "admit-user",
                "roomCode", "room-1",
                "userId", 1,
                "targetUserId", 2,
                "data", Map.of("roomId", 10)
        ));

        Map<String, Object> admit = firstMessage(sentMessages(waiting), "admit-user");
        assertEquals(1, admit.get("userId"));
        assertEquals("主持人", admit.get("userName"));
        assertEquals(2, admit.get("targetUserId"));
        assertEquals(Map.of("roomId", 10), admit.get("data"));
    }

    @Test
    void rejectTargetsWaitingUserAndRemovesThemFromWaitingRoom() throws Exception {
        WebSocketSession host = session("host");
        WebSocketSession waiting = session("waiting");
        join(host, 1L, "主持人");
        send(waiting, Map.of(
                "type", "join-request",
                "roomCode", "room-1",
                "userId", 2,
                "userName", "观众"
        ));
        clearInvocations(waiting);

        send(host, Map.of(
                "type", "reject-user",
                "roomCode", "room-1",
                "userId", 1,
                "targetUserId", 2,
                "data", Map.of("reason", "拒绝")
        ));

        Map<String, Object> rejected = firstMessage(sentMessages(waiting), "reject-user");
        assertEquals(1, rejected.get("userId"));
        assertEquals("主持人", rejected.get("userName"));
        assertEquals(2, rejected.get("targetUserId"));
        assertEquals(Map.of("reason", "拒绝"), rejected.get("data"));

        clearInvocations(waiting);
        send(host, Map.of(
                "type", "admit-user",
                "roomCode", "room-1",
                "userId", 1,
                "targetUserId", 2
        ));

        verify(waiting, never()).sendMessage(any(TextMessage.class));
    }

    @Test
    void joinRequestMovesWaitingSessionOutOfPreviousRoom() throws Exception {
        WebSocketSession hostOne = session("host-one");
        WebSocketSession hostTwo = session("host-two");
        WebSocketSession waiting = session("waiting");
        join(hostOne, "room-1", 1L, "主持人一");
        join(hostTwo, "room-2", 2L, "主持人二");

        send(waiting, Map.of(
                "type", "join-request",
                "roomCode", "room-1",
                "userId", 3,
                "userName", "观众"
        ));
        send(waiting, Map.of(
                "type", "join-request",
                "roomCode", "room-2",
                "userId", 3,
                "userName", "观众"
        ));
        clearInvocations(waiting);

        send(hostOne, Map.of(
                "type", "admit-user",
                "roomCode", "room-1",
                "userId", 1,
                "targetUserId", 3
        ));

        verify(waiting, never()).sendMessage(any(TextMessage.class));

        send(hostTwo, Map.of(
                "type", "admit-user",
                "roomCode", "room-2",
                "userId", 2,
                "targetUserId", 3
        ));

        Map<String, Object> admit = firstMessage(sentMessages(waiting), "admit-user");
        assertEquals(2, admit.get("userId"));
        assertEquals("主持人二", admit.get("userName"));
        assertEquals(3, admit.get("targetUserId"));
    }

    private WebSocketSession session(String id) {
        WebSocketSession session = mock(WebSocketSession.class);
        when(session.getId()).thenReturn(id);
        when(session.isOpen()).thenReturn(true);
        return session;
    }

    private void join(WebSocketSession session, Long userId, String userName) throws Exception {
        join(session, "room-1", userId, userName);
    }

    private void join(WebSocketSession session, String roomCode, Long userId, String userName) throws Exception {
        send(session, Map.of(
                "type", "join",
                "roomCode", roomCode,
                "userId", userId,
                "userName", userName
        ));
    }

    private void send(WebSocketSession session, Map<String, Object> payload) throws Exception {
        handler.handleTextMessage(session, new TextMessage(objectMapper.writeValueAsString(payload)));
    }

    private List<Map<String, Object>> sentMessages(WebSocketSession session) throws Exception {
        ArgumentCaptor<TextMessage> captor = ArgumentCaptor.forClass(TextMessage.class);
        verify(session, org.mockito.Mockito.atLeastOnce()).sendMessage(captor.capture());
        return captor.getAllValues().stream()
                .map(TextMessage::getPayload)
                .map(this::readPayload)
                .toList();
    }

    private Map<String, Object> readPayload(String payload) {
        try {
            return objectMapper.readValue(payload, new TypeReference<>() {});
        } catch (Exception e) {
            throw new AssertionError(e);
        }
    }

    private Map<String, Object> firstMessage(List<Map<String, Object>> messages, String type) {
        return messages.stream()
                .filter(message -> type.equals(message.get("type")))
                .findFirst()
                .orElseThrow(() -> new AssertionError("Missing message type: " + type));
    }

    private boolean hasMessage(List<Map<String, Object>> messages, String type, Map<String, Object> fields) {
        return messages.stream().anyMatch(message -> {
            if (!type.equals(message.get("type"))) return false;
            return fields.entrySet().stream().allMatch(field -> field.getValue().equals(message.get(field.getKey())));
        });
    }

    private boolean hasViewerCount(List<Map<String, Object>> messages, int count) {
        return messages.stream()
                .filter(message -> "viewer-count".equals(message.get("type")))
                .map(message -> (Map<String, Object>) message.get("data"))
                .anyMatch(data -> Integer.valueOf(count).equals(data.get("count")));
    }
}
