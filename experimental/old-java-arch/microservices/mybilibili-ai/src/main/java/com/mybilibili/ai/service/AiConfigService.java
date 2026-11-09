package com.mybilibili.ai.service;

import com.mybilibili.ai.config.DynamicChatClient;
import com.mybilibili.ai.entity.AiApiConfig;
import com.mybilibili.ai.util.AiUsageLogger;
import lombok.extern.slf4j.Slf4j;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.LinkedHashMap;
import java.util.Map;

@Slf4j
@Service
public class AiConfigService {

    @Autowired
    private DynamicChatClient dynamicChatClient;

    @Autowired
    private AiApiConfigService aiApiConfigService;

    @Autowired
    private AiUsageLogger aiUsageLogger;

    public Map<String, Object> testConnection(Map<String, String> config) {
        long startTime = System.currentTimeMillis();
        Map<String, Object> result = new LinkedHashMap<>();

        try {
            String prompt = config != null && config.get("text") != null && !config.get("text").isEmpty()
                    ? config.get("text") : "你好，请回复'API测试成功'";

            // 优先按 channelId 直接测，其次按 feature 绑定，最后用第一个活跃渠道
            ChatClient client = null;
            String model = null;
            if (config != null && config.get("channelId") != null) {
                Long channelId = Long.valueOf(config.get("channelId"));
                AiApiConfig channelConfig = aiApiConfigService.getById(channelId);
                model = modelOf(channelConfig);
                client = dynamicChatClient.getClientById(channelId);
            }
            if (client == null) {
                String feature = config != null ? config.get("feature") : null;
                if (feature != null) {
                    AiApiConfig featureConfig = aiApiConfigService.getConfigForFeature(feature);
                    model = modelOf(featureConfig);
                    client = dynamicChatClient.getClient(feature);
                }
            }
            if (client == null) {
                AiApiConfig fallbackConfig = firstActiveConfig();
                model = modelOf(fallbackConfig);
                client = fallbackConfig != null ? dynamicChatClient.getClientById(fallbackConfig.getId()) : null;
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

            aiUsageLogger.log("TEST", model, null, null, System.currentTimeMillis() - startTime, responseContent != null, null);

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

    private AiApiConfig firstActiveConfig() {
        for (AiApiConfig config : aiApiConfigService.listAll()) {
            if (Boolean.TRUE.equals(config.getEnabled())) {
                return config;
            }
        }
        return null;
    }

    private String modelOf(AiApiConfig config) {
        return config != null ? config.getModel() : null;
    }
}
