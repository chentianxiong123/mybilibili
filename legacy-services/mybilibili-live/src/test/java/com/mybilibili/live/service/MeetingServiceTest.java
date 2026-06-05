package com.mybilibili.live.service;

import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.live.common.LiveConstants;
import com.mybilibili.live.entity.MeetingParticipant;
import com.mybilibili.live.entity.MeetingRoom;
import com.mybilibili.live.mapper.MeetingParticipantMapper;
import com.mybilibili.live.mapper.MeetingRoomMapper;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.time.LocalDateTime;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertSame;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.reset;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class MeetingServiceTest {

    @Mock
    private MeetingRoomMapper roomMapper;

    @Mock
    private MeetingParticipantMapper participantMapper;

    @InjectMocks
    private MeetingService meetingService;

    @Test
    void reserveCreatesPendingReservationWithNormalizedFields() {
        when(roomMapper.selectOne(any())).thenReturn(null);
        when(roomMapper.insert(any(MeetingRoom.class))).thenReturn(1);
        LocalDateTime start = LocalDateTime.now().plusDays(1).withHour(10).withMinute(0).withSecond(0).withNano(0);
        LocalDateTime end = start.plusHours(1);

        MeetingRoom room = meetingService.reserve(" 周会 ", 7L, " 主持人 ", start, end, "  评审  ");

        assertEquals("周会", room.getRoomName());
        assertEquals("主持人", room.getCreatorName());
        assertEquals(start, room.getScheduledStart());
        assertEquals(end, room.getScheduledEnd());
        assertEquals("评审", room.getScheduledReason());
        assertEquals(LiveConstants.MeetingStatus.PENDING_APPROVAL, room.getStatus());
        verify(roomMapper).insert(room);
        verify(roomMapper).updateById(room);
    }

    @Test
    void reserveRejectsInvalidReservationWindowBeforePersisting() {
        LocalDateTime start = LocalDateTime.now().plusDays(1).withSecond(0).withNano(0);
        LocalDateTime shortEnd = start.plusMinutes(30);
        LocalDateTime longEnd = start.plusHours(9);
        LocalDateTime pastStart = LocalDateTime.now().minusHours(1);

        assertReserveRejected(start, shortEnd, "单次预约最少1小时");
        assertReserveRejected(start, longEnd, "单次预约最长8小时");
        assertReserveRejected(pastStart, pastStart.plusHours(1), "预约开始时间不能早于现在");
        verify(roomMapper, never()).insert(any(MeetingRoom.class));
    }

    @Test
    void approveReservationOnlyAllowsPendingScheduledRoom() {
        MeetingRoom notScheduled = room(1L, LiveConstants.MeetingStatus.PENDING_APPROVAL);
        when(roomMapper.selectById(1L)).thenReturn(notScheduled);

        assertFalse(meetingService.approveReservation(1L));
        verify(roomMapper, never()).updateById(any(MeetingRoom.class));

        reset(roomMapper);
        MeetingRoom pending = room(2L, LiveConstants.MeetingStatus.PENDING_APPROVAL);
        pending.setScheduledStart(LocalDateTime.now().plusHours(1));
        when(roomMapper.selectById(2L)).thenReturn(pending);

        assertTrue(meetingService.approveReservation(2L));
        assertEquals(LiveConstants.MeetingStatus.NOT_STARTED, pending.getStatus());
        assertNotNull(pending.getUpdateTime());
        verify(roomMapper).updateById(pending);
    }

    @Test
    void rejectReservationOnlyAllowsPendingRooms() {
        MeetingRoom rejected = room(3L, LiveConstants.MeetingStatus.PENDING_APPROVAL);
        rejected.setScheduledStart(LocalDateTime.now().plusHours(1));
        when(roomMapper.selectById(3L)).thenReturn(rejected);

        assertTrue(meetingService.rejectReservation(3L));
        assertEquals(LiveConstants.MeetingStatus.REJECTED, rejected.getStatus());
        assertNotNull(rejected.getEndTime());
        assertNotNull(rejected.getUpdateTime());
        verify(roomMapper).updateById(rejected);

        reset(roomMapper);
        MeetingRoom nonPending = room(4L, LiveConstants.MeetingStatus.NOT_STARTED);
        when(roomMapper.selectById(4L)).thenReturn(nonPending);

        assertFalse(meetingService.rejectReservation(4L));
        verify(roomMapper, never()).updateById(any(MeetingRoom.class));
    }

    @Test
    void joinRoomStartsMeetingAndCreatesHostParticipant() {
        MeetingRoom room = room(1L, LiveConstants.MeetingStatus.NOT_STARTED);
        room.setCreatorId(7L);
        room.setMaxParticipants(5);
        when(roomMapper.selectById(1L)).thenReturn(room);
        when(participantMapper.selectOne(any())).thenReturn(null);
        when(participantMapper.selectCount(any())).thenReturn(0L);
        when(participantMapper.insert(any(MeetingParticipant.class))).thenReturn(1);

        MeetingParticipant participant = meetingService.joinRoom(1L, 7L, " 主持人 ");

        assertEquals(7L, participant.getUserId());
        assertEquals("主持人", participant.getUserName());
        assertEquals(LiveConstants.ParticipantRole.HOST, participant.getRole());
        assertEquals(LiveConstants.SwitchState.OFF, participant.getAudioEnabled());
        assertEquals(LiveConstants.SwitchState.OFF, participant.getVideoEnabled());
        assertEquals(LiveConstants.SwitchState.OFF, participant.getScreenShareEnabled());
        assertNotNull(participant.getJoinTime());
        assertEquals(LiveConstants.MeetingStatus.IN_PROGRESS, room.getStatus());
        assertNotNull(room.getStartTime());
        verify(participantMapper).insert(participant);
        verify(roomMapper).updateById(room);
    }

    @Test
    void joinRoomReturnsExistingActiveParticipantWithoutChangingCapacity() {
        MeetingRoom room = room(1L, LiveConstants.MeetingStatus.IN_PROGRESS);
        MeetingParticipant existing = new MeetingParticipant();
        existing.setId(3L);
        existing.setRoomId(1L);
        existing.setUserId(8L);
        when(roomMapper.selectById(1L)).thenReturn(room);
        when(participantMapper.selectOne(any())).thenReturn(existing);

        MeetingParticipant participant = meetingService.joinRoom(1L, 8L, "观众");

        assertSame(existing, participant);
        verify(participantMapper, never()).selectCount(any());
        verify(participantMapper, never()).insert(any(MeetingParticipant.class));
        verify(roomMapper, never()).updateById(any(MeetingRoom.class));
    }

    @Test
    void joinRoomBlocksReviewTerminalAndInvalidStatuses() {
        assertJoinBlocked(LiveConstants.MeetingStatus.PENDING_APPROVAL, "会议预约待审批");
        assertJoinBlocked(LiveConstants.MeetingStatus.REJECTED, "会议预约已被拒绝");
        assertJoinBlocked(LiveConstants.MeetingStatus.ENDED, "会议已结束");
        assertJoinBlocked(99, "会议状态无效");
    }

    @Test
    void leaveRoomEndsMeetingWhenLastActiveParticipantLeaves() {
        MeetingParticipant existing = new MeetingParticipant();
        existing.setId(4L);
        existing.setRoomId(1L);
        existing.setUserId(8L);
        MeetingRoom room = room(1L, LiveConstants.MeetingStatus.IN_PROGRESS);
        when(participantMapper.selectOne(any())).thenReturn(existing);
        when(participantMapper.updateById(any(MeetingParticipant.class))).thenReturn(1);
        when(participantMapper.selectCount(any())).thenReturn(0L);
        when(roomMapper.selectById(1L)).thenReturn(room);

        meetingService.leaveRoom(1L, 8L);

        assertNotNull(existing.getLeaveTime());
        assertEquals(LiveConstants.MeetingStatus.ENDED, room.getStatus());
        assertNotNull(room.getEndTime());
        verify(participantMapper).updateById(existing);
        verify(roomMapper).updateById(room);
    }

    @Test
    void leaveAllReportsMissingRoomsAndEndsExistingRooms() {
        when(roomMapper.selectById(1L)).thenReturn(null);

        assertFalse(meetingService.leaveAll(1L));
        verify(participantMapper, never()).update(any(), any());

        reset(roomMapper, participantMapper);
        MeetingRoom room = room(2L, LiveConstants.MeetingStatus.IN_PROGRESS);
        when(roomMapper.selectById(2L)).thenReturn(room);

        assertTrue(meetingService.leaveAll(2L));

        assertEquals(LiveConstants.MeetingStatus.ENDED, room.getStatus());
        assertNotNull(room.getEndTime());
        verify(participantMapper).update(any(), any());
        verify(roomMapper).updateById(room);
    }

    private void assertJoinBlocked(Integer status, String message) {
        reset(roomMapper, participantMapper);
        MeetingRoom room = room(1L, status);
        when(roomMapper.selectById(1L)).thenReturn(room);

        BusinessException ex = assertThrows(BusinessException.class,
                () -> meetingService.joinRoom(1L, 8L, "观众"));

        assertEquals(message, ex.getMessage());
        verify(participantMapper, never()).insert(any(MeetingParticipant.class));
    }

    private void assertReserveRejected(LocalDateTime start, LocalDateTime end, String message) {
        reset(roomMapper);

        BusinessException ex = assertThrows(BusinessException.class,
                () -> meetingService.reserve("周会", 7L, "主持人", start, end, "评审"));

        assertEquals(message, ex.getMessage());
        verify(roomMapper, never()).selectOne(any());
        verify(roomMapper, never()).insert(any(MeetingRoom.class));
        verify(roomMapper, never()).updateById(any(MeetingRoom.class));
    }

    private MeetingRoom room(Long id, Integer status) {
        MeetingRoom room = new MeetingRoom();
        room.setId(id);
        room.setRoomName("会议室");
        room.setRoomCode("123456");
        room.setCreatorId(7L);
        room.setCreatorName("主持人");
        room.setMaxParticipants(5);
        room.setStatus(status);
        return room;
    }
}
