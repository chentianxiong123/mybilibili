package com.mybilibili.web.config;

import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.ResourceHandlerRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

import java.io.File;

@Configuration
public class WebMvcConfig implements WebMvcConfigurer {

    @Override
    public void addResourceHandlers(ResourceHandlerRegistry registry) {
        // 获取项目根目录（mybilibili目录）
        String projectRoot = System.getProperty("user.dir");
        
        // 处理各种可能的启动路径情况
        // 情况1: 从mybilibili-java/mybilibili-web/target启动（jar包运行）
        // 情况2: 从mybilibili-java启动（IDE运行）
        // 情况3: 从mybilibili启动
        
        File currentDir = new File(projectRoot);
        
        // 向上查找直到找到包含uploads目录的mybilibili目录
        while (currentDir != null) {
            // 检查当前目录是否是mybilibili目录（包含uploads子目录）
            File uploadsDir = new File(currentDir, "uploads");
            if (uploadsDir.exists() && uploadsDir.isDirectory()) {
                projectRoot = currentDir.getAbsolutePath();
                break;
            }
            
            // 检查mybilibili子目录
            File mybilibiliDir = new File(currentDir, "mybilibili");
            if (mybilibiliDir.exists() && mybilibiliDir.isDirectory()) {
                File mybilibiliUploadsDir = new File(mybilibiliDir, "uploads");
                if (mybilibiliUploadsDir.exists() && mybilibiliUploadsDir.isDirectory()) {
                    projectRoot = mybilibiliDir.getAbsolutePath();
                    break;
                }
            }
            
            // 向上级目录查找
            currentDir = currentDir.getParentFile();
        }
        
        // 如果找不到，使用默认路径
        if (currentDir == null) {
            // 尝试使用绝对路径
            projectRoot = "d:/files/mybilibili";
        }
        
        String uploadPath = projectRoot + File.separator + "uploads";

        // 调试日志
        System.out.println("【WebMvcConfig调试】projectRoot: " + projectRoot);
        System.out.println("【WebMvcConfig调试】uploadPath: " + uploadPath);

        // 检查目录是否存在
        File uploadDir = new File(uploadPath);
        System.out.println("【WebMvcConfig调试】uploadDir.exists(): " + uploadDir.exists());
        System.out.println("【WebMvcConfig调试】uploadDir.isDirectory(): " + uploadDir.isDirectory());

        // 检查具体文件是否存在
        File coverFile = new File(uploadPath + File.separator + "manuscripts" + File.separator + "10" + File.separator + "cover.jpg");
        System.out.println("【WebMvcConfig调试】coverFile.exists(): " + coverFile.exists());
        System.out.println("【WebMvcConfig调试】coverFile.absolutePath: " + coverFile.getAbsolutePath());

        // 映射 /uploads/** 到文件系统
        String resourceLocation = "file:" + uploadPath.replace("\\", "/") + "/";
        System.out.println("【WebMvcConfig调试】resourceLocation: " + resourceLocation);

        registry.addResourceHandler("/uploads/**")
                .addResourceLocations(resourceLocation);

        // 保留旧的映射（兼容旧数据）
        registry.addResourceHandler("/covers/**")
                .addResourceLocations("file:" + uploadPath.replace("\\", "/") + "/covers/");

        registry.addResourceHandler("/videos/**")
                .addResourceLocations("file:" + uploadPath.replace("\\", "/") + "/videos/");

        System.out.println("【WebMvcConfig调试】静态资源映射完成");
    }
}
