package com.mybilibili.live.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.live.common.AuthUser;
import com.mybilibili.live.common.RequestUserResolver;
import com.mybilibili.live.entity.LiveLinkmic;
import com.mybilibili.live.service.LiveLinkmicService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.util.List;

@Slf4j
@RestController
@RequestMapping("/live/linkmic")
public class LiveLinkmicController {

    @Autowired
    private LiveLinkmicService linkmicService;

    @Autowired
    private RequestUserResolver userResolver;

    @PostMapping("/apply/{roomId}")
    public Result<?> applyLinkmic(@PathVariable Long roomId, HttpServletRequest request) {
        AuthUser user = userResolver.requireUser(request);
        LiveLinkmic linkmic = linkmicService.applyLinkmic(roomId, user.id(), user.username());
        return Result.success(linkmic);
    }

    @PostMapping("/accept/{linkmicId}")
    public Result<?> acceptLinkmic(@PathVariable Long linkmicId, HttpServletRequest request) {
        AuthUser user = userResolver.requireUser(request);
        linkmicService.acceptLinkmic(linkmicId, user.id());
        return Result.success("已同意连麦");
    }

    @PostMapping("/reject/{linkmicId}")
    public Result<?> rejectLinkmic(@PathVariable Long linkmicId, HttpServletRequest request) {
        AuthUser user = userResolver.requireUser(request);
        linkmicService.rejectLinkmic(linkmicId, user.id());
        return Result.success("已拒绝连麦");
    }

    @PostMapping("/disconnect/{linkmicId}")
    public Result<?> disconnectLinkmic(@PathVariable Long linkmicId, HttpServletRequest request) {
        AuthUser user = userResolver.requireUser(request);
        linkmicService.disconnectLinkmic(linkmicId, user.id());
        return Result.success("已断开连麦");
    }

    @PostMapping("/toggle-audio/{linkmicId}")
    public Result<?> toggleAudio(@PathVariable Long linkmicId, @RequestParam boolean enabled, HttpServletRequest request) {
        AuthUser user = userResolver.requireUser(request);
        linkmicService.toggleAudio(linkmicId, user.id(), enabled);
        return Result.success(enabled ? "麦克风已开启" : "麦克风已关闭");
    }

    @PostMapping("/toggle-video/{linkmicId}")
    public Result<?> toggleVideo(@PathVariable Long linkmicId, @RequestParam boolean enabled, HttpServletRequest request) {
        AuthUser user = userResolver.requireUser(request);
        linkmicService.toggleVideo(linkmicId, user.id(), enabled);
        return Result.success(enabled ? "摄像头已开启" : "摄像头已关闭");
    }

    @GetMapping("/active/{roomId}")
    public Result<?> getActiveLinkmics(@PathVariable Long roomId) {
        List<LiveLinkmic> linkmics = linkmicService.getActiveLinkmics(roomId);
        return Result.success(linkmics);
    }

    @GetMapping("/pending/{roomId}")
    public Result<?> getPendingApplications(@PathVariable Long roomId, HttpServletRequest request) {
        AuthUser user = userResolver.requireUser(request);
        List<LiveLinkmic> applications = linkmicService.getPendingApplications(roomId, user.id());
        return Result.success(applications);
    }

    @GetMapping("/queue-position/{roomId}")
    public Result<?> getQueuePosition(@PathVariable Long roomId, HttpServletRequest request) {
        AuthUser user = userResolver.requireUser(request);
        int position = linkmicService.getQueuePosition(roomId, user.id());
        return Result.success(position);
    }
}
