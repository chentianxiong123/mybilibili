package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.config.UploadFilePathConfig;
import com.mybilibili.ai.service.VideoTranscodeService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.io.BufferedReader;
import java.io.InputStreamReader;

@Service
public class VideoTranscodeServiceImpl implements VideoTranscodeService {

    @Autowired
    private UploadFilePathConfig uploadFilePathConfig;

    @Override
    public boolean transcode(Integer manuscriptId, Integer videoId) {
        try {
            String sourceVideoPath = uploadFilePathConfig.getVideoSourcePath(manuscriptId, videoId);
            String transcodedDir = uploadFilePathConfig.getTranscodedDir(manuscriptId, videoId);

            java.io.File videoFile = new java.io.File(sourceVideoPath);
            if (!videoFile.exists()) {
                System.err.println("[转码] 源视频文件不存在: " + sourceVideoPath);
                return false;
            }

            uploadFilePathConfig.ensureDirectoryExists(transcodedDir);

            boolean hdSuccess = transcodeToQuality(sourceVideoPath, transcodedDir, "hd", 1920, 1080, 600);
            boolean sdSuccess = transcodeToQuality(sourceVideoPath, transcodedDir, "sd", 1280, 720, 300);
            boolean ldSuccess = transcodeToQuality(sourceVideoPath, transcodedDir, "ld", 854, 480, 200);

            return hdSuccess || sdSuccess || ldSuccess;

        } catch (Exception e) {
            System.err.println("[转码] 转码异常: " + e.getMessage());
            e.printStackTrace();
            return false;
        }
    }

    private boolean transcodeToQuality(String sourcePath, String outputDir, String quality, int width, int height, int bitrate) {
        try {
            String fileName;
            switch (quality) {
                case "hd": fileName = "1080p.mp4"; break;
                case "sd": fileName = "720p.mp4"; break;
                case "ld": fileName = "480p.mp4"; break;
                default: fileName = quality + ".mp4";
            }
            String outputPath = outputDir + "/" + fileName;

            System.out.println("[转码] 开始转码 " + quality + ": " + outputPath);

            String scaleFilter = "scale=w=" + width + ":h=" + height + ":force_original_aspect_ratio=decrease,pad=" + width + ":" + height + ":(ow-iw)/2:(oh-ih)/2";

            ProcessBuilder pb = new ProcessBuilder(
                "ffmpeg",
                "-i", sourcePath,
                "-vf", scaleFilter,
                "-c:v", "libx264",
                "-preset", "fast",
                "-b:v", bitrate + "k",
                "-c:a", "aac",
                "-b:a", "96k",
                "-movflags", "+faststart",
                "-y",
                outputPath
            );
            pb.redirectErrorStream(true);

            Process process = pb.start();

            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    System.out.println("[ffmpeg-" + quality + "] " + line);
                }
            }

            int exitCode = process.waitFor();
            System.out.println("[转码] " + quality + " 退出码: " + exitCode);

            return exitCode == 0;

        } catch (Exception e) {
            System.err.println("[转码] " + quality + " 转码失败: " + e.getMessage());
            e.printStackTrace();
            return false;
        }
    }

    @Override
    public String getVideoPath(Integer manuscriptId, Integer videoId, String quality) {
        return uploadFilePathConfig.getTranscodedDir(manuscriptId, videoId) + "/" + quality + ".mp4";
    }
}
