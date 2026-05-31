package com.mybilibili.live.service;

import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.live.common.LiveConstants;
import com.mybilibili.live.entity.LiveLinkmic;
import com.mybilibili.live.entity.LiveRoom;
import com.mybilibili.live.mapper.LiveLinkmicMapper;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.time.LocalDateTime;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class LiveLinkmicServiceTest {

    @Mock
    private LiveLinkmicMapper linkmicMapper;

    @Mock
    private LiveRoomService liveRoomService;

    @InjectMocks
    private LiveLinkmicService liveLinkmicService;

    @Test
    void applyLinkmicCreatesTrimmedPendingRequestForLiveRoom() {
        LiveRoom room = liveRoom(1, 7, LiveConstants.RoomStatus.LIVE);
        when(liveRoomService.getById(1)).thenReturn(room);
        when(linkmicMapper.selectOne(any())).thenReturn(null);

        LiveLinkmic linkmic = liveLinkmicService.applyLinkmic(1L, 8L, "  观众A  ");

        assertEquals(1L, linkmic.getRoomId());
        assertEquals(7L, linkmic.getStreamerId());
        assertEquals(8L, linkmic.getViewerId());
        assertEquals("观众A", linkmic.getViewerName());
        assertEquals(LiveConstants.LinkmicStatus.PENDING, linkmic.getStatus());
        assertEquals(LiveConstants.SwitchState.OFF, linkmic.getAudioEnabled());
        assertEquals(LiveConstants.SwitchState.OFF, linkmic.getVideoEnabled());
        assertNotNull(linkmic.getApplyTime());
        verify(linkmicMapper).insert(linkmic);
    }

    @Test
    void applyLinkmicRejectsNonLiveRoomAndStreamerSelfRequest() {
        LiveRoom offline = liveRoom(1, 7, LiveConstants.RoomStatus.OFFLINE);
        when(liveRoomService.getById(1)).thenReturn(offline);

        BusinessException ex = assertThrows(BusinessException.class,
                () -> liveLinkmicService.applyLinkmic(1L, 8L, "观众"));
        assertEquals("直播未开播，暂不能申请连麦", ex.getMessage());
        verify(linkmicMapper, never()).insert(any(LiveLinkmic.class));

        LiveRoom liveRoom = liveRoom(1, 7, LiveConstants.RoomStatus.LIVE);
        when(liveRoomService.getById(1)).thenReturn(liveRoom);

        BusinessException selfEx = assertThrows(BusinessException.class,
                () -> liveLinkmicService.applyLinkmic(1L, 7L, "主播"));
        assertEquals("主播无需申请连麦", selfEx.getMessage());
    }

    @Test
    void acceptLinkmicOnlyMovesPendingRequestsIntoConnectedState() {
        LiveLinkmic linkmic = linkmic(5L, 7L, 8L, LiveConstants.LinkmicStatus.PENDING);
        when(linkmicMapper.selectById(5L)).thenReturn(linkmic);
        when(linkmicMapper.selectCount(any())).thenReturn(0L);

        liveLinkmicService.acceptLinkmic(5L, 7L);

        assertEquals(LiveConstants.LinkmicStatus.CONNECTED, linkmic.getStatus());
        assertNotNull(linkmic.getConnectTime());
        verify(linkmicMapper).updateById(linkmic);
    }

    @Test
    void acceptLinkmicRejectsWhenConnectedCapacityIsFull() {
        LiveLinkmic linkmic = linkmic(6L, 7L, 8L, LiveConstants.LinkmicStatus.PENDING);
        linkmic.setMaxLinkmics(1);
        when(linkmicMapper.selectById(6L)).thenReturn(linkmic);
        when(linkmicMapper.selectCount(any())).thenReturn(1L);

        BusinessException ex = assertThrows(BusinessException.class,
                () -> liveLinkmicService.acceptLinkmic(6L, 7L));

        assertEquals("当前连麦人数已满", ex.getMessage());
        assertEquals(LiveConstants.LinkmicStatus.PENDING, linkmic.getStatus());
        verify(linkmicMapper, never()).updateById(any(LiveLinkmic.class));
    }

    @Test
    void acceptAndDisconnectRespectStateAndOperatorGuards() {
        LiveLinkmic connected = linkmic(5L, 7L, 8L, LiveConstants.LinkmicStatus.CONNECTED);
        when(linkmicMapper.selectById(5L)).thenReturn(connected);

        BusinessException acceptEx = assertThrows(BusinessException.class,
                () -> liveLinkmicService.acceptLinkmic(5L, 7L));
        assertEquals("该连麦申请状态不可同意", acceptEx.getMessage());

        BusinessException authEx = assertThrows(BusinessException.class,
                () -> liveLinkmicService.disconnectLinkmic(5L, 99L));
        assertEquals(403, authEx.getCode());
        assertEquals("无权操作该连麦", authEx.getMessage());

        liveLinkmicService.disconnectLinkmic(5L, 8L);
        assertEquals(LiveConstants.LinkmicStatus.DISCONNECTED, connected.getStatus());
        assertNotNull(connected.getEndTime());
        verify(linkmicMapper).updateById(connected);
    }

    @Test
    void mediaTogglesRequireConnectedLinkmicAndAuthorizedOperator() {
        LiveLinkmic pending = linkmic(9L, 7L, 8L, LiveConstants.LinkmicStatus.PENDING);
        when(linkmicMapper.selectById(9L)).thenReturn(pending);

        BusinessException stateEx = assertThrows(BusinessException.class,
                () -> liveLinkmicService.toggleAudio(9L, 8L, true));
        assertEquals("连麦未连接，不能切换媒体状态", stateEx.getMessage());
        verify(linkmicMapper, never()).updateById(any(LiveLinkmic.class));

        LiveLinkmic connected = linkmic(10L, 7L, 8L, LiveConstants.LinkmicStatus.CONNECTED);
        when(linkmicMapper.selectById(10L)).thenReturn(connected);

        BusinessException authEx = assertThrows(BusinessException.class,
                () -> liveLinkmicService.toggleVideo(10L, 99L, true));
        assertEquals(403, authEx.getCode());
        assertEquals("无权操作该连麦", authEx.getMessage());

        liveLinkmicService.toggleAudio(10L, 7L, true);
        liveLinkmicService.toggleVideo(10L, 8L, false);

        assertEquals(LiveConstants.SwitchState.ON, connected.getAudioEnabled());
        assertEquals(LiveConstants.SwitchState.OFF, connected.getVideoEnabled());
        verify(linkmicMapper, times(2)).updateById(connected);
    }

    private LiveRoom liveRoom(Integer id, Integer userId, String status) {
        LiveRoom room = new LiveRoom();
        room.setId(id);
        room.setUserId(userId);
        room.setStatus(status);
        return room;
    }

    private LiveLinkmic linkmic(Long id, Long streamerId, Long viewerId, Integer status) {
        LiveLinkmic linkmic = new LiveLinkmic();
        linkmic.setId(id);
        linkmic.setRoomId(1L);
        linkmic.setStreamerId(streamerId);
        linkmic.setViewerId(viewerId);
        linkmic.setStatus(status);
        linkmic.setApplyTime(LocalDateTime.now());
        return linkmic;
    }
}
