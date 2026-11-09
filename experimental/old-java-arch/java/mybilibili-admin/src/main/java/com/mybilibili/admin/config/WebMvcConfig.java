package com.mybilibili.admin.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.ResourceHandlerRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

import java.io.File;

@Configuration
public class WebMvcConfig implements WebMvcConfigurer {

    @Value("${upload.base-path:d:/files/mybilibili/uploads}")
    private String uploadBasePath;

    @Override
    public void addResourceHandlers(ResourceHandlerRegistry registry) {
        // 使用配置文件中的路径，或自动检测
        String uploadPath = uploadBasePath;
        
        // 如果配置的路径不存在，尝试自动检测
        File uploadDir = new File(uploadPath);
        if (!uploadDir.exists()) {
            // 获取项目根目录
            String projectRoot = System.getProperty("user.dir");
            File currentDir = new File(projectRoot);

            // 向上查找直到找到包含uploads目录的mybilibili目录
            while (currentDir != null) {
                File uploadsDir = new File(currentDir, "uploads");
                if (uploadsDir.exists() && uploadsDir.isDirectory()) {
                    uploadPath = uploadsDir.getAbsolutePath();
                    break;
                }

                // 检查mybilibili子目录
                File mybilibiliDir = new File(currentDir, "mybilibili");
                if (mybilibiliDir.exists() && mybilibiliDir.isDirectory()) {
                    File mybilibiliUploadsDir = new File(mybilibiliDir, "uploads");
                    if (mybilibiliUploadsDir.exists() && mybilibiliUploadsDir.isDirectory()) {
                        uploadPath = mybilibiliUploadsDir.getAbsolutePath();
                        break;
                    }
                }

                // 向上级目录查找
                currentDir = currentDir.getParentFile();
            }
        }

        // 调试日志
        System.out.println("【Admin WebMvcConfig】uploadPath: " + uploadPath);
        System.out.println("【Admin WebMvcConfig】uploadDir.exists(): " + new File(uploadPath).exists());

        // 映射 /uploads/** 到文件系统
        String resourceLocation = "file:" + uploadPath.replace("\\", "/") + "/";
        System.out.println("【Admin WebMvcConfig】resourceLocation: " + resourceLocation);

        registry.addResourceHandler("/uploads/**")
                .addResourceLocations(resourceLocation);

        System.out.println("【Admin WebMvcConfig】静态资源映射完成");
    }
}
