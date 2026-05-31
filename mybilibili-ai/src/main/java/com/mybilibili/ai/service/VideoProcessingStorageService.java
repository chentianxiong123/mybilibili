package com.mybilibili.ai.service;

import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.storage.StorageKeys;
import com.mybilibili.common.storage.StorageService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.io.InputStream;
import java.net.URI;
import java.net.URLDecoder;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardCopyOption;
import java.util.Comparator;
import java.util.List;
import java.util.Locale;
import java.util.stream.Stream;

@Slf4j
@Service
public class VideoProcessingStorageService {

    private final StorageService storageService;
    private final VideoMapper videoMapper;

    @Value("${video.processing.work-dir:mybilibili-processing}")
    private String workDir;

    @Value("${minio.bucket-name:mybilibili}")
    private String bucketName;

    public VideoProcessingStorageService(StorageService storageService, VideoMapper videoMapper) {
        this.storageService = storageService;
        this.videoMapper = videoMapper;
    }

    public Path materializeSourceVideo(Integer manuscriptId, Integer videoId) throws IOException {
        String key = getSourceObjectKey(videoId);
        String extension = getRequiredExtension(key);
        Path sourcePath = getSourceVideoPath(manuscriptId, videoId, extension);

        if (Files.exists(sourcePath) && Files.size(sourcePath) > 0) {
            return sourcePath;
        }

        Files.createDirectories(sourcePath.getParent());
        try (InputStream input = storageService.download(key)) {
            Files.copy(input, sourcePath, StandardCopyOption.REPLACE_EXISTING);
        }
        log.info("[视频处理存储] 源视频已从MinIO拉取: videoId={}, key={}", videoId, key);
        return sourcePath;
    }

    public Path getTranscodedDir(Integer manuscriptId, Integer videoId) {
        return getVideoWorkDir(manuscriptId, videoId).resolve("transcoded");
    }

    public Path getAudioPath(Integer manuscriptId, Integer videoId) {
        return getVideoWorkDir(manuscriptId, videoId).resolve("audio").resolve("audio.wav");
    }

    public Path getSubtitlePath(Integer manuscriptId, Integer videoId) {
        return getVideoWorkDir(manuscriptId, videoId).resolve("subtitles").resolve("zh-CN.srt");
    }

    public Path getSubtitleDir(Integer manuscriptId, Integer videoId) {
        return getSubtitlePath(manuscriptId, videoId).getParent();
    }

    public Path getSummaryPath(Integer manuscriptId, Integer videoId) {
        return getVideoWorkDir(manuscriptId, videoId).resolve("summary").resolve("ai-summary.txt");
    }

    public void recreateDirectory(Path dir) throws IOException {
        deleteDirectory(dir);
        Files.createDirectories(dir);
    }

    public String uploadHlsDirectory(Integer manuscriptId, Integer videoId, String quality, Path qualityDir) throws IOException {
        Path playlist = qualityDir.resolve("playlist.m3u8");
        if (!Files.exists(playlist)) {
            throw new IOException("HLS播放列表不存在: " + playlist);
        }

        List<Path> files;
        try (Stream<Path> stream = Files.walk(qualityDir)) {
            files = stream.filter(Files::isRegularFile).toList();
        }

        for (Path file : files) {
            String relative = qualityDir.relativize(file).toString().replace('\\', '/');
            String key = StorageKeys.videoHlsObject(manuscriptId, videoId, quality, relative);
            uploadFile(file, key, contentType(file));
        }

        return storageService.getPublicUrl(StorageKeys.videoHlsPlaylist(manuscriptId, videoId, quality));
    }

    public String uploadAudio(Integer manuscriptId, Integer videoId, Path audioPath) throws IOException {
        return uploadFile(audioPath, StorageKeys.videoAudio(manuscriptId, videoId), "audio/wav");
    }

    public String uploadSubtitle(Integer manuscriptId, Integer videoId, Path subtitlePath) throws IOException {
        return uploadFile(subtitlePath, StorageKeys.videoSubtitle(manuscriptId, videoId, "zh-CN"), "application/x-subrip; charset=utf-8");
    }

    public String uploadSummary(Integer manuscriptId, Integer videoId, Path summaryPath) throws IOException {
        return uploadFile(summaryPath, StorageKeys.videoSummary(manuscriptId, videoId), "text/plain; charset=utf-8");
    }

    public String getPublicUrl(String key) {
        return storageService.getPublicUrl(key);
    }

    public void cleanupWorkDir(Integer manuscriptId, Integer videoId) {
        deleteDirectory(getVideoWorkDir(manuscriptId, videoId));
    }

    public void deleteDirectory(Path dir) {
        if (dir == null || !Files.exists(dir)) {
            return;
        }
        try (Stream<Path> stream = Files.walk(dir)) {
            stream.sorted(Comparator.reverseOrder()).forEach(path -> {
                try {
                    Files.deleteIfExists(path);
                } catch (IOException e) {
                    log.warn("[视频处理存储] 删除临时文件失败: {}", path, e);
                }
            });
        } catch (IOException e) {
            log.warn("[视频处理存储] 删除临时目录失败: {}", dir, e);
        }
    }

    private Path getVideoWorkDir(Integer manuscriptId, Integer videoId) {
        return Path.of(workDir, "manuscripts", String.valueOf(manuscriptId), "videos", String.valueOf(videoId));
    }

    private Path getSourceVideoPath(Integer manuscriptId, Integer videoId, String extension) {
        return getVideoWorkDir(manuscriptId, videoId).resolve("source").resolve("video" + extension);
    }

    private String uploadFile(Path file, String key, String contentType) throws IOException {
        if (!Files.exists(file) || Files.size(file) == 0) {
            throw new IOException("待上传文件不存在或为空: " + file);
        }
        try (InputStream input = Files.newInputStream(file)) {
            return storageService.upload(key, input, Files.size(file), contentType);
        }
    }

    private String getSourceObjectKey(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            throw new IllegalStateException("视频不存在: " + videoId);
        }
        String sourceUrl = video.getSourceVideoUrl();
        if (sourceUrl == null || sourceUrl.isBlank()) {
            throw new IllegalStateException("源视频地址为空: " + videoId);
        }
        return extractObjectKey(sourceUrl);
    }

    private String extractObjectKey(String sourceUrl) {
        if (!sourceUrl.contains("://") && !sourceUrl.startsWith("/")) {
            return sourceUrl;
        }
        if (sourceUrl.contains("/uploads/")) {
            throw new IllegalStateException("源视频仍是旧本地uploads地址，不能进入转码: " + sourceUrl);
        }

        URI uri = URI.create(sourceUrl);
        String path = uri.getPath();
        String bucketPrefix = "/" + bucketName + "/";
        int bucketIndex = path.indexOf(bucketPrefix);
        if (bucketIndex < 0) {
            throw new IllegalStateException("源视频不是当前MinIO bucket地址: " + sourceUrl);
        }
        String key = path.substring(bucketIndex + bucketPrefix.length());
        if (key.isBlank()) {
            throw new IllegalStateException("源视频对象Key为空: " + sourceUrl);
        }
        return URLDecoder.decode(key, StandardCharsets.UTF_8);
    }

    private String getRequiredExtension(String key) {
        int slashIndex = key.lastIndexOf('/');
        String fileName = slashIndex >= 0 ? key.substring(slashIndex + 1) : key;
        int dotIndex = fileName.lastIndexOf('.');
        if (dotIndex < 0 || dotIndex == fileName.length() - 1) {
            throw new IllegalStateException("源视频对象缺少文件扩展名: " + key);
        }
        return fileName.substring(dotIndex).toLowerCase(Locale.ROOT);
    }

    private String contentType(Path file) {
        String name = file.getFileName().toString().toLowerCase(Locale.ROOT);
        if (name.endsWith(".m3u8")) {
            return "application/vnd.apple.mpegurl";
        }
        if (name.endsWith(".ts")) {
            return "video/mp2t";
        }
        if (name.endsWith(".mp4")) {
            return "video/mp4";
        }
        if (name.endsWith(".wav")) {
            return "audio/wav";
        }
        if (name.endsWith(".srt")) {
            return "application/x-subrip; charset=utf-8";
        }
        if (name.endsWith(".txt")) {
            return "text/plain; charset=utf-8";
        }
        return "application/octet-stream";
    }
}
