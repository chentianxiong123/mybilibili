package com.mybilibili.ai.service;

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

        public String getAudioPath() { return audioPath; }
        public void setAudioPath(String audioPath) { this.audioPath = audioPath; }
        public String getLanguage() { return language; }
        public void setLanguage(String language) { this.language = language; }

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
    }

    @Override
    default String getType() { return "STT"; }
}