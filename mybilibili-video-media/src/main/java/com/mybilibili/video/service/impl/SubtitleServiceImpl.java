package com.mybilibili.video.service.impl;

import com.mybilibili.common.entity.Subtitle;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.storage.StorageKeys;
import com.mybilibili.common.storage.StorageService;
import com.mybilibili.video.mapper.VideoMapper;
import com.mybilibili.video.repository.SubtitleRepository;
import com.mybilibili.video.service.SubtitleService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;

import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Date;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Optional;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

@Service
public class SubtitleServiceImpl implements SubtitleService {

    private static final int STATUS_PENDING = 0;
    private static final int STATUS_APPROVED = 1;
    private static final int STATUS_REJECTED = 2;
    private static final int STATUS_SYSTEM = 3;

    private static final List<String> SYSTEM_LANGUAGES = List.of("zh-CN");

    @Autowired
    private SubtitleRepository subtitleRepository;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private StorageService storageService;

    @Value("${minio.bucket-name:mybilibili}")
    private String bucketName;

    @Override
    public List<Subtitle> getSubtitlesByVideoId(Integer videoId) {
        List<Subtitle> subtitles = subtitleRepository.findByVideoId(videoId);
        return subtitles.stream()
                .filter(this::isAvailableSubtitle)
                .collect(Collectors.toList());
    }

    @Override
    public List<Map<String, Object>> getVideosWithSubtitleInfo() {
        List<Video> videos = videoMapper.selectAll();
        List<Map<String, Object>> result = new ArrayList<>();
        for (Video video : videos) {
            List<Subtitle> subtitles = subtitleRepository.findByVideoId(video.getId());
            List<Map<String, Object>> pendingImportSubtitles = scanSystemSubtitles(video.getId());
            long pendingImportCount = pendingImportSubtitles.stream()
                    .filter(item -> "pending".equals(item.get("status")))
                    .count();

            Map<String, Object> row = new LinkedHashMap<>();
            row.put("id", video.getId());
            row.put("manuscriptId", video.getManuscriptId());
            row.put("title", video.getTitle());
            row.put("durationSeconds", video.getDurationSeconds());
            row.put("processStatus", video.getProcessStatus());
            row.put("processProgress", video.getProcessProgress());
            row.put("processStage", video.getProcessStage());
            row.put("hasSubtitle", video.getHasSubtitle());
            row.put("approvedCount", countByStatus(subtitles, STATUS_APPROVED));
            row.put("pendingCount", countByStatus(subtitles, STATUS_PENDING));
            row.put("rejectedCount", countByStatus(subtitles, STATUS_REJECTED));
            row.put("systemCount", countByStatus(subtitles, STATUS_SYSTEM));
            row.put("hasDefaultSubtitle", subtitles.stream().anyMatch(s -> Boolean.TRUE.equals(s.getIsDefault())));
            row.put("pendingImportCount", pendingImportCount);
            row.put("pendingImportSubtitles", pendingImportSubtitles);
            result.add(row);
        }
        return result;
    }

    private long countByStatus(List<Subtitle> subtitles, int status) {
        return subtitles.stream()
                .filter(subtitle -> subtitle.getStatus() != null && subtitle.getStatus() == status)
                .count();
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
        Subtitle saved = subtitleRepository.save(subtitle);
        if (isAvailableSubtitle(saved)) {
            refreshVideoSubtitleFlag(saved.getVideoId());
        }
        return saved;
    }

    @Override
    public Subtitle uploadSrtFile(Integer videoId, MultipartFile file, String language, Boolean isDefault, Integer uploadedBy) {
        if (file == null || file.isEmpty()) {
            throw new IllegalArgumentException("字幕文件不能为空");
        }
        try {
            String srtContent = new String(file.getBytes(), StandardCharsets.UTF_8);
            return saveParsedSrt(videoId, srtContent, language, languageName(language), uploadedBy, "admin", STATUS_APPROVED, isDefault, null);
        } catch (Exception e) {
            throw new IllegalArgumentException("读取字幕文件失败: " + e.getMessage(), e);
        }
    }

    @Override
    public Subtitle parseAndSaveSrt(Integer videoId, String srtContent, String language, String languageName, Integer uploadedBy) {
        return saveParsedSrt(videoId, srtContent, language, languageName, uploadedBy, "user", STATUS_APPROVED, false, null);
    }

    @Override
    public Subtitle importSrtFilePath(Integer videoId, String srtFilePath, String language, Boolean isDefault, Integer uploadedBy) {
        if (srtFilePath == null || srtFilePath.isBlank()) {
            throw new IllegalArgumentException("SRT文件路径不能为空");
        }
        String srtContent = readSrtContent(srtFilePath);
        return saveParsedSrt(videoId, srtContent, language, languageName(language), uploadedBy, "admin", STATUS_APPROVED, isDefault, srtFilePath);
    }

    @Override
    public List<Subtitle> getPendingSubtitles() {
        return subtitleRepository.findByStatus(STATUS_PENDING);
    }

    @Override
    public Subtitle approveSubtitle(String subtitleId, Integer reviewerId) {
        Subtitle subtitle = requireSubtitle(subtitleId);
        subtitle.setStatus(STATUS_APPROVED);
        subtitle.setReviewerId(reviewerId);
        subtitle.setReviewTime(new Date());
        subtitle.setReviewReason(null);
        Subtitle saved = subtitleRepository.save(subtitle);
        refreshVideoSubtitleFlag(saved.getVideoId());
        return saved;
    }

    @Override
    public Subtitle rejectSubtitle(String subtitleId, String reason, Integer reviewerId) {
        Subtitle subtitle = requireSubtitle(subtitleId);
        subtitle.setStatus(STATUS_REJECTED);
        subtitle.setIsDefault(false);
        subtitle.setReviewerId(reviewerId);
        subtitle.setReviewTime(new Date());
        subtitle.setReviewReason(reason);
        Subtitle saved = subtitleRepository.save(subtitle);
        refreshVideoSubtitleFlag(saved.getVideoId());
        return saved;
    }

    @Override
    public Map<String, Object> previewSubtitle(String subtitleId) {
        Subtitle subtitle = requireSubtitle(subtitleId);
        Map<String, Object> result = new LinkedHashMap<>();
        result.put("subtitle", subtitle);
        result.put("content", subtitle.getContent() == null ? List.of() : subtitle.getContent());
        return result;
    }

    @Override
    public List<Map<String, Object>> scanSystemSubtitles(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null || video.getManuscriptId() == null) {
            return List.of();
        }

        List<Map<String, Object>> result = new ArrayList<>();
        for (String language : SYSTEM_LANGUAGES) {
            String objectKey = StorageKeys.videoSubtitle(video.getManuscriptId(), videoId, language);
            if (!storageService.exists(objectKey)) {
                continue;
            }
            boolean imported = subtitleRepository.findFirstByVideoIdAndLanguage(videoId, language).isPresent();
            Map<String, Object> item = new LinkedHashMap<>();
            item.put("language", language);
            item.put("languageName", languageName(language));
            item.put("fileName", language + ".srt");
            item.put("fileSize", 0);
            item.put("objectKey", objectKey);
            item.put("status", imported ? "imported" : "pending");
            result.add(item);
        }
        return result;
    }

    @Override
    public Subtitle importSystemSubtitle(Integer videoId, String language, Boolean isDefault, Integer uploadedBy) {
        Video video = videoMapper.selectById(videoId);
        if (video == null || video.getManuscriptId() == null) {
            throw new IllegalArgumentException("视频不存在");
        }
        String objectKey = StorageKeys.videoSubtitle(video.getManuscriptId(), videoId, language);
        if (!storageService.exists(objectKey)) {
            throw new IllegalArgumentException("系统字幕文件不存在");
        }
        String srtContent = readStorageObject(objectKey);
        return saveParsedSrt(videoId, srtContent, language, languageName(language), uploadedBy, "whisper", STATUS_SYSTEM, isDefault, objectKey);
    }

    private Subtitle saveParsedSrt(Integer videoId,
                                   String srtContent,
                                   String language,
                                   String languageName,
                                   Integer uploadedBy,
                                   String source,
                                   Integer status,
                                   Boolean isDefault,
                                   String filePath) {
        List<Subtitle.SubtitleItem> items = parseSrtContent(srtContent);
        if (items.isEmpty()) {
            throw new IllegalArgumentException("字幕内容为空或格式不正确");
        }

        Subtitle subtitle = new Subtitle();
        subtitle.setVideoId(videoId);
        subtitle.setLanguage(language);
        subtitle.setLanguageName(languageName == null || languageName.isBlank() ? languageName(language) : languageName);
        subtitle.setFormat("srt");
        subtitle.setContent(items);
        subtitle.setUploadedBy(uploadedBy);
        subtitle.setSource(source);
        subtitle.setStatus(status);
        subtitle.setIsDefault(Boolean.TRUE.equals(isDefault));
        subtitle.setFilePath(filePath);

        Subtitle saved = uploadSubtitle(subtitle);
        if (Boolean.TRUE.equals(isDefault)) {
            setDefaultSubtitleById(saved.getId());
            return requireSubtitle(saved.getId());
        }
        return saved;
    }

    private List<Subtitle.SubtitleItem> parseSrtContent(String srtContent) {
        List<Subtitle.SubtitleItem> items = new ArrayList<>();
        if (srtContent == null || srtContent.isBlank()) {
            return items;
        }

        srtContent = srtContent.replace("\r\n", "\n").replace("\r", "\n");
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
        Subtitle subtitle = requireSubtitle(subtitleId);
        subtitleRepository.deleteById(subtitleId);
        refreshVideoSubtitleFlag(subtitle.getVideoId());
    }

    @Override
    public void setDefaultSubtitle(Integer videoId, String language) {
        List<Subtitle> subtitles = subtitleRepository.findByVideoId(videoId);
        for (Subtitle subtitle : subtitles) {
            subtitle.setIsDefault(isAvailableSubtitle(subtitle) && subtitle.getLanguage().equals(language));
            subtitleRepository.save(subtitle);
        }
    }

    @Override
    public void setDefaultSubtitleById(String subtitleId) {
        Subtitle target = requireSubtitle(subtitleId);
        if (!isAvailableSubtitle(target)) {
            throw new IllegalArgumentException("只有审核通过或系统生成字幕可以设为默认");
        }
        List<Subtitle> subtitles = subtitleRepository.findByVideoId(target.getVideoId());
        for (Subtitle subtitle : subtitles) {
            subtitle.setIsDefault(subtitle.getId().equals(subtitleId));
            subtitleRepository.save(subtitle);
        }
    }

    private Subtitle requireSubtitle(String subtitleId) {
        return subtitleRepository.findById(subtitleId)
                .orElseThrow(() -> new IllegalArgumentException("字幕不存在"));
    }

    private void refreshVideoSubtitleFlag(Integer videoId) {
        List<Subtitle> subtitles = subtitleRepository.findByVideoId(videoId);
        boolean hasAvailableSubtitle = subtitles.stream().anyMatch(this::isAvailableSubtitle);
        videoMapper.updateHasSubtitle(videoId, hasAvailableSubtitle ? 1 : 0);
    }

    private boolean isAvailableSubtitle(Subtitle subtitle) {
        return subtitle != null
                && subtitle.getStatus() != null
                && (subtitle.getStatus() == STATUS_APPROVED || subtitle.getStatus() == STATUS_SYSTEM);
    }

    private String readSrtContent(String srtFilePath) {
        try {
            Path localPath = Path.of(srtFilePath);
            if (Files.exists(localPath)) {
                return Files.readString(localPath, StandardCharsets.UTF_8);
            }
        } catch (Exception ignored) {
            // If this is not a valid local path, try it as an object-storage key.
        }

        String objectKey = normalizeStorageKey(srtFilePath);
        if (storageService.exists(objectKey)) {
            return readStorageObject(objectKey);
        }
        throw new IllegalArgumentException("SRT文件不存在");
    }

    private String readStorageObject(String objectKey) {
        try (InputStream input = storageService.download(objectKey)) {
            return new String(input.readAllBytes(), StandardCharsets.UTF_8);
        } catch (Exception e) {
            throw new IllegalArgumentException("读取对象存储字幕失败: " + e.getMessage(), e);
        }
    }

    private String normalizeStorageKey(String value) {
        String normalized = value.trim().replace('\\', '/');
        String bucketPrefix = "/" + bucketName + "/";
        int bucketIndex = normalized.indexOf(bucketPrefix);
        if (bucketIndex >= 0) {
            return normalized.substring(bucketIndex + bucketPrefix.length());
        }
        if (normalized.startsWith(bucketName + "/")) {
            return normalized.substring(bucketName.length() + 1);
        }
        if (normalized.startsWith("/")) {
            return normalized.substring(1);
        }
        return normalized;
    }

    private String languageName(String language) {
        if (language == null || language.isBlank()) {
            return "未知";
        }
        return switch (language.toLowerCase(Locale.ROOT)) {
            case "zh-cn" -> "简体中文";
            case "zh-tw" -> "繁体中文";
            case "en" -> "English";
            case "ja" -> "日本語";
            case "ko" -> "한국어";
            default -> language;
        };
    }
}
