package com.mybilibili.live.controller;

import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.vo.Result;
import com.mybilibili.live.common.AuthUser;
import com.mybilibili.live.common.LiveConstants;
import com.mybilibili.live.common.RequestUserResolver;
import com.mybilibili.live.dto.ScheduleLiveRoomRequest;
import com.mybilibili.live.dto.SrsHookRequest;
import com.mybilibili.live.entity.LiveRoom;
import com.mybilibili.live.service.LiveRoomService;
import jakarta.servlet.http.HttpServletRequest;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Date;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.ArgumentMatchers.isNull;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class LiveRoomControllerTest {

    @Mock
    private LiveRoomService liveRoomService;

    @Mock
    private RequestUserResolver userResolver;

    @Mock
    private HttpServletRequest request;

    @InjectMocks
    private LiveRoomController controller;

    @Test
    void scheduleRoomConvertsTimestampRequestToDate() {
        long timestamp = System.currentTimeMillis() + 60_000;
        ScheduleLiveRoomRequest body = new ScheduleLiveRoomRequest();
        body.setScheduledAt(timestamp);
        givenOwnedRoom(7, 42);

        Result<?> result = controller.scheduleRoom(7, body, request);

        ArgumentCaptor<Date> scheduledAt = ArgumentCaptor.forClass(Date.class);
        verify(liveRoomService).scheduleRoom(eq(7), scheduledAt.capture());
        assertEquals(timestamp, scheduledAt.getValue().getTime());
        assertEquals(200, result.getCode());
        assertEquals("定时开播已设置", result.getData());
    }

    @Test
    void scheduleRoomAllowsNullRequestToCancelSchedule() {
        givenOwnedRoom(7, 42);

        Result<?> result = controller.scheduleRoom(7, null, request);

        verify(liveRoomService).scheduleRoom(eq(7), isNull());
        assertEquals(200, result.getCode());
        assertEquals("定时开播已取消", result.getData());
    }

    @Test
    void scheduleRoomRejectsNonOwnerBeforePersisting() {
        when(userResolver.requireUser(request)).thenReturn(new AuthUser(99L, "viewer"));
        when(liveRoomService.getById(7)).thenReturn(liveRoom(7, 42));

        BusinessException ex = assertThrows(BusinessException.class,
                () -> controller.scheduleRoom(7, null, request));

        assertEquals(403, ex.getCode());
        assertEquals("无权操作", ex.getMessage());
    }

    @Test
    void srsHookMapsPublishAndUnpublishActionsToRoomStatus() {
        when(liveRoomService.getByStreamKey("stream-1")).thenReturn(liveRoom(7, 42));

        Map<String, Object> publishResponse = controller.srsHook(srsHook("on_publish", "stream-1"));
        Map<String, Object> unpublishResponse = controller.srsHook(srsHook("on_unpublish", "stream-1"));

        assertEquals(0, publishResponse.get("code"));
        assertEquals(0, unpublishResponse.get("code"));
        verify(liveRoomService).updateStatus(7, LiveConstants.RoomStatus.LIVE);
        verify(liveRoomService).updateStatus(7, LiveConstants.RoomStatus.OFFLINE);
    }

    @Test
    void srsHookIgnoresUnsupportedActionsBeforeRoomLookup() {
        Map<String, Object> response = controller.srsHook(srsHook("on_play", "stream-1"));

        assertEquals(0, response.get("code"));
        verify(liveRoomService, never()).getByStreamKey(any());
        verify(liveRoomService, never()).updateStatus(any(), any());
    }

    @Test
    void srsHookAcknowledgesMissingRoomsWithoutStatusUpdate() {
        when(liveRoomService.getByStreamKey("missing-stream")).thenReturn(null);

        Map<String, Object> response = controller.srsHook(srsHook("on_publish", "missing-stream"));

        assertEquals(0, response.get("code"));
        verify(liveRoomService, never()).updateStatus(any(), any());
    }

    private void givenOwnedRoom(Integer roomId, Integer userId) {
        when(userResolver.requireUser(request)).thenReturn(new AuthUser(userId.longValue(), "owner"));
        when(liveRoomService.getById(roomId)).thenReturn(liveRoom(roomId, userId));
    }

    private LiveRoom liveRoom(Integer roomId, Integer userId) {
        LiveRoom room = new LiveRoom();
        room.setId(roomId);
        room.setUserId(userId);
        return room;
    }

    private SrsHookRequest srsHook(String action, String stream) {
        SrsHookRequest request = new SrsHookRequest();
        request.setAction(action);
        request.setStream(stream);
        return request;
    }
}
