package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.config.DeepSeekConfig;
import com.mybilibili.ai.service.AiConfigService;
import com.mybilibili.ai.service.ContentReviewService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.util.*;

@Service
public class ContentReviewServiceImpl implements ContentReviewService {

    @Autowired
    private AiConfigService aiConfigService;

    private final RestTemplate restTemplate = new RestTemplate();

    private static final String SYSTEM_PROMPT =
        "你是一个内容审核助手，负责对用户举报的内容进行审核评估。" +
        "请根据举报原因分析内容，判断是否违规，并给出风险等级。\n\n" +
        "审核标准：\n" +
        "1. HIGH（高风险）：包含违法信息、色情内容、暴力恐怖、仇恨言论、严重辱骂攻击\n" +
        "2. MEDIUM（中风险）：包含广告营销、轻微辱骂、引战内容、未经证实的谣言\n" +
        "3. LOW（低风险）：无明显违规，正常讨论内容\n\n" +
        "注意：仅根据提供的文字内容进行判断，不要过度解读。\n\n" +
        "请严格按照以下JSON格式输出审核结果（不要包含markdown代码块标记）：\n" +
        "{\"verdict\": \"违规描述或'未发现违规'\", \"riskLevel\": \"HIGH/MEDIUM/LOW\"}";

    @Override
    public Map<String, Object> reviewContent(String content, String reason) {
        Map<String, Object> result = new HashMap<>();
        try {
            String apiKey = aiConfigService.getApiKey();
            if (apiKey == null || apiKey.isEmpty()) {
                result.put("status", "FAILED");
                result.put("verdict", "AI服务未配置API密钥");
                result.put("riskLevel", null);
                return result;
            }

            String userPrompt = "举报原因：" + reason + "\n\n被举报内容：\n" + content;

            Map<String, Object> requestBody = new HashMap<>();
            requestBody.put("model", "deepseek-chat");

            List<Map<String, String>> messages = new ArrayList<>();
            Map<String, String> systemMsg = new HashMap<>();
            systemMsg.put("role", "system");
            systemMsg.put("content", SYSTEM_PROMPT);
            messages.add(systemMsg);

            Map<String, String> userMsg = new HashMap<>();
            userMsg.put("role", "user");
            userMsg.put("content", userPrompt);
            messages.add(userMsg);

            requestBody.put("messages", messages);
            requestBody.put("max_tokens", 500);
            requestBody.put("temperature", 0.3);

            HttpHeaders headers = new HttpHeaders();
            headers.setContentType(MediaType.APPLICATION_JSON);
            headers.set("Authorization", "Bearer " + apiKey);

            HttpEntity<Map<String, Object>> request = new HttpEntity<>(requestBody, headers);
            ResponseEntity<Map> response = restTemplate.postForEntity(
                "https://api.qnaigc.com/v1/chat/completions", request, Map.class);

            if (response.getStatusCode() == HttpStatus.OK && response.getBody() != null) {
                String responseContent = parseResponse(response.getBody());
                Map<String, String> parsed = parseJsonResult(responseContent);

                result.put("status", "COMPLETED");
                result.put("verdict", parsed.getOrDefault("verdict", responseContent));
                result.put("riskLevel", parsed.getOrDefault("riskLevel", "LOW"));
            } else {
                result.put("status", "FAILED");
                result.put("verdict", "API返回错误: " + response.getStatusCode());
                result.put("riskLevel", null);
            }
        } catch (Exception e) {
            result.put("status", "FAILED");
            result.put("verdict", "审核调用异常: " + e.getMessage());
            result.put("riskLevel", null);
        }
        return result;
    }

    private String parseResponse(Map<String, Object> responseBody) {
        try {
            List<Map<String, Object>> choices = (List<Map<String, Object>>) responseBody.get("choices");
            if (choices != null && !choices.isEmpty()) {
                Map<String, Object> firstChoice = choices.get(0);
                Map<String, Object> message = (Map<String, Object>) firstChoice.get("message");
                if (message != null) {
                    return (String) message.get("content");
                }
            }
        } catch (Exception ignored) {}
        return "";
    }

    private Map<String, String> parseJsonResult(String responseContent) {
        Map<String, String> parsed = new HashMap<>();
        if (responseContent == null || responseContent.isEmpty()) {
            return parsed;
        }

        // 尝试提取JSON
        try {
            // 查找JSON开始和结束位置
            int start = responseContent.indexOf('{');
            int end = responseContent.lastIndexOf('}');
            if (start != -1 && end > start) {
                String json = responseContent.substring(start, end + 1);
                com.fasterxml.jackson.databind.ObjectMapper mapper = new com.fasterxml.jackson.databind.ObjectMapper();
                Map<String, String> map = mapper.readValue(json, Map.class);
                return map;
            }
        } catch (Exception ignored) {}

        // 失败则尝试正则提取
        String cleaned = responseContent.replaceAll("[\\n\\r]", " ");
        java.util.regex.Matcher verdictMatcher = java.util.regex.Pattern.compile("\"verdict\"\\s*:\\s*\"([^\"]+)\"").matcher(cleaned);
        java.util.regex.Matcher riskMatcher = java.util.regex.Pattern.compile("\"riskLevel\"\\s*:\\s*\"([^\"]+)\"").matcher(cleaned);

        if (verdictMatcher.find()) {
            parsed.put("verdict", verdictMatcher.group(1));
        }
        if (riskMatcher.find()) {
            parsed.put("riskLevel", riskMatcher.group(1));
        }

        return parsed;
    }
}