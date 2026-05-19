package com.mybilibili.common.utils;

import org.springframework.web.multipart.MultipartFile;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.util.UUID;

public class FileUploadUtils {

    private final String projectFolder;

    public FileUploadUtils(String projectFolder) {
        this.projectFolder = projectFolder;
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
        String basePath = projectFolder.endsWith("/") ? projectFolder : projectFolder + "/";
        String fullPath = basePath + folder + "/";

        File directory = new File(fullPath);
        if (!directory.exists()) {
            if (!directory.mkdirs()) {
                throw new IOException("无法创建目录: " + fullPath);
            }
        }

        String originalFilename = file.getOriginalFilename();
        String extension = originalFilename.substring(originalFilename.lastIndexOf("."));
        String fileName = UUID.randomUUID().toString() + extension;

        File destFile = new File(fullPath + fileName);
        if (destFile.exists()) {
            destFile.delete();
        }

        try {
            byte[] bytes = file.getBytes();
            FileOutputStream fos = new FileOutputStream(destFile);
            fos.write(bytes);
            fos.close();
        } catch (Exception e) {
            file.transferTo(destFile);
        }

        return "/" + folder + "/" + fileName;
    }

    public String getFullPath(String relativePath) {
        String cleanPath = relativePath.startsWith("/") ? relativePath.substring(1) : relativePath;
        String normalizedPath = cleanPath.replace("/", File.separator);
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