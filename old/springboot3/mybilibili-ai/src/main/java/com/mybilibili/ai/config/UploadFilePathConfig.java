package com.mybilibili.ai.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Configuration;

@Configuration
public class UploadFilePathConfig {

    @Value("${upload.base-path:d:/files/mybilibili/uploads}")
    private String basePath;

    public String getBasePath() {
        return basePath;
    }

    public String getVideoSourceDir(Integer manuscriptId, Integer videoId) {
        return basePath + java.io.File.separator + "manuscripts" + java.io.File.separator
            + manuscriptId + java.io.File.separator + "videos" + java.io.File.separator
            + videoId + java.io.File.separator + "source";
    }

    public String getVideoSourcePath(Integer manuscriptId, Integer videoId) {
        java.io.File sourceDir = new java.io.File(getVideoSourceDir(manuscriptId, videoId));
        if (sourceDir.exists()) {
            java.io.File[] files = sourceDir.listFiles((dir, name) ->
                name.endsWith(".mp4") || name.endsWith(".avi") ||
                name.endsWith(".mov") || name.endsWith(".mkv"));
            if (files != null && files.length > 0) {
                return files[0].getAbsolutePath();
            }
        }
        return getVideoSourceDir(manuscriptId, videoId) + java.io.File.separator + "video.mp4";
    }

    public String getVideoAudioDir(Integer manuscriptId, Integer videoId) {
        return basePath + java.io.File.separator + "manuscripts" + java.io.File.separator
            + manuscriptId + java.io.File.separator + "videos" + java.io.File.separator
            + videoId + java.io.File.separator + "audio";
    }

    public String getAudioPath(Integer manuscriptId, Integer videoId) {
        return getVideoAudioDir(manuscriptId, videoId) + java.io.File.separator + "audio.wav";
    }

    public String getVideoSubtitleDir(Integer manuscriptId, Integer videoId) {
        return basePath + java.io.File.separator + "manuscripts" + java.io.File.separator
            + manuscriptId + java.io.File.separator + "videos" + java.io.File.separator
            + videoId + java.io.File.separator + "subtitles";
    }

    public String getChineseSubtitlePath(Integer manuscriptId, Integer videoId) {
        return getVideoSubtitleDir(manuscriptId, videoId) + java.io.File.separator + "zh-CN.srt";
    }

    public String getAiSummaryDir(Integer manuscriptId, Integer videoId) {
        return basePath + java.io.File.separator + "manuscripts" + java.io.File.separator
            + manuscriptId + java.io.File.separator + "videos" + java.io.File.separator
            + videoId + java.io.File.separator + "summary";
    }

    public String getAiSummaryPath(Integer manuscriptId, Integer videoId) {
        return getAiSummaryDir(manuscriptId, videoId) + java.io.File.separator + "ai-summary.txt";
    }

    public String getVideoDir(Integer manuscriptId, Integer videoId) {
        return basePath + java.io.File.separator + "manuscripts" + java.io.File.separator
            + manuscriptId + java.io.File.separator + "videos" + java.io.File.separator
            + videoId;
    }

    public String getTranscodedDir(Integer manuscriptId, Integer videoId) {
        return getVideoDir(manuscriptId, videoId) + java.io.File.separator + "transcoded";
    }

    public void ensureDirectoryExists(String path) {
        java.io.File dir = new java.io.File(path);
        if (!dir.exists()) {
            dir.mkdirs();
        }
    }
}