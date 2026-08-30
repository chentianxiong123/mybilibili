package com.mybilibili.ai.controller;

import com.mybilibili.ai.service.AiConfigService;
import com.mybilibili.common.vo.Result;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/ai/admin/config")
public class AiConfigAdminController {

    @Autowired
    private AiConfigService aiConfigService;

    @GetMapping
    public Result<Map<String, Object>> getConfig() {
        try {
            return Result.success("获取成功", aiConfigService.getConfig());
        } catch (Exception e) {
            return Result.error("获取配置失败：" + e.getMessage());
        }
    }

    @PutMapping
    public Result<Void> updateConfig(@RequestBody Map<String, Object> updates) {
        try {
            aiConfigService.updateConfig(updates);
            return Result.success("配置更新成功", null);
        } catch (Exception e) {
            return Result.error("更新配置失败：" + e.getMessage());
        }
    }

    @PostMapping("/test")
    public Result<Map<String, Object>> testConnection(@RequestBody(required = false) Map<String, String> request) {
        try {
            Map<String, Object> testResult = aiConfigService.testConnection(request);
            if (Boolean.TRUE.equals(testResult.get("success"))) {
                return Result.success(testResult);
            } else {
                return Result.error(500, (String) testResult.get("message"));
            }
        } catch (Exception e) {
            return Result.error("测试连接失败：" + e.getMessage());
        }
    }
}