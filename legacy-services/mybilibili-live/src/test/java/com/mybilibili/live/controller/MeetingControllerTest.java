package com.mybilibili.live.controller;

import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.vo.Result;
import com.mybilibili.live.common.AuthUser;
import com.mybilibili.live.common.RequestUserResolver;
import com.mybilibili.live.dto.ReserveMeetingRequest;
import com.mybilibili.live.entity.MeetingRoom;
import com.mybilibili.live.service.MeetingService;
import jakarta.servlet.http.HttpServletRequest;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.time.LocalDateTime;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertSame;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class MeetingControllerTest {

    @Mock
    private MeetingService meetingService;

    @Mock
    private RequestUserResolver userResolver;

    @Mock
    private HttpServletRequest request;

    @InjectMocks
    private MeetingController controller;

    @Test
    void reserveMeetingParsesRequestWindowAndDelegatesToService() {
        ReserveMeetingRequest body = reserveRequest("2026-05-31T10:00:00", "2026-05-31T11:00:00");
        MeetingRoom room = new MeetingRoom();
        when(userResolver.requireUser(request)).thenReturn(new AuthUser(7L, "主持人"));
        when(meetingService.reserve(eq("排期会"), eq(7L), eq("主持人"), any(), any(), eq("同步排期")))
                .thenReturn(room);

        Result<?> result = controller.reserveMeeting(body, request);

        ArgumentCaptor<LocalDateTime> start = ArgumentCaptor.forClass(LocalDateTime.class);
        ArgumentCaptor<LocalDateTime> end = ArgumentCaptor.forClass(LocalDateTime.class);
        verify(meetingService).reserve(eq("排期会"), eq(7L), eq("主持人"),
                start.capture(), end.capture(), eq("同步排期"));
        assertEquals(LocalDateTime.of(2026, 5, 31, 10, 0), start.getValue());
        assertEquals(LocalDateTime.of(2026, 5, 31, 11, 0), end.getValue());
        assertEquals(200, result.getCode());
        assertEquals("预约提交成功，等待管理员审批", result.getMessage());
        assertSame(room, result.getData());
    }

    @Test
    void reserveMeetingRejectsNullBodyBeforeServiceCall() {
        when(userResolver.requireUser(request)).thenReturn(new AuthUser(7L, "主持人"));

        BusinessException ex = assertThrows(BusinessException.class,
                () -> controller.reserveMeeting(null, request));

        assertEquals("预约参数不能为空", ex.getMessage());
        verify(meetingService, never()).reserve(any(), any(), any(), any(), any(), any());
    }

    @Test
    void reserveMeetingRejectsMissingTimeRangeBeforeServiceCall() {
        ReserveMeetingRequest body = reserveRequest(" ", "2026-05-31T11:00:00");
        when(userResolver.requireUser(request)).thenReturn(new AuthUser(7L, "主持人"));

        BusinessException ex = assertThrows(BusinessException.class,
                () -> controller.reserveMeeting(body, request));

        assertEquals("请选择预约时间段", ex.getMessage());
        verify(meetingService, never()).reserve(any(), any(), any(), any(), any(), any());
    }

    @Test
    void reserveMeetingRejectsInvalidDateTimeBeforeServiceCall() {
        ReserveMeetingRequest body = reserveRequest("2026/05/31 10:00", "2026-05-31T11:00:00");
        when(userResolver.requireUser(request)).thenReturn(new AuthUser(7L, "主持人"));

        BusinessException ex = assertThrows(BusinessException.class,
                () -> controller.reserveMeeting(body, request));

        assertEquals("预约时间格式无效", ex.getMessage());
        verify(meetingService, never()).reserve(any(), any(), any(), any(), any(), any());
    }

    private ReserveMeetingRequest reserveRequest(String start, String end) {
        ReserveMeetingRequest request = new ReserveMeetingRequest();
        request.setRoomName("排期会");
        request.setScheduledStart(start);
        request.setScheduledEnd(end);
        request.setReason("同步排期");
        return request;
    }
}
