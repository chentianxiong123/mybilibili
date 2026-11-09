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

/**
 * STT 提供者管理器。
 * 自动扫描所有 SttProvider 实现，按 TRANSCRIBE 绑定的渠道配置匹配提供者。
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

    @PostConstruct
    public void init() {
        refreshProviders();
    }

    /**
     * 获取 STT 提供者。
     * 只使用 TRANSCRIBE 明确绑定的渠道，不做运行时降级。
     */
    public SttProvider getProvider() {
        AiApiConfig config = getTranscribeConfig();
        return getProvider(config);
    }

    /**
     * 获取指定渠道的 STT 提供者
     */
    public SttProvider getProviderByChannelId(Long channelId) {
        AiApiConfig config = aiApiConfigMapper.selectById(channelId);
        if (config == null || !config.getEnabled()) return null;
        return getProvider(config);
    }

    /**
     * 刷新提供者缓存
     */
    public void refreshProviders() {
        SttProvider provider = getProvider();
        if (provider != null) {
            log.info("STT 提供者已加载: {}", provider.getName());
        } else {
            log.info("STT 提供者未配置");
        }
    }

    /**
     * 通过 API 转写音频，返回纯文本或 SRT 格式
     */
    public String transcribeWithApi(String audioPath, String language) {
        AiApiConfig config = getTranscribeConfig();
        SttProvider provider = getProvider(config);
        if (provider == null) {
            log.warn("[DynamicSttClient] 无可用的 STT 提供者");
            return null;
        }
        Object result = provider.invoke(SttProvider.TranscribeRequest.of(audioPath, language, config));
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

    private AiApiConfig getTranscribeConfig() {
        AiBinding binding = aiBindingMapper.selectByFeature("TRANSCRIBE");
        if (binding == null) {
            return null;
        }
        AiApiConfig config = aiApiConfigMapper.selectById(binding.getApiConfigId());
        if (config == null || !config.getEnabled()) {
            return null;
        }
        return config;
    }

    private SttProvider getProvider(AiApiConfig config) {
        if (config == null) {
            return null;
        }
        SttProvider provider = findProvider(config);
        return provider != null && provider.isAvailable() ? provider : null;
    }

    private SttProvider findProvider(AiApiConfig config) {
        if (config == null) {
            return null;
        }
        for (SttProvider candidate : providers) {
            if (candidate.supports(config)) {
                return candidate;
            }
        }
        return null;
    }

    private String formatSrtTime(double seconds) {
        int h = (int) (seconds / 3600);
        int m = (int) ((seconds % 3600) / 60);
        int s = (int) (seconds % 60);
        int ms = (int) ((seconds % 1) * 1000);
        return String.format("%02d:%02d:%02d,%03d", h, m, s, ms);
    }
}
