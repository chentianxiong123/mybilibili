package com.mybilibili.danmaku.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.danmaku.dto.SendDanmakuDTO;
import com.mybilibili.danmaku.service.DanmakuService;
import com.mybilibili.danmaku.vo.DanmakuVO;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

@RestController
@RequestMapping("/danmaku")
@Tag(name = "弹幕接口", description = "弹幕的发送、获取、删除等操作")
public class DanmakuController {

    @Autowired
    private DanmakuService danmakuService;

    @GetMapping("/video/{videoId}")
    @Operation(summary = "获取视频弹幕", description = "获取指定视频的弹幕列表")
    public Result<List<DanmakuVO>> getDanmakus(@PathVariable Integer videoId) {
        return danmakuService.getDanmakus(videoId);
    }

    @GetMapping("/video/{videoId}/time-range")
    @Operation(summary = "获取指定时间范围的弹幕", description = "获取视频在指定时间范围内的弹幕")
    public Result<List<DanmakuVO>> getDanmakusByTimeRange(
            @PathVariable Integer videoId,
            @RequestParam Double startTime,
            @RequestParam Double endTime) {
        return danmakuService.getDanmakusByTimeRange(videoId, startTime, endTime);
    }

    @PostMapping
    @Operation(summary = "发送弹幕", description = "向指定视频发送弹幕")
    public Result<Void> sendDanmaku(@RequestBody SendDanmakuDTO dto, HttpServletRequest request) {
        Integer userId = JwtUtils.getUserIdFromRequest(request);
        return danmakuService.sendDanmaku(dto, userId);
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除弹幕", description = "删除指定的弹幕")
    public Result<Void> deleteDanmaku(@PathVariable String id, HttpServletRequest request) {
        Integer userId = JwtUtils.getUserIdFromRequest(request);
        return danmakuService.deleteDanmaku(id, userId);
    }

    @GetMapping("/video/{videoId}/count")
    @Operation(summary = "获取弹幕数量", description = "获取指定视频的弹幕数量")
    public Result<Long> getDanmakuCount(@PathVariable Integer videoId) {
        return danmakuService.getDanmakuCount(videoId);
    }
}