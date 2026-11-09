package com.mybilibili.web.utils;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.io.File;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.UUID;

/**
 * 上传文件路径工具类 - 新目录结构
 * /uploads/
 *   ├── avatars/          # 用户头像
 *   └── manuscripts/      # 稿件目录
 *       └── {manuscript_id}/
 *           ├── cover.jpg               # 稿件封面
 *           └── videos/
 *               └── {video_id}/
 *                   ├── source/         # 视频源
 *                   ├── transcoded/     # 转码视频
 *                   ├── audio/          # 音频
 *                   ├── subtitles/      # 字幕
 *                   └── summary/        # 总结文档
 */
@Component
public class UploadFilePathUtils {

    @Value("${upload.base-path}")
    private String basePath;

    // ==================== 基础路径方法 ====================

    /**
     * 获取上传根目录
     */
    public String getBasePath() {
        return basePath;
    }

    /**
     * 获取头像根目录：/uploads/avatars/
     */
    public String getAvatarDir() {
        return basePath + File.separator + "avatars";
    }

    /**
     * 获取用户头像目录：/uploads/avatars/{user_id}/
     */
    public String getUserAvatarDir(Integer userId) {
        return getAvatarDir() + File.separator + userId;
    }

    /**
     * 获取稿件根目录：/uploads/manuscripts/
     */
    public String getManuscriptsDir() {
        return basePath + File.separator + "manuscripts";
    }

    /**
     * 获取稿件目录：/uploads/manuscripts/{manuscript_id}/
     */
    public String getManuscriptDir(Integer manuscriptId) {
        return getManuscriptsDir() + File.separator + manuscriptId;
    }

    /**
     * 获取稿件视频根目录：/uploads/manuscripts/{manuscript_id}/videos/
     */
    public String getManuscriptVideosDir(Integer manuscriptId) {
        return getManuscriptDir(manuscriptId) + File.separator + "videos";
    }

    /**
     * 获取视频目录：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/
     */
    public String getVideoDir(Integer manuscriptId, Integer videoId) {
        return getManuscriptVideosDir(manuscriptId) + File.separator + videoId;
    }

    /**
     * 获取视频源目录：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/source/
     */
    public String getVideoSourceDir(Integer manuscriptId, Integer videoId) {
        return getVideoDir(manuscriptId, videoId) + File.separator + "source";
    }

    /**
     * 获取转码视频目录：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/transcoded/
     */
    public String getVideoTranscodedDir(Integer manuscriptId, Integer videoId) {
        return getVideoDir(manuscriptId, videoId) + File.separator + "transcoded";
    }

    /**
     * 获取音频目录：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/audio/
     */
    public String getVideoAudioDir(Integer manuscriptId, Integer videoId) {
        return getVideoDir(manuscriptId, videoId) + File.separator + "audio";
    }

    /**
     * 获取字幕目录：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/subtitles/
     */
    public String getVideoSubtitleDir(Integer manuscriptId, Integer videoId) {
        return getVideoDir(manuscriptId, videoId) + File.separator + "subtitles";
    }

    /**
     * 获取总结文档目录：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/summary/
     */
    public String getVideoSummaryDir(Integer manuscriptId, Integer videoId) {
        return getVideoDir(manuscriptId, videoId) + File.separator + "summary";
    }

    // ==================== 具体文件路径方法 ====================

    /**
     * 获取头像路径：/uploads/avatars/{user_id}/avatar.jpg
     */
    public String getAvatarPath(Integer userId) {
        return getUserAvatarDir(userId) + File.separator + "avatar.jpg";
    }

    /**
     * 获取稿件封面路径：/uploads/manuscripts/{manuscript_id}/cover.jpg
     */
    public String getManuscriptCoverPath(Integer manuscriptId) {
        return getManuscriptDir(manuscriptId) + File.separator + "cover.jpg";
    }

    /**
     * 获取视频源路径：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/source/video.mp4
     */
    public String getVideoSourcePath(Integer manuscriptId, Integer videoId, String ext) {
        return getVideoSourceDir(manuscriptId, videoId) + File.separator + "video" + ext;
    }

    /**
     * 获取视频源路径（默认mp4）：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/source/video.mp4
     */
    public String getVideoSourcePath(Integer manuscriptId, Integer videoId) {
        return getVideoSourcePath(manuscriptId, videoId, ".mp4");
    }

    /**
     * 获取转码视频路径：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/transcoded/{resolution}.mp4
     * @param resolution 分辨率：1080p, 720p, 480p
     */
    public String getTranscodedVideoPath(Integer manuscriptId, Integer videoId, String resolution) {
        return getVideoTranscodedDir(manuscriptId, videoId) + File.separator + resolution + ".mp4";
    }

    /**
     * 获取高清视频路径：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/transcoded/1080p.mp4
     */
    public String getHdVideoPath(Integer manuscriptId, Integer videoId) {
        return getTranscodedVideoPath(manuscriptId, videoId, "1080p");
    }

    /**
     * 获取标清视频路径：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/transcoded/720p.mp4
     */
    public String getSdVideoPath(Integer manuscriptId, Integer videoId) {
        return getTranscodedVideoPath(manuscriptId, videoId, "720p");
    }

    /**
     * 获取流畅视频路径：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/transcoded/480p.mp4
     */
    public String getLdVideoPath(Integer manuscriptId, Integer videoId) {
        return getTranscodedVideoPath(manuscriptId, videoId, "480p");
    }

    /**
     * 获取音频路径：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/audio/audio.wav
     */
    public String getAudioPath(Integer manuscriptId, Integer videoId) {
        return getVideoAudioDir(manuscriptId, videoId) + File.separator + "audio.wav";
    }

    /**
     * 获取字幕路径：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/subtitles/{lang}.srt
     * @param lang 语言代码：zh-CN, en 等
     */
    public String getSubtitlePath(Integer manuscriptId, Integer videoId, String lang) {
        return getVideoSubtitleDir(manuscriptId, videoId) + File.separator + lang + ".srt";
    }

    /**
     * 获取中文字幕路径：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/subtitles/zh-CN.srt
     */
    public String getChineseSubtitlePath(Integer manuscriptId, Integer videoId) {
        return getSubtitlePath(manuscriptId, videoId, "zh-CN");
    }

    /**
     * 获取总结文件路径：/uploads/manuscripts/{manuscript_id}/videos/{video_id}/summary/summary.txt
     */
    public String getSummaryPath(Integer manuscriptId, Integer videoId) {
        return getVideoSummaryDir(manuscriptId, videoId) + File.separator + "summary.txt";
    }

    // ==================== 目录操作工具方法 ====================

    /**
     * 确保目录存在，不存在则创建
     */
    public void ensureDirectoryExists(String path) {
        File dir = new File(path);
        if (!dir.exists()) {
            dir.mkdirs();
        }
    }

    /**
     * 创建头像根目录
     */
    public void createAvatarDirectory() {
        ensureDirectoryExists(getAvatarDir());
    }

    /**
     * 创建用户头像目录
     */
    public void createUserAvatarDirectory(Integer userId) {
        ensureDirectoryExists(getUserAvatarDir(userId));
    }

    /**
     * 创建稿件目录（仅稿件根目录和封面）
     */
    public void createManuscriptDirectory(Integer manuscriptId) {
        ensureDirectoryExists(getManuscriptDir(manuscriptId));
    }

    /**
     * 创建视频所有需要的目录
     */
    public void createVideoDirectories(Integer manuscriptId, Integer videoId) {
        ensureDirectoryExists(getVideoSourceDir(manuscriptId, videoId));
        ensureDirectoryExists(getVideoTranscodedDir(manuscriptId, videoId));
        ensureDirectoryExists(getVideoAudioDir(manuscriptId, videoId));
        ensureDirectoryExists(getVideoSubtitleDir(manuscriptId, videoId));
        ensureDirectoryExists(getVideoSummaryDir(manuscriptId, videoId));
    }

    /**
     * 删除稿件所有相关文件
     */
    public void deleteManuscriptDirectory(Integer manuscriptId) {
        File manuscriptDir = new File(getManuscriptDir(manuscriptId));
        if (manuscriptDir.exists()) {
            deleteDirectory(manuscriptDir);
        }
    }

    /**
     * 删除视频所有相关文件
     */
    public void deleteVideoDirectory(Integer manuscriptId, Integer videoId) {
        File videoDir = new File(getVideoDir(manuscriptId, videoId));
        if (videoDir.exists()) {
            deleteDirectory(videoDir);
        }
    }

    /**
     * 递归删除目录
     */
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

    /**
     * 检查文件是否存在
     */
    public boolean fileExists(String path) {
        File file = new File(path);
        return file.exists() && file.isFile();
    }

    /**
     * 检查目录是否存在
     */
    public boolean directoryExists(String path) {
        File dir = new File(path);
        return dir.exists() && dir.isDirectory();
    }

    // ==================== 访问URL方法（用于前端）====================

    /**
     * 获取头像访问URL
     */
    public String getAvatarUrl(Integer userId) {
        return "/uploads/avatars/" + userId + "/avatar.jpg";
    }

    /**
     * 获取稿件封面访问URL
     */
    public String getManuscriptCoverUrl(Integer manuscriptId) {
        return "/uploads/manuscripts/" + manuscriptId + "/cover.jpg";
    }

    /**
     * 获取视频访问URL
     */
    public String getVideoUrl(Integer manuscriptId, Integer videoId, String resolution) {
        return "/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/transcoded/" + resolution + ".mp4";
    }

    /**
     * 获取高清视频URL
     */
    public String getHdVideoUrl(Integer manuscriptId, Integer videoId) {
        return getVideoUrl(manuscriptId, videoId, "1080p");
    }

    /**
     * 获取标清视频URL
     */
    public String getSdVideoUrl(Integer manuscriptId, Integer videoId) {
        return getVideoUrl(manuscriptId, videoId, "720p");
    }

    /**
     * 获取流畅视频URL
     */
    public String getLdVideoUrl(Integer manuscriptId, Integer videoId) {
        return getVideoUrl(manuscriptId, videoId, "480p");
    }

    /**
     * 获取源视频访问URL
     */
    public String getVideoSourceUrl(Integer manuscriptId, Integer videoId, String ext) {
        return "/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/source/video" + ext;
    }

    /**
     * 获取源视频访问URL（默认mp4）
     */
    public String getVideoSourceUrl(Integer manuscriptId, Integer videoId) {
        return getVideoSourceUrl(manuscriptId, videoId, ".mp4");
    }

    /**
     * 获取字幕访问URL
     */
    public String getSubtitleUrl(Integer manuscriptId, Integer videoId, String lang) {
        return "/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/subtitles/" + lang + ".srt";
    }

    // ==================== 通用图片存储方法（扁平化存储）====================

    /**
     * 获取图片根目录：/uploads/images/
     */
    public String getImagesDir() {
        return basePath + File.separator + "images";
    }

    /**
     * 生成图片文件名：{timestamp}_{uuid}.{ext}
     * 格式示例：20260315143025_a1b2c3d4-e5f6-7890-abcd-ef1234567890.jpg
     *
     * @param originalFilename 原始文件名，用于获取扩展名
     * @return 生成的唯一文件名
     */
    public String generateImageFileName(String originalFilename) {
        String extension = "";
        if (originalFilename != null && originalFilename.contains(".")) {
            extension = originalFilename.substring(originalFilename.lastIndexOf(".")).toLowerCase();
        }
        String timestamp = new SimpleDateFormat("yyyyMMddHHmmss").format(new Date());
        String uuid = UUID.randomUUID().toString();
        return timestamp + "_" + uuid + extension;
    }

    /**
     * 获取图片文件完整路径：/uploads/images/{filename}
     *
     * @param fileName 文件名
     * @return 完整路径
     */
    public String getImagePath(String fileName) {
        return getImagesDir() + File.separator + fileName;
    }

    /**
     * 获取图片访问URL：/uploads/images/{filename}
     *
     * @param fileName 文件名
     * @return 访问URL
     */
    public String getImageUrl(String fileName) {
        return "/uploads/images/" + fileName;
    }

    /**
     * 创建图片存储目录
     */
    public void createImagesDirectory() {
        ensureDirectoryExists(getImagesDir());
    }

    /**
     * 校验是否为允许的图片类型
     *
     * @param contentType 文件MIME类型
     * @return 是否允许
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
     * 校验是否为允许的图片扩展名
     *
     * @param fileName 文件名
     * @return 是否允许
     */
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
