package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.config.WhisperConfig;
import com.mybilibili.ai.service.SttProvider;
import com.mybilibili.ai.service.SttProvider.TranscribeRequest;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.io.BufferedReader;
import java.io.InputStreamReader;

/**
 * 本地 Whisper CLI STT 提供者。
 */
@Slf4j
@Component
public class WhisperLocalProvider implements SttProvider {

    @Autowired
    private WhisperConfig whisperConfig;

    @Override
    public String getName() {
        return "whisper-local";
    }

    @Override
    public boolean isAvailable() {
        try {
            ProcessBuilder pb = new ProcessBuilder(whisperConfig.getCliPath(), "--version");
            pb.redirectErrorStream(true);
            Process process = pb.start();
            int exitCode = process.waitFor();
            return exitCode == 0;
        } catch (Exception e) {
            return false;
        }
    }

    @Override
    public Object invoke(TranscribeRequest request) {
        return transcribe(request);
    }

    public String transcribe(TranscribeRequest request) {
        try {
            String audioPath = request.getAudioPath();
            String lang = request.getLanguage() != null ? request.getLanguage() : whisperConfig.getLanguage();

            ProcessBuilder pb = new ProcessBuilder(
                whisperConfig.getCliPath(),
                "-m", whisperConfig.getModelPath(),
                "-l", lang,
                "--threads", String.valueOf(whisperConfig.getThreads()),
                "--fp16", "false",
                "-otxt", "-o", ".",
                audioPath
            );
            pb.redirectErrorStream(true);
            Process process = pb.start();

            StringBuilder output = new StringBuilder();
            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    if (!line.startsWith("[")) {
                        output.append(line).append("\n");
                    }
                }
            }

            int exitCode = process.waitFor();
            if (exitCode == 0) {
                String txtPath = audioPath + ".txt";
                java.io.File txtFile = new java.io.File(txtPath);
                if (txtFile.exists()) {
                    String content = java.nio.file.Files.readString(txtFile.toPath());
                    txtFile.delete();
                    return content.trim();
                }
                return output.toString().trim();
            }
            return null;
        } catch (Exception e) {
            log.warn("[Whisper] 转写异常: {}", e.getMessage());
            return null;
        }
    }
}
