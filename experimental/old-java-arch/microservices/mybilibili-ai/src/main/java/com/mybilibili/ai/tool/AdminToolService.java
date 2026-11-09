package com.mybilibili.ai.tool;

import com.mybilibili.ai.mapper.AdminStatsMapper;
import com.mybilibili.ai.mapper.AiUsageLogMapper;
import org.springframework.ai.tool.annotation.Tool;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.time.LocalDate;
import java.time.format.DateTimeFormatter;
import java.util.HashMap;
import java.util.Map;

@Component
public class AdminToolService {

    private static final DateTimeFormatter DATE_FORMATTER = DateTimeFormatter.ofPattern("yyyy-MM-dd");

    private final AdminStatsMapper adminStatsMapper;
    private final AiUsageLogMapper aiUsageLogMapper;

    @Autowired
    public AdminToolService(AdminStatsMapper adminStatsMapper, AiUsageLogMapper aiUsageLogMapper) {
        this.adminStatsMapper = adminStatsMapper;
        this.aiUsageLogMapper = aiUsageLogMapper;
    }

    @Tool(name = "getOverviewStats", description = "获取平台概览统计：用户总数、稿件总数、评论总数等核心指标")
    public StatsData getOverviewStats() {
        try {
            Map<String, Object> data = adminStatsMapper.selectOverviewStats();
            return new StatsData("overview", data, "number", "平台数据概览");
        } catch (Exception e) {
            return createErrorStatsData("overview", "number", "平台数据概览");
        }
    }

    @Tool(name = "getUserGrowth", description = "获取用户增长趋势，参数 days 表示最近天数（默认7）")
    public StatsData getUserGrowth(int days) {
        try {
            int normalizedDays = normalizePositive(days, 7, 90);
            LocalDate startDate = LocalDate.now().minusDays(normalizedDays);
            String startStr = startDate.format(DATE_FORMATTER);
            Map<String, Object> data = new HashMap<>();
            data.put("items", adminStatsMapper.selectUserGrowth(startStr));
            data.put("days", normalizedDays);
            return new StatsData("user_growth", data, "line", "用户增长趋势");
        } catch (Exception e) {
            return createErrorStatsData("user_growth", "line", "用户增长趋势");
        }
    }

    @Tool(name = "getManuscriptStats", description = "获取稿件状态分布统计（审核通过/待审核/已下架等）")
    public StatsData getManuscriptStats() {
        try {
            Map<String, Object> data = new HashMap<>();
            data.put("items", adminStatsMapper.selectManuscriptStatusStats());
            return new StatsData("manuscript_stats", data, "pie", "稿件状态统计");
        } catch (Exception e) {
            return createErrorStatsData("manuscript_stats", "pie", "稿件状态统计");
        }
    }

    @Tool(name = "getAiUsageOverview", description = "获取 AI 功能用量概览（总调用次数、Token消耗等）")
    public StatsData getAiUsageOverview() {
        try {
            Map<String, Object> data = aiUsageLogMapper.selectOverview();
            return new StatsData("ai_usage", data, "number", "AI使用概览");
        } catch (Exception e) {
            return createErrorStatsData("ai_usage", "number", "AI使用概览");
        }
    }

    @Tool(name = "getHotVideos", description = "获取热门视频列表，参数 limit 表示返回数量（默认10）")
    public StatsData getHotVideos(int limit) {
        try {
            int normalizedLimit = normalizePositive(limit, 10, 50);
            Map<String, Object> data = new HashMap<>();
            data.put("items", adminStatsMapper.selectHotVideos(normalizedLimit));
            data.put("limit", normalizedLimit);
            return new StatsData("hot_videos", data, "table", "热门视频排行");
        } catch (Exception e) {
            return createErrorStatsData("hot_videos", "table", "热门视频排行");
        }
    }

    private int normalizePositive(int value, int defaultValue, int maxValue) {
        if (value <= 0) {
            return defaultValue;
        }
        return Math.min(value, maxValue);
    }

    private StatsData createErrorStatsData(String type, String chartType, String title) {
        Map<String, Object> errorData = new HashMap<>();
        errorData.put("error", "获取数据失败");
        return new StatsData(type, errorData, chartType, title);
    }
}
