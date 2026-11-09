package com.mybilibili.live.controller;

import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.vo.Result;
import com.mybilibili.live.common.AuthUser;
import com.mybilibili.live.common.RequestUserResolver;
import com.mybilibili.live.dto.ReserveMeetingRequest;
import com.mybilibili.live.entity.MeetingRoom;
import com.mybilibili.live.entity.MeetingParticipant;
import com.mybilibili.live.service.MeetingService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.time.LocalDateTime;
import java.util.List;

@Slf4j
@RestController
@RequestMapping("/meeting")
public class MeetingController {

    @Autowired
    private MeetingService meetingService;

    @Autowired
    private RequestUserResolver userResolver;

    @PostMapping("/create")
    public Result<?> createRoom(@RequestParam String roomName, HttpServletRequest request) {
        AuthUser user = userResolver.requireUser(request);
        MeetingRoom room = meetingService.createRoom(roomName, user.id(), user.username());
        return Result.success(room);
    }

    @GetMapping("/room/{roomCode}")
    public Result<?> getRoom(@PathVariable String roomCode) {
        MeetingRoom room = meetingService.getRoomByCode(roomCode);
        if (room == null) {
            return Result.error("会议室不存在");
        }
        return Result.success(room);
    }

    @GetMapping("/my-rooms")
    public Result<?> getMyRooms(HttpServletRequest request) {
        AuthUser user = userResolver.requireUser(request);
        List<MeetingRoom> rooms = meetingService.getMyRooms(user.id());
        return Result.success(rooms);
    }

    @PostMapping("/join/{roomCode}")
    public Result<?> joinRoom(@PathVariable String roomCode, HttpServletRequest request) {
        AuthUser user = userResolver.requireUser(request);
        MeetingRoom room = meetingService.getRoomByCode(roomCode);
        if (room == null) {
            return Result.error("会议室不存在");
        }
        MeetingParticipant participant = meetingService.joinRoom(room.getId(), user.id(), user.username());
        return Result.success(participant);
    }

    @PostMapping("/leave/{roomId}")
    public Result<?> leaveRoom(@PathVariable Long roomId, HttpServletRequest request) {
        AuthUser user = userResolver.requireUser(request);
        meetingService.leaveRoom(roomId, user.id());
        return Result.success(null);
    }

    @GetMapping("/participants/{roomId}")
    public Result<?> getParticipants(@PathVariable Long roomId) {
        List<MeetingParticipant> participants = meetingService.getParticipants(roomId);
        return Result.success(participants);
    }

    @PostMapping("/end/{roomId}")
    public Result<?> endRoom(@PathVariable Long roomId, HttpServletRequest request) {
        AuthUser user = userResolver.requireUser(request);
        boolean ok = meetingService.endRoom(roomId, user.id());
        if (!ok) return Result.error("无权限结束会议或会议不存在");
        return Result.success("会议已结束");
    }

    @PostMapping("/reserve")
    public Result<?> reserveMeeting(@RequestBody(required = false) ReserveMeetingRequest body, HttpServletRequest request) {
        AuthUser user = userResolver.requireUser(request);
        if (body == null) throw new BusinessException("预约参数不能为空");
        LocalDateTime start = body.requireScheduledStart();
        LocalDateTime end = body.requireScheduledEnd();

        MeetingRoom room = meetingService.reserve(body.getRoomName(), user.id(), user.username(), start, end, body.getReason());
        return Result.success("预约提交成功，等待管理员审批", room);
    }
}
