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
        private String accessKey = "";
        private String secretKey = "";
        private String bucketName = "mybilibili";
        private String publicEndpoint = "";

        public String getEndpoint() { return endpoint; }
        public void setEndpoint(String endpoint) { this.endpoint = endpoint; }
        public String getAccessKey() { return accessKey; }
        public void setAccessKey(String accessKey) { this.accessKey = accessKey; }
        public String getSecretKey() { return secretKey; }
        public void setSecretKey(String secretKey) { this.secretKey = secretKey; }
        public String getBucketName() { return bucketName; }
        public void setBucketName(String bucketName) { this.bucketName = bucketName; }
        public String getPublicEndpoint() { return publicEndpoint; }
        public void setPublicEndpoint(String publicEndpoint) { this.publicEndpoint = publicEndpoint; }

        String effectivePublicEndpoint() {
            return publicEndpoint == null || publicEndpoint.isBlank() ? endpoint : publicEndpoint;
        }
    }

    @Bean
    @ConditionalOnProperty(name = "minio.endpoint")
    public MinioClient minioClient(MinioProperties properties) {
        requireText(properties.getAccessKey(), "minio.access-key");
        requireText(properties.getSecretKey(), "minio.secret-key");
        return MinioClient.builder()
                .endpoint(properties.getEndpoint())
                .credentials(properties.getAccessKey(), properties.getSecretKey())
                .build();
    }

    @Bean
    @ConditionalOnProperty(name = "minio.endpoint")
    public StorageService storageService(MinioClient minioClient, MinioProperties properties) {
        MinioClient publicMinioClient = MinioClient.builder()
                .endpoint(properties.effectivePublicEndpoint())
                .credentials(properties.getAccessKey(), properties.getSecretKey())
                .build();
        log.info("Initializing MinIO storage: endpoint={}, publicEndpoint={}, bucket={}",
                properties.getEndpoint(), properties.effectivePublicEndpoint(), properties.getBucketName());
        return new MinioStorageService(minioClient, publicMinioClient, properties.getBucketName(),
                properties.effectivePublicEndpoint());
    }

    private void requireText(String value, String propertyName) {
        if (value == null || value.isBlank()) {
            throw new IllegalStateException(propertyName + " is required");
        }
    }
}
