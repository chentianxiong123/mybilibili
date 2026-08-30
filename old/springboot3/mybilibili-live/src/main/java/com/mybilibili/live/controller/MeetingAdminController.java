package com.mybilibili.live.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.live.service.MeetingService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import jakarta.servlet.http.HttpServletRequest;

@RestController
@RequestMapping("/admin/meeting")
public class MeetingAdminController {

    @Autowired
    private MeetingService meetingService;

    @GetMapping("/rooms")
    public Result<?> listRooms(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "10") int size,
            @RequestParam(required = false) Integer status) {
        return Result.success(meetingService.getAllRooms(page, size, status));
    }

    @GetMapping("/pending")
    public Result<?> getPendingReservations() {
        return Result.success(meetingService.getPendingReservations());
    }

    @PostMapping("/approve/{roomId}")
    public Result<?> approve(@PathVariable Long roomId, HttpServletRequest request) {
        Long adminId = Long.parseLong(request.getHeader("X-User-Id"));
        return meetingService.approveReservation(roomId, adminId)
                ? Result.success("审批通过") : Result.error("审批失败");
    }

    @PostMapping("/reject/{roomId}")
    public Result<?> reject(@PathVariable Long roomId, HttpServletRequest request) {
        Long adminId = Long.parseLong(request.getHeader("X-User-Id"));
        return meetingService.rejectReservation(roomId, adminId)
                ? Result.success("已拒绝") : Result.error("操作失败");
    }

    @PostMapping("/end/{roomId}")
    public Result<?> forceEnd(@PathVariable Long roomId) {
        meetingService.leaveAll(roomId);
        return Result.success("已强制结束会议");
    }
}