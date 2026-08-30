package com.mybilibili.video.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.*;
import com.mybilibili.video.service.CreatorStatsService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/creator/stats")
@Tag(name = "创作者数据统计", description = "创作者数据中心相关接口")
public class CreatorStatsController {

    @Autowired
    private CreatorStatsService creatorStatsService;

    @GetMapping("/overview")
    @Operation(summary = "获取创作者数据概览", description = "获取创作者的核心数据指标汇总")
    public Result<CreatorOverviewVO> getOverview(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            CreatorOverviewVO overview = creatorStatsService.getCreatorOverview(userId);
            return Result.success("获取成功", overview);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/trend")
    @Operation(summary = "获取数据趋势", description = "获取指定天数内的数据趋势")
    public Result<TrendDataVO> getTrend(
            @Parameter(description = "天数") @RequestParam(defaultValue = "7") Integer days,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            TrendDataVO trend = creatorStatsService.getPlayTrend(userId, days);
            return Result.success("获取成功", trend);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/ranking")
    @Operation(summary = "获取稿件排行", description = "获取创作者稿件排行")
    public Result<Map<String, Object>> getRanking(
            @Parameter(description = "排序方式") @RequestParam(defaultValue = "views") String sortBy,
            @Parameter(description = "数量限制") @RequestParam(defaultValue = "10") Integer limit,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            List<ManuscriptRankVO> ranking = creatorStatsService.getManuscriptRanking(userId, sortBy, limit);
            
            Map<String, Object> result = new HashMap<>();
            result.put("list", ranking);
            result.put("total", ranking.size());
            
            return Result.success("获取成功", result);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/latest-comments")
    @Operation(summary = "获取最新评论", description = "获取创作者收到的最新评论")
    public Result<List<LatestCommentVO>> getLatestComments(
            @Parameter(description = "数量限制") @RequestParam(defaultValue = "5") Integer limit,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            List<LatestCommentVO> comments = creatorStatsService.getLatestComments(userId, limit);
            return Result.success("获取成功", comments);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/fans-ranking")
    @Operation(summary = "获取粉丝排行", description = "获取粉丝互动或观看排行")
    public Result<List<FansRankingVO>> getFansRanking(
            @Parameter(description = "排行类型") @RequestParam(defaultValue = "view") String type,
            @Parameter(description = "数量限制") @RequestParam(defaultValue = "10") Integer limit,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            List<FansRankingVO> ranking = creatorStatsService.getFansRanking(userId, type, limit);
            return Result.success("获取成功", ranking);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/fans-trend")
    @Operation(summary = "获取粉丝增长趋势", description = "获取指定天数内的粉丝增长趋势数据")
    public Result<FansTrendVO> getFansTrend(
            @Parameter(description = "天数") @RequestParam(defaultValue = "30") Integer days,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            FansTrendVO trend = creatorStatsService.getFansTrend(userId, days);
            return Result.success("获取成功", trend);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/manuscript-trend")
    @Operation(summary = "获取稿件播放趋势", description = "获取所有稿件按上传时间排列的播放量/弹幕趋势")
    public Result<ManuscriptTrendVO> getManuscriptTrend(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            ManuscriptTrendVO trend = creatorStatsService.getManuscriptTrend(userId);
            return Result.success("获取成功", trend);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error(e.getMessage());
        }
    }
}
