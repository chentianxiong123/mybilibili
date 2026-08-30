package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.entity.AiApiConfig;
import com.mybilibili.ai.service.SttProvider;
import com.mybilibili.ai.service.SttProvider.TranscribeRequest;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.*;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

import java.util.Map;

/**
 * API 类型的 STT 提供者。
 * 通过 ai_api_configs 表中 type='STT' 的渠道配置，
 * 调用第三方 Whisper API（如 Groq、OpenAI 等兼容接口）。
 */
@Component
public class WhisperApiProvider implements SttProvider {

    @Autowired
    private RestTemplate restTemplate;

    @Override
    public String getName() {
        return "whisper-api";
    }

    @Override
    public boolean isAvailable() {
        return true;
    }

    @Override
    public Object invoke(TranscribeRequest request) {
        return null;
    }

    /**
     * 调用 Whisper API
     */
    public String transcribe(AiApiConfig config, String audioPath, String language) {
        try {
            java.io.File audioFile = new java.io.File(audioPath);
            if (!audioFile.exists()) {
                System.err.println("[WhisperApi] 音频文件不存在: " + audioPath);
                return null;
            }

            // 构造 form-data
            org.springframework.util.MultiValueMap<String, Object> form =
                    new org.springframework.util.LinkedMultiValueMap<>();
            form.add("file", new org.springframework.core.io.FileSystemResource(audioFile));
            form.add("model", config.getModel() != null ? config.getModel() : "whisper-1");
            if (language != null) {
                form.add("language", language);
            }

            HttpHeaders headers = new HttpHeaders();
            headers.setContentType(MediaType.MULTIPART_FORM_DATA);
            headers.set("Authorization", "Bearer " + config.getApiKey());

            HttpEntity<org.springframework.util.MultiValueMap<String, Object>> entity =
                    new HttpEntity<>(form, headers);

            ResponseEntity<Map> response = restTemplate.exchange(
                    config.getBaseUrl() + "/v1/audio/transcriptions",
                    HttpMethod.POST,
                    entity,
                    Map.class
            );

            if (response.getStatusCode() == HttpStatus.OK && response.getBody() != null) {
                return (String) response.getBody().get("text");
            }
            return null;
        } catch (Exception e) {
            System.err.println("[WhisperApi] 转写异常: " + e.getMessage());
            return null;
        }
    }
}