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
        list.forEach(this::maskApiKey);
        return Result.success(list);
    }

    @GetMapping("/type/{type}")
    public Result<List<AiApiConfig>> listByType(@PathVariable String type) {
        List<AiApiConfig> list = aiApiConfigService.listByType(type);
        list.forEach(this::maskApiKey);
        return Result.success(list);
    }

    @GetMapping("/{id}")
    public Result<AiApiConfig> getById(@PathVariable Long id) {
        AiApiConfig config = aiApiConfigService.getById(id);
        maskApiKey(config);
        return Result.success(config);
    }

    @PostMapping
    public Result<AiApiConfig> create(@RequestBody AiApiConfig config) {
        if (config.getName() == null || config.getName().isEmpty()) return Result.error("渠道名称不能为空");
        if (config.getType() == null || config.getType().isEmpty()) return Result.error("类型不能为空");
        if (config.getBaseUrl() == null || config.getBaseUrl().isEmpty()) return Result.error("API地址不能为空");
        if (config.getApiKey() == null || config.getApiKey().isEmpty()) return Result.error("API密钥不能为空");
        if (config.getModel() == null || config.getModel().isEmpty()) return Result.error("模型名称不能为空");
        AiApiConfig saved = aiApiConfigService.create(config);
        maskApiKey(saved);
        return Result.success(saved);
    }

    @PutMapping("/{id}")
    public Result<AiApiConfig> update(@PathVariable Long id, @RequestBody AiApiConfig config) {
        AiApiConfig saved = aiApiConfigService.update(id, config);
        maskApiKey(saved);
        return Result.success(saved);
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

    @GetMapping("/types")
    public Result<List<String>> getAvailableTypes() {
        return Result.success(List.of("LLM", "ASR", "TTS", "IMAGE", "MODERATION"));
    }

    @GetMapping("/features")
    public Result<List<Map<String, String>>> getAvailableFeatures() {
        return Result.success(List.of(
                Map.of("value", "CHAT", "label", "AI 客服", "type", "LLM"),
                Map.of("value", "REVIEW", "label", "内容审核", "type", "MODERATION"),
                Map.of("value", "SUMMARY", "label", "视频摘要", "type", "LLM"),
                Map.of("value", "ADMIN", "label", "管理助手", "type", "LLM"),
                Map.of("value", "TRANSCRIBE", "label", "语音转写", "type", "ASR")
        ));
    }

    private void maskApiKey(AiApiConfig config) {
        if (config == null || config.getApiKey() == null) {
            return;
        }
        String apiKey = config.getApiKey();
        if (apiKey.length() <= 8) {
            config.setApiKey("****");
            return;
        }
        config.setApiKey(apiKey.substring(0, 6) + "****" + apiKey.substring(apiKey.length() - 4));
    }
}
