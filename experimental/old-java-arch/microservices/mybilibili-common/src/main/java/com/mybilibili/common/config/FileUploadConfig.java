package com.mybilibili.common.config;

import com.mybilibili.common.utils.UploadFilePathUtils;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class FileUploadConfig {

    @Configuration
    @ConditionalOnProperty(name = "upload.base-path")
    public static class UploadPathConfig {
        @Value("${upload.base-path}")
        private String uploadBasePath;

        @Bean
        public UploadFilePathUtils uploadFilePathUtils() {
            return new UploadFilePathUtils(uploadBasePath);
        }
    }
}
