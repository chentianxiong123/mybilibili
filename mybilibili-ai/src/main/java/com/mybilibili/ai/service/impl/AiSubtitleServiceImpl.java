package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.config.UploadFilePathConfig;
import com.mybilibili.ai.config.WhisperConfig;
import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.repository.SubtitleRepository;
import com.mybilibili.ai.service.AiSubtitleService;
import com.mybilibili.ai.utils.SubtitleTextUtils;
import com.mybilibili.common.entity.Subtitle;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
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

    private static final Logger log = LoggerFactory.getLogger(AiSubtitleServiceImpl.class);

    @Autowired
    private UploadFilePathConfig uploadFilePathConfig;

    @Autowired
    private WhisperConfig whisperConfig;

    @Autowired
    private SubtitleRepository subtitleRepository;

    @Autowired
    private VideoMapper videoMapper;

    private static final Pattern TIME_PATTERN = Pattern.compile(
            "(\\d{2}):\\s*(\\d{2}):\\s*(\\d{2})[,.](\\d{3})\\s*-->\\s*(\\d{2}):\\s*(\\d{2}):\\s*(\\d{2})[,.](\\d{3})"
    );

    @Override
    public boolean generateSubtitle(Integer manuscriptId, Integer videoId) {
        return generateSubtitle(manuscriptId, videoId, null);
    }

    @Override
    public boolean generateSubtitle(Integer manuscriptId, Integer videoId, ProgressListener progressListener) {
        try {
            String sourceVideoPath = uploadFilePathConfig.getVideoSourcePath(manuscriptId, videoId);
            String audioPath = uploadFilePathConfig.getAudioPath(manuscriptId, videoId);
            String subtitleDir = uploadFilePathConfig.getVideoSubtitleDir(manuscriptId, videoId);

            uploadFilePathConfig.ensureDirectoryExists(uploadFilePathConfig.getVideoAudioDir(manuscriptId, videoId));
            uploadFilePathConfig.ensureDirectoryExists(subtitleDir);

            pushProgress(progressListener, 5, "提取音频中");

            if (!extractAudio(sourceVideoPath, audioPath)) {
                pushProgress(progressListener, 0, "音频提取失败");
                return false;
            }
            pushProgress(progressListener, 30, "音频提取完成，开始Whisper识别");

            if (!generateSubtitleWithWhisper(manuscriptId, videoId, audioPath, subtitleDir, progressListener)) {
                pushProgress(progressListener, 0, "Whisper识别失败");
                return false;
            }
            pushProgress(progressListener, 90, "Whisper识别完成，导入字幕");

            String subtitlePath = uploadFilePathConfig.getChineseSubtitlePath(manuscriptId, videoId);
            if (importSystemSubtitle(videoId, subtitlePath)) {
                pushProgress(progressListener, 100, "字幕生成完成");
                return true;
            } else {
                pushProgress(progressListener, 0, "字幕导入失败");
                return false;
            }

        } catch (Exception e) {
            pushProgress(progressListener, 0, "字幕生成异常");
            return false;
        }
    }

    private void pushProgress(ProgressListener listener, int percent, String stageText) {
        if (listener != null) {
            listener.onProgress(percent, stageText);
        }
    }

    private boolean importSystemSubtitle(Integer videoId, String subtitlePath) {
        try {
            log.info("[字幕导入] 开始导入, videoId={}, subtitlePath={}", videoId, subtitlePath);
            java.io.File subtitleFile = new java.io.File(subtitlePath);
            if (!subtitleFile.exists()) {
                log.error("[字幕导入] 字幕文件不存在: {}", subtitlePath);
                return false;
            }

            String srtContent = new String(Files.readAllBytes(Paths.get(subtitlePath)), StandardCharsets.UTF_8);
            log.info("[字幕导入] 读取字幕文件成功, 长度={}", srtContent.length());

            List<Subtitle.SubtitleItem> items = parseSrtContent(srtContent);
            log.info("[字幕导入] 解析字幕条目, count={}", items.size());

            if (items.isEmpty()) {
                log.error("[字幕导入] 字幕内容为空");
                return false;
            }

            List<Subtitle> existing = subtitleRepository.findAllByVideoIdAndLanguage(videoId, "zh-CN");
            if (existing != null && !existing.isEmpty()) {
                log.info("[字幕导入] 删除旧字幕记录, count={}", existing.size());
                subtitleRepository.deleteAll(existing);
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

            log.info("[字幕导入] 保存字幕到MongoDB, videoId={}", videoId);
            subtitleRepository.save(subtitle);
            log.info("[字幕导入] 保存成功");
            return true;

        } catch (Exception e) {
            log.error("[字幕导入] 异常: {}", e.getMessage(), e);
            return false;
        }
    }

    private List<Subtitle.SubtitleItem> parseSrtContent(String srtContent) {
        List<Subtitle.SubtitleItem> items = new ArrayList<>();
        if (srtContent == null || srtContent.isEmpty()) {
            return items;
        }

        srtContent = srtContent.replace("\r\n", "\n").replace("\r", "\n");
        String[] blocks = srtContent.split("\n\n+");

        for (int blockIndex = 0; blockIndex < blocks.length; blockIndex++) {
            String block = blocks[blockIndex];
            String[] lines = block.split("\n");
            if (lines.length < 3) {
                continue;
            }

            try {
                Subtitle.SubtitleItem item = new Subtitle.SubtitleItem();

                String indexLine = lines[0].trim();
                item.setIndex(Integer.parseInt(indexLine));

                String timeLine = lines[1].trim();
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
                    item.setStartTime(startTime);
                    item.setEndTime(endTime);
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
                items.add(item);
            } catch (Exception e) {
                // skip
            }
        }
        return items;
    }

    private double parseTimeToSeconds(String hour, String minute, String second, String millis) {
        return Double.parseDouble(hour) * 3600 +
               Double.parseDouble(minute) * 60 +
               Double.parseDouble(second) +
               Double.parseDouble(millis) / 1000.0;
    }

    private boolean extractAudio(String videoPath, String audioPath) {
        try {
            java.io.File videoFile = new java.io.File(videoPath);
            if (!videoFile.exists()) {
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
                    // silent
                }
            }

            int exitCode = process.waitFor();
            return exitCode == 0;

        } catch (Exception e) {
            return false;
        }
    }

    private boolean generateSubtitleWithWhisper(Integer manuscriptId, Integer videoId, String audioPath, String outputDir, ProgressListener progressListener) {
        try {
            String[] cmd = {
                    whisperConfig.getCliPath(),
                    "-m", whisperConfig.getModelPath(),
                    "-f", audioPath,
                    "-l", whisperConfig.getLanguage(),
                    "-osrt",
                    "-of", "zh-CN",
                    "-t", String.valueOf(whisperConfig.getThreads()),
                    "-pp",
                    "--initial-prompt", "以下是普通话的句子。"
            };

            log.info("[Whisper] 开始识别, audioPath={}, outputDir={}", audioPath, outputDir);
            log.info("[Whisper] 命令: {}", String.join(" ", cmd));

            ProcessBuilder pb = new ProcessBuilder(cmd);
            pb.directory(new java.io.File(outputDir));
            pb.redirectErrorStream(true);

            Process process = pb.start();

            long startTime = System.currentTimeMillis();
            int lastPushedPercent = 30;
            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    log.info("[Whisper] {}", line);
                    long elapsed = System.currentTimeMillis() - startTime;
                    // Map elapsed time to 30-90 range (the Whisper portion of overall progress)
                    int percent = 30 + (int) Math.min(59, elapsed / 3000);
                    if (percent > lastPushedPercent) {
                        pushProgress(progressListener, percent, "Whisper识别中 " + (percent - 30) * 100 / 60 + "%");
                        lastPushedPercent = percent;
                    }
                }
            }

            int exitCode = process.waitFor();
            log.info("[Whisper] 完成, exitCode={}", exitCode);
            return exitCode == 0;

        } catch (Exception e) {
            log.error("[Whisper] 异常: {}", e.getMessage(), e);
            return false;
        }
    }

    @Override
    public String getSubtitlePath(Integer manuscriptId, Integer videoId) {
        return uploadFilePathConfig.getChineseSubtitlePath(manuscriptId, videoId);
    }
}
