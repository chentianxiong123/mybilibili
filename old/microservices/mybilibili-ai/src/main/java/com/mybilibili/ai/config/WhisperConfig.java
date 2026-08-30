package com.mybilibili.ai.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Configuration;

@Configuration
public class WhisperConfig {

    @Value("${ai.whisper.cli-path:whisper}")
    private String cliPath;

    @Value("${ai.whisper.model-path:base}")
    private String modelPath;

    @Value("${ai.whisper.language:zh}")
    private String language;

    @Value("${ai.whisper.threads:4}")
    private int threads;

    public String getCliPath() {
        return cliPath;
    }

    public String getModelPath() {
        return modelPath;
    }

    public String getLanguage() {
        return language;
    }

    public int getThreads() {
        return threads;
    }
}