package com.mybilibili.ai.tool;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.ai.mapper.AiUsageLogMapper;
import org.springframework.ai.tool.annotation.Tool;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

import java.time.LocalDate;
import java.time.format.DateTimeFormatter;
import java.util.HashMap;
import java.util.Map;

@Component
public class AdminToolService {

    private static final String GATEWAY_BASE_URL = "http://localhost:8080/api";
    private static final DateTimeFormatter DATE_FORMATTER = DateTimeFormatter.ofPattern("yyyy-MM-dd");

    private final AiUsageLogMapper aiUsageLogMapper;
    private final RestTemplate restTemplate;
    private final ObjectMapper objectMapper;

    @Autowired
    public AdminToolService(AiUsageLogMapper aiUsageLogMapper) {
        this.aiUsageLogMapper = aiUsageLogMapper;
        this.restTemplate = new RestTemplate();
        this.objectMapper = new ObjectMapper();
    }

    @Tool(name = "getOverviewStats", description = "获取平台概览统计：用户总数、稿件总数、评论总数等核心指标")
    public StatsData getOverviewStats() {
        try {
            String url = GATEWAY_BASE_URL + "/statistics/overview";
            String response = restTemplate.getForObject(url, String.class);
            Map<String, Object> data = objectMapper.readValue(response, Map.class);
            return new StatsData("overview", data, "number", "平台数据概览");
        } catch (Exception e) {
            return createErrorStatsData("overview", "number", "平台数据概览");
        }
    }

    @Tool(name = "getUserGrowth", description = "获取用户增长趋势，参数 days 表示最近天数（默认7）")
    public StatsData getUserGrowth(int days) {
        try {
            LocalDate endDate = LocalDate.now();
            LocalDate startDate = endDate.minusDays(days);
            String startStr = startDate.format(DATE_FORMATTER);
            String endStr = endDate.format(DATE_FORMATTER);

            String url = GATEWAY_BASE_URL + "/statistics/user/growth?startDate=" + startStr + "&endDate=" + endStr;
            String response = restTemplate.getForObject(url, String.class);
            Map<String, Object> data = objectMapper.readValue(response, Map.class);
            return new StatsData("user_growth", data, "line", "用户增长趋势");
        } catch (Exception e) {
            return createErrorStatsData("user_growth", "line", "用户增长趋势");
        }
    }

    @Tool(name = "getManuscriptStats", description = "获取稿件状态分布统计（审核通过/待审核/已下架等）")
    public StatsData getManuscriptStats() {
        try {
            String url = GATEWAY_BASE_URL + "/statistics/manuscript/status";
            String response = restTemplate.getForObject(url, String.class);
            Map<String, Object> data = objectMapper.readValue(response, Map.class);
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
            String url = GATEWAY_BASE_URL + "/statistics/video/hot?limit=" + limit;
            String response = restTemplate.getForObject(url, String.class);
            Map<String, Object> data = objectMapper.readValue(response, Map.class);
            return new StatsData("hot_videos", data, "table", "热门视频排行");
        } catch (Exception e) {
            return createErrorStatsData("hot_videos", "table", "热门视频排行");
        }
    }

    private StatsData createErrorStatsData(String type, String chartType, String title) {
        Map<String, Object> errorData = new HashMap<>();
        errorData.put("error", "获取数据失败");
        return new StatsData(type, errorData, chartType, title);
    }
}