package com.mybilibili.live.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.live.common.RequestUserResolver;
import com.mybilibili.live.service.MeetingService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import jakarta.servlet.http.HttpServletRequest;

@RestController
@RequestMapping("/admin/meeting")
public class MeetingAdminController {

    @Autowired
    private MeetingService meetingService;

    @Autowired
    private RequestUserResolver userResolver;

    @GetMapping("/rooms")
    public Result<?> listRooms(
            HttpServletRequest request,
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "10") int size,
            @RequestParam(required = false) Integer status) {
        userResolver.requireAdmin(request);
        return Result.success(meetingService.getAllRooms(page, size, status));
    }

    @GetMapping("/pending")
    public Result<?> getPendingReservations(HttpServletRequest request) {
        userResolver.requireAdmin(request);
        return Result.success(meetingService.getPendingReservations());
    }

    @PostMapping("/approve/{roomId}")
    public Result<?> approve(@PathVariable Long roomId, HttpServletRequest request) {
        userResolver.requireAdmin(request);
        return meetingService.approveReservation(roomId)
                ? Result.success("审批通过") : Result.error("审批失败");
    }

    @PostMapping("/reject/{roomId}")
    public Result<?> reject(@PathVariable Long roomId, HttpServletRequest request) {
        userResolver.requireAdmin(request);
        return meetingService.rejectReservation(roomId)
                ? Result.success("已拒绝") : Result.error("操作失败");
    }

    @PostMapping("/end/{roomId}")
    public Result<?> forceEnd(@PathVariable Long roomId, HttpServletRequest request) {
        userResolver.requireAdmin(request);
        return meetingService.leaveAll(roomId)
                ? Result.success("已强制结束会议") : Result.error("会议不存在");
    }
}
