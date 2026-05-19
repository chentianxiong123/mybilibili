package com.mybilibili.live.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.live.entity.LiveLinkmic;
import com.mybilibili.live.service.LiveLinkmicService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

@Slf4j
@RestController
@RequestMapping("/live/linkmic")
public class LiveLinkmicController {

    @Autowired
    private LiveLinkmicService linkmicService;

    @PostMapping("/apply/{roomId}")
    public Result<?> applyLinkmic(@PathVariable Long roomId, HttpServletRequest request) {
        Long viewerId = Long.parseLong(request.getHeader("X-User-Id"));
        String viewerName = request.getHeader("X-Username");
        LiveLinkmic linkmic = linkmicService.applyLinkmic(roomId, viewerId, viewerName);
        if (linkmic == null) {
            return Result.error("连麦人数已满，请稍后再试");
        }
        return Result.success(linkmic);
    }

    @PostMapping("/accept/{linkmicId}")
    public Result<?> acceptLinkmic(@PathVariable Long linkmicId) {
        linkmicService.acceptLinkmic(linkmicId);
        return Result.success("已同意连麦");
    }

    @PostMapping("/reject/{linkmicId}")
    public Result<?> rejectLinkmic(@PathVariable Long linkmicId) {
        linkmicService.rejectLinkmic(linkmicId);
        return Result.success("已拒绝连麦");
    }

    @PostMapping("/disconnect/{linkmicId}")
    public Result<?> disconnectLinkmic(@PathVariable Long linkmicId) {
        linkmicService.disconnectLinkmic(linkmicId);
        return Result.success("已断开连麦");
    }

    @PostMapping("/toggle-audio/{linkmicId}")
    public Result<?> toggleAudio(@PathVariable Long linkmicId, @RequestParam boolean enabled) {
        linkmicService.toggleAudio(linkmicId, enabled);
        return Result.success(enabled ? "麦克风已开启" : "麦克风已关闭");
    }

    @PostMapping("/toggle-video/{linkmicId}")
    public Result<?> toggleVideo(@PathVariable Long linkmicId, @RequestParam boolean enabled) {
        linkmicService.toggleVideo(linkmicId, enabled);
        return Result.success(enabled ? "摄像头已开启" : "摄像头已关闭");
    }

    @GetMapping("/active/{roomId}")
    public Result<?> getActiveLinkmics(@PathVariable Long roomId) {
        List<LiveLinkmic> linkmics = linkmicService.getActiveLinkmics(roomId);
        return Result.success(linkmics);
    }

    @GetMapping("/pending/{roomId}")
    public Result<?> getPendingApplications(@PathVariable Long roomId, HttpServletRequest request) {
        Long streamerId = Long.parseLong(request.getHeader("X-User-Id"));
        List<LiveLinkmic> applications = linkmicService.getPendingApplications(roomId, streamerId);
        return Result.success(applications);
    }
}