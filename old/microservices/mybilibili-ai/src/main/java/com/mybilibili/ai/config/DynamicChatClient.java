package com.mybilibili.ai.config;

import com.mybilibili.ai.entity.AiApiConfig;
import com.mybilibili.ai.service.AiApiConfigService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.openai.OpenAiChatModel;
import org.springframework.ai.openai.OpenAiChatOptions;
import org.springframework.ai.openai.api.OpenAiApi;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import jakarta.annotation.PostConstruct;
import java.util.concurrent.ConcurrentHashMap;

/**
 * 多渠道 ChatClient 管理器。
 * 每个渠道（ai_api_configs 表的一行）对应一个独立的 ChatClient。
 * 每个 AI 功能（CHAT/REVIEW/SUMMARY）绑定到一个渠道。
 */
@Slf4j
@Component
public class DynamicChatClient {

    @Autowired
    private AiApiConfigService aiApiConfigService;

    /** channelId → ChatClient 缓存 */
    private final ConcurrentHashMap<Long, ChatClient> clientCache = new ConcurrentHashMap<>();

    @PostConstruct
    public void init() {
        rebuildAll();
    }

    /** 按渠道 ID 直接获取 ChatClient */
    public ChatClient getClientById(Long channelId) {
        AiApiConfig config = aiApiConfigService.getById(channelId);
        if (config == null || !config.getEnabled()) return null;
        return clientCache.computeIfAbsent(config.getId(), id -> buildClient(config));
    }

    /** 获取第一个活跃渠道的 ChatClient */
    public ChatClient getFirstActiveClient() {
        for (AiApiConfig config : aiApiConfigService.listAll()) {
            if (config.getEnabled()) {
                return clientCache.computeIfAbsent(config.getId(), id -> buildClient(config));
            }
        }
        return null;
    }

    /** 获取功能绑定的渠道对应的 ChatClient */
    public ChatClient getClient(String feature) {
        AiApiConfig config = aiApiConfigService.getConfigForFeature(feature);
        if (config == null || !config.getEnabled()) {
            log.warn("功能 [{}] 未绑定或渠道已禁用", feature);
            return null;
        }
        return clientCache.computeIfAbsent(config.getId(), id -> buildClient(config));
    }

    /** 重建指定渠道的 ChatClient（管理员修改配置后调用） */
    public void rebuildChannel(Long channelId) {
        AiApiConfig config = aiApiConfigService.getById(channelId);
        if (config == null) {
            clientCache.remove(channelId);
            return;
        }
        clientCache.put(channelId, buildClient(config));
        log.info("渠道 [{}] ChatClient 已重建: model={}", config.getName(), config.getModel());
    }

    /** 重建所有渠道的 ChatClient */
    public void rebuildAll() {
        clientCache.clear();
        for (AiApiConfig config : aiApiConfigService.listAll()) {
            if (config.getEnabled()) {
                clientCache.put(config.getId(), buildClient(config));
            }
        }
        log.info("所有渠道 ChatClient 已重建，共 {} 个活跃渠道", clientCache.size());
    }

    /** 从数据库构建 ChatClient */
    private ChatClient buildClient(AiApiConfig config) {
        OpenAiApi openAiApi = OpenAiApi.builder()
                .apiKey(config.getApiKey())
                .baseUrl(config.getBaseUrl())
                .build();

        OpenAiChatOptions options = OpenAiChatOptions.builder()
                .model(config.getModel())
                .maxTokens(config.getMaxTokens())
                .temperature(config.getTemperature())
                .build();

        OpenAiChatModel chatModel = OpenAiChatModel.builder()
                .openAiApi(openAiApi)
                .defaultOptions(options)
                .build();

        return ChatClient.builder(chatModel).build();
    }
}
