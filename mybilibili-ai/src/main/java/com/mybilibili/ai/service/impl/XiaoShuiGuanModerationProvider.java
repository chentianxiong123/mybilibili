package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.service.ModerationProvider;
import com.mybilibili.ai.service.ModerationProvider.ModerateRequest;
import com.mybilibili.ai.service.ModerationProvider.ModerationResult;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.*;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

import java.util.HashMap;
import java.util.Map;

/**
 * 小水管 API 内容审核提供者。
 * 支持 qwen3guard（预过滤）和 llama-guard（细审核）双模型。
 *
 * 用法：
 * - qwen3guard：弹幕预过滤，拦截 spam/广告/低质量内容
 * - llama-guard：举报内容深度审核，识别辱骂/政治/色情等深层违规
 */
public class XiaoShuiGuanModerationProvider implements ModerationProvider {

    private final String baseUrl;
    private final String apiKey;
    private final String defaultModel;
    private final RestTemplate restTemplate;

    public XiaoShuiGuanModerationProvider(String baseUrl, String apiKey, String defaultModel) {
        this.baseUrl = baseUrl;
        this.apiKey = apiKey;
        this.defaultModel = defaultModel != null ? defaultModel : "qwen3guard";
        this.restTemplate = new RestTemplate();
    }

    @Override
    public String getName() {
        return "小水管";
    }

    @Override
    public boolean isAvailable() {
        return apiKey != null && !apiKey.isEmpty();
    }

    @Override
    public Object invoke(ModerateRequest request) {
        return moderate(request);
    }

    /**
     * 执行内容审核
     * @param request 审核请求
     * @return 审核结果
     */
    public ModerationResult moderate(ModerateRequest request) {
        if (request.getText() == null || request.getText().isEmpty()) {
            return new ModerationResult(true, null);
        }

        try {
            String scene = request.getScene();
            String model = selectModel(scene);

            // 构建请求
            Map<String, Object> chatRequest = new HashMap<>();
            chatRequest.put("model", model);

            // 构建消息
            Map<String, String> userMessage = new HashMap<>();
            userMessage.put("role", "user");

            String prompt = buildPrompt(request.getText(), scene);
            userMessage.put("content", prompt);

            chatRequest.put("messages", new Map[]{userMessage});

            // 发送请求
            HttpHeaders headers = new HttpHeaders();
            headers.setContentType(MediaType.APPLICATION_JSON);
            headers.set("Authorization", "Bearer " + apiKey);

            HttpEntity<Map<String, Object>> entity = new HttpEntity<>(chatRequest, headers);
            ResponseEntity<Map> response = restTemplate.exchange(
                    baseUrl + "/v1/chat/completions",
                    HttpMethod.POST,
                    entity,
                    Map.class
            );

            if (response.getStatusCode() == HttpStatus.OK && response.getBody() != null) {
                return parseResponse(response.getBody());
            }

            return new ModerationResult(false, "审核服务异常");
        } catch (Exception e) {
            System.err.println("[小水管] 审核异常: " + e.getMessage());
            return new ModerationResult(false, "审核异常: " + e.getMessage());
        }
    }

    /**
     * 根据场景选择模型
     * - COMMENT/REPLY: qwen3guard (预过滤，拦 spam/广告)
     * - REPORT: llama-guard (细审核，识别辱骂/政治/色情)
     */
    private String selectModel(String scene) {
        if ("REPORT".equals(scene)) {
            return "llama-guard".equals(defaultModel) ? "llama-guard" : "nemoguard";
        }
        return "qwen3guard";
    }

    /**
     * 构建审核 prompt
     */
    private String buildPrompt(String text, String scene) {
        if ("DANMAKU".equals(scene)) {
            return "请判断以下弹幕内容是否违规（spam/广告/低质量）：\n" + text + "\n\n只返回 JSON：{\"passed\": true/false, \"reason\": \"原因\"}";
        } else {
            return "请判断以下内容是否违规（辱骂/政治敏感/色情等深层违规）：\n" + text + "\n\n只返回 JSON：{\"passed\": true/false, \"reason\": \"原因\", \"violationTypes\": [\"类型列表\"]}";
        }
    }

    /**
     * 解析响应
     */
    private ModerationResult parseResponse(Map body) {
        try {
            var choices = (java.util.List<?>) body.get("choices");
            if (choices == null || choices.isEmpty()) {
                return new ModerationResult(false, "响应格式异常");
            }

            var firstChoice = (Map<?, ?>) choices.get(0);
            var message = (Map<?, ?>) firstChoice.get("message");
            String content = (String) message.get("content");

            // 解析 JSON 响应
            if (content != null) {
                // 尝试解析 JSON
                int jsonStart = content.indexOf("{");
                int jsonEnd = content.lastIndexOf("}");
                if (jsonStart >= 0 && jsonEnd > jsonStart) {
                    String jsonStr = content.substring(jsonStart, jsonEnd + 1);
                    var mapper = new com.fasterxml.jackson.databind.ObjectMapper();
                    var result = mapper.readTree(jsonStr);

                    boolean passed = result.has("passed") ? result.get("passed").asBoolean() : true;
                    String reason = result.has("reason") ? result.get("reason").asText() : null;

                    ModerationResult mr = new ModerationResult(passed, reason);

                    if (result.has("violationTypes")) {
                        java.util.List<String> types = new java.util.ArrayList<>();
                        result.get("violationTypes").forEach(v -> types.add(v.asText()));
                        mr.setViolationTypes(types);
                    }

                    return mr;
                }
            }

            return new ModerationResult(true, null);
        } catch (Exception e) {
            System.err.println("[小水管] 解析响应异常: " + e.getMessage());
            return new ModerationResult(false, "解析响应异常");
        }
    }
}