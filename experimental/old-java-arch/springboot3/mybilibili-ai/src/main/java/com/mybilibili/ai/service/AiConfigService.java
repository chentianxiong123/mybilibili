package com.mybilibili.ai.service;

import com.mybilibili.ai.config.DeepSeekConfig;
import com.mybilibili.ai.config.DynamicChatClient;
import com.mybilibili.ai.util.AiUsageLogger;
import lombok.extern.slf4j.Slf4j;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.*;

@Slf4j
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

    @Autowired
    private DynamicChatClient dynamicChatClient;

    @Autowired
    private AiUsageLogger aiUsageLogger;

    public String getApiKey() {
        String redisKey = redisTemplate.opsForValue().get(REDIS_KEY_API_KEY);
        if (redisKey != null && !redisKey.isEmpty()) {
            return redisKey;
        }
        return deepSeekConfig.getApiKey();
    }

    /** 按属性短名取配置，如 "api-url"、"model" */
    public String getConfigValue(String shortKey) {
        String redisKey = "ai:config:deepseek-" + shortKey;
        String value = redisTemplate.opsForValue().get(redisKey);
        if (value != null && !value.isEmpty()) return value;
        return switch (shortKey) {
            case "api-url" -> deepSeekConfig.getApiUrl();
            case "model" -> deepSeekConfig.getModel();
            default -> null;
        };
    }

    public int getConfigInt(String shortKey, int defaultValue) {
        String redisKey = "ai:config:deepseek-" + shortKey;
        String value = redisTemplate.opsForValue().get(redisKey);
        if (value != null && !value.isEmpty()) {
            try { return Integer.parseInt(value); } catch (NumberFormatException ignored) {}
        }
        if ("max-tokens".equals(shortKey)) return deepSeekConfig.getMaxTokens();
        return defaultValue;
    }

    public double getConfigDouble(String shortKey, double defaultValue) {
        String redisKey = "ai:config:deepseek-" + shortKey;
        String value = redisTemplate.opsForValue().get(redisKey);
        if (value != null && !value.isEmpty()) {
            try { return Double.parseDouble(value); } catch (NumberFormatException ignored) {}
        }
        if ("temperature".equals(shortKey)) return deepSeekConfig.getTemperature();
        return defaultValue;
    }

    public Map<String, Object> getConfig() {
        Map<String, Object> config = new LinkedHashMap<>();
        config.put("apiKey", maskApiKey(getApiKey()));
        config.put("apiUrl", normalizeApiUrl(getConfigValue("api-url")));
        config.put("model", getConfigValue("model"));
        config.put("maxTokens", getConfigInt("max-tokens", 2000));
        config.put("temperature", getConfigDouble("temperature", 0.7));
        config.put("hasApiKey", getApiKey() != null && !getApiKey().isEmpty());
        return config;
    }

    /**
     * 统一 URL 为根地址格式，如 https://api.qnaigc.com
     * Spring AI OpenAiApi 会自动拼 /v1/chat/completions
     */
    private String normalizeApiUrl(String url) {
        if (url == null) return url;
        return url.replaceAll("/chat/completions$", "")
                  .replaceAll("/v1$", "")
                  .replaceAll("/$", "");
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
                redisTemplate.opsForValue().set(REDIS_KEY_API_URL, normalizeApiUrl(value));
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

        // 配置变更后重建所有 ChatClient
        dynamicChatClient.rebuildAll();
    }

    public Map<String, Object> testConnection(Map<String, String> config) {
        long startTime = System.currentTimeMillis();
        Map<String, Object> result = new LinkedHashMap<>();

        try {
            String prompt = config != null && config.get("text") != null && !config.get("text").isEmpty()
                    ? config.get("text") : "你好，请回复'API测试成功'";

            // 优先按 channelId 直接测，其次按 feature 绑定，最后用第一个活跃渠道
            ChatClient client = null;
            if (config != null && config.get("channelId") != null) {
                Long channelId = Long.valueOf(config.get("channelId"));
                client = dynamicChatClient.getClientById(channelId);
            }
            if (client == null) {
                String feature = config != null ? config.get("feature") : null;
                if (feature != null) {
                    client = dynamicChatClient.getClient(feature);
                }
            }
            if (client == null) {
                client = dynamicChatClient.getFirstActiveClient();
            }
            if (client == null) {
                result.put("success", false);
                result.put("message", "没有可用的 API 渠道，请先配置渠道");
                return result;
            }

            String responseContent = client.prompt()
                    .user(prompt)
                    .call()
                    .content();

            aiUsageLogger.log("TEST", "deepseek-r1", null, null, System.currentTimeMillis() - startTime, responseContent != null, null);

            long responseTime = System.currentTimeMillis() - startTime;
            result.put("responseTime", responseTime + "ms");

            if (responseContent != null && !responseContent.isEmpty()) {
                result.put("success", true);
                result.put("message", "API连接成功");
                result.put("response", responseContent);
            } else {
                result.put("success", false);
                result.put("message", "API返回空响应");
            }

        } catch (Exception e) {
            log.error("AI连接测试失败: {}", e.getMessage(), e);
            aiUsageLogger.log("TEST", null, null, null, System.currentTimeMillis() - startTime, false, e.getMessage());
            long responseTime = System.currentTimeMillis() - startTime;
            result.put("success", false);
            result.put("message", "API调用异常: " + e.getMessage());
            result.put("responseTime", responseTime + "ms");
        }

        return result;
    }

    private String maskApiKey(String apiKey) {
        if (apiKey == null || apiKey.length() < 8) {
            return apiKey;
        }
        return apiKey.substring(0, 6) + "****" + apiKey.substring(apiKey.length() - 4);
    }
}
