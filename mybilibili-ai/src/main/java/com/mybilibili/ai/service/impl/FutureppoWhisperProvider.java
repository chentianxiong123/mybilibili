package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.entity.AiApiConfig;
import com.mybilibili.ai.service.AiApiConfigService;
import com.mybilibili.ai.service.SttProvider;
import com.mybilibili.ai.service.SttProvider.TranscribeRequest;
import com.mybilibili.ai.util.AiUsageLogger;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.core.io.FileSystemResource;
import org.springframework.http.*;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;
import org.springframework.web.client.RestTemplate;

import java.io.File;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.*;

/**
 * Futureppo Whisper ASR 提供者。
 * 使用 /v1/audio/transcriptions 接口，model=whisper-large-v3。
 * 返回 verbose_json 格式带 segments（start/end/text），转为 SRT。
 */
@Component
public class FutureppoWhisperProvider implements SttProvider {

    private final RestTemplate restTemplate = new RestTemplate();

    @Autowired(required = false)
    private AiApiConfigService aiApiConfigService;

    @Autowired(required = false)
    private AiUsageLogger aiUsageLogger;

    private volatile AiApiConfig activeConfig;

    @Override
    public String getName() {
        return "futureppo-whisper";
    }

    @Override
    public boolean isAvailable() {
        loadConfig();
        return activeConfig != null && activeConfig.getEnabled();
    }

    private void loadConfig() {
        if (activeConfig != null) return;
        if (aiApiConfigService == null) return;
        activeConfig = aiApiConfigService.getConfigForFeature("TRANSCRIBE");
    }

    @Override
    public Object invoke(TranscribeRequest request) {
        return transcribe(request.getAudioPath(), request.getLanguage());
    }

    public String transcribe(String audioPath, String language) {
        long start = System.currentTimeMillis();
        try {
            loadConfig();

            File audioFile = new File(audioPath);
            if (!audioFile.exists()) {
                System.err.println("[FutureppoWhisper] 音频文件不存在: " + audioPath);
                return null;
            }

            String baseUrl;
            String apiKey;
            String model;
            if (activeConfig != null) {
                baseUrl = activeConfig.getBaseUrl() != null ? activeConfig.getBaseUrl() : "";
                apiKey = activeConfig.getApiKey() != null ? activeConfig.getApiKey() : "";
                model = activeConfig.getModel() != null ? activeConfig.getModel() : "whisper-large-v3";
            } else {
                System.err.println("[FutureppoWhisper] 未找到 TRANSCRIBE 渠道配置");
                return null;
            }

            HttpHeaders headers = new HttpHeaders();
            headers.setContentType(MediaType.MULTIPART_FORM_DATA);
            headers.set("Authorization", "Bearer " + apiKey);

            MultiValueMap<String, Object> form = new LinkedMultiValueMap<>();
            form.add("file", new FileSystemResource(audioFile));
            form.add("model", model);
            form.add("language", language != null ? language : "zh");
            form.add("response_format", "verbose_json");

            HttpEntity<MultiValueMap<String, Object>> entity = new HttpEntity<>(form, headers);

            ResponseEntity<Map<String, Object>> response = restTemplate.exchange(
                    baseUrl + "/v1/audio/transcriptions",
                    HttpMethod.POST,
                    entity,
                    new org.springframework.core.ParameterizedTypeReference<>() {}
            );

            if (response.getStatusCode() == HttpStatus.OK && response.getBody() != null) {
                Map<String, Object> body = response.getBody();
                String srt = buildSrtFromVerboseJson(body);
                if (srt != null && !srt.isEmpty()) {
                    if (aiUsageLogger != null) {
                        aiUsageLogger.log("TRANSCRIBE", model, null, null, System.currentTimeMillis() - start, true, null);
                    }
                    return srt;
                }
                // verbose_json 没有 segments 时，使用接口返回的 text 字段。
                Object textObj = body.get("text");
                if (textObj != null) {
                    if (aiUsageLogger != null) {
                        aiUsageLogger.log("TRANSCRIBE", model, null, null, System.currentTimeMillis() - start, true, null);
                    }
                    return textObj.toString();
                }
                if (aiUsageLogger != null) {
                    aiUsageLogger.log("TRANSCRIBE", model, null, null, System.currentTimeMillis() - start, false, "empty response");
                }
                return null;
            }
            if (aiUsageLogger != null) {
                aiUsageLogger.log("TRANSCRIBE", model, null, null, System.currentTimeMillis() - start, false, "non-200 status");
            }
            return null;
        } catch (Exception e) {
            System.err.println("[FutureppoWhisper] 转写异常: " + e.getMessage());
            if (aiUsageLogger != null) {
                aiUsageLogger.log("TRANSCRIBE", "whisper-large-v3", null, null, System.currentTimeMillis() - start, false, e.getMessage());
            }
            return null;
        }
    }

    @SuppressWarnings("unchecked")
    private String buildSrtFromVerboseJson(Map<String, Object> body) {
        Object segmentsObj = body.get("segments");
        if (segmentsObj == null) return null;

        List<Map<String, Object>> segments = (List<Map<String, Object>>) segmentsObj;
        if (segments.isEmpty()) return null;

        StringBuilder srt = new StringBuilder();
        int index = 1;
        for (Map<String, Object> seg : segments) {
            Object startObj = seg.get("start");
            Object endObj = seg.get("end");
            Object textObj = seg.get("text");
            if (startObj == null || endObj == null || textObj == null) continue;

            double startSec = parseDouble(startObj);
            double endSec = parseDouble(endObj);
            String text = textObj.toString().trim();
            if (text.isEmpty()) continue;

            srt.append(index++).append("\n");
            srt.append(formatSrtTime(startSec)).append(" --> ").append(formatSrtTime(endSec)).append("\n");
            srt.append(text).append("\n\n");
        }
        return srt.toString();
    }

    private double parseDouble(Object val) {
        if (val instanceof Number) return ((Number) val).doubleValue();
        return Double.parseDouble(val.toString());
    }

    private String formatSrtTime(double seconds) {
        int h = (int) (seconds / 3600);
        int m = (int) ((seconds % 3600) / 60);
        int s = (int) (seconds % 60);
        int ms = (int) ((seconds % 1) * 1000);
        return String.format("%02d:%02d:%02d,%03d", h, m, s, ms);
    }
}
