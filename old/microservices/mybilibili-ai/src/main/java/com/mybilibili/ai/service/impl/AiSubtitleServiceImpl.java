package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.config.DynamicSttClient;
import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.repository.SubtitleRepository;
import com.mybilibili.ai.service.AiSubtitleService;
import com.mybilibili.ai.service.VideoProcessingStorageService;
import com.mybilibili.ai.utils.SubtitleTextUtils;
import com.mybilibili.common.entity.Subtitle;
import com.mybilibili.common.entity.Video;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardOpenOption;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Service
public class AiSubtitleServiceImpl implements AiSubtitleService {

    private static final Logger log = LoggerFactory.getLogger(AiSubtitleServiceImpl.class);

    @Autowired
    private VideoProcessingStorageService processingStorageService;

    @Autowired
    private SubtitleRepository subtitleRepository;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private DynamicSttClient dynamicSttClient;

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
            Path audioPath = processingStorageService.materializeAudio(manuscriptId, videoId);
            Path subtitleDir = processingStorageService.getSubtitleDir(manuscriptId, videoId);

            Files.createDirectories(subtitleDir);

            pushProgress(progressListener, 30, "音频已准备，开始语音识别");

            boolean recognized = generateSubtitleWithApi(manuscriptId, videoId, audioPath.toString(), subtitleDir.toString(), progressListener);

            if (!recognized) {
                pushProgress(progressListener, 0, "语音识别失败");
                return false;
            }
            pushProgress(progressListener, 90, "语音识别完成，导入字幕");

            Path subtitlePath = processingStorageService.getSubtitlePath(manuscriptId, videoId);
            if (importSystemSubtitle(videoId, subtitlePath.toString())) {
                processingStorageService.uploadSubtitle(manuscriptId, videoId, subtitlePath);
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

    private boolean generateSubtitleWithApi(Integer manuscriptId, Integer videoId, String audioPath, String outputDir, ProgressListener progressListener) {
        try {
            log.info("[ApiStt] 开始识别, audioPath={}", audioPath);

            String result = dynamicSttClient.transcribeWithApi(audioPath, "zh");
            if (result == null || result.isEmpty()) {
                log.error("[ApiStt] 识别结果为空");
                return false;
            }

            // 判断是否为 SRT 格式（带时间戳）
            if (dynamicSttClient.isSrtFormat(result)) {
                // 直接写入 SRT 文件
                java.io.File outputFile = new java.io.File(outputDir, "zh-CN.srt");
                Files.writeString(outputFile.toPath(), result, StandardOpenOption.CREATE, StandardOpenOption.TRUNCATE_EXISTING);
                log.info("[ApiStt] SRT 文件已写入（带时间戳）: {}", outputFile.getAbsolutePath());
                return true;
            }

            // 纯文本模式，估算时间戳
            log.info("[ApiStt] 识别成功（纯文本），估算时间戳");
            Video video = videoMapper.selectById(videoId);
            double duration = video != null && video.getDurationSeconds() != null ? video.getDurationSeconds() : 0;
            String srtContent = dynamicSttClient.textToSrt(result, duration);

            java.io.File outputFile = new java.io.File(outputDir, "zh-CN.srt");
            Files.writeString(outputFile.toPath(), srtContent, StandardOpenOption.CREATE, StandardOpenOption.TRUNCATE_EXISTING);
            log.info("[ApiStt] SRT 文件已写入（估算时间戳）: {}", outputFile.getAbsolutePath());
            return true;

        } catch (Exception e) {
            log.error("[ApiStt] 异常: {}", e.getMessage(), e);
            return false;
        }
    }

    @Override
    public String getSubtitlePath(Integer manuscriptId, Integer videoId) {
        return processingStorageService.getSubtitlePath(manuscriptId, videoId).toString();
    }
}
