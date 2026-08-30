package com.mybilibili.live.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.live.service.LiveRoomService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/admin/live")
public class LiveAdminController {

    @Autowired
    private LiveRoomService liveRoomService;

    @GetMapping("/rooms")
    public Result<?> getRooms(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "10") int size,
            @RequestParam(required = false) String status) {
        return Result.success(liveRoomService.listAll(page, size, status));
    }

    @GetMapping("/rooms/{id}")
    public Result<?> getRoom(@PathVariable Integer id) {
        return Result.success(liveRoomService.getById(id));
    }

    @PutMapping("/rooms/{id}/status")
    public Result<?> updateStatus(@PathVariable Integer id, @RequestBody Map<String, String> body) {
        liveRoomService.updateStatus(id, body.get("status"));
        return Result.success("状态更新成功");
    }

    @GetMapping("/stats")
    public Result<?> getStats() {
        return Result.success(liveRoomService.getAdminStats());
    }
}