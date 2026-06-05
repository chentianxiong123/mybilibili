package com.mybilibili.videomedia.process;

import com.mybilibili.common.entity.Video;
import com.mybilibili.common.storage.StorageKeys;
import com.mybilibili.common.storage.StorageService;
import com.mybilibili.video.mapper.VideoMapper;
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

@Service
public class VideoMediaProcessingStorageService {

    private final StorageService storageService;
    private final VideoMapper videoMapper;

    @Value("${video.processing.work-dir:${upload.base-path:mybilibili-processing}}")
    private String workDir;

    @Value("${minio.bucket-name:mybilibili}")
    private String bucketName;

    public VideoMediaProcessingStorageService(StorageService storageService, VideoMapper videoMapper) {
        this.storageService = storageService;
        this.videoMapper = videoMapper;
    }

    public Path materializeSourceVideo(Integer manuscriptId, Integer videoId) throws IOException {
        String key = getSourceObjectKey(videoId);
        String extension = getRequiredExtension(key);
        Path sourcePath = getVideoWorkDir(manuscriptId, videoId).resolve("source").resolve("video" + extension);
        if (Files.exists(sourcePath) && Files.size(sourcePath) > 0) {
            return sourcePath;
        }

        Files.createDirectories(sourcePath.getParent());
        try (InputStream input = storageService.download(key)) {
            Files.copy(input, sourcePath, StandardCopyOption.REPLACE_EXISTING);
        }
        return sourcePath;
    }

    public Path getTranscodedDir(Integer manuscriptId, Integer videoId) {
        return getVideoWorkDir(manuscriptId, videoId).resolve("transcoded");
    }

    public Path getAudioPath(Integer manuscriptId, Integer videoId) {
        return getVideoWorkDir(manuscriptId, videoId).resolve("audio").resolve("audio.wav");
    }

    public void recreateDirectory(Path dir) throws IOException {
        deleteDirectory(dir);
        Files.createDirectories(dir);
    }

    public String uploadHlsDirectory(Integer manuscriptId, Integer videoId, String quality, Path qualityDir) throws IOException {
        Path playlist = qualityDir.resolve("playlist.m3u8");
        if (!Files.exists(playlist)) {
            throw new IOException("HLS playlist missing: " + playlist);
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
                } catch (IOException ignored) {
                    // Best effort cleanup.
                }
            });
        } catch (IOException ignored) {
            // Best effort cleanup.
        }
    }

    private Path getVideoWorkDir(Integer manuscriptId, Integer videoId) {
        return Path.of(workDir, "manuscripts", String.valueOf(manuscriptId), "videos", String.valueOf(videoId));
    }

    private String uploadFile(Path file, String key, String contentType) throws IOException {
        if (!Files.exists(file) || Files.size(file) == 0) {
            throw new IOException("File missing or empty: " + file);
        }
        try (InputStream input = Files.newInputStream(file)) {
            return storageService.upload(key, input, Files.size(file), contentType);
        }
    }

    private String getSourceObjectKey(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null || video.getSourceVideoUrl() == null || video.getSourceVideoUrl().isBlank()) {
            throw new IllegalStateException("Source video missing: " + videoId);
        }
        return extractObjectKey(video.getSourceVideoUrl());
    }

    private String extractObjectKey(String sourceUrl) {
        if (!sourceUrl.contains("://") && !sourceUrl.startsWith("/")) {
            return sourceUrl;
        }
        URI uri = URI.create(sourceUrl);
        String path = uri.getPath();
        String bucketPrefix = "/" + bucketName + "/";
        int bucketIndex = path.indexOf(bucketPrefix);
        if (bucketIndex < 0) {
            throw new IllegalStateException("Source video is not in current bucket: " + sourceUrl);
        }
        String key = path.substring(bucketIndex + bucketPrefix.length());
        if (key.isBlank()) {
            throw new IllegalStateException("Source object key is blank: " + sourceUrl);
        }
        return URLDecoder.decode(key, StandardCharsets.UTF_8);
    }

    private String getRequiredExtension(String key) {
        int slashIndex = key.lastIndexOf('/');
        String fileName = slashIndex >= 0 ? key.substring(slashIndex + 1) : key;
        int dotIndex = fileName.lastIndexOf('.');
        if (dotIndex < 0 || dotIndex == fileName.length() - 1) {
            throw new IllegalStateException("Source object key has no extension: " + key);
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
        if (name.endsWith(".wav")) {
            return "audio/wav";
        }
        return "application/octet-stream";
    }
}
