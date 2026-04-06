package com.mybilibili.admin.utils;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.io.BufferedReader;
import java.io.File;
import java.io.InputStreamReader;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Component
public class FFmpegUtils {

    private static final Logger logger = LoggerFactory.getLogger(FFmpegUtils.class);

    /**
     * 视频转码回调接口
     */
    public interface VideoTranscodeCallback {
        void onTranscodeComplete(String hdPath, String sdPath, String ldPath);
        void onTranscodeError(String errorMessage);
    }

    /**
     * 转码视频为多个清晰度
     * @param inputPath 输入视频路径
     * @param outputDir 输出目录
     * @param videoId 视频ID
     * @param callback 回调接口
     */
    public void transcodeVideo(String inputPath, String outputDir, Integer videoId, VideoTranscodeCallback callback) {
        new Thread(() -> {
            try {
                logger.info("[FFmpeg] 开始转码视频: " + videoId);
                logger.info("[FFmpeg] 输入: " + inputPath);
                logger.info("[FFmpeg] 输出目录: " + outputDir);

                // 确保输出目录存在
                new File(outputDir).mkdirs();

                String hdPath = outputDir + File.separator + "1080p.mp4";
                String sdPath = outputDir + File.separator + "720p.mp4";
                String ldPath = outputDir + File.separator + "480p.mp4";

                // 1. 转码高清 1080p (降低码率：3000k -> 1500k)
                logger.info("[FFmpeg] 转码高清 1080p...");
                boolean hdSuccess = transcodeToResolution(inputPath, hdPath, 1920, 1080, 1500);
                if (!hdSuccess) {
                    callback.onTranscodeError("高清转码失败");
                    return;
                }

                // 2. 转码标清 720p (降低码率：1500k -> 800k)
                logger.info("[FFmpeg] 转码标清 720p...");
                boolean sdSuccess = transcodeToResolution(inputPath, sdPath, 1280, 720, 800);
                if (!sdSuccess) {
                    callback.onTranscodeError("标清转码失败");
                    return;
                }

                // 3. 转码流畅 480p (降低码率：800k -> 400k)
                logger.info("[FFmpeg] 转码流畅 480p...");
                boolean ldSuccess = transcodeToResolution(inputPath, ldPath, 854, 480, 400);
                if (!ldSuccess) {
                    callback.onTranscodeError("流畅转码失败");
                    return;
                }

                logger.info("[FFmpeg] 转码完成: " + videoId);
                callback.onTranscodeComplete(hdPath, sdPath, ldPath);

            } catch (Exception e) {
                logger.error("[FFmpeg] 转码异常: " + e.getMessage(), e);
                callback.onTranscodeError(e.getMessage());
            }
        }).start();
    }

    /**
     * 转码为指定分辨率
     */
    private boolean transcodeToResolution(String inputPath, String outputPath, int width, int height, int videoBitrate) {
        try {
            ProcessBuilder builder = new ProcessBuilder(
                "ffmpeg",
                "-i", inputPath,
                "-vf", "scale=w=" + width + ":h=" + height + ":force_original_aspect_ratio=decrease,pad=" + width + ":" + height + ":(ow-iw)/2:(oh-ih)/2",
                "-c:v", "libx264",
                "-b:v", videoBitrate + "k",
                "-c:a", "aac",
                "-b:a", "128k",
                "-movflags", "+faststart",
                "-y",
                outputPath
            );
            builder.redirectErrorStream(true);

            Process process = builder.start();

            // 读取输出
            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    if (line.contains("time=") || line.contains("size=")) {
                        logger.info("FFmpeg: " + line);
                    }
                }
            }

            int exitCode = process.waitFor();
            logger.info("[FFmpeg] 转码退出码: " + exitCode);

            return exitCode == 0 && new File(outputPath).exists();

        } catch (Exception e) {
            logger.error("[FFmpeg] 转码失败: " + e.getMessage(), e);
            return false;
        }
    }

    /**
     * 获取视频时长（秒）
     */
    public int getVideoDuration(String videoPath) {
        try {
            ProcessBuilder builder = new ProcessBuilder(
                "ffprobe",
                "-v", "error",
                "-show_entries", "format=duration",
                "-of", "default=noprint_wrappers=1:nokey=1",
                videoPath
            );
            builder.redirectErrorStream(true);

            Process process = builder.start();

            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                String line = reader.readLine();
                if (line != null) {
                    double duration = Double.parseDouble(line.trim());
                    return (int) duration;
                }
            }

            process.waitFor();
        } catch (Exception e) {
            System.err.println("[FFmpeg] 获取视频时长失败: " + e.getMessage());
        }
        return 0;
    }

    /**
     * 从视频提取音频
     */
    public boolean extractAudio(String videoPath, String audioPath) {
        try {
            logger.info("[FFmpeg] 提取音频: " + videoPath + " -> " + audioPath);

            // 确保目录存在
            new File(audioPath).getParentFile().mkdirs();

            ProcessBuilder builder = new ProcessBuilder(
                "ffmpeg",
                "-i", videoPath,
                "-vn",                    // 去掉视频
                "-acodec", "pcm_s16le",  // WAV编码（无损）
                "-ar", "16000",           // 采样率16kHz（Whisper推荐）
                "-ac", "1",               // 单声道
                "-y",                     // 覆盖
                audioPath
            );
            builder.redirectErrorStream(true);

            Process process = builder.start();

            // 读取输出
            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    if (line.contains("time=") || line.contains("size=")) {
                        logger.info("FFmpeg: " + line);
                    }
                }
            }

            int exitCode = process.waitFor();
            logger.info("[FFmpeg] 音频提取退出码: " + exitCode);

            return exitCode == 0 && new File(audioPath).exists();

        } catch (Exception e) {
            logger.error("[FFmpeg] 音频提取失败: " + e.getMessage(), e);
            return false;
        }
    }
}
