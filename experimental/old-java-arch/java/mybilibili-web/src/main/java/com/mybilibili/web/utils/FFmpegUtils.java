package com.mybilibili.web.utils;

import org.springframework.stereotype.Component;

import java.io.BufferedReader;
import java.io.File;
import java.io.IOException;
import java.io.InputStreamReader;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

@Component
public class FFmpegUtils {

    private final ExecutorService executorService = Executors.newFixedThreadPool(5);

    /**
     * 转码视频为不同分辨率
     * @param inputPath 输入视频路径
     * @param outputDir 输出目录
     * @param videoId 视频ID
     * @param callback 转码完成回调
     */
    public void transcodeVideo(String inputPath, String outputDir, Integer videoId, VideoTranscodeCallback callback) {
        executorService.submit(() -> {
            try {
                // 创建输出目录
                File outputDirectory = new File(outputDir);
                if (!outputDirectory.exists()) {
                    outputDirectory.mkdirs();
                }

                String hdOutputPath = null;
                String sdOutputPath = null;
                String ldOutputPath = null;
                boolean anySuccess = false;

                // 高清：1080p, 2Mbps
                try {
                    hdOutputPath = outputDir + File.separator + "1080p.mp4";
                    transcode(inputPath, hdOutputPath, "1920x1080", "2000k");
                    anySuccess = true;
                } catch (Exception e) {
                    System.err.println("高清转码失败: " + e.getMessage());
                }

                // 标清：720p, 1Mbps
                try {
                    sdOutputPath = outputDir + File.separator + "720p.mp4";
                    transcode(inputPath, sdOutputPath, "1280x720", "1000k");
                    anySuccess = true;
                } catch (Exception e) {
                    System.err.println("标清转码失败: " + e.getMessage());
                }

                // 流畅：480p, 500Kbps
                try {
                    ldOutputPath = outputDir + File.separator + "480p.mp4";
                    transcode(inputPath, ldOutputPath, "854x480", "500k");
                    anySuccess = true;
                } catch (Exception e) {
                    System.err.println("流畅转码失败: " + e.getMessage());
                }

                // 回调通知转码完成
                if (callback != null) {
                    if (anySuccess) {
                        callback.onTranscodeComplete(hdOutputPath, sdOutputPath, ldOutputPath);
                    } else {
                        callback.onTranscodeError("所有分辨率转码都失败了");
                    }
                }
            } catch (Exception e) {
                e.printStackTrace();
                if (callback != null) {
                    callback.onTranscodeError(e.getMessage());
                }
            }
        });
    }

    /**
     * 执行转码命令
     * @param inputPath 输入路径
     * @param outputPath 输出路径
     * @param resolution 分辨率
     * @param bitrate 比特率
     */
    private void transcode(String inputPath, String outputPath, String resolution, String bitrate) throws IOException, InterruptedException {
        System.out.println("执行转码命令: ffmpeg -i " + inputPath + " -s " + resolution + " -b:v " + bitrate + " -c:v libx264 -c:a aac -b:a 128k -y " + outputPath);
        
        ProcessBuilder builder = new ProcessBuilder(
            "ffmpeg",
            "-i", inputPath,
            "-s", resolution,
            "-b:v", bitrate,
            "-c:v", "libx264",
            "-c:a", "aac",
            "-b:a", "128k",
            "-y",
            outputPath
        );
        builder.redirectErrorStream(true);
        
        Process process = builder.start();
        
        // 读取输出
        try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
            String line;
            while ((line = reader.readLine()) != null) {
                if (line.contains("frame=") || line.contains("time=") || line.contains("speed=")) {
                    System.out.println("FFmpeg: " + line);
                }
            }
        }
        
        int exitCode = process.waitFor();
        System.out.println("FFmpeg退出码: " + exitCode);

        // 检查转码是否成功
        if (exitCode != 0) {
            throw new RuntimeException("转码失败 (退出码=" + exitCode + "): " + inputPath + " -> " + outputPath);
        }
        
        // 检查输出文件是否存在且大小合理
        File outputFile = new File(outputPath);
        if (!outputFile.exists()) {
            throw new RuntimeException("转码失败: 输出文件不存在: " + outputPath);
        }
        if (outputFile.length() < 1024) {
            throw new RuntimeException("转码失败: 输出文件太小 (" + outputFile.length() + "字节): " + outputPath);
        }
        System.out.println("转码成功: " + outputPath + " (" + outputFile.length() + "字节)");
    }

    /**
     * 生成视频封面
     * @param inputPath 输入视频路径
     * @param outputPath 输出封面路径
     */
    private void generateCover(String inputPath, String outputPath) throws IOException, InterruptedException {
        extractCover(inputPath, outputPath, 1);
    }

    /**
     * 提取视频封面
     * @param inputPath 输入视频路径
     * @param outputPath 输出封面路径
     * @param seconds 第几秒提取
     */
    public void extractCover(String inputPath, String outputPath, int seconds) throws IOException, InterruptedException {
        String timeStr = String.format("00:00:%02d", seconds);
        
        ProcessBuilder builder = new ProcessBuilder(
            "ffmpeg",
            "-ss", timeStr,
            "-i", inputPath,
            "-vframes", "1",
            "-y",
            outputPath
        );
        builder.redirectErrorStream(true);
        
        Process process = builder.start();
        
        try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
            String line;
            while ((line = reader.readLine()) != null) {
                if (line.contains("frame=")) {
                    System.out.println("FFmpeg: " + line);
                }
            }
        }
        
        int exitCode = process.waitFor();
        if (exitCode != 0) {
            throw new RuntimeException("生成封面失败 (退出码=" + exitCode + "): " + inputPath);
        }
        
        File outputFile = new File(outputPath);
        if (!outputFile.exists() || outputFile.length() < 100) {
            throw new RuntimeException("封面生成失败: " + outputPath);
        }
        System.out.println("封面生成成功: " + outputPath + " (" + outputFile.length() + "字节)");
    }

    /**
     * 解析视频时长
     * @param videoPath 视频路径
     * @return 时长（秒）
     */
    public int getVideoDuration(String videoPath) throws IOException, InterruptedException {
        String[] cmd = {
            "ffprobe",
            "-v", "error",
            "-show_entries", "format=duration",
            "-of", "default=noprint_wrappers=1:nokey=1",
            videoPath
        };

        Process process = Runtime.getRuntime().exec(cmd);
        try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
            String line = reader.readLine();
            if (line != null) {
                return (int) Math.round(Double.parseDouble(line));
            }
        }
        return 0;
    }

    /**
     * 转码回调接口
     */
    public interface VideoTranscodeCallback {
        void onTranscodeComplete(String hdPath, String sdPath, String ldPath);
        void onTranscodeError(String errorMessage);
    }

    /**
     * 单分辨率转码
     */
    public void transcodeSingleResolution(String inputPath, String outputPath, String resolution, String bitrate) throws IOException, InterruptedException {
        transcode(inputPath, outputPath, resolution, bitrate);
    }
}
