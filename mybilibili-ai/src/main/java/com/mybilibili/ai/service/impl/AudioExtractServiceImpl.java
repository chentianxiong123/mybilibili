package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.service.AudioExtractService;
import com.mybilibili.ai.service.VideoProcessingStorageService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.nio.file.Files;
import java.nio.file.Path;

@Service
public class AudioExtractServiceImpl implements AudioExtractService {

    @Autowired
    private VideoProcessingStorageService processingStorageService;

    @Override
    public boolean extractAudio(Integer manuscriptId, Integer videoId) {
        try {
            Path sourceVideoPath = processingStorageService.materializeSourceVideo(manuscriptId, videoId);
            Path audioPath = processingStorageService.getAudioPath(manuscriptId, videoId);
            Files.createDirectories(audioPath.getParent());

            System.out.println("[音频提取] 开始提取音频: " + audioPath);

            ProcessBuilder pb = new ProcessBuilder(
                "ffmpeg",
                "-i", sourceVideoPath.toString(),
                "-vn",
                "-acodec", "pcm_s16le",
                "-ar", "16000",
                "-ac", "1",
                "-y",
                audioPath.toString()
            );
            pb.redirectErrorStream(true);

            Process process = pb.start();

            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    System.out.println("[ffmpeg-audio] " + line);
                }
            }

            int exitCode = process.waitFor();
            System.out.println("[音频提取] 退出码: " + exitCode);

            if (exitCode != 0 || !Files.exists(audioPath) || Files.size(audioPath) == 0) {
                return false;
            }

            processingStorageService.uploadAudio(manuscriptId, videoId, audioPath);
            return true;

        } catch (Exception e) {
            System.err.println("[音频提取] 异常: " + e.getMessage());
            e.printStackTrace();
            return false;
        }
    }

    @Override
    public String getAudioPath(Integer manuscriptId, Integer videoId) {
        return processingStorageService.getAudioPath(manuscriptId, videoId).toString();
    }
}
