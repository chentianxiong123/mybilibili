package com.mybilibili.ai.controller;

import com.mybilibili.ai.entity.AiApiConfig;
import com.mybilibili.ai.service.AiApiConfigService;
import com.mybilibili.common.vo.Result;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/ai/admin/channels")
public class AiChannelController {

    @Autowired
    private AiApiConfigService aiApiConfigService;

    @GetMapping
    public Result<List<AiApiConfig>> listAll() {
        List<AiApiConfig> list = aiApiConfigService.listAll();
        // 隐藏 apiKey，只返回掩码
        list.forEach(c -> {
            if (c.getApiKey() != null && c.getApiKey().length() > 8) {
                c.setApiKey(c.getApiKey().substring(0, 6) + "****" + c.getApiKey().substring(c.getApiKey().length() - 4));
            }
        });
        return Result.success(list);
    }

    @GetMapping("/{id}")
    public Result<AiApiConfig> getById(@PathVariable Long id) {
        AiApiConfig config = aiApiConfigService.getById(id);
        if (config == null) return Result.error("渠道不存在");
        return Result.success(config);
    }

    @PostMapping
    public Result<AiApiConfig> create(@RequestBody AiApiConfig config) {
        if (config.getName() == null || config.getName().isEmpty()) return Result.error("渠道名称不能为空");
        if (config.getBaseUrl() == null || config.getBaseUrl().isEmpty()) return Result.error("API地址不能为空");
        if (config.getApiKey() == null || config.getApiKey().isEmpty()) return Result.error("API密钥不能为空");
        if (config.getModel() == null || config.getModel().isEmpty()) return Result.error("模型名称不能为空");
        return Result.success(aiApiConfigService.create(config));
    }

    @PutMapping("/{id}")
    public Result<AiApiConfig> update(@PathVariable Long id, @RequestBody AiApiConfig config) {
        return Result.success(aiApiConfigService.update(id, config));
    }

    @DeleteMapping("/{id}")
    public Result<Void> delete(@PathVariable Long id) {
        aiApiConfigService.delete(id);
        return Result.success();
    }

    @PutMapping("/{id}/toggle")
    public Result<Void> toggle(@PathVariable Long id) {
        aiApiConfigService.toggleEnabled(id);
        return Result.success();
    }

    @GetMapping("/bindings")
    public Result<Map<String, Long>> getBindings() {
        return Result.success(aiApiConfigService.getAllBindings());
    }

    @PostMapping("/bindings/{feature}")
    public Result<Void> bindFeature(@PathVariable String feature, @RequestBody Map<String, Long> body) {
        Long configId = body.get("configId");
        if (configId == null) return Result.error("configId 不能为空");
        aiApiConfigService.bindFeature(feature, configId);
        return Result.success();
    }
}
