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

    @PutMapping("/{id}")
    @Operation(summary = "更新直播间（封面/名称/分类）")
    public Result<?> updateRoom(@PathVariable Integer id, @RequestBody Map<String, String> request,
                                HttpServletRequest httpRequest) {
        Integer userId = getUserId(httpRequest);
        if (userId == null) return Result.error("未登录");
        LiveRoom room = liveRoomService.getById(id);
        if (room == null) return Result.error("直播间不存在");
        if (!room.getUserId().equals(userId)) return Result.error("无权操作");
        if (request.containsKey("coverUrl")) room.setCoverUrl(request.get("coverUrl"));
        if (request.containsKey("roomName")) room.setRoomName(request.get("roomName"));
        if (request.containsKey("category")) room.setCategory(request.get("category"));
        liveRoomService.updateRoom(room);
        return Result.success(room);
    }

    @PutMapping("/{id}/schedule")
    @Operation(summary = "设置定时开播时间")
    public Result<?> scheduleRoom(@PathVariable Integer id, @RequestBody Map<String, Object> request,
                                  HttpServletRequest httpRequest) {
        Integer userId = getUserId(httpRequest);
        if (userId == null) return Result.error("未登录");
        LiveRoom room = liveRoomService.getById(id);
        if (room == null) return Result.error("直播间不存在");
        if (!room.getUserId().equals(userId)) return Result.error("无权操作");
        Object scheduledAtObj = request.get("scheduledAt");
        if (scheduledAtObj == null) {
            room.setScheduledAt(null);
        } else {
            long ts = Long.parseLong(String.valueOf(scheduledAtObj));
            room.setScheduledAt(new java.util.Date(ts));
        }
        liveRoomService.updateRoom(room);
        return Result.success(room);
    }

    /**
     * SRS HTTP 回调，配置示例 (srs.conf):
     *   http_hooks {
     *       enabled on;
     *       on_publish http://host:port/live/room/srs/hook;
     *       on_unpublish http://host:port/live/room/srs/hook;
     *   }
     * 回调请求体由 SRS 发出，包含 action / stream / app / vhost 等字段。
     * 这里按 stream（=streamKey）匹配房间，更新 status。
     * 必须返回 {code:0} 给 SRS，否则 SRS 会拒绝推流。
     */
    @PostMapping("/srs/hook")
    @Operation(summary = "SRS 推流/断流回调", hidden = true)
    public Map<String, Object> srsHook(@RequestBody Map<String, Object> body) {
        Object action = body.get("action");
        Object stream = body.get("stream");
        if (stream != null) {
            LiveRoom room = liveRoomService.getByStreamKey(String.valueOf(stream));
            if (room != null) {
                if ("on_publish".equals(String.valueOf(action))) {
                    liveRoomService.updateStatus(room.getId(), "live");
                } else if ("on_unpublish".equals(String.valueOf(action))) {
                    liveRoomService.updateStatus(room.getId(), "offline");
                }
            }
        }
        java.util.Map<String, Object> resp = new java.util.HashMap<>();
        resp.put("code", 0);
        return resp;
    }

    private Integer getUserId(HttpServletRequest request) {
        String uid = request.getHeader("X-User-Id");
        if (uid != null) {
            try { return Integer.parseInt(uid); } catch (NumberFormatException ignored) {}
        }
        return null;
    }
}