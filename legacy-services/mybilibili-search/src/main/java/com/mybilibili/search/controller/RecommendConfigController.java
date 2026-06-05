package com.mybilibili.search.controller;

import com.mybilibili.common.audit.AuditLogRecorder;
import com.mybilibili.common.vo.Result;
import com.mybilibili.search.entity.RecommendConfig;
import com.mybilibili.search.service.RecommendConfigService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;

@RestController
@RequestMapping("/search/admin/recommend-config")
public class RecommendConfigController {

    @Autowired
    private RecommendConfigService configService;

    @Autowired
    private AuditLogRecorder auditLogRecorder;

    @GetMapping
    public Result<RecommendConfig> getConfig() {
        return Result.success("获取成功", configService.getConfig());
    }

    @PutMapping
    public Result<RecommendConfig> updateConfig(@RequestBody RecommendConfig config, HttpServletRequest request) {
        String adminName = request.getHeader("X-Username");
        if (adminName == null || adminName.isBlank()) adminName = "admin";
        try {
            RecommendConfig updated = configService.updateConfig(config, adminName);
            Result<RecommendConfig> result = Result.success("更新成功", updated);
            auditLogRecorder.record(request, "recommend", "recommend_config_update", "recommend_config", "default", result);
            return result;
        } catch (Exception e) {
            Result<RecommendConfig> result = Result.error("更新失败: " + e.getMessage());
            auditLogRecorder.record(request, "recommend", "recommend_config_update", "recommend_config", "default", result);
            return result;
        }
    }

    @PostMapping("/reset")
    public Result<RecommendConfig> resetConfig(HttpServletRequest request) {
        String adminName = request.getHeader("X-Username");
        if (adminName == null || adminName.isBlank()) adminName = "admin";
        try {
            RecommendConfig defaults = RecommendConfig.defaults();
            RecommendConfig updated = configService.updateConfig(defaults, adminName);
            Result<RecommendConfig> result = Result.success("已重置为默认值", updated);
            auditLogRecorder.record(request, "recommend", "recommend_config_reset", "recommend_config", "default", result);
            return result;
        } catch (Exception e) {
            Result<RecommendConfig> result = Result.error("重置失败: " + e.getMessage());
            auditLogRecorder.record(request, "recommend", "recommend_config_reset", "recommend_config", "default", result);
            return result;
        }
    }
}
