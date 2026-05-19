package com.mybilibili.live.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.live.entity.MeetingRoom;
import com.mybilibili.live.entity.MeetingParticipant;
import com.mybilibili.live.service.MeetingService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

@Slf4j
@RestController
@RequestMapping("/meeting")
public class MeetingController {

    @Autowired
    private MeetingService meetingService;

    @PostMapping("/create")
    public Result<?> createRoom(@RequestParam String roomName, HttpServletRequest request) {
        Long userId = Long.parseLong(request.getHeader("X-User-Id"));
        String userName = request.getHeader("X-Username");
        MeetingRoom room = meetingService.createRoom(roomName, userId, userName);
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
        Long userId = Long.parseLong(request.getHeader("X-User-Id"));
        List<MeetingRoom> rooms = meetingService.getMyRooms(userId);
        return Result.success(rooms);
    }

    @PostMapping("/join/{roomCode}")
    public Result<?> joinRoom(@PathVariable String roomCode, HttpServletRequest request) {
        Long userId = Long.parseLong(request.getHeader("X-User-Id"));
        String userName = request.getHeader("X-Username");
        MeetingRoom room = meetingService.getRoomByCode(roomCode);
        if (room == null) {
            return Result.error("会议室不存在");
        }
        MeetingParticipant participant = meetingService.joinRoom(room.getId(), userId, userName);
        return Result.success(participant);
    }

    @PostMapping("/leave/{roomId}")
    public Result<?> leaveRoom(@PathVariable Long roomId, HttpServletRequest request) {
        Long userId = Long.parseLong(request.getHeader("X-User-Id"));
        meetingService.leaveRoom(roomId, userId);
        return Result.success(null);
    }

    @GetMapping("/participants/{roomId}")
    public Result<?> getParticipants(@PathVariable Long roomId) {
        List<MeetingParticipant> participants = meetingService.getParticipants(roomId);
        return Result.success(participants);
    }
}