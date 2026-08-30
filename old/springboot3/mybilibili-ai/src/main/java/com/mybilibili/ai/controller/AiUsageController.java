package com.mybilibili.ai.controller;

import com.mybilibili.ai.mapper.AiUsageLogMapper;
import com.mybilibili.common.vo.Result;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/ai/usage")
public class AiUsageController {

    @Autowired
    private AiUsageLogMapper aiUsageLogMapper;

    @GetMapping("/overview")
    public Result<Map<String, Object>> getOverview() {
        return Result.success(aiUsageLogMapper.selectOverview());
    }

    @GetMapping("/features")
    public Result<List<Map<String, Object>>> getFeatureStats() {
        return Result.success(aiUsageLogMapper.selectFeatureStats());
    }

    @GetMapping("/daily")
    public Result<List<Map<String, Object>>> getDailyStats(@RequestParam(defaultValue = "7") int days) {
        String startDate = new java.text.SimpleDateFormat("yyyy-MM-dd")
                .format(new java.util.Date(System.currentTimeMillis() - days * 86400000L));
        return Result.success(aiUsageLogMapper.selectDailyStats(startDate));
    }
}