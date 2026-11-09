package com.mybilibili.common.utils;

import java.io.File;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.UUID;

public class UploadFilePathUtils {

    private final String basePath;

    public UploadFilePathUtils(String basePath) {
        this.basePath = basePath;
    }

    public String getBasePath() {
        return basePath;
    }

    public String getAvatarDir() {
        return basePath + File.separator + "avatars";
    }

    public String getUserAvatarDir(Integer userId) {
        return getAvatarDir() + File.separator + userId;
    }

    public String getAvatarPath(Integer userId) {
        return getUserAvatarDir(userId) + File.separator + "avatar.jpg";
    }

    public void ensureDirectoryExists(String path) {
        File dir = new File(path);
        if (!dir.exists()) {
            dir.mkdirs();
        }
    }

    public void createAvatarDirectory() {
        ensureDirectoryExists(getAvatarDir());
    }

    public void createUserAvatarDirectory(Integer userId) {
        ensureDirectoryExists(getUserAvatarDir(userId));
    }

    public boolean fileExists(String path) {
        File file = new File(path);
        return file.exists() && file.isFile();
    }

    public String getAvatarUrl(Integer userId) {
        return unsupportedUploadsUrl();
    }

    public String getManuscriptsDir() {
        return basePath + File.separator + "manuscripts";
    }

    public String getManuscriptDir(Integer manuscriptId) {
        return getManuscriptsDir() + File.separator + manuscriptId;
    }

    public String getManuscriptVideosDir(Integer manuscriptId) {
        return getManuscriptDir(manuscriptId) + File.separator + "videos";
    }

    public String getVideoDir(Integer manuscriptId, Integer videoId) {
        return getManuscriptVideosDir(manuscriptId) + File.separator + videoId;
    }

    public String getVideoSourceDir(Integer manuscriptId, Integer videoId) {
        return getVideoDir(manuscriptId, videoId) + File.separator + "source";
    }

    public String getVideoTranscodedDir(Integer manuscriptId, Integer videoId) {
        return getVideoDir(manuscriptId, videoId) + File.separator + "transcoded";
    }

    public String getVideoAudioDir(Integer manuscriptId, Integer videoId) {
        return getVideoDir(manuscriptId, videoId) + File.separator + "audio";
    }

    public String getVideoSubtitleDir(Integer manuscriptId, Integer videoId) {
        return getVideoDir(manuscriptId, videoId) + File.separator + "subtitles";
    }

    public String getVideoSummaryDir(Integer manuscriptId, Integer videoId) {
        return getVideoDir(manuscriptId, videoId) + File.separator + "summary";
    }

    public String getManuscriptCoverPath(Integer manuscriptId) {
        return getManuscriptDir(manuscriptId) + File.separator + "cover.jpg";
    }

    public String getVideoSourcePath(Integer manuscriptId, Integer videoId, String ext) {
        return getVideoSourceDir(manuscriptId, videoId) + File.separator + "video" + ext;
    }

    public String getVideoSourcePath(Integer manuscriptId, Integer videoId) {
        return getVideoSourcePath(manuscriptId, videoId, ".mp4");
    }

    public String getTranscodedVideoPath(Integer manuscriptId, Integer videoId, String resolution) {
        return getVideoTranscodedDir(manuscriptId, videoId) + File.separator + resolution + ".mp4";
    }

    public String getHdVideoPath(Integer manuscriptId, Integer videoId) {
        return getTranscodedVideoPath(manuscriptId, videoId, "1080p");
    }

    public String getSdVideoPath(Integer manuscriptId, Integer videoId) {
        return getTranscodedVideoPath(manuscriptId, videoId, "720p");
    }

    public String getLdVideoPath(Integer manuscriptId, Integer videoId) {
        return getTranscodedVideoPath(manuscriptId, videoId, "480p");
    }

    public String getAudioPath(Integer manuscriptId, Integer videoId) {
        return getVideoAudioDir(manuscriptId, videoId) + File.separator + "audio.wav";
    }

    public String getSubtitlePath(Integer manuscriptId, Integer videoId, String lang) {
        return getVideoSubtitleDir(manuscriptId, videoId) + File.separator + lang + ".srt";
    }

    public String getChineseSubtitlePath(Integer manuscriptId, Integer videoId) {
        return getSubtitlePath(manuscriptId, videoId, "zh-CN");
    }

    public String getSummaryPath(Integer manuscriptId, Integer videoId) {
        return getVideoSummaryDir(manuscriptId, videoId) + File.separator + "ai-summary.txt";
    }

    public void createManuscriptDirectory(Integer manuscriptId) {
        ensureDirectoryExists(getManuscriptDir(manuscriptId));
    }

    public void createVideoDirectories(Integer manuscriptId, Integer videoId) {
        ensureDirectoryExists(getVideoSourceDir(manuscriptId, videoId));
        ensureDirectoryExists(getVideoTranscodedDir(manuscriptId, videoId));
        ensureDirectoryExists(getVideoAudioDir(manuscriptId, videoId));
        ensureDirectoryExists(getVideoSubtitleDir(manuscriptId, videoId));
        ensureDirectoryExists(getVideoSummaryDir(manuscriptId, videoId));
    }

    public void deleteManuscriptDirectory(Integer manuscriptId) {
        File manuscriptDir = new File(getManuscriptDir(manuscriptId));
        if (manuscriptDir.exists()) {
            deleteDirectory(manuscriptDir);
        }
    }

    public void deleteVideoDirectory(Integer manuscriptId, Integer videoId) {
        File videoDir = new File(getVideoDir(manuscriptId, videoId));
        if (videoDir.exists()) {
            deleteDirectory(videoDir);
        }
    }

    private void deleteDirectory(File dir) {
        File[] files = dir.listFiles();
        if (files != null) {
            for (File file : files) {
                if (file.isDirectory()) {
                    deleteDirectory(file);
                } else {
                    file.delete();
                }
            }
        }
        dir.delete();
    }

    public String getManuscriptCoverUrl(Integer manuscriptId) {
        return unsupportedUploadsUrl();
    }

    public String getVideoUrl(Integer manuscriptId, Integer videoId, String resolution) {
        return unsupportedUploadsUrl();
    }

    public String getHdVideoUrl(Integer manuscriptId, Integer videoId) {
        return getVideoUrl(manuscriptId, videoId, "1080p");
    }

    public String getSdVideoUrl(Integer manuscriptId, Integer videoId) {
        return getVideoUrl(manuscriptId, videoId, "720p");
    }

    public String getLdVideoUrl(Integer manuscriptId, Integer videoId) {
        return getVideoUrl(manuscriptId, videoId, "480p");
    }

    public String getVideoSourceUrl(Integer manuscriptId, Integer videoId, String ext) {
        return unsupportedUploadsUrl();
    }

    public String getVideoSourceUrl(Integer manuscriptId, Integer videoId) {
        return getVideoSourceUrl(manuscriptId, videoId, ".mp4");
    }

    public String getSubtitleUrl(Integer manuscriptId, Integer videoId, String lang) {
        return unsupportedUploadsUrl();
    }

    public String getImagesDir() {
        return basePath + File.separator + "images";
    }

    public String generateImageFileName(String originalFilename) {
        String extension = "";
        if (originalFilename != null && originalFilename.contains(".")) {
            extension = originalFilename.substring(originalFilename.lastIndexOf(".")).toLowerCase();
        }
        String timestamp = new SimpleDateFormat("yyyyMMddHHmmss").format(new Date());
        String uuid = UUID.randomUUID().toString();
        return timestamp + "_" + uuid + extension;
    }

    public String getImagePath(String fileName) {
        return getImagesDir() + File.separator + fileName;
    }

    public String getImageUrl(String fileName) {
        return unsupportedUploadsUrl();
    }

    private String unsupportedUploadsUrl() {
        throw new UnsupportedOperationException("Local /uploads URLs are disabled; use StorageService URLs instead");
    }

    public void createImagesDirectory() {
        ensureDirectoryExists(getImagesDir());
    }

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

    public boolean isValidImageExtension(String fileName) {
        if (fileName == null || !fileName.contains(".")) {
            return false;
        }
        String extension = fileName.substring(fileName.lastIndexOf(".")).toLowerCase();
        return extension.equals(".jpg") ||
               extension.equals(".jpeg") ||
               extension.equals(".png") ||
               extension.equals(".gif") ||
               extension.equals(".webp");
    }
}
