package com.mybilibili.interaction.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.ResourceHandlerRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

import java.io.File;

@Configuration
public class WebMvcConfig implements WebMvcConfigurer {

    @Value("${project.folder}")
    private String projectFolder;

    @Override
    public void addResourceHandlers(ResourceHandlerRegistry registry) {
        String uploadPath = projectFolder.replace("/", File.separator).replace("\\", File.separator);
        if (!uploadPath.endsWith(File.separator)) {
            uploadPath = uploadPath + File.separator;
        }

        System.out.println("[WebMvcConfig] uploadPath: " + uploadPath);

        registry.addResourceHandler("/covers/**")
                .addResourceLocations("file:" + uploadPath + "covers" + File.separator);

        registry.addResourceHandler("/videos/**")
                .addResourceLocations("file:" + uploadPath + "videos" + File.separator);

        System.out.println("[WebMvcConfig] Static resource mapping complete:");
        System.out.println("[WebMvcConfig] /covers/** -> file:" + uploadPath + "covers" + File.separator);
        System.out.println("[WebMvcConfig] /videos/** -> file:" + uploadPath + "videos" + File.separator);
    }
}
