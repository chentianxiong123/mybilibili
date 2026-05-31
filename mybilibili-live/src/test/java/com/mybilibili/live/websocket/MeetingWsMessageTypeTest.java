package com.mybilibili.live.websocket;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

class MeetingWsMessageTypeTest {

    @Test
    void forwardedPredicateIncludesRoomAndWebRtcMessagesOnly() {
        assertTrue(MeetingWsMessageType.isForwarded(MeetingWsMessageType.OFFER));
        assertTrue(MeetingWsMessageType.isForwarded(MeetingWsMessageType.CHAT));
        assertTrue(MeetingWsMessageType.isForwarded(MeetingWsMessageType.MEETING_ENDED));
        assertTrue(MeetingWsMessageType.isForwarded(MeetingWsMessageType.LINKMIC_APPLY));

        assertFalse(MeetingWsMessageType.isForwarded(MeetingWsMessageType.PING));
        assertFalse(MeetingWsMessageType.isForwarded(MeetingWsMessageType.JOIN));
        assertFalse(MeetingWsMessageType.isForwarded(MeetingWsMessageType.LEAVE));
        assertFalse(MeetingWsMessageType.isForwarded(MeetingWsMessageType.JOIN_REQUEST));
        assertFalse(MeetingWsMessageType.isForwarded("unknown-message"));
        assertFalse(MeetingWsMessageType.isForwarded(null));
    }
}
