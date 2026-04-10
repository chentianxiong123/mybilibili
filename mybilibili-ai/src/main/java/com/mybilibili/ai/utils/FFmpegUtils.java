package com.mybilibili.ai.utils;

import org.springframework.stereotype.Component;

import java.io.BufferedReader;
import java.io.File;
import java.io.InputStreamReader;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Component
public class FFmpegUtils {

    public interface VideoTranscodeCallback {
        void onTranscodeComplete(String hdPath, String sdPath, String ldPath);
        void onTranscodeError(String errorMessage);
    }

    public void transcodeVideo(String inputPath, String outputDir, Integer videoId, VideoTranscodeCallback callback) {
        new Thread(() -> {
            try {
                System.out.println("[FFmpeg] 开始转码视频: " + videoId);
                System.out.println("[FFmpeg] 输入: " + inputPath);
                System.out.println("[FFmpeg] 输出目录: " + outputDir);

                new File(outputDir).mkdirs();

                String hdPath = outputDir + File.separator + "1080p.mp4";
                String sdPath = outputDir + File.separator + "720p.mp4";
                String ldPath = outputDir + File.separator + "480p.mp4";

                System.out.println("[FFmpeg] 转码高清 1080p...");
                boolean hdSuccess = transcodeToResolution(inputPath, hdPath, 1920, 1080, 1500);
                if (!hdSuccess) {
                    callback.onTranscodeError("高清转码失败");
                    return;
                }

                System.out.println("[FFmpeg] 转码标清 720p...");
                boolean sdSuccess = transcodeToResolution(inputPath, sdPath, 1280, 720, 800);
                if (!sdSuccess) {
                    callback.onTranscodeError("标清转码失败");
                    return;
                }

                System.out.println("[FFmpeg] 转码流畅 480p...");
                boolean ldSuccess = transcodeToResolution(inputPath, ldPath, 854, 480, 400);
                if (!ldSuccess) {
                    callback.onTranscodeError("流畅转码失败");
                    return;
                }

                System.out.println("[FFmpeg] 转码完成: " + videoId);
                callback.onTranscodeComplete(hdPath, sdPath, ldPath);

            } catch (Exception e) {
                System.err.println("[FFmpeg] 转码异常: " + e.getMessage());
                e.printStackTrace();
                callback.onTranscodeError(e.getMessage());
            }
        }).start();
    }

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

            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    if (line.contains("time=") || line.contains("size=")) {
                        System.out.println("FFmpeg: " + line);
                    }
                }
            }

            int exitCode = process.waitFor();
            System.out.println("[FFmpeg] 转码退出码: " + exitCode);

            return exitCode == 0 && new File(outputPath).exists();

        } catch (Exception e) {
            System.err.println("[FFmpeg] 转码失败: " + e.getMessage());
            e.printStackTrace();
            return false;
        }
    }

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

    public boolean extractAudio(String videoPath, String audioPath) {
        try {
            System.out.println("[FFmpeg] 提取音频: " + videoPath + " -> " + audioPath);

            new File(audioPath).getParentFile().mkdirs();

            ProcessBuilder builder = new ProcessBuilder(
                "ffmpeg",
                "-i", videoPath,
                "-vn",
                "-acodec", "pcm_s16le",
                "-ar", "16000",
                "-ac", "1",
                "-y",
                audioPath
            );
            builder.redirectErrorStream(true);

            Process process = builder.start();

            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    if (line.contains("time=") || line.contains("size=")) {
                        System.out.println("FFmpeg: " + line);
                    }
                }
            }

            int exitCode = process.waitFor();
            System.out.println("[FFmpeg] 音频提取退出码: " + exitCode);

            return exitCode == 0 && new File(audioPath).exists();

        } catch (Exception e) {
            System.err.println("[FFmpeg] 音频提取失败: " + e.getMessage());
            e.printStackTrace();
            return false;
        }
    }
}
