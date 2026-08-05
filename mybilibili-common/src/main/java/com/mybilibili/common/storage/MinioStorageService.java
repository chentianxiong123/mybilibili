package com.mybilibili.common.storage;

import io.minio.*;
import io.minio.http.Method;
import io.minio.messages.Item;
import lombok.extern.slf4j.Slf4j;

import java.io.InputStream;
import java.util.concurrent.TimeUnit;

@Slf4j
public class MinioStorageService implements StorageService {

    private final MinioClient minioClient;
    private final MinioClient publicMinioClient;
    private final String bucketName;
    private final String publicEndpoint;

    public MinioStorageService(MinioClient minioClient, MinioClient publicMinioClient, String bucketName, String publicEndpoint) {
        this.minioClient = minioClient;
        this.publicMinioClient = publicMinioClient;
        this.bucketName = bucketName;
        this.publicEndpoint = publicEndpoint;
        initBucket();
    }

    private void initBucket() {
        try {
            boolean exists = minioClient.bucketExists(BucketExistsArgs.builder().bucket(bucketName).build());
            if (!exists) {
                minioClient.makeBucket(MakeBucketArgs.builder().bucket(bucketName).build());
                String policy = """
                    {
                      "Version": "2012-10-17",
                      "Statement": [
                        {
                          "Effect": "Allow",
                          "Principal": {"AWS": ["*"]},
                          "Action": ["s3:GetObject"],
                          "Resource": ["arn:aws:s3:::%s/*"]
                        }
                      ]
                    }
                    """.formatted(bucketName);
                minioClient.setBucketPolicy(SetBucketPolicyArgs.builder().bucket(bucketName).config(policy).build());
                log.info("MinIO bucket created with public read policy: {}", bucketName);
            }
        } catch (Exception e) {
            log.error("MinIO bucket initialization failed: {}", e.getMessage(), e);
        }
    }

    @Override
    public String upload(String key, InputStream data, long size, String contentType) {
        try {
            minioClient.putObject(PutObjectArgs.builder()
                    .bucket(bucketName)
                    .object(key)
                    .stream(data, size, -1)
                    .contentType(contentType)
                    .build());
            return getPublicUrl(key);
        } catch (Exception e) {
            throw new RuntimeException("MinIO upload failed: " + key, e);
        }
    }

    @Override
    public String upload(String key, InputStream data, String contentType) {
        try {
            byte[] bytes = data.readAllBytes();
            return upload(key, new java.io.ByteArrayInputStream(bytes), bytes.length, contentType);
        } catch (Exception e) {
            throw new RuntimeException("MinIO upload failed: " + key, e);
        }
    }

    @Override
    public InputStream download(String key) {
        try {
            return minioClient.getObject(GetObjectArgs.builder()
                    .bucket(bucketName)
                    .object(key)
                    .build());
        } catch (Exception e) {
            throw new RuntimeException("MinIO download failed: " + key, e);
        }
    }

    @Override
    public String getPresignedUrl(String key, int expirySeconds) {
        try {
            return publicMinioClient.getPresignedObjectUrl(GetPresignedObjectUrlArgs.builder()
                    .method(Method.GET)
                    .bucket(bucketName)
                    .object(key)
                    .expiry(expirySeconds, TimeUnit.SECONDS)
                    .build());
        } catch (Exception e) {
            throw new RuntimeException("MinIO presigned URL failed: " + key, e);
        }
    }

    @Override
    public String getPublicUrl(String key) {
        String base = publicEndpoint.endsWith("/") ? publicEndpoint.substring(0, publicEndpoint.length() - 1) : publicEndpoint;
        return base + "/" + bucketName + "/" + key;
    }

    @Override
    public boolean exists(String key) {
        try {
            minioClient.statObject(StatObjectArgs.builder().bucket(bucketName).object(key).build());
            return true;
        } catch (Exception e) {
            return false;
        }
    }

    @Override
    public void delete(String key) {
        try {
            minioClient.removeObject(RemoveObjectArgs.builder().bucket(bucketName).object(key).build());
        } catch (Exception e) {
            log.warn("MinIO delete failed: {}", key, e);
        }
    }

    @Override
    public void deletePrefix(String prefix) {
        try {
            Iterable<io.minio.Result<Item>> results = minioClient.listObjects(ListObjectsArgs.builder()
                    .bucket(bucketName)
                    .prefix(prefix)
                    .recursive(true)
                    .build());
            for (io.minio.Result<Item> result : results) {
                Item item = result.get();
                minioClient.removeObject(RemoveObjectArgs.builder().bucket(bucketName).object(item.objectName()).build());
            }
        } catch (Exception e) {
            log.warn("MinIO prefix delete failed: {}", prefix, e);
        }
    }

    @Override
    public void copy(String sourceKey, String targetKey) {
        try {
            minioClient.copyObject(CopyObjectArgs.builder()
                    .bucket(bucketName)
                    .source(CopySource.builder().bucket(bucketName).object(sourceKey).build())
                    .object(targetKey)
                    .build());
        } catch (Exception e) {
            throw new RuntimeException("MinIO copy failed: " + sourceKey + " -> " + targetKey, e);
        }
    }
}
