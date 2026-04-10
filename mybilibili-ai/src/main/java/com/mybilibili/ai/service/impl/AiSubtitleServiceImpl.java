package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.config.UploadFilePathConfig;
import com.mybilibili.ai.config.WhisperConfig;
import com.mybilibili.ai.repository.SubtitleRepository;
import com.mybilibili.ai.service.AiSubtitleService;
import com.mybilibili.ai.utils.SubtitleTextUtils;
import com.mybilibili.common.entity.Subtitle;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Service
public class AiSubtitleServiceImpl implements AiSubtitleService {

    @Autowired
    private UploadFilePathConfig uploadFilePathConfig;

    @Autowired
    private WhisperConfig whisperConfig;

    @Autowired
    private SubtitleRepository subtitleRepository;

    private static final Pattern TIME_PATTERN = Pattern.compile(
            "(\\d{2}):\\s*(\\d{2}):\\s*(\\d{2})[,.](\\d{3})\\s*-->\\s*(\\d{2}):\\s*(\\d{2}):\\s*(\\d{2})[,.](\\d{3})"
    );

    @Override
    public boolean generateSubtitle(Integer manuscriptId, Integer videoId) {
        try {
            String sourceVideoPath = uploadFilePathConfig.getVideoSourcePath(manuscriptId, videoId);
            String audioPath = uploadFilePathConfig.getAudioPath(manuscriptId, videoId);
            String subtitleDir = uploadFilePathConfig.getVideoSubtitleDir(manuscriptId, videoId);

            uploadFilePathConfig.ensureDirectoryExists(uploadFilePathConfig.getVideoAudioDir(manuscriptId, videoId));
            uploadFilePathConfig.ensureDirectoryExists(subtitleDir);

            if (!extractAudio(sourceVideoPath, audioPath)) {
                System.err.println("[字幕生成] 音频提取失败");
                return false;
            }

            if (!generateSubtitleWithWhisper(audioPath, subtitleDir)) {
                System.err.println("[字幕生成] Whisper生成字幕失败");
                return false;
            }

            String subtitlePath = uploadFilePathConfig.getChineseSubtitlePath(manuscriptId, videoId);
            if (importSystemSubtitle(videoId, subtitlePath)) {
                System.out.println("[字幕生成] 字幕已存入MongoDB");
                return true;
            } else {
                System.err.println("[字幕生成] 字幕存入MongoDB失败");
                return false;
            }

        } catch (Exception e) {
            System.err.println("[字幕生成] 生成字幕异常: " + e.getMessage());
            e.printStackTrace();
            return false;
        }
    }

    private boolean importSystemSubtitle(Integer videoId, String subtitlePath) {
        try {
            java.io.File subtitleFile = new java.io.File(subtitlePath);
            if (!subtitleFile.exists()) {
                System.err.println("[MongoDB] 字幕文件不存在: " + subtitlePath);
                return false;
            }

            String srtContent = new String(Files.readAllBytes(Paths.get(subtitlePath)), StandardCharsets.UTF_8);
            System.out.println("[MongoDB] SRT文件内容长度: " + srtContent.length());
            System.out.println("[MongoDB] SRT文件前500字符: " + srtContent.substring(0, Math.min(500, srtContent.length())));
            
            List<Subtitle.SubtitleItem> items = parseSrtContent(srtContent);
            System.out.println("[MongoDB] 解析后字幕条数: " + items.size());
            
            if (items.isEmpty()) {
                System.err.println("[MongoDB] 警告: 解析出的字幕条数为0!");
                return false;
            }

            Subtitle subtitle = new Subtitle();
            subtitle.setVideoId(videoId);
            subtitle.setLanguage("zh-CN");
            subtitle.setLanguageName("中文");
            subtitle.setFormat("srt");
            subtitle.setContent(items);
            subtitle.setIsDefault(true);
            subtitle.setSource("whisper");
            subtitle.setStatus(3);
            subtitle.setUploadTime(new Date());

            subtitleRepository.save(subtitle);
            System.out.println("[MongoDB] 字幕保存成功，videoId: " + videoId);
            return true;

        } catch (Exception e) {
            System.err.println("[MongoDB] 保存字幕失败: " + e.getMessage());
            e.printStackTrace();
            return false;
        }
    }

    private List<Subtitle.SubtitleItem> parseSrtContent(String srtContent) {
        List<Subtitle.SubtitleItem> items = new ArrayList<>();
        if (srtContent == null || srtContent.isEmpty()) {
            System.err.println("[解析SRT] SRT内容为空");
            return items;
        }

        srtContent = srtContent.replace("\r\n", "\n").replace("\r", "\n");
        String[] blocks = srtContent.split("\n\n+");
        System.out.println("[解析SRT] 分割后的块数: " + blocks.length);
        
        for (int blockIndex = 0; blockIndex < blocks.length; blockIndex++) {
            String block = blocks[blockIndex];
            String[] lines = block.split("\n");
            System.out.println("[解析SRT] 块" + blockIndex + " 行数: " + lines.length);
            if (lines.length < 3) {
                System.out.println("[解析SRT] 块" + blockIndex + " 行数不足3，跳过");
                continue;
            }

            try {
                Subtitle.SubtitleItem item = new Subtitle.SubtitleItem();

                String indexLine = lines[0].trim();
                System.out.println("[解析SRT] 块" + blockIndex + " index: " + indexLine);
                item.setIndex(Integer.parseInt(indexLine));

                String timeLine = lines[1].trim();
                System.out.println("[解析SRT] 块" + blockIndex + " 时间行: " + timeLine);
                Matcher matcher = TIME_PATTERN.matcher(timeLine);
                if (matcher.find()) {
                    double startTime = parseTimeToSeconds(
                        matcher.group(1), matcher.group(2),
                        matcher.group(3), matcher.group(4)
                    );
                    double endTime = parseTimeToSeconds(
                        matcher.group(5), matcher.group(6),
                        matcher.group(7), matcher.group(8)
                    );
                    System.out.println("[解析SRT] 块" + blockIndex + " 时间: " + startTime + " -> " + endTime);
                    item.setStartTime(startTime);
                    item.setEndTime(endTime);
                } else {
                    System.out.println("[解析SRT] 块" + blockIndex + " 时间行匹配失败");
                }

                StringBuilder textBuilder = new StringBuilder();
                for (int i = 2; i < lines.length; i++) {
                    String text = lines[i].trim();
                    if (!text.isEmpty()) {
                        if (textBuilder.length() > 0) {
                            textBuilder.append("\\n");
                        }
                        textBuilder.append(text);
                    }
                }
                item.setText(textBuilder.toString());
                System.out.println("[解析SRT] 块" + blockIndex + " 文本: " + textBuilder.toString().substring(0, Math.min(50, textBuilder.length())));

                items.add(item);
            } catch (Exception e) {
                System.err.println("[解析SRT] 块" + blockIndex + " 解析失败: " + e.getMessage());
            }
        }

        return items;
    }

    private double parseTimeToSeconds(String hours, String minutes, String seconds, String millis) {
        int h = Integer.parseInt(hours);
        int m = Integer.parseInt(minutes);
        int s = Integer.parseInt(seconds);
        int ms = Integer.parseInt(millis);
        return h * 3600 + m * 60 + s + ms / 1000.0;
    }

    private boolean extractAudio(String videoPath, String audioPath) {
        try {
            java.io.File videoFile = new java.io.File(videoPath);
            if (!videoFile.exists()) {
                System.err.println("[音频提取] 源视频文件不存在: " + videoPath);
                return false;
            }

            java.io.File audioFile = new java.io.File(audioPath);
            java.io.File audioDir = audioFile.getParentFile();
            if (audioDir != null && !audioDir.exists()) {
                audioDir.mkdirs();
            }

            ProcessBuilder pb = new ProcessBuilder(
                    "ffmpeg",
                    "-i", videoPath,
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
                    System.out.println("[ffmpeg] " + line);
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

    private boolean generateSubtitleWithWhisper(String audioPath, String outputDir) {
        try {
            System.out.println("[Whisper] 开始生成字幕...");
            System.out.println("[Whisper] 音频: " + audioPath);
            System.out.println("[Whisper] 输出目录: " + outputDir);

            String[] cmd = {
                    whisperConfig.getCliPath(),
                    "-m", whisperConfig.getModelPath(),
                    "-f", audioPath,
                    "-l", whisperConfig.getLanguage(),
                    "-osrt",
                    "-of", "zh-CN",
                    "-t", String.valueOf(whisperConfig.getThreads()),
                    "-pp"
            };

            ProcessBuilder pb = new ProcessBuilder(cmd);
            pb.directory(new java.io.File(outputDir));
            pb.redirectErrorStream(true);

            Process process = pb.start();

            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    System.out.println("[Whisper] " + line);
                }
            }

            int exitCode = process.waitFor();
            System.out.println("[Whisper] 退出码: " + exitCode);

            return exitCode == 0;

        } catch (Exception e) {
            System.err.println("[Whisper] 生成字幕异常: " + e.getMessage());
            e.printStackTrace();
            return false;
        }
    }

    @Override
    public String getSubtitlePath(Integer manuscriptId, Integer videoId) {
        return uploadFilePathConfig.getChineseSubtitlePath(manuscriptId, videoId);
    }
}
