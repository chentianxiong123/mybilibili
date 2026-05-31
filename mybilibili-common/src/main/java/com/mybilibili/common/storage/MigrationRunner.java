package com.mybilibili.common.storage;

import io.minio.*;
import java.nio.file.*;
import java.nio.file.attribute.BasicFileAttributes;

public class MigrationRunner {

    public static void main(String[] args) throws Exception {
        String endpoint = args.length > 0 ? args[0] : "http://127.0.0.1:9000";
        String accessKey = args.length > 1 ? args[1] : "REDACTED_MINIO_CREDENTIAL";
        String secretKey = args.length > 2 ? args[2] : "REDACTED_MINIO_CREDENTIAL";
        String bucketName = args.length > 3 ? args[3] : "mybilibili";
        String localUploadsPath = args.length > 4 ? args[4] : "d:/files/mybilibili/uploads";

        log("=== MinIO Migration Tool ===");
        log("Endpoint: " + endpoint);
        log("Bucket: " + bucketName);
        log("Local path: " + localUploadsPath);

        MinioClient minioClient = MinioClient.builder()
                .endpoint(endpoint)
                .credentials(accessKey, secretKey)
                .build();

        boolean exists = minioClient.bucketExists(BucketExistsArgs.builder().bucket(bucketName).build());
        if (!exists) {
            minioClient.makeBucket(MakeBucketArgs.builder().bucket(bucketName).build());
            String policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Action\":[\"s3:GetObject\"],\"Resource\":[\"arn:aws:s3:::" + bucketName + "/*\"]}]}";
            minioClient.setBucketPolicy(SetBucketPolicyArgs.builder().bucket(bucketName).config(policy).build());
            log("Bucket created: " + bucketName);
        } else {
            log("Bucket exists: " + bucketName);
        }

        Path uploadsDir = Path.of(localUploadsPath);
        if (!Files.exists(uploadsDir)) {
            log("ERROR: Directory not found: " + uploadsDir);
            System.exit(1);
        }

        int[] counts = {0, 0, 0};
        Files.walkFileTree(uploadsDir, new SimpleFileVisitor<>() {
            @Override
            public FileVisitResult visitFile(Path file, BasicFileAttributes attrs) {
                try {
                    Path relativePath = uploadsDir.relativize(file);
                    String key = relativePath.toString().replace('\\', '/');
                    try {
                        minioClient.statObject(StatObjectArgs.builder().bucket(bucketName).object(key).build());
                        counts[1]++;
                        return FileVisitResult.CONTINUE;
                    } catch (Exception ignored) {}
                    String contentType = detectContentType(file.getFileName().toString());
                    try (var stream = Files.newInputStream(file)) {
                        minioClient.putObject(PutObjectArgs.builder()
                                .bucket(bucketName)
                                .object(key)
                                .stream(stream, attrs.size(), -1)
                                .contentType(contentType)
                                .build());
                        counts[0]++;
                        if (counts[0] % 50 == 0) {
                            log("Progress: uploaded=" + counts[0] + " skipped=" + counts[1] + " failed=" + counts[2]);
                        }
                    }
                } catch (Exception e) {
                    counts[2]++;
                    log("FAILED: " + file + " -> " + e.getMessage());
                }
                return FileVisitResult.CONTINUE;
            }
        });

        log("=== Migration Complete ===");
        log("Uploaded: " + counts[0] + ", Skipped: " + counts[1] + ", Failed: " + counts[2]);
    }

    private static String detectContentType(String name) {
        String lower = name.toLowerCase();
        if (lower.endsWith(".jpg") || lower.endsWith(".jpeg")) return "image/jpeg";
        if (lower.endsWith(".png")) return "image/png";
        if (lower.endsWith(".gif")) return "image/gif";
        if (lower.endsWith(".mp4")) return "video/mp4";
        if (lower.endsWith(".webm")) return "video/webm";
        if (lower.endsWith(".m3u8")) return "application/vnd.apple.mpegurl";
        if (lower.endsWith(".ts")) return "video/mp2t";
        if (lower.endsWith(".wav")) return "audio/wav";
        if (lower.endsWith(".srt") || lower.endsWith(".txt")) return "text/plain";
        if (lower.endsWith(".json")) return "application/json";
        return "application/octet-stream";
    }

    private static void log(String msg) {
        System.out.println("[" + java.time.LocalTime.now().toString().substring(0, 12) + "] " + msg);
    }
}
