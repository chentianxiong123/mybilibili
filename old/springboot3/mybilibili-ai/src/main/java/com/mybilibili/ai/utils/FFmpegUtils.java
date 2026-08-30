package com.mybilibili.ai.utils;

import org.springframework.stereotype.Component;

import java.io.BufferedReader;
import java.io.File;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.Locale;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Component
public class FFmpegUtils {

    public interface VideoTranscodeCallback {
        void onTranscodeComplete(String hdPath, String sdPath, String ldPath);
        void onTranscodeError(String errorMessage);
    }

    public interface ProgressListener {
        void onProgress(double ratio);
    }

    private static final Pattern TIME_PATTERN = Pattern.compile("time=(\\d+):(\\d+):(\\d+(?:\\.\\d+)?)");

    public void transcodeVideo(String inputPath, String outputDir, Integer videoId, VideoTranscodeCallback callback) {
        new Thread(() -> {
            try {
                System.out.println("[FFmpeg] 开始HLS转码视频: " + videoId);
                System.out.println("[FFmpeg] 输入: " + inputPath);
                System.out.println("[FFmpeg] 输出目录: " + outputDir);

                File outputFile = new File(outputDir);
                if (outputFile.exists()) {
                    deleteDirectory(outputFile);
                }
                outputFile.mkdirs();

                String hdDir = outputDir + File.separator + "1080p";
                String sdDir = outputDir + File.separator + "720p";
                String ldDir = outputDir + File.separator + "480p";

                new File(hdDir).mkdirs();
                new File(sdDir).mkdirs();
                new File(ldDir).mkdirs();

                System.out.println("[FFmpeg] 转码 1080p...");
                boolean hdSuccess = transcodeToHLS(inputPath, hdDir, "playlist", 1920, 1080);
                if (!hdSuccess) {
                    callback.onTranscodeError("1080p转码失败");
                    return;
                }

                System.out.println("[FFmpeg] 转码 720p...");
                boolean sdSuccess = transcodeToHLS(inputPath, sdDir, "playlist", 1280, 720);
                if (!sdSuccess) {
                    callback.onTranscodeError("720p转码失败");
                    return;
                }

                System.out.println("[FFmpeg] 转码 480p...");
                boolean ldSuccess = transcodeToHLS(inputPath, ldDir, "playlist", 854, 480);
                if (!ldSuccess) {
                    callback.onTranscodeError("480p转码失败");
                    return;
                }

                System.out.println("[FFmpeg] HLS转码完成: " + videoId);
                callback.onTranscodeComplete(hdDir, sdDir, ldDir);

            } catch (Exception e) {
                System.err.println("[FFmpeg] 转码异常: " + e.getMessage());
                e.printStackTrace();
                callback.onTranscodeError(e.getMessage());
            }
        }).start();
    }

    public boolean transcodeToHLS(String inputPath, String outputDir, String playlistName, int width, int height) {
        return transcodeToHLS(inputPath, outputDir, playlistName, width, height, null);
    }

    public boolean transcodeToHLS(String inputPath, String outputDir, String playlistName, int width, int height, ProgressListener progressListener) {
        try {
            System.out.println("[FFmpeg-HLS] 开始转码 " + width + "p: " + inputPath + " -> " + outputDir);

            new File(outputDir).mkdirs();

            String m3u8Path = outputDir + File.separator + playlistName + ".m3u8";
            double durationSeconds = getVideoDurationSeconds(inputPath);

            ProcessBuilder builder = new ProcessBuilder(
                "ffmpeg",
                "-i", inputPath,
                "-vf", "scale=" + width + ":-2",
                "-c:v", "libx264",
                "-preset", "fast",
                "-b:v", "2000k",
                "-c:a", "aac",
                "-b:a", "128k",
                "-hls_time", "10",
                "-hls_list_size", "0",
                "-hls_segment_filename", outputDir + File.separator + "%03d.ts",
                "-progress", "pipe:1",
                "-nostats",
                "-y",
                m3u8Path
            );
            builder.redirectErrorStream(true);

            Process process = builder.start();

            double lastRatio = -1;
            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    System.out.println("[FFmpeg-HLS] " + line);
                    double ratio = parseProgressRatio(line, durationSeconds);
                    if (ratio >= 0 && ratio > lastRatio) {
                        lastRatio = ratio;
                        if (progressListener != null) {
                            progressListener.onProgress(ratio);
                        }
                    }
                }
            }

            int exitCode = process.waitFor();
            System.out.println("[FFmpeg-HLS] " + width + "p 转码退出码: " + exitCode);

            if (exitCode == 0 && progressListener != null) {
                progressListener.onProgress(1.0d);
            }

            return exitCode == 0 && new File(m3u8Path).exists();

        } catch (Exception e) {
            System.err.println("[FFmpeg-HLS] " + width + "p 转码失败: " + e.getMessage());
            e.printStackTrace();
            return false;
        }
    }

    public int getVideoDuration(String videoPath) {
        return (int) getVideoDurationSeconds(videoPath);
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

    private double getVideoDurationSeconds(String videoPath) {
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

            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
                String line = reader.readLine();
                if (line != null) {
                    return Double.parseDouble(line.trim());
                }
            }

            process.waitFor();
        } catch (Exception e) {
            System.err.println("[FFmpeg] 获取视频时长失败: " + e.getMessage());
        }
        return 0;
    }

    private double parseProgressRatio(String line, double durationSeconds) {
        if (durationSeconds <= 0) {
            Matcher matcher = TIME_PATTERN.matcher(line);
            if (matcher.find()) {
                return -1;
            }
            return parseOutTimeRatioWithoutDuration(line);
        }

        if (line.startsWith("out_time_ms=")) {
            String value = line.substring("out_time_ms=".length()).trim();
            try {
                double outTimeSeconds = Long.parseLong(value) / 1_000_000d;
                return normalizeRatio(outTimeSeconds / durationSeconds);
            } catch (NumberFormatException ignored) {
                return -1;
            }
        }

        if (line.startsWith("out_time=")) {
            String value = line.substring("out_time=".length()).trim();
            return normalizeRatio(parseTimestampSeconds(value) / durationSeconds);
        }

        Matcher matcher = TIME_PATTERN.matcher(line);
        if (matcher.find()) {
            return normalizeRatio(parseTimestampSeconds(matcher.group(0).substring("time=".length())) / durationSeconds);
        }

        if (line.toLowerCase(Locale.ROOT).startsWith("progress=end")) {
            return 1.0d;
        }

        return -1;
    }

    private double parseOutTimeRatioWithoutDuration(String line) {
        if (line.toLowerCase(Locale.ROOT).startsWith("progress=end")) {
            return 1.0d;
        }
        return -1;
    }

    private double parseTimestampSeconds(String value) {
        String[] parts = value.split(":");
        if (parts.length != 3) {
            return 0;
        }
        double hours = Double.parseDouble(parts[0]);
        double minutes = Double.parseDouble(parts[1]);
        double seconds = Double.parseDouble(parts[2]);
        return hours * 3600 + minutes * 60 + seconds;
    }

    private double normalizeRatio(double ratio) {
        if (Double.isNaN(ratio) || Double.isInfinite(ratio)) {
            return -1;
        }
        if (ratio < 0) {
            return 0;
        }
        if (ratio > 1) {
            return 1;
        }
        return ratio;
    }

    private boolean deleteDirectory(File dir) {
        if (dir.exists()) {
            File[] files = dir.listFiles();
            if (files != null) {
                for (File file : files) {
                    if (file.isDirectory()) {
                        deleteDirectory(file);
                    } else {
                        file.delete();
                    }
                }
            }
        }
        return dir.delete();
    }
}
