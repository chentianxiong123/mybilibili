package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.entity.AiApiConfig;
import com.mybilibili.ai.service.SttProvider;
import com.mybilibili.ai.service.SttProvider.TranscribeRequest;
import com.mybilibili.ai.util.AiUsageLogger;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.http.*;
import org.springframework.web.client.RestTemplate;

import java.io.File;
import java.util.Date;

/**
 * 智谱 ASR (glm-asr) 语音转文字提供者。
 * 使用小水管 API 平台的 /v1/audio/transcriptions 接口。
 */
@Slf4j
@Component
public class GlmAsrProvider implements SttProvider {

    private final RestTemplate restTemplate;

    public GlmAsrProvider(RestTemplate restTemplate) {
        this.restTemplate = restTemplate;
    }

    @Autowired(required = false)
    private AiUsageLogger aiUsageLogger;

    @Override
    public String getName() {
        return "glm-asr";
    }

    @Override
    public boolean isAvailable() {
        return true;
    }

    @Override
    public Object invoke(TranscribeRequest request) {
        return transcribe(request.getConfig(), request.getAudioPath(), request.getLanguage());
    }

    /**
     * 调用智谱 ASR API
     */
    public String transcribe(AiApiConfig config, String audioPath, String language) {
        long start = System.currentTimeMillis();
        String model = null;
        try {
            File audioFile = new File(audioPath);
            if (!audioFile.exists()) {
                log.warn("[GlmAsr] 音频文件不存在: {}", audioPath);
                return null;
            }

            String baseUrl;
            String apiKey;
            if (config != null) {
                baseUrl = config.getBaseUrl() != null ? config.getBaseUrl() : "";
                apiKey = config.getApiKey() != null ? config.getApiKey() : "";
                model = config.getModel() != null ? config.getModel() : "glm-asr";
            } else {
                log.warn("[GlmAsr] 未找到 TRANSCRIBE 渠道配置");
                return null;
            }

            org.springframework.util.MultiValueMap<String, Object> form =
                    new org.springframework.util.LinkedMultiValueMap<>();
            form.add("file", new org.springframework.core.io.FileSystemResource(audioFile));
            form.add("model", model);
            if (language != null) {
                form.add("language", language);
            }

            HttpHeaders headers = new HttpHeaders();
            headers.setContentType(MediaType.MULTIPART_FORM_DATA);
            headers.set("Authorization", "Bearer " + apiKey);

            HttpEntity<org.springframework.util.MultiValueMap<String, Object>> entity =
                    new HttpEntity<>(form, headers);

            ResponseEntity<java.util.Map> response = restTemplate.exchange(
                    baseUrl + "/v1/audio/transcriptions",
                    HttpMethod.POST,
                    entity,
                    java.util.Map.class
            );

            if (response.getStatusCode() == HttpStatus.OK && response.getBody() != null) {
                String text = (String) response.getBody().get("text");
                if (aiUsageLogger != null) {
                    aiUsageLogger.log("TRANSCRIBE", model, null, null, System.currentTimeMillis() - start, true, null);
                }
                return text;
            }
            if (aiUsageLogger != null) {
                aiUsageLogger.log("TRANSCRIBE", model, null, null, System.currentTimeMillis() - start, false, "empty response");
            }
            return null;
        } catch (Exception e) {
            log.warn("[GlmAsr] 转写异常: {}", e.getMessage());
            if (aiUsageLogger != null) {
                aiUsageLogger.log("TRANSCRIBE", model, null, null, System.currentTimeMillis() - start, false, e.getMessage());
            }
            return null;
        }
    }
}
