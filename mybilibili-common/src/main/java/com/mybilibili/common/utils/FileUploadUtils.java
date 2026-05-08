package com.mybilibili.common.utils;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.web.multipart.MultipartFile;
import javax.annotation.PostConstruct;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.UUID;

@Component
public class FileUploadUtils {

    @Value("${project.folder}")
    private String projectFolder;
    
    public FileUploadUtils() {
        System.out.println("FileUploadUtils constructor called");
    }
    
    @PostConstruct
    public void init() {
        System.out.println("FileUploadUtils initialized with projectFolder: " + projectFolder);
    }

    private final String VIDEO_FOLDER = "videos";
    private final String COVER_FOLDER = "covers";

    public String uploadVideo(MultipartFile file) throws IOException {
        return uploadFile(file, VIDEO_FOLDER);
    }

    public String uploadCover(MultipartFile file) throws IOException {
        return uploadFile(file, COVER_FOLDER);
    }

    private String uploadFile(MultipartFile file, String folder) throws IOException {
        // 打印projectFolder配置
        System.out.println("projectFolder配置: " + projectFolder);
        
        // 确保基础路径正确处理
        String basePath = projectFolder.endsWith("/") ? projectFolder : projectFolder + "/";
        
        // 直接使用根目录，不创建子目录
        String fullPath = basePath + folder + "/";
        
        System.out.println("文件上传根目录: " + basePath);
        System.out.println("使用目录: " + fullPath);
        
        // 尝试创建目录
        File directory = new File(fullPath);
        System.out.println("目录绝对路径: " + directory.getAbsolutePath());
        System.out.println("目录是否存在: " + directory.exists());
        System.out.println("目录是否可写: " + directory.canWrite());
        
        if (!directory.exists()) {
            System.out.println("尝试创建目录");
            boolean created = directory.mkdirs();
            if (!created) {
                System.out.println("目录创建失败");
                System.out.println("目录父目录是否存在: " + directory.getParentFile().exists());
                System.out.println("目录父目录是否可写: " + directory.getParentFile().canWrite());
                throw new IOException("无法创建目录: " + fullPath);
            } else {
                System.out.println("目录创建成功");
                System.out.println("创建后目录是否存在: " + directory.exists());
                System.out.println("创建后目录是否可写: " + directory.canWrite());
            }
        } else {
            System.out.println("目录已存在");
        }

        // 生成文件名
        String originalFilename = file.getOriginalFilename();
        System.out.println("原始文件名: " + originalFilename);
        String extension = originalFilename.substring(originalFilename.lastIndexOf("."));
        String fileName = UUID.randomUUID().toString() + extension;
        System.out.println("生成的文件名: " + fileName);

        // 保存文件
        File destFile = new File(fullPath + fileName);
        System.out.println("尝试保存文件: " + destFile.getAbsolutePath());
        System.out.println("文件父目录是否存在: " + destFile.getParentFile().exists());
        System.out.println("文件父目录是否可写: " + destFile.getParentFile().canWrite());
        
        try {
            // 先检查文件是否存在，如果存在则删除
            if (destFile.exists()) {
                System.out.println("文件已存在，尝试删除");
                boolean deleted = destFile.delete();
                if (!deleted) {
                    System.out.println("文件删除失败");
                }
            }
            
            // 尝试使用transferTo方法保存文件
            System.out.println("开始保存文件...");
            System.out.println("文件大小: " + file.getSize());
            System.out.println("文件类型: " + file.getContentType());
            
            // 尝试使用字节流方式保存文件
            try {
                System.out.println("尝试使用字节流方式保存文件...");
                byte[] bytes = file.getBytes();
                FileOutputStream fos = new FileOutputStream(destFile);
                fos.write(bytes);
                fos.close();
                System.out.println("字节流保存文件成功");
            } catch (Exception e) {
                System.out.println("字节流保存文件失败: " + e.getMessage());
                e.printStackTrace();
                // 回退到transferTo方法
                System.out.println("回退到transferTo方法...");
                file.transferTo(destFile);
                System.out.println("transferTo方法保存文件成功");
            }
            
            System.out.println("文件保存成功");
            System.out.println("文件是否存在: " + destFile.exists());
            System.out.println("文件大小: " + destFile.length());
        } catch (Exception e) {
            System.out.println("文件保存失败: " + e.getMessage());
            e.printStackTrace();
            throw new IOException("文件保存失败: " + e.getMessage());
        }

        // 返回以 / 开头的绝对路径
        return "/" + folder + "/" + fileName;
    }
    
    /**
     * 创建目录，确保所有父目录都存在
     * @param path 目录路径
     * @return 创建的目录对象，如果创建失败返回null
     */
    private File createDirectory(String path) {
        File directory = new File(path);
        System.out.println("创建目录: " + directory.getAbsolutePath());
        
        // 确保所有父目录都存在
        if (!directory.exists()) {
            boolean created = directory.mkdirs();
            if (!created) {
                System.out.println("目录创建失败");
                System.out.println("目录是否可写: " + directory.canWrite());
                return null;
            } else {
                System.out.println("目录创建成功");
            }
        } else {
            System.out.println("目录已存在");
        }
        
        return directory;
    }

    public String getFullPath(String relativePath) {
        String cleanPath = relativePath.startsWith("/") ? relativePath.substring(1) : relativePath;
        // 统一使用系统默认的文件分隔符
        String normalizedPath = cleanPath.replace("/", File.separator);
        // 确保basePath也使用系统默认的文件分隔符
        String basePath = projectFolder.replace("/", File.separator).replace("\\", File.separator);
        if (!basePath.endsWith(File.separator)) {
            basePath = basePath + File.separator;
        }
        return basePath + normalizedPath;
    }

    public String getAccessUrl(String relativePath) {
        if (relativePath.startsWith(projectFolder)) {
            return "/" + relativePath.substring(projectFolder.length());
        }
        return "/" + relativePath;
    }

    public String getVideoDirectory() {
        return projectFolder + VIDEO_FOLDER + "/";
    }
}