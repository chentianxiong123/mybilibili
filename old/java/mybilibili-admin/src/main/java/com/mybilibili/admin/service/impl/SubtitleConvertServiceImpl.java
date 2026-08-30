package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.repository.SubtitleRepository;
import com.mybilibili.admin.service.SubtitleConvertService;
import com.mybilibili.common.entity.Subtitle;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * 字幕转换服务实现
 */
@Service
public class SubtitleConvertServiceImpl implements SubtitleConvertService {

    @Autowired
    private SubtitleRepository subtitleRepository;

    @Override
    public Subtitle saveSrtToMongo(Integer videoId, String srtContent, String language,
                                   String languageName, String source, Integer uploadedBy) {
        try {
            System.out.println("[字幕转换] 开始保存字幕到 MongoDB: videoId=" + videoId + ", language=" + language);

            // 解析 SRT 内容
            List<Subtitle.SubtitleItem> items = parseSrtContent(srtContent);
            System.out.println("[字幕转换] 解析完成，共 " + items.size() + " 条字幕");

            // 先删除已存在的同语言字幕
            deleteSubtitleFromMongo(videoId, language);

            // 创建字幕实体
            Subtitle subtitle = new Subtitle();
            subtitle.setVideoId(videoId);
            subtitle.setLanguage(language);
            subtitle.setLanguageName(languageName != null ? languageName : language);
            subtitle.setFormat("srt");
            subtitle.setContent(items);
            subtitle.setIsDefault(true);
            subtitle.setUploadedBy(uploadedBy != null ? uploadedBy : 0);
            subtitle.setUploadTime(new Date());
            subtitle.setStatus(1); // 已审核
            subtitle.setSource(source != null ? source : "unknown");
            subtitle.setVersion(1);

            // 保存到 MongoDB
            Subtitle saved = subtitleRepository.save(subtitle);
            System.out.println("[字幕转换] 字幕保存成功: id=" + saved.getId());

            return saved;

        } catch (Exception e) {
            System.err.println("[字幕转换] 保存字幕失败: " + e.getMessage());
            e.printStackTrace();
            throw new RuntimeException("保存字幕失败: " + e.getMessage(), e);
        }
    }

    @Override
    public Subtitle saveSrtFileToMongo(Integer videoId, String srtFilePath, String language,
                                       String languageName, String source, Integer uploadedBy) {
        try {
            System.out.println("[字幕转换] 从文件读取字幕: " + srtFilePath);

            // 读取文件内容
            File file = new File(srtFilePath);
            if (!file.exists()) {
                throw new RuntimeException("字幕文件不存在: " + srtFilePath);
            }

            String srtContent = new String(Files.readAllBytes(Paths.get(srtFilePath)));
            System.out.println("[字幕转换] 文件读取完成，大小: " + srtContent.length() + " 字符");

            return saveSrtToMongo(videoId, srtContent, language, languageName, source, uploadedBy);

        } catch (IOException e) {
            System.err.println("[字幕转换] 读取字幕文件失败: " + e.getMessage());
            throw new RuntimeException("读取字幕文件失败: " + e.getMessage(), e);
        }
    }

    @Override
    public List<Subtitle.SubtitleItem> parseSrtContent(String srtContent) {
        List<Subtitle.SubtitleItem> items = new ArrayList<>();

        if (srtContent == null || srtContent.trim().isEmpty()) {
            System.out.println("[字幕转换] SRT 内容为空");
            return items;
        }

        String[] blocks = srtContent.split("\\n\\s*\\n");
        Pattern timePattern = Pattern.compile("(\\d{2}):\\s*(\\d{2}):\\s*(\\d{2})[,.](\\d{3})\\s*-->\\s*(\\d{2}):\\s*(\\d{2}):\\s*(\\d{2})[,.](\\d{3})");

        int successCount = 0;
        int failCount = 0;

        for (String block : blocks) {
            block = block.trim();
            if (block.isEmpty()) continue;

            String[] lines = block.split("\\n");
            if (lines.length < 3) {
                failCount++;
                continue;
            }

            try {
                Integer index = Integer.parseInt(lines[0].trim());

                Matcher matcher = timePattern.matcher(lines[1]);
                if (!matcher.find()) {
                    failCount++;
                    continue;
                }

                double startTime = parseTimeToSeconds(
                    matcher.group(1), matcher.group(2),
                    matcher.group(3), matcher.group(4)
                );
                double endTime = parseTimeToSeconds(
                    matcher.group(5), matcher.group(6),
                    matcher.group(7), matcher.group(8)
                );

                StringBuilder textBuilder = new StringBuilder();
                for (int i = 2; i < lines.length; i++) {
                    if (textBuilder.length() > 0) {
                        textBuilder.append("\n");
                    }
                    textBuilder.append(lines[i].trim());
                }

                Subtitle.SubtitleItem item = new Subtitle.SubtitleItem();
                item.setIndex(index);
                item.setStartTime(startTime);
                item.setEndTime(endTime);
                item.setText(textBuilder.toString());
                items.add(item);
                successCount++;

            } catch (Exception e) {
                failCount++;
                System.err.println("[字幕转换] 解析字幕块失败: " + e.getMessage());
            }
        }

        System.out.println("[字幕转换] 解析完成: 成功 " + successCount + " 条, 失败 " + failCount + " 条");
        return items;
    }

    @Override
    public void deleteSubtitleFromMongo(Integer videoId, String language) {
        try {
            subtitleRepository.findByVideoIdAndLanguage(videoId, language)
                .ifPresent(existing -> {
                    subtitleRepository.delete(existing);
                    System.out.println("[字幕转换] 删除已存在字幕: videoId=" + videoId + ", language=" + language);
                });
        } catch (Exception e) {
            System.err.println("[字幕转换] 删除字幕失败: " + e.getMessage());
        }
    }

    @Override
    public boolean existsSubtitle(Integer videoId, String language) {
        return subtitleRepository.findByVideoIdAndLanguage(videoId, language).isPresent();
    }

    /**
     * 将时间字符串转换为秒
     */
    private double parseTimeToSeconds(String hours, String minutes, String seconds, String millis) {
        int h = Integer.parseInt(hours);
        int m = Integer.parseInt(minutes);
        int s = Integer.parseInt(seconds);
        int ms = Integer.parseInt(millis);
        return h * 3600 + m * 60 + s + ms / 1000.0;
    }
}
