package com.mybilibili.admin.controller;

import com.mybilibili.admin.service.StatisticsService;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/statistics")
@Tag(name = "统计分析接口", description = "视频播放量、用户增长、评论数量等统计分析")
public class StatisticsController {

    @Autowired
    private StatisticsService statisticsService;

    @GetMapping("/video/play")
    @Operation(summary = "视频播放量统计", description = "统计指定时间范围内的视频播放量")
    public Result<Map<String, Object>> getVideoPlayStatistics(
            @RequestParam String startDate,
            @RequestParam String endDate) {
        try {
            Map<String, Object> result = statisticsService.getVideoPlayStatistics(startDate, endDate);
            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/growth")
    @Operation(summary = "用户增长统计", description = "统计指定时间范围内的用户增长情况")
    public Result<Map<String, Object>> getUserGrowthStatistics(
            @RequestParam String startDate,
            @RequestParam String endDate) {
        try {
            Map<String, Object> result = statisticsService.getUserGrowthStatistics(startDate, endDate);
            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/comment")
    @Operation(summary = "评论数量统计", description = "统计指定时间范围内的评论数量")
    public Result<Map<String, Object>> getCommentStatistics(
            @RequestParam String startDate,
            @RequestParam String endDate) {
        try {
            Map<String, Object> result = statisticsService.getCommentStatistics(startDate, endDate);
            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/video/hot")
    @Operation(summary = "热门视频分析", description = "获取热门视频列表")
    public Result<List<Map<String, Object>>> getHotVideos(
            @RequestParam(defaultValue = "10") int limit) {
        try {
            List<Map<String, Object>> result = statisticsService.getHotVideos(limit);
            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/overview")
    @Operation(summary = "获取总览统计", description = "获取用户数、稿件数、视频数、播放量等总览数据")
    public Result<Map<String, Object>> getOverviewStatistics() {
        try {
            Map<String, Object> result = statisticsService.getOverviewStatistics();
            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/manuscript/status")
    @Operation(summary = "稿件状态分布", description = "获取各状态的稿件数量分布")
    public Result<List<Map<String, Object>>> getManuscriptStatusStatistics() {
        try {
            List<Map<String, Object>> result = statisticsService.getManuscriptStatusStatistics();
            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/manuscript/recent")
    @Operation(summary = "最近稿件", description = "获取最近上传的稿件列表")
    public Result<List<Map<String, Object>>> getRecentManuscripts(
            @RequestParam(defaultValue = "10") int limit) {
        try {
            List<Map<String, Object>> result = statisticsService.getRecentManuscripts(limit);
            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}