package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.config.UploadFilePathConfig;
import com.mybilibili.ai.service.AudioExtractService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.io.BufferedReader;
import java.io.InputStreamReader;

@Service
public class AudioExtractServiceImpl implements AudioExtractService {

    @Autowired
    private UploadFilePathConfig uploadFilePathConfig;

    @Override
    public boolean extractAudio(Integer manuscriptId, Integer videoId) {
        try {
            String sourceVideoPath = uploadFilePathConfig.getVideoSourcePath(manuscriptId, videoId);
            String audioPath = getAudioPath(manuscriptId, videoId);

            java.io.File videoFile = new java.io.File(sourceVideoPath);
            if (!videoFile.exists()) {
                System.err.println("[音频提取] 源视频文件不存在: " + sourceVideoPath);
                return false;
            }

            java.io.File audioFile = new java.io.File(audioPath);
            java.io.File audioDir = audioFile.getParentFile();
            if (audioDir != null && !audioDir.exists()) {
                audioDir.mkdirs();
            }

            System.out.println("[音频提取] 开始提取音频: " + audioPath);

            ProcessBuilder pb = new ProcessBuilder(
                "ffmpeg",
                "-i", sourceVideoPath,
                "-vn",
                "-acodec", "pcm_s16le",
                "-ar", "16000",
                "-ac", "1",
                "-y",
                audioPath
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

            return exitCode == 0;

        } catch (Exception e) {
            System.err.println("[音频提取] 异常: " + e.getMessage());
            e.printStackTrace();
            return false;
        }
    }

    @Override
    public String getAudioPath(Integer manuscriptId, Integer videoId) {
        return uploadFilePathConfig.getAudioPath(manuscriptId, videoId);
    }
}
