package com.mybilibili.live.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.live.common.RequestUserResolver;
import com.mybilibili.live.dto.UpdateLiveRoomStatusRequest;
import com.mybilibili.live.entity.LiveRoom;
import com.mybilibili.live.service.LiveRoomService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;

@RestController
@RequestMapping("/admin/live")
public class LiveAdminController {

    @Autowired
    private LiveRoomService liveRoomService;

    @Autowired
    private RequestUserResolver userResolver;

    @GetMapping("/rooms")
    public Result<?> getRooms(
            HttpServletRequest request,
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "10") int size,
            @RequestParam(required = false) String status) {
        userResolver.requireAdmin(request);
        return Result.success(liveRoomService.listAll(page, size, status));
    }

    @GetMapping("/rooms/{id}")
    public Result<?> getRoom(@PathVariable Integer id, HttpServletRequest request) {
        userResolver.requireAdmin(request);
        LiveRoom room = liveRoomService.getById(id);
        return room == null ? Result.error("直播间不存在") : Result.success(room);
    }

    @PutMapping("/rooms/{id}/status")
    public Result<?> updateStatus(@PathVariable Integer id,
                                  @RequestBody UpdateLiveRoomStatusRequest body,
                                  HttpServletRequest request) {
        userResolver.requireAdmin(request);
        liveRoomService.updateStatus(id, body == null ? null : body.getStatus());
        return Result.success("状态更新成功");
    }

    @GetMapping("/stats")
    public Result<?> getStats(HttpServletRequest request) {
        userResolver.requireAdmin(request);
        return Result.success(liveRoomService.getAdminStats());
    }
}
