package com.mybilibili.videomedia.process;

import org.springframework.stereotype.Component;

import java.io.BufferedReader;
import java.io.File;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;

@Component
public class FfmpegMediaTool {

    public interface ProgressListener {
        void onProgress(double ratio);
    }

    public boolean transcodeToHLS(String inputPath,
                                  String outputDir,
                                  String playlistName,
                                  int width,
                                  int height,
                                  ProgressListener listener) {
        try {
            new File(outputDir).mkdirs();
            String playlistPath = outputDir + File.separator + playlistName + ".m3u8";
            double durationSeconds = getDurationSeconds(inputPath);

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
                    playlistPath
            );
            builder.redirectErrorStream(true);

            Process process = builder.start();
            double lastRatio = -1;
            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    double ratio = parseProgressRatio(line, durationSeconds);
                    if (ratio >= 0 && ratio > lastRatio) {
                        lastRatio = ratio;
                        if (listener != null) {
                            listener.onProgress(ratio);
                        }
                    }
                }
            }

            int exitCode = process.waitFor();
            if (exitCode == 0 && listener != null) {
                listener.onProgress(1.0d);
            }
            return exitCode == 0 && new File(playlistPath).exists();
        } catch (Exception e) {
            return false;
        }
    }

    public boolean extractAudio(String inputPath, String audioPath) {
        try {
            File audioFile = new File(audioPath);
            File parent = audioFile.getParentFile();
            if (parent != null) {
                parent.mkdirs();
            }

            ProcessBuilder builder = new ProcessBuilder(
                    "ffmpeg",
                    "-i", inputPath,
                    "-vn",
                    "-acodec", "pcm_s16le",
                    "-ar", "16000",
                    "-ac", "1",
                    "-y",
                    audioPath
            );
            builder.redirectErrorStream(true);

            Process process = builder.start();
            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
                while (reader.readLine() != null) {
                    // Drain process output to avoid blocking.
                }
            }
            return process.waitFor() == 0 && audioFile.exists() && audioFile.length() > 0;
        } catch (Exception e) {
            return false;
        }
    }

    private double getDurationSeconds(String inputPath) {
        try {
            ProcessBuilder builder = new ProcessBuilder(
                    "ffprobe",
                    "-v", "error",
                    "-show_entries", "format=duration",
                    "-of", "default=noprint_wrappers=1:nokey=1",
                    inputPath
            );
            builder.redirectErrorStream(true);
            Process process = builder.start();
            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
                String line = reader.readLine();
                if (line != null && !line.isBlank()) {
                    return Double.parseDouble(line.trim());
                }
            }
            process.waitFor();
        } catch (Exception ignored) {
            // Progress becomes coarse when duration cannot be probed.
        }
        return 0;
    }

    private double parseProgressRatio(String line, double durationSeconds) {
        if (line == null) {
            return -1;
        }
        if ("progress=end".equalsIgnoreCase(line.trim())) {
            return 1.0d;
        }
        if (durationSeconds <= 0 || !line.startsWith("out_time_ms=")) {
            return -1;
        }
        try {
            long outTimeMicros = Long.parseLong(line.substring("out_time_ms=".length()).trim());
            double ratio = outTimeMicros / 1_000_000d / durationSeconds;
            if (Double.isNaN(ratio) || Double.isInfinite(ratio)) {
                return -1;
            }
            return Math.max(0, Math.min(1, ratio));
        } catch (NumberFormatException e) {
            return -1;
        }
    }
}
