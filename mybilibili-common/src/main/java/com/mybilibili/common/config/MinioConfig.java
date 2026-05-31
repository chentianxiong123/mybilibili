package com.mybilibili.common.config;

import com.mybilibili.common.storage.MinioStorageService;
import com.mybilibili.common.storage.StorageService;
import io.minio.MinioClient;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.stereotype.Component;

@Slf4j
@Configuration
public class MinioConfig {

    @Component
    @ConfigurationProperties(prefix = "minio")
    public static class MinioProperties {
        private String endpoint = "http://127.0.0.1:9000";
        private String accessKey = "REDACTED_MINIO_CREDENTIAL";
        private String secretKey = "REDACTED_MINIO_CREDENTIAL";
        private String bucketName = "mybilibili";

        public String getEndpoint() { return endpoint; }
        public void setEndpoint(String endpoint) { this.endpoint = endpoint; }
        public String getAccessKey() { return accessKey; }
        public void setAccessKey(String accessKey) { this.accessKey = accessKey; }
        public String getSecretKey() { return secretKey; }
        public void setSecretKey(String secretKey) { this.secretKey = secretKey; }
        public String getBucketName() { return bucketName; }
        public void setBucketName(String bucketName) { this.bucketName = bucketName; }
    }

    @Bean
    @ConditionalOnProperty(name = "minio.endpoint")
    public MinioClient minioClient(MinioProperties properties) {
        return MinioClient.builder()
                .endpoint(properties.getEndpoint())
                .credentials(properties.getAccessKey(), properties.getSecretKey())
                .build();
    }

    @Bean
    @ConditionalOnProperty(name = "minio.endpoint")
    public StorageService storageService(MinioClient minioClient, MinioProperties properties) {
        log.info("Initializing MinIO storage: endpoint={}, bucket={}", properties.getEndpoint(), properties.getBucketName());
        return new MinioStorageService(minioClient, properties.getBucketName(), properties.getEndpoint());
    }
}
