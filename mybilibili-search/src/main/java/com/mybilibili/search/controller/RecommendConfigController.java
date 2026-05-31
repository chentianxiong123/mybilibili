package com.mybilibili.search.controller;

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

    @GetMapping
    public Result<RecommendConfig> getConfig() {
        return Result.success("获取成功", configService.getConfig());
    }

    @PutMapping
    public Result<RecommendConfig> updateConfig(@RequestBody RecommendConfig config, HttpServletRequest request) {
        String adminName = request.getHeader("X-User-Name");
        if (adminName == null) adminName = "admin";
        RecommendConfig updated = configService.updateConfig(config, adminName);
        return Result.success("更新成功", updated);
    }

    @PostMapping("/reset")
    public Result<RecommendConfig> resetConfig(HttpServletRequest request) {
        String adminName = request.getHeader("X-User-Name");
        if (adminName == null) adminName = "admin";
        RecommendConfig defaults = RecommendConfig.defaults();
        RecommendConfig updated = configService.updateConfig(defaults, adminName);
        return Result.success("已重置为默认值", updated);
    }
}
