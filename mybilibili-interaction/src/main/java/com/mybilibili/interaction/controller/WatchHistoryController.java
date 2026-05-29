package com.mybilibili.interaction.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.WatchHistoryVO;
import com.mybilibili.interaction.service.WatchHistoryService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.util.List;

@RestController
@RequestMapping("/watch-history")
@Tag(name = "浏览历史相关接口", description = "浏览历史记录的添加、查询、删除等操作")
public class WatchHistoryController {

    @Autowired
    private WatchHistoryService watchHistoryService;

    @GetMapping
    @Operation(summary = "获取浏览历史列表", description = "获取当前用户的浏览历史记录，支持分页")
    @SecurityRequirement(name = "JWT")
    public Result<List<WatchHistoryVO>> getWatchHistoryList(
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            List<WatchHistoryVO> histories = watchHistoryService.getWatchHistoryList(userId, page, size);
            return Result.success("获取成功", histories);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping
    @Operation(summary = "记录浏览历史", description = "记录或更新视频浏览历史")
    @SecurityRequirement(name = "JWT")
    public Result<?> recordWatchHistory(
            @RequestParam Integer videoId,
            @RequestParam Integer progressSeconds,
            @RequestParam Integer videoDuration,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            watchHistoryService.recordWatchHistory(userId, videoId, progressSeconds, videoDuration);
            return Result.success("记录成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping
    @Operation(summary = "清空浏览历史", description = "清空当前用户的所有浏览历史记录")
    @SecurityRequirement(name = "JWT")
    public Result<?> clearWatchHistory(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            watchHistoryService.clearWatchHistory(userId);
            return Result.success("清空成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除单条浏览历史", description = "删除指定的浏览历史记录")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteWatchHistory(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            watchHistoryService.deleteWatchHistory(id, userId);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
