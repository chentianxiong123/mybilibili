package com.mybilibili.ai.config;

import com.mybilibili.ai.entity.AiApiConfig;
import com.mybilibili.ai.entity.AiBinding;
import com.mybilibili.ai.mapper.AiApiConfigMapper;
import com.mybilibili.ai.mapper.AiBindingMapper;
import com.mybilibili.ai.service.SttProvider;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import jakarta.annotation.PostConstruct;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

/**
 * STT 提供者管理器。
 * 自动扫描所有 SttProvider 实现，按类型匹配 ai_api_configs 表中的渠道配置。
 */
@Slf4j
@Component
public class DynamicSttClient {

    @Autowired
    private AiApiConfigMapper aiApiConfigMapper;

    @Autowired
    private AiBindingMapper aiBindingMapper;

    @Autowired
    private List<SttProvider> providers;

    /** type → SttProvider 缓存 */
    private final Map<String, SttProvider> providerCache = new ConcurrentHashMap<>();

    @PostConstruct
    public void init() {
        refreshProviders();
    }

    /**
     * 获取 STT 提供者
     * 优先使用绑定配置，其次降级到本地 Whisper
     */
    public SttProvider getProvider() {
        SttProvider provider = providerCache.get("STT");
        if (provider != null) {
            return provider;
        }
        // 降级：返回本地 Whisper
        return providers.stream()
                .filter(p -> p.getName().contains("local") && p.isAvailable())
                .findFirst()
                .orElse(null);
    }

    /**
     * 获取指定渠道的 STT 提供者
     */
    public SttProvider getProviderByChannelId(Long channelId) {
        AiApiConfig config = aiApiConfigMapper.selectById(channelId);
        if (config == null) return null;

        // 本地 Whisper 渠道
        if ("本地 Whisper".equals(config.getName()) || !config.getEnabled()) {
            // 返回本地 Whisper 提供者
            for (SttProvider provider : providers) {
                if (provider.getName().contains("local") && provider.isAvailable()) {
                    return provider;
                }
            }
        }

        if (!config.getEnabled()) return null;

        // 查找匹配类型和渠道的提供者
        for (SttProvider provider : providers) {
            if (provider.getName().equalsIgnoreCase(config.getName())
                    || config.getName().toLowerCase().contains(provider.getName().replace("-", "_"))) {
                return provider;
            }
        }
        return null;
    }

    /**
     * 刷新提供者缓存
     */
    public void refreshProviders() {
        List<AiApiConfig> sttConfigs = aiApiConfigMapper.selectList(
                new com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper<AiApiConfig>()
                        .eq(AiApiConfig::getType, "ASR")
                        .eq(AiApiConfig::getEnabled, true)
        );

        providerCache.clear();
        for (AiApiConfig config : sttConfigs) {
            // 匹配对应名称的提供者
            for (SttProvider provider : providers) {
                String providerName = provider.getName().toLowerCase();
                String configName = config.getName().toLowerCase();
                if (providerName.equals(configName)
                        || configName.contains(providerName)
                        || providerName.contains(configName.replace("-", "_"))) {
                    providerCache.put("STT", provider);
                    log.info("STT 提供者已加载: {} (渠道: {})", provider.getName(), config.getName());
                }
            }
        }

        if (providerCache.isEmpty()) {
            // 无 API 配置时，查找本地提供者
            providers.stream()
                    .filter(SttProvider::isAvailable)
                    .findFirst()
                    .ifPresent(p -> {
                        providerCache.put("STT", p);
                        log.info("STT 提供者: {} (本地)", p.getName());
                    });
        }
    }

    /**
     * 通过 API 转写音频，返回纯文本或 SRT 格式
     */
    public String transcribeWithApi(String audioPath, String language) {
        // 优先使用 TRANSCRIBE 绑定的渠道
        AiBinding binding = aiBindingMapper.selectByFeature("TRANSCRIBE");
        if (binding != null) {
            AiApiConfig config = aiApiConfigMapper.selectById(binding.getApiConfigId());
            if (config != null && config.getEnabled()) {
                for (SttProvider provider : providers) {
                    if ("glm-asr".equals(provider.getName())) {
                        Object result = provider.invoke(SttProvider.TranscribeRequest.of(audioPath, language));
                        return result != null ? result.toString() : null;
                    }
                    // futureppo-whisper 返回 SRT 格式，带精确时间戳
                    if ("futureppo-whisper".equals(provider.getName())) {
                        Object result = provider.invoke(SttProvider.TranscribeRequest.of(audioPath, language));
                        return result != null ? result.toString() : null;
                    }
                }
            }
        }
        // 降级
        SttProvider provider = getProvider();
        if (provider == null) {
            System.err.println("[DynamicSttClient] 无可用的 STT 提供者");
            return null;
        }
        Object result = provider.invoke(SttProvider.TranscribeRequest.of(audioPath, language));
        return result != null ? result.toString() : null;
    }

    /**
     * 判断转写结果是否为 SRT 格式（带时间戳）
     */
    public boolean isSrtFormat(String result) {
        if (result == null || result.isEmpty()) return false;
        return result.contains("-->");
    }

    /**
     * 将纯文本转换为简化的 SRT 格式（无精确时间戳，每行一段字幕）
     */
    public String textToSrt(String text, double durationSeconds) {
        if (text == null || text.isEmpty()) return "";
        String[] lines = text.split("\n");
        StringBuilder srt = new StringBuilder();
        double totalDuration = durationSeconds > 0 ? durationSeconds : lines.length * 3.0;
        double perLine = totalDuration / Math.max(lines.length, 1);
        for (int i = 0; i < lines.length; i++) {
            String line = lines[i].trim();
            if (line.isEmpty()) continue;
            double start = i * perLine;
            double end = (i + 1) * perLine;
            srt.append(i + 1).append("\n");
            srt.append(formatSrtTime(start)).append(" --> ").append(formatSrtTime(end)).append("\n");
            srt.append(line).append("\n\n");
        }
        return srt.toString();
    }

    private String formatSrtTime(double seconds) {
        int h = (int) (seconds / 3600);
        int m = (int) ((seconds % 3600) / 60);
        int s = (int) (seconds % 60);
        int ms = (int) ((seconds % 1) * 1000);
        return String.format("%02d:%02d:%02d,%03d", h, m, s, ms);
    }
}