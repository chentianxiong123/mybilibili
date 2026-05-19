package com.mybilibili.live.controller;

import com.mybilibili.live.entity.LiveRoom;
import com.mybilibili.live.service.LiveRoomService;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/live/room")
@Tag(name = "直播管理")
public class LiveRoomController {

    @Autowired
    private LiveRoomService liveRoomService;

    @PostMapping("/create")
    @Operation(summary = "创建直播间")
    public Result<LiveRoom> createRoom(@RequestBody Map<String, String> request,
                                        HttpServletRequest httpRequest) {
        Integer userId = getUserId(httpRequest);
        if (userId == null) return Result.error("未登录");
        String roomName = request.get("roomName");
        LiveRoom existing = liveRoomService.getByUserId(userId);
        if (existing != null) {
            return Result.success(existing);
        }
        return Result.success(liveRoomService.createRoom(userId, roomName));
    }

    @GetMapping("/my")
    @Operation(summary = "获取我的直播间")
    public Result<LiveRoom> getMyRoom(HttpServletRequest httpRequest) {
        Integer userId = getUserId(httpRequest);
        if (userId == null) return Result.error("未登录");
        LiveRoom room = liveRoomService.getByUserId(userId);
        return Result.success(room);
    }

    @GetMapping("/list")
    @Operation(summary = "获取直播列表")
    public Result<List<LiveRoom>> getLiveList() {
        return Result.success(liveRoomService.getLiveList());
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取直播间详情")
    public Result<LiveRoom> getRoom(@PathVariable Integer id) {
        LiveRoom room = liveRoomService.getById(id);
        if (room == null) return Result.error("直播间不存在");
        return Result.success(room);
    }

    @PutMapping("/{id}/status")
    @Operation(summary = "更新直播间状态 (offline/live)", hidden = true)
    public Result<?> updateStatus(@PathVariable Integer id, @RequestBody Map<String, String> request) {
        String status = request.get("status");
        liveRoomService.updateStatus(id, status);
        return Result.success("状态已更新");
    }

    private Integer getUserId(HttpServletRequest request) {
        String uid = request.getHeader("X-User-Id");
        if (uid != null) {
            try { return Integer.parseInt(uid); } catch (NumberFormatException ignored) {}
        }
        return null;
    }
}