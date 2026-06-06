package com.mybilibili.video.service.impl;

import com.mybilibili.common.entity.Subtitle;
import com.mybilibili.video.repository.SubtitleRepository;
import com.mybilibili.video.service.SubtitleService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.Optional;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

@Service
public class SubtitleServiceImpl implements SubtitleService {

    @Autowired
    private SubtitleRepository subtitleRepository;

    @Override
    public List<Subtitle> getSubtitlesByVideoId(Integer videoId) {
        List<Subtitle> subtitles = subtitleRepository.findByVideoId(videoId);
        return subtitles.stream()
                .filter(s -> s.getStatus() != null && (s.getStatus() == 1 || s.getStatus() == 3))
                .collect(Collectors.toList());
    }

    @Override
    public Subtitle getSubtitleByVideoIdAndLanguage(Integer videoId, String language) {
        Optional<Subtitle> subtitleOpt = subtitleRepository.findFirstByVideoIdAndLanguage(videoId, language);
        return subtitleOpt.orElse(null);
    }

    @Override
    public Subtitle uploadSubtitle(Subtitle subtitle) {
        if (subtitle.getUploadTime() == null) {
            subtitle.setUploadTime(new Date());
        }
        if (subtitle.getStatus() == null) {
            subtitle.setStatus(1);
        }
        if (subtitle.getVersion() == null) {
            subtitle.setVersion(1);
        }
        if (subtitle.getIsDefault() == null) {
            subtitle.setIsDefault(false);
        }
        return subtitleRepository.save(subtitle);
    }

    @Override
    public Subtitle parseAndSaveSrt(Integer videoId, String srtContent, String language, String languageName, Integer uploadedBy) {
        List<Subtitle.SubtitleItem> items = parseSrtContent(srtContent);

        Subtitle subtitle = new Subtitle();
        subtitle.setVideoId(videoId);
        subtitle.setLanguage(language);
        subtitle.setLanguageName(languageName);
        subtitle.setFormat("srt");
        subtitle.setContent(items);
        subtitle.setUploadedBy(uploadedBy);
        subtitle.setSource("user");
        subtitle.setIsDefault(false);

        return uploadSubtitle(subtitle);
    }

    private List<Subtitle.SubtitleItem> parseSrtContent(String srtContent) {
        List<Subtitle.SubtitleItem> items = new ArrayList<>();

        String[] blocks = srtContent.split("\\n\\s*\\n");
        Pattern timePattern = Pattern.compile("(\\d{2}):\\s*(\\d{2}):\\s*(\\d{2})[,.](\\d{3})\\s*-->\\s*(\\d{2}):\\s*(\\d{2}):\\s*(\\d{2})[,.](\\d{3})");

        for (String block : blocks) {
            block = block.trim();
            if (block.isEmpty()) continue;

            String[] lines = block.split("\\n");
            if (lines.length < 3) continue;

            try {
                Integer index = Integer.parseInt(lines[0].trim());

                Matcher matcher = timePattern.matcher(lines[1]);
                if (!matcher.find()) continue;

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
                        textBuilder.append("\\n");
                    }
                    textBuilder.append(lines[i].trim());
                }

                Subtitle.SubtitleItem item = new Subtitle.SubtitleItem();
                item.setIndex(index);
                item.setStartTime(startTime);
                item.setEndTime(endTime);
                item.setText(textBuilder.toString());
                items.add(item);

            } catch (Exception e) {
                continue;
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

    @Override
    public void deleteSubtitle(String subtitleId) {
        subtitleRepository.deleteById(subtitleId);
    }

    @Override
    public void setDefaultSubtitle(Integer videoId, String language) {
        List<Subtitle> subtitles = subtitleRepository.findByVideoId(videoId);
        for (Subtitle subtitle : subtitles) {
            subtitle.setIsDefault(subtitle.getLanguage().equals(language));
            subtitleRepository.save(subtitle);
        }
    }
}
