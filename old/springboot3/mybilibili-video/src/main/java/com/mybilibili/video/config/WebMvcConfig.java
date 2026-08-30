package com.mybilibili.video.config;

import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.ResourceHandlerRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

import java.io.File;

@Configuration
public class WebMvcConfig implements WebMvcConfigurer {

    @Override
    public void addResourceHandlers(ResourceHandlerRegistry registry) {
        String projectRoot = findProjectRoot();
        String uploadPath = projectRoot + File.separator + "uploads";

        System.out.println("[WebMvcConfig] projectRoot: " + projectRoot);
        System.out.println("[WebMvcConfig] uploadPath: " + uploadPath);

        String resourceLocation = "file:" + uploadPath + File.separator;

        registry.addResourceHandler("/uploads/**")
                .addResourceLocations(resourceLocation);

        System.out.println("[WebMvcConfig] Static resource mapping complete: /uploads/** -> " + resourceLocation);
    }

    private String findProjectRoot() {
        String userDir = System.getProperty("user.dir");
        File currentDir = new File(userDir);

        File uploadsDir = new File(currentDir, "uploads");
        if (uploadsDir.exists() && uploadsDir.isDirectory()) {
            return currentDir.getAbsolutePath();
        }

        File parentDir = currentDir.getParentFile();
        if (parentDir != null) {
            uploadsDir = new File(parentDir, "uploads");
            if (uploadsDir.exists() && uploadsDir.isDirectory()) {
                return parentDir.getAbsolutePath();
            }
        }

        return "d:/files/mybilibili";
    }
}
