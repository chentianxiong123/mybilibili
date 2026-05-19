package com.mybilibili.ai.service;

import com.mybilibili.ai.config.DeepSeekConfig;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.util.*;

@Service
public class AiConfigService {

    private static final String REDIS_KEY_API_KEY = "ai:config:deepseek-api-key";
    private static final String REDIS_KEY_API_URL = "ai:config:deepseek-api-url";
    private static final String REDIS_KEY_MODEL = "ai:config:deepseek-model";
    private static final String REDIS_KEY_MAX_TOKENS = "ai:config:deepseek-max-tokens";
    private static final String REDIS_KEY_TEMPERATURE = "ai:config:deepseek-temperature";

    @Autowired
    private DeepSeekConfig deepSeekConfig;

    @Autowired
    private StringRedisTemplate redisTemplate;

    private final RestTemplate restTemplate = new RestTemplate();

    public String getApiKey() {
        String redisKey = redisTemplate.opsForValue().get(REDIS_KEY_API_KEY);
        if (redisKey != null && !redisKey.isEmpty()) {
            return redisKey;
        }
        return deepSeekConfig.getApiKey();
    }

    public Map<String, Object> getConfig() {
        Map<String, Object> config = new LinkedHashMap<>();
        config.put("apiKey", maskApiKey(getApiKey()));
        config.put("apiUrl", getConfigValue(REDIS_KEY_API_URL, deepSeekConfig.getApiUrl()));
        config.put("model", getConfigValue(REDIS_KEY_MODEL, deepSeekConfig.getModel()));
        config.put("maxTokens", getConfigInt(REDIS_KEY_MAX_TOKENS, deepSeekConfig.getMaxTokens()));
        config.put("temperature", getConfigDouble(REDIS_KEY_TEMPERATURE, deepSeekConfig.getTemperature()));
        config.put("hasApiKey", getApiKey() != null && !getApiKey().isEmpty());
        return config;
    }

    public void updateConfig(Map<String, Object> updates) {
        if (updates.containsKey("apiKey")) {
            String value = (String) updates.get("apiKey");
            if (value != null && !value.isEmpty()) {
                redisTemplate.opsForValue().set(REDIS_KEY_API_KEY, value);
            }
        }
        if (updates.containsKey("apiUrl")) {
            String value = (String) updates.get("apiUrl");
            if (value != null && !value.isEmpty()) {
                redisTemplate.opsForValue().set(REDIS_KEY_API_URL, value);
            }
        }
        if (updates.containsKey("model")) {
            String value = (String) updates.get("model");
            if (value != null && !value.isEmpty()) {
                redisTemplate.opsForValue().set(REDIS_KEY_MODEL, value);
            }
        }
        if (updates.containsKey("maxTokens")) {
            Object value = updates.get("maxTokens");
            if (value != null) {
                redisTemplate.opsForValue().set(REDIS_KEY_MAX_TOKENS, String.valueOf(value));
            }
        }
        if (updates.containsKey("temperature")) {
            Object value = updates.get("temperature");
            if (value != null) {
                redisTemplate.opsForValue().set(REDIS_KEY_TEMPERATURE, String.valueOf(value));
            }
        }
    }

    public Map<String, Object> testConnection(String testText) {
        long startTime = System.currentTimeMillis();
        Map<String, Object> result = new LinkedHashMap<>();

        try {
            String apiKey = getApiKey();
            if (apiKey == null || apiKey.isEmpty()) {
                result.put("success", false);
                result.put("message", "API密钥未配置");
                return result;
            }

            String apiUrl = getConfigValue(REDIS_KEY_API_URL, deepSeekConfig.getApiUrl());
            String model = getConfigValue(REDIS_KEY_MODEL, deepSeekConfig.getModel());

            Map<String, Object> requestBody = new HashMap<>();
            requestBody.put("model", model);

            List<Map<String, String>> messages = new ArrayList<>();
            Map<String, String> message = new HashMap<>();
            message.put("role", "user");
            message.put("content", testText != null && !testText.isEmpty() ? testText : "你好，请回复'API测试成功'");
            messages.add(message);

            requestBody.put("messages", messages);
            requestBody.put("max_tokens", 100);

            HttpHeaders headers = new HttpHeaders();
            headers.setContentType(MediaType.APPLICATION_JSON);
            headers.set("Authorization", "Bearer " + apiKey);

            HttpEntity<Map<String, Object>> request = new HttpEntity<>(requestBody, headers);
            ResponseEntity<Map> response = restTemplate.postForEntity(apiUrl, request, Map.class);

            long responseTime = System.currentTimeMillis() - startTime;
            result.put("responseTime", responseTime + "ms");

            if (response.getStatusCode() == HttpStatus.OK && response.getBody() != null) {
                String responseContent = parseResponse(response.getBody());
                result.put("success", true);
                result.put("message", "API连接成功");
                result.put("response", responseContent);
            } else {
                result.put("success", false);
                result.put("message", "API返回错误，状态码: " + response.getStatusCode());
            }

        } catch (Exception e) {
            long responseTime = System.currentTimeMillis() - startTime;
            result.put("success", false);
            result.put("message", "API调用异常: " + e.getMessage());
            result.put("responseTime", responseTime + "ms");
        }

        return result;
    }

    private String getConfigValue(String redisKey, String defaultValue) {
        String value = redisTemplate.opsForValue().get(redisKey);
        return value != null && !value.isEmpty() ? value : defaultValue;
    }

    private int getConfigInt(String redisKey, int defaultValue) {
        String value = redisTemplate.opsForValue().get(redisKey);
        if (value != null && !value.isEmpty()) {
            try {
                return Integer.parseInt(value);
            } catch (NumberFormatException ignored) {}
        }
        return defaultValue;
    }

    private double getConfigDouble(String redisKey, double defaultValue) {
        String value = redisTemplate.opsForValue().get(redisKey);
        if (value != null && !value.isEmpty()) {
            try {
                return Double.parseDouble(value);
            } catch (NumberFormatException ignored) {}
        }
        return defaultValue;
    }

    private String maskApiKey(String apiKey) {
        if (apiKey == null || apiKey.length() < 8) {
            return apiKey;
        }
        return apiKey.substring(0, 6) + "****" + apiKey.substring(apiKey.length() - 4);
    }

    @SuppressWarnings("unchecked")
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
            return "无法解析API响应";
        } catch (Exception e) {
            return "解析响应异常: " + e.getMessage();
        }
    }
}