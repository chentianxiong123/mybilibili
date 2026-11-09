package com.mybilibili.admin.utils;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.io.File;

@Component
public class UploadFilePathUtils {

    @Value("${upload.base-path:d:/files/mybilibili/uploads}")
    private String basePath;

    public String getBasePath() {
        return basePath;
    }

    /**
     * 获取视频源文件目录
     */
    public String getVideoSourceDir(Integer manuscriptId, Integer videoId) {
        return basePath + File.separator + "manuscripts" + File.separator 
            + manuscriptId + File.separator + "videos" + File.separator 
            + videoId + File.separator + "source";
    }

    /**
     * 获取视频源文件路径
     */
    public String getVideoSourcePath(Integer manuscriptId, Integer videoId) {
        // 查找源目录中的视频文件
        File sourceDir = new File(getVideoSourceDir(manuscriptId, videoId));
        if (sourceDir.exists()) {
            File[] files = sourceDir.listFiles((dir, name) -> 
                name.endsWith(".mp4") || name.endsWith(".avi") || 
                name.endsWith(".mov") || name.endsWith(".mkv"));
            if (files != null && files.length > 0) {
                return files[0].getAbsolutePath();
            }
        }
        return getVideoSourceDir(manuscriptId, videoId) + File.separator + "video.mp4";
    }

    /**
     * 获取视频转码输出目录
     */
    public String getVideoTranscodedDir(Integer manuscriptId, Integer videoId) {
        return basePath + File.separator + "manuscripts" + File.separator 
            + manuscriptId + File.separator + "videos" + File.separator 
            + videoId + File.separator + "transcoded";
    }

    /**
     * 获取音频目录
     */
    public String getVideoAudioDir(Integer manuscriptId, Integer videoId) {
        return basePath + File.separator + "manuscripts" + File.separator 
            + manuscriptId + File.separator + "videos" + File.separator 
            + videoId + File.separator + "audio";
    }

    /**
     * 获取音频文件路径
     */
    public String getAudioPath(Integer manuscriptId, Integer videoId) {
        return getVideoAudioDir(manuscriptId, videoId) + File.separator + "audio.wav";
    }

    /**
     * 获取字幕目录
     */
    public String getVideoSubtitleDir(Integer manuscriptId, Integer videoId) {
        return basePath + File.separator + "manuscripts" + File.separator 
            + manuscriptId + File.separator + "videos" + File.separator 
            + videoId + File.separator + "subtitles";
    }

    /**
     * 获取中文字幕路径
     */
    public String getChineseSubtitlePath(Integer manuscriptId, Integer videoId) {
        return getVideoSubtitleDir(manuscriptId, videoId) + File.separator + "zh-CN.srt";
    }

    /**
     * 获取字幕路径
     */
    public String getSubtitlePath(Integer manuscriptId, Integer videoId, String language) {
        return getVideoSubtitleDir(manuscriptId, videoId) + File.separator + language + ".srt";
    }

    /**
     * 获取高清视频URL
     */
    public String getHdVideoUrl(Integer manuscriptId, Integer videoId) {
        return "/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/transcoded/1080p.mp4";
    }

    /**
     * 获取标清视频URL
     */
    public String getSdVideoUrl(Integer manuscriptId, Integer videoId) {
        return "/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/transcoded/720p.mp4";
    }

    /**
     * 获取流畅视频URL
     */
    public String getLdVideoUrl(Integer manuscriptId, Integer videoId) {
        return "/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/transcoded/480p.mp4";
    }

    /**
     * 获取用户上传字幕目录
     */
    public String getUserSubtitleDir(Integer manuscriptId, Integer videoId, String uploadId) {
        return basePath + File.separator + "manuscripts" + File.separator
            + manuscriptId + File.separator + "videos" + File.separator
            + videoId + File.separator + "user-subtitles" + File.separator
            + uploadId;
    }

    /**
     * 确保目录存在
     */
    public void ensureDirectoryExists(String path) {
        File dir = new File(path);
        if (!dir.exists()) {
            dir.mkdirs();
        }
    }

    /**
     * 获取AI摘要目录（使用现有的summary文件夹）
     */
    public String getAiSummaryDir(Integer manuscriptId, Integer videoId) {
        return basePath + File.separator + "manuscripts" + File.separator
            + manuscriptId + File.separator + "videos" + File.separator
            + videoId + File.separator + "summary";
    }

    /**
     * 获取AI摘要文件路径
     */
    public String getAiSummaryPath(Integer manuscriptId, Integer videoId) {
        return getAiSummaryDir(manuscriptId, videoId) + File.separator + "ai-summary.txt";
    }

    /**
     * 获取AI摘要访问URL
     */
    public String getAiSummaryUrl(Integer manuscriptId, Integer videoId) {
        return "/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/summary/ai-summary.txt";
    }

    // ==================== 图片上传相关方法 ====================

    /**
     * 获取图片存储目录
     */
    public String getImagesDir() {
        return basePath + File.separator + "images";
    }

    /**
     * 创建图片目录
     */
    public void createImagesDirectory() {
        ensureDirectoryExists(getImagesDir());
    }

    /**
     * 生成图片文件名
     */
    public String generateImageFileName(String originalFilename) {
        String extension = "";
        if (originalFilename != null && originalFilename.contains(".")) {
            extension = originalFilename.substring(originalFilename.lastIndexOf("."));
        }
        return System.currentTimeMillis() + "_" + (int)(Math.random() * 10000) + extension;
    }

    /**
     * 获取图片完整路径
     */
    public String getImagePath(String fileName) {
        return getImagesDir() + File.separator + fileName;
    }

    /**
     * 获取图片访问URL
     */
    public String getImageUrl(String fileName) {
        return "/uploads/images/" + fileName;
    }

    /**
     * 校验是否为有效的图片类型
     */
    public boolean isValidImageType(String contentType) {
        if (contentType == null) {
            return false;
        }
        return contentType.equals("image/jpeg") ||
               contentType.equals("image/jpg") ||
               contentType.equals("image/png") ||
               contentType.equals("image/gif") ||
               contentType.equals("image/webp");
    }

    /**
     * 校验是否为有效的图片扩展名
     */
    public boolean isValidImageExtension(String filename) {
        if (filename == null) {
            return false;
        }
        String lowerName = filename.toLowerCase();
        return lowerName.endsWith(".jpg") ||
               lowerName.endsWith(".jpeg") ||
               lowerName.endsWith(".png") ||
               lowerName.endsWith(".gif") ||
               lowerName.endsWith(".webp");
    }
}
