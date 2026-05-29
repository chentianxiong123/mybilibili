package com.mybilibili.live.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.live.entity.MeetingRoom;
import com.mybilibili.live.entity.MeetingParticipant;
import com.mybilibili.live.service.MeetingService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Map;

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

    @PostMapping("/end/{roomId}")
    public Result<?> endRoom(@PathVariable Long roomId, HttpServletRequest request) {
        Long userId = Long.parseLong(request.getHeader("X-User-Id"));
        boolean ok = meetingService.endRoom(roomId, userId);
        if (!ok) return Result.error("无权限结束会议或会议不存在");
        return Result.success("会议已结束");
    }

    @PostMapping("/reserve")
    public Result<?> reserveMeeting(@RequestBody Map<String, Object> body, HttpServletRequest request) {
        Long userId = Long.parseLong(request.getHeader("X-User-Id"));
        String userName = request.getHeader("X-Username");
        String roomName = (String) body.get("roomName");
        String startStr = (String) body.get("scheduledStart");
        String endStr = (String) body.get("scheduledEnd");
        String reason = (String) body.get("reason");
        if (roomName == null || roomName.isBlank()) return Result.error("会议室名称不能为空");
        if (startStr == null || endStr == null) return Result.error("请选择预约时间段");
        LocalDateTime start = LocalDateTime.parse(startStr);
        LocalDateTime end = LocalDateTime.parse(endStr);
        if (end.isBefore(start)) return Result.error("结束时间不能早于开始时间");
        if (start.isBefore(LocalDateTime.now())) return Result.error("预约开始时间不能早于现在");
        long hours = java.time.Duration.between(start, end).toHours();
        if (hours > 8) return Result.error("单次预约最长8小时");
        if (hours < 1) return Result.error("单次预约最少1小时");
        try {
            MeetingRoom room = meetingService.reserve(roomName, userId, userName, start, end, reason);
            return Result.success("预约提交成功，等待管理员审批", room);
        } catch (Exception e) {
            return Result.error("预约失败: " + e.getMessage());
        }
    }
}