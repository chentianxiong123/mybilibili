package com.mybilibili.live.service.impl;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.live.common.LiveConstants;
import com.mybilibili.live.entity.LiveRoom;
import com.mybilibili.live.mapper.LiveRoomMapper;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Date;
import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.isNull;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.reset;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@SuppressWarnings("unchecked")
class LiveRoomServiceImplTest {

    @Mock
    private LiveRoomMapper liveRoomMapper;

    @InjectMocks
    private LiveRoomServiceImpl liveRoomService;

    @Test
    void createRoomNormalizesDefaultsAndGeneratesStreamKey() {
        when(liveRoomMapper.insert(any(LiveRoom.class))).thenReturn(1);

        LiveRoom room = liveRoomService.createRoom(7, "  我的频道  ");

        assertEquals(7, room.getUserId());
        assertEquals("我的频道", room.getRoomName());
        assertEquals(LiveConstants.RoomStatus.OFFLINE, room.getStatus());
        assertEquals(0, room.getViewerCount());
        assertNotNull(room.getStreamKey());
        assertEquals(32, room.getStreamKey().length());
        assertFalse(room.getStreamKey().contains("-"));
        verify(liveRoomMapper).insert(room);
    }

    @Test
    void updateStatusValidatesRoomAndResetsViewerCount() {
        LiveRoom existing = room(1, LiveConstants.RoomStatus.OFFLINE, 88);
        when(liveRoomMapper.selectById(1)).thenReturn(existing);

        liveRoomService.updateStatus(1, LiveConstants.RoomStatus.LIVE);

        ArgumentCaptor<LiveRoom> captor = ArgumentCaptor.forClass(LiveRoom.class);
        verify(liveRoomMapper).updateById(captor.capture());
        LiveRoom update = captor.getValue();
        assertEquals(1, update.getId());
        assertEquals(LiveConstants.RoomStatus.LIVE, update.getStatus());
        assertEquals(0, update.getViewerCount());
    }

    @Test
    void updateStatusRejectsInvalidOrMissingRoom() {
        BusinessException invalid = assertThrows(BusinessException.class,
                () -> liveRoomService.updateStatus(1, "bad"));
        assertEquals("直播间状态无效", invalid.getMessage());
        verify(liveRoomMapper, never()).selectById(any());

        when(liveRoomMapper.selectById(1)).thenReturn(null);
        BusinessException missing = assertThrows(BusinessException.class,
                () -> liveRoomService.updateStatus(1, LiveConstants.RoomStatus.LIVE));
        assertEquals("直播间不存在", missing.getMessage());
        verify(liveRoomMapper, never()).updateById(any(LiveRoom.class));
    }

    @Test
    void updateViewerCountClampsAtZeroAndIgnoresMissingRoom() {
        LiveRoom room = room(1, LiveConstants.RoomStatus.LIVE, 2);
        when(liveRoomMapper.selectById(1)).thenReturn(room);

        liveRoomService.updateViewerCount(1, -5);

        assertEquals(0, room.getViewerCount());
        verify(liveRoomMapper).updateById(room);

        reset(liveRoomMapper);
        when(liveRoomMapper.selectById(2)).thenReturn(null);
        liveRoomService.updateViewerCount(2, 1);
        verify(liveRoomMapper, never()).updateById(any(LiveRoom.class));
    }

    @Test
    void updateRoomRequiresIdAndNormalizesProvidedFields() {
        BusinessException missing = assertThrows(BusinessException.class,
                () -> liveRoomService.updateRoom(new LiveRoom()));
        assertEquals("直播间不存在", missing.getMessage());

        LiveRoom room = room(3, LiveConstants.RoomStatus.OFFLINE, 0);
        room.setRoomName("  新标题  ");
        room.setCoverUrl("  https://example.com/cover.jpg  ");
        room.setCategory("  学习  ");

        liveRoomService.updateRoom(room);

        assertEquals("新标题", room.getRoomName());
        assertEquals("https://example.com/cover.jpg", room.getCoverUrl());
        assertEquals("学习", room.getCategory());
        verify(liveRoomMapper).updateById(room);
    }

    @Test
    void updateRoomRejectsBlankRoomNameBeforePersisting() {
        LiveRoom room = room(4, LiveConstants.RoomStatus.OFFLINE, 0);
        room.setRoomName("   ");
        room.setCoverUrl("   ");
        room.setCategory("   ");

        BusinessException ex = assertThrows(BusinessException.class,
                () -> liveRoomService.updateRoom(room));

        assertEquals("直播间名称不能为空", ex.getMessage());
        verify(liveRoomMapper, never()).updateById(any(LiveRoom.class));
    }

    @Test
    void updateRoomConvertsBlankOptionalMetadataToNull() {
        LiveRoom room = room(5, LiveConstants.RoomStatus.OFFLINE, 0);
        room.setRoomName("有效标题");
        room.setCoverUrl("   ");
        room.setCategory("   ");

        liveRoomService.updateRoom(room);

        assertNull(room.getCoverUrl());
        assertNull(room.getCategory());
        verify(liveRoomMapper).updateById(room);
    }

    @Test
    void scheduleRoomRejectsMissingOrPastRoomAndAcceptsFutureTime() {
        when(liveRoomMapper.selectById(1)).thenReturn(null);
        BusinessException missing = assertThrows(BusinessException.class,
                () -> liveRoomService.scheduleRoom(1, new Date(System.currentTimeMillis() + 60_000)));
        assertEquals("直播间不存在", missing.getMessage());

        LiveRoom room = room(2, LiveConstants.RoomStatus.OFFLINE, 0);
        when(liveRoomMapper.selectById(2)).thenReturn(room);
        Date past = new Date(System.currentTimeMillis() - 60_000);
        BusinessException pastEx = assertThrows(BusinessException.class,
                () -> liveRoomService.scheduleRoom(2, past));
        assertEquals("定时开播时间必须晚于当前时间", pastEx.getMessage());

        Date future = new Date(System.currentTimeMillis() + 60_000);
        liveRoomService.scheduleRoom(2, future);
        assertEquals(future, room.getScheduledAt());
        verify(liveRoomMapper).updateById(room);
    }

    @Test
    void listAllValidatesStatusAndClampsPagination() {
        BusinessException invalid = assertThrows(BusinessException.class,
                () -> liveRoomService.listAll(1, 20, "bad"));
        assertEquals("直播间状态无效", invalid.getMessage());

        when(liveRoomMapper.selectPage(any(Page.class), any())).thenAnswer(invocation -> invocation.getArgument(0));
        Page<LiveRoom> page = (Page<LiveRoom>) liveRoomService.listAll(0, 500, LiveConstants.RoomStatus.LIVE);

        assertEquals(1, page.getCurrent());
        assertEquals(100, page.getSize());
        verify(liveRoomMapper).selectPage(any(Page.class), any(LambdaQueryWrapper.class));
    }

    @Test
    void getAdminStatsCountsRoomsAndSumsLiveViewers() {
        when(liveRoomMapper.selectCount(isNull())).thenReturn(4L);
        when(liveRoomMapper.selectCount(any(LambdaQueryWrapper.class))).thenReturn(2L);
        when(liveRoomMapper.selectList(any(LambdaQueryWrapper.class))).thenReturn(List.of(
                room(1, LiveConstants.RoomStatus.LIVE, 10),
                room(2, LiveConstants.RoomStatus.LIVE, null),
                room(3, LiveConstants.RoomStatus.LIVE, 7)
        ));

        Map<String, Object> stats = liveRoomService.getAdminStats();

        assertEquals(4L, stats.get("totalRooms"));
        assertEquals(2L, stats.get("onlineRooms"));
        assertEquals(17, stats.get("totalViewers"));
    }

    private LiveRoom room(Integer id, String status, Integer viewerCount) {
        LiveRoom room = new LiveRoom();
        room.setId(id);
        room.setUserId(7);
        room.setRoomName("直播间");
        room.setStreamKey("stream-" + id);
        room.setStatus(status);
        room.setViewerCount(viewerCount);
        return room;
    }
}
