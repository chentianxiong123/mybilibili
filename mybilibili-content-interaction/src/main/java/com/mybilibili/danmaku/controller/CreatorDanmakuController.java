package com.mybilibili.danmaku.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.danmaku.entity.Danmaku;
import com.mybilibili.danmaku.repository.DanmakuRepository;
import com.mybilibili.danmaku.service.DanmakuService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/creator/danmaku")
@Tag(name = "创作者弹幕管理", description = "创作者管理自己稿件的弹幕")
public class CreatorDanmakuController {

    @Autowired
    private DanmakuService danmakuService;

    @Autowired
    private DanmakuRepository danmakuRepository;

    @GetMapping("/list")
    @Operation(summary = "获取创作者弹幕列表", description = "获取当前用户所有视频的弹幕列表，支持分页")
    public Result<Map<String, Object>> getCreatorDanmakuList(
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            @RequestParam(value = "videoId", required = false) Integer videoId,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            Map<String, Object> result = danmakuService.getCreatorDanmakuList(userId, videoId, page, size);
            return Result.success("获取成功", result);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{danmakuId}")
    @Operation(summary = "删除弹幕", description = "创作者删除自己稿件下的弹幕")
    public Result<?> deleteDanmaku(
            @PathVariable String danmakuId,
            HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            danmakuService.deleteDanmakuByCreator(danmakuId, userId);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error(e.getMessage());
        }
    }

    private Integer getUserIdFromRequest(HttpServletRequest request) {
        String userIdStr = request.getHeader("X-User-Id");
        if (userIdStr != null && !userIdStr.isEmpty()) {
            try {
                return Integer.parseInt(userIdStr);
            } catch (NumberFormatException e) {
                return null;
            }
        }
        return null;
    }

    @GetMapping("/debug/all")
    @Operation(summary = "调试接口-获取所有弹幕", description = "获取MongoDB中所有弹幕数据（调试用）")
    public Result<Map<String, Object>> debugGetAllDanmaku() {
        try {
            List<Danmaku> allDanmaku = danmakuRepository.findAll();
            long count = danmakuRepository.count();

            Map<String, Object> result = new HashMap<>();
            result.put("totalCount", count);
            result.put("danmakuList", allDanmaku);

            return Result.success("获取成功", result);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error("获取失败: " + e.getMessage());
        }
    }

    @GetMapping("/debug/by-video/{videoId}")
    @Operation(summary = "调试接口-根据视频ID获取弹幕", description = "获取指定视频的所有弹幕（调试用）")
    public Result<Map<String, Object>> debugGetDanmakuByVideoId(@PathVariable Integer videoId) {
        try {
            List<Danmaku> danmakuList = danmakuRepository.findByVideoId(videoId);

            Map<String, Object> result = new HashMap<>();
            result.put("videoId", videoId);
            result.put("count", danmakuList.size());
            result.put("danmakuList", danmakuList);

            return Result.success("获取成功", result);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error("获取失败: " + e.getMessage());
        }
    }
}
