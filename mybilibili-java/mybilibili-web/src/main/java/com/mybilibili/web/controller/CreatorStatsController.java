package com.mybilibili.web.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.CreatorStatsVO;
import com.mybilibili.common.vo.LatestCommentVO;
import com.mybilibili.common.vo.RankingVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.web.service.CreatorStatsService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.time.LocalDate;
import java.time.format.DateTimeFormatter;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 创作中心统计接口
 */
@RestController
@RequestMapping("/creator")
@Tag(name = "创作中心统计接口", description = "创作者统计数据、最新评论、排行榜等")
public class CreatorStatsController {

    @Autowired
    private CreatorStatsService creatorStatsService;

    /**
     * 获取创作者统计数据
     */
    @GetMapping("/stats")
    @Operation(summary = "获取创作者统计数据", description = "获取粉丝总数、总播放量、总评论数、总点赞数、总分享数、总收藏数、总投币数")
    @SecurityRequirement(name = "JWT")
    public Result<CreatorStatsVO> getCreatorStats(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            CreatorStatsVO stats = creatorStatsService.getCreatorStats(userId);
            return Result.success("获取成功", stats);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    /**
     * 获取数据概览
     */
    @GetMapping("/stats/overview")
    @Operation(summary = "获取数据概览", description = "获取创作者数据概览，包含增量数据")
    @SecurityRequirement(name = "JWT")
    public Result<Map<String, Object>> getStatsOverview(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            Map<String, Object> overview = creatorStatsService.getStatsOverview(userId);
            return Result.success("获取成功", overview);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    /**
     * 获取趋势数据
     */
    @GetMapping("/stats/trend")
    @Operation(summary = "获取趋势数据", description = "获取近N天的数据趋势")
    @SecurityRequirement(name = "JWT")
    public Result<Map<String, Object>> getStatsTrend(
            @RequestParam(value = "days", required = false, defaultValue = "7") Integer days,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            Map<String, Object> trend = creatorStatsService.getStatsTrend(userId, days);
            return Result.success("获取成功", trend);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    /**
     * 获取稿件排行
     */
    @GetMapping("/stats/ranking")
    @Operation(summary = "获取稿件排行", description = "获取稿件排行榜")
    @SecurityRequirement(name = "JWT")
    public Result<Map<String, Object>> getManuscriptRanking(
            @RequestParam(value = "sortBy", required = false, defaultValue = "views") String sortBy,
            @RequestParam(value = "limit", required = false, defaultValue = "10") Integer limit,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<RankingVO> list = creatorStatsService.getManuscriptRanking(userId, sortBy, limit);
            Map<String, Object> result = new HashMap<>();
            result.put("list", list);
            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    /**
     * 获取最新评论列表
     */
    @GetMapping("/latest-comments")
    @Operation(summary = "获取最新评论列表", description = "获取当前用户稿件下的最新评论")
    @SecurityRequirement(name = "JWT")
    public Result<List<LatestCommentVO>> getLatestComments(
            @RequestParam(value = "limit", required = false, defaultValue = "10") Integer limit,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<LatestCommentVO> comments = creatorStatsService.getLatestComments(userId, limit);
            return Result.success("获取成功", comments);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    /**
     * 获取排行榜数据
     */
    @GetMapping("/rankings")
    @Operation(summary = "获取排行榜数据", description = "获取观看排行、互动排行")
    @SecurityRequirement(name = "JWT")
    public Result<Map<String, Object>> getRankings(
            @RequestParam(value = "limit", required = false, defaultValue = "10") Integer limit,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<RankingVO> viewRankings = creatorStatsService.getViewRankings(userId, limit);
            List<RankingVO> interactionRankings = creatorStatsService.getInteractionRankings(userId, limit);
            Map<String, Object> result = new HashMap<>();
            result.put("viewRankings", viewRankings);
            result.put("interactionRankings", interactionRankings);
            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
