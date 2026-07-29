package com.mybilibili.ai.service;

import com.mybilibili.ai.entity.AiApiConfig;

/**
 * STT 提供者接口（语音转文字）
 */
public interface SttProvider extends AiServiceProvider<SttProvider.TranscribeRequest> {

    /**
     * 转写请求
     */
    class TranscribeRequest {
        private String audioPath;
        private String language;
        private AiApiConfig config;

        public String getAudioPath() { return audioPath; }
        public void setAudioPath(String audioPath) { this.audioPath = audioPath; }
        public String getLanguage() { return language; }
        public void setLanguage(String language) { this.language = language; }
        public AiApiConfig getConfig() { return config; }
        public void setConfig(AiApiConfig config) { this.config = config; }

        public static TranscribeRequest of(String audioPath) {
            TranscribeRequest req = new TranscribeRequest();
            req.audioPath = audioPath;
            return req;
        }

        public static TranscribeRequest of(String audioPath, String language) {
            TranscribeRequest req = new TranscribeRequest();
            req.audioPath = audioPath;
            req.language = language;
            return req;
        }

        public static TranscribeRequest of(String audioPath, String language, AiApiConfig config) {
            TranscribeRequest req = of(audioPath, language);
            req.config = config;
            return req;
        }
    }

    @Override
    default String getType() { return "ASR"; }

    default boolean supports(AiApiConfig config) {
        if (config == null) {
            return false;
        }
        String providerName = normalize(getName());
        String configName = normalize(config.getName());
        String model = normalize(config.getModel());
        return contains(configName, providerName)
                || contains(model, providerName)
                || contains(providerName, configName)
                || contains(providerName, model);
    }

    private static boolean contains(String value, String token) {
        return token != null && !token.isBlank() && value != null && value.contains(token);
    }

    private static String normalize(String value) {
        return value == null ? "" : value.toLowerCase().replace("_", "-").replace(" ", "-");
    }
}
