package com.mybilibili.common.storage;

import lombok.extern.slf4j.Slf4j;

import java.io.IOException;
import java.nio.file.*;
import java.nio.file.attribute.BasicFileAttributes;
import java.util.concurrent.atomic.AtomicInteger;

@Slf4j
public class StorageMigrationTool {

    private final StorageService storageService;
    private final Path localBasePath;

    public StorageMigrationTool(StorageService storageService, Path localBasePath) {
        this.storageService = storageService;
        this.localBasePath = localBasePath;
    }

    public MigrationResult migrateAll() throws IOException {
        AtomicInteger uploaded = new AtomicInteger(0);
        AtomicInteger skipped = new AtomicInteger(0);
        AtomicInteger failed = new AtomicInteger(0);

        Path uploadsDir = localBasePath.resolve("uploads");
        if (!Files.exists(uploadsDir)) {
            log.warn("Local uploads directory not found: {}", uploadsDir);
            return new MigrationResult(0, 0, 0);
        }

        Files.walkFileTree(uploadsDir, new SimpleFileVisitor<>() {
            @Override
            public FileVisitResult visitFile(Path file, BasicFileAttributes attrs) throws IOException {
                try {
                    Path relativePath = uploadsDir.relativize(file);
                    String key = relativePath.toString().replace('\\', '/');

                    if (storageService.exists(key)) {
                        skipped.incrementAndGet();
                        return FileVisitResult.CONTINUE;
                    }

                    String contentType = detectContentType(file);
                    try (var stream = Files.newInputStream(file)) {
                        storageService.upload(key, stream, attrs.size(), contentType);
                        uploaded.incrementAndGet();
                        log.debug("Migrated: {} -> {}", file, key);
                    }
                } catch (Exception e) {
                    failed.incrementAndGet();
                    log.error("Failed to migrate: {}", file, e);
                }
                return FileVisitResult.CONTINUE;
            }
        });

        MigrationResult result = new MigrationResult(uploaded.get(), skipped.get(), failed.get());
        log.info("Migration complete: uploaded={}, skipped={}, failed={}", result.uploaded(), result.skipped(), result.failed());
        return result;
    }

    private String detectContentType(Path file) {
        String name = file.getFileName().toString().toLowerCase();
        if (name.endsWith(".jpg") || name.endsWith(".jpeg")) return "image/jpeg";
        if (name.endsWith(".png")) return "image/png";
        if (name.endsWith(".gif")) return "image/gif";
        if (name.endsWith(".mp4")) return "video/mp4";
        if (name.endsWith(".webm")) return "video/webm";
        if (name.endsWith(".m3u8")) return "application/vnd.apple.mpegurl";
        if (name.endsWith(".ts")) return "video/mp2t";
        if (name.endsWith(".wav")) return "audio/wav";
        if (name.endsWith(".mp3")) return "audio/mpeg";
        if (name.endsWith(".srt")) return "text/plain";
        if (name.endsWith(".txt")) return "text/plain";
        return "application/octet-stream";
    }

    public record MigrationResult(int uploaded, int skipped, int failed) {}
}
