package com.mybilibili.live.controller;

import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.vo.Result;
import com.mybilibili.live.common.AuthUser;
import com.mybilibili.live.common.LiveConstants;
import com.mybilibili.live.common.RequestUserResolver;
import com.mybilibili.live.dto.CreateLiveRoomRequest;
import com.mybilibili.live.dto.ScheduleLiveRoomRequest;
import com.mybilibili.live.dto.SrsHookRequest;
import com.mybilibili.live.dto.UpdateLiveRoomRequest;
import com.mybilibili.live.dto.UpdateLiveRoomStatusRequest;
import com.mybilibili.live.entity.LiveRoom;
import com.mybilibili.live.service.LiveRoomService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.util.Date;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/live/room")
@Tag(name = "直播管理")
public class LiveRoomController {

    @Autowired
    private LiveRoomService liveRoomService;

    @Autowired
    private RequestUserResolver userResolver;

    @PostMapping("/create")
    @Operation(summary = "创建直播间")
    public Result<LiveRoom> createRoom(@RequestBody(required = false) CreateLiveRoomRequest request,
                                        HttpServletRequest httpRequest) {
        AuthUser user = userResolver.requireUser(httpRequest);
        Integer userId = user.intId();
        String roomName = request == null ? null : request.getRoomName();
        LiveRoom existing = liveRoomService.getByUserId(userId);
        if (existing != null) {
            return Result.success(existing);
        }
        return Result.success(liveRoomService.createRoom(userId, roomName));
    }

    @GetMapping("/my")
    @Operation(summary = "获取我的直播间")
    public Result<LiveRoom> getMyRoom(HttpServletRequest httpRequest) {
        AuthUser user = userResolver.requireUser(httpRequest);
        LiveRoom room = liveRoomService.getByUserId(user.intId());
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
    public Result<?> updateStatus(@PathVariable Integer id,
                                  @RequestBody UpdateLiveRoomStatusRequest request,
                                  HttpServletRequest httpRequest) {
        AuthUser user = userResolver.requireUser(httpRequest);
        requireOwnedRoom(id, user);
        liveRoomService.updateStatus(id, request == null ? null : request.getStatus());
        return Result.success("状态已更新");
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新直播间（封面/名称/分类）")
    public Result<?> updateRoom(@PathVariable Integer id, @RequestBody UpdateLiveRoomRequest request,
                                HttpServletRequest httpRequest) {
        AuthUser user = userResolver.requireUser(httpRequest);
        LiveRoom room = requireOwnedRoom(id, user);
        if (request != null) {
            if (request.getCoverUrl() != null) {
                room.setCoverUrl(request.getCoverUrl());
            }
            if (request.getRoomName() != null) {
                room.setRoomName(request.getRoomName());
            }
            if (request.getCategory() != null) {
                room.setCategory(request.getCategory());
            }
        }
        liveRoomService.updateRoom(room);
        return Result.success(room);
    }

    @PutMapping("/{id}/schedule")
    @Operation(summary = "设置定时开播时间")
    public Result<?> scheduleRoom(@PathVariable Integer id, @RequestBody(required = false) ScheduleLiveRoomRequest request,
                                  HttpServletRequest httpRequest) {
        AuthUser user = userResolver.requireUser(httpRequest);
        requireOwnedRoom(id, user);
        Date scheduledAt = request == null || request.getScheduledAt() == null
                ? null
                : new Date(request.getScheduledAt());
        liveRoomService.scheduleRoom(id, scheduledAt);
        return Result.success("定时开播已" + (scheduledAt == null ? "取消" : "设置"));
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
    public Map<String, Object> srsHook(@RequestBody SrsHookRequest body) {
        applySrsHook(body);
        return Map.of("code", 0);
    }

    private void applySrsHook(SrsHookRequest body) {
        if (body == null || !StringUtils.hasText(body.getStream())) {
            return;
        }
        String status = LiveConstants.SrsAction.roomStatusFor(body.getAction());
        if (status == null) {
            return;
        }
        LiveRoom room = liveRoomService.getByStreamKey(body.getStream());
        if (room != null) {
            liveRoomService.updateStatus(room.getId(), status);
        }
    }

    private LiveRoom requireOwnedRoom(Integer id, AuthUser user) {
        LiveRoom room = liveRoomService.getById(id);
        if (room == null) {
            throw new BusinessException("直播间不存在");
        }
        if (!room.getUserId().equals(user.intId())) {
            throw new BusinessException(403, "无权操作");
        }
        return room;
    }

}
