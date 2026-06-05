package com.mybilibili.live.common;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.junit.jupiter.api.Assertions.assertEquals;

class LiveConstantsTest {

    @Test
    void meetingStatusPredicatesDescribeAllowedTransitions() {
        assertTrue(LiveConstants.MeetingStatus.isValid(LiveConstants.MeetingStatus.NOT_STARTED));
        assertTrue(LiveConstants.MeetingStatus.isValid(LiveConstants.MeetingStatus.IN_PROGRESS));
        assertTrue(LiveConstants.MeetingStatus.isValid(LiveConstants.MeetingStatus.ENDED));
        assertTrue(LiveConstants.MeetingStatus.isValid(LiveConstants.MeetingStatus.PENDING_APPROVAL));
        assertTrue(LiveConstants.MeetingStatus.isValid(LiveConstants.MeetingStatus.REJECTED));

        assertTrue(LiveConstants.MeetingStatus.isJoinable(LiveConstants.MeetingStatus.NOT_STARTED));
        assertTrue(LiveConstants.MeetingStatus.isJoinable(LiveConstants.MeetingStatus.IN_PROGRESS));
        assertFalse(LiveConstants.MeetingStatus.isJoinable(LiveConstants.MeetingStatus.PENDING_APPROVAL));
        assertFalse(LiveConstants.MeetingStatus.isJoinable(LiveConstants.MeetingStatus.REJECTED));
        assertFalse(LiveConstants.MeetingStatus.isJoinable(LiveConstants.MeetingStatus.ENDED));
        assertFalse(LiveConstants.MeetingStatus.isValid(99));
    }

    @Test
    void srsActionsMapOnlyPublishLifecycleEventsToRoomStatus() {
        assertEquals(LiveConstants.RoomStatus.LIVE,
                LiveConstants.SrsAction.roomStatusFor(LiveConstants.SrsAction.ON_PUBLISH));
        assertEquals(LiveConstants.RoomStatus.OFFLINE,
                LiveConstants.SrsAction.roomStatusFor(LiveConstants.SrsAction.ON_UNPUBLISH));
        assertNull(LiveConstants.SrsAction.roomStatusFor("on_play"));
        assertNull(LiveConstants.SrsAction.roomStatusFor(null));
    }

    @Test
    void linkmicStatusPredicatesDescribePendingAndFinishedStates() {
        assertTrue(LiveConstants.LinkmicStatus.isPending(LiveConstants.LinkmicStatus.PENDING));
        assertTrue(LiveConstants.LinkmicStatus.isConnected(LiveConstants.LinkmicStatus.CONNECTED));
        assertTrue(LiveConstants.LinkmicStatus.isFinished(LiveConstants.LinkmicStatus.DISCONNECTED));
        assertTrue(LiveConstants.LinkmicStatus.isFinished(LiveConstants.LinkmicStatus.REJECTED));

        assertFalse(LiveConstants.LinkmicStatus.isPending(LiveConstants.LinkmicStatus.CONNECTED));
        assertFalse(LiveConstants.LinkmicStatus.isFinished(LiveConstants.LinkmicStatus.PENDING));
    }
}
