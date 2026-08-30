package com.mybilibili.web.utils;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.io.File;

/**
 * 视频文件路径工具类 - 新目录结构
 * /uploads/{video_id}/
 *   ├── source/       # 原始素材
 *   ├── videos/       # 转码后的视频
 *   ├── thumbnails/   # 封面/缩略图
 *   ├── audios/       # 音频文件
 *   ├── subtitles/    # 字幕文件
 *   └── meta/         # 衍生数据（总结等）
 */
@Component
public class VideoFilePathUtils {

    @Value("${upload.video-dir}")
    private String videoBaseDir;

    @Value("${upload.source-dir:source}")
    private String sourceDirName;

    @Value("${upload.transcoded-dir:videos}")
    private String transcodedDirName;

    @Value("${upload.thumbnail-dir:thumbnails}")
    private String thumbnailDirName;

    @Value("${upload.audio-dir:audios}")
    private String audioDirName;

    @Value("${upload.subtitle-dir:subtitles}")
    private String subtitleDirName;

    @Value("${upload.meta-dir:meta}")
    private String metaDirName;

    // ==================== 基础路径方法 ====================

    /**
     * 获取视频根目录：/uploads/{video_id}/
     */
    public String getVideoDir(Integer videoId) {
        return videoBaseDir + File.separator + videoId;
    }

    /**
     * 获取原始素材目录：/uploads/{video_id}/source/
     */
    public String getSourceDir(Integer videoId) {
        return getVideoDir(videoId) + File.separator + sourceDirName;
    }

    /**
     * 获取转码视频目录：/uploads/{video_id}/videos/
     */
    public String getTranscodedDir(Integer videoId) {
        return getVideoDir(videoId) + File.separator + transcodedDirName;
    }

    /**
     * 获取缩略图目录：/uploads/{video_id}/thumbnails/
     */
    public String getThumbnailDir(Integer videoId) {
        return getVideoDir(videoId) + File.separator + thumbnailDirName;
    }

    /**
     * 获取音频目录：/uploads/{video_id}/audios/
     */
    public String getAudioDir(Integer videoId) {
        return getVideoDir(videoId) + File.separator + audioDirName;
    }

    /**
     * 获取字幕目录：/uploads/{video_id}/subtitles/
     */
    public String getSubtitleDir(Integer videoId) {
        return getVideoDir(videoId) + File.separator + subtitleDirName;
    }

    /**
     * 获取元数据目录：/uploads/{video_id}/meta/
     */
    public String getMetaDir(Integer videoId) {
        return getVideoDir(videoId) + File.separator + metaDirName;
    }

    // ==================== 具体文件路径方法 ====================

    /**
     * 获取原始视频路径：/uploads/{video_id}/source/video.mp4
     */
    public String getSourceVideoPath(Integer videoId, String ext) {
        return getSourceDir(videoId) + File.separator + "video" + ext;
    }

    /**
     * 获取原始封面路径：/uploads/{video_id}/source/poster.jpg
     */
    public String getSourcePosterPath(Integer videoId) {
        return getSourceDir(videoId) + File.separator + "poster.jpg";
    }

    /**
     * 获取转码视频路径：/uploads/{video_id}/videos/1080p.mp4
     * @param resolution 分辨率：1080p, 720p, 480p
     */
    public String getTranscodedVideoPath(Integer videoId, String resolution) {
        return getTranscodedDir(videoId) + File.separator + resolution + ".mp4";
    }

    /**
     * 获取高清视频路径：/uploads/{video_id}/videos/1080p.mp4
     */
    public String getHdVideoPath(Integer videoId) {
        return getTranscodedVideoPath(videoId, "1080p");
    }

    /**
     * 获取标清视频路径：/uploads/{video_id}/videos/720p.mp4
     */
    public String getSdVideoPath(Integer videoId) {
        return getTranscodedVideoPath(videoId, "720p");
    }

    /**
     * 获取流畅视频路径：/uploads/{video_id}/videos/480p.mp4
     */
    public String getLdVideoPath(Integer videoId) {
        return getTranscodedVideoPath(videoId, "480p");
    }

    /**
     * 获取封面路径：/uploads/{video_id}/thumbnails/cover.jpg
     */
    public String getCoverPath(Integer videoId) {
        return getThumbnailDir(videoId) + File.separator + "cover.jpg";
    }

    /**
     * 获取音频路径：/uploads/{video_id}/audios/audio.wav
     */
    public String getAudioPath(Integer videoId) {
        return getAudioDir(videoId) + File.separator + "audio.wav";
    }

    /**
     * 获取字幕路径：/uploads/{video_id}/subtitles/zh-CN.srt
     * @param lang 语言代码：zh-CN, en 等
     */
    public String getSubtitlePath(Integer videoId, String lang) {
        return getSubtitleDir(videoId) + File.separator + lang + ".srt";
    }

    /**
     * 获取中文字幕路径：/uploads/{video_id}/subtitles/zh-CN.srt
     */
    public String getChineseSubtitlePath(Integer videoId) {
        return getSubtitlePath(videoId, "zh-CN");
    }

    /**
     * 获取总结文件路径：/uploads/{video_id}/meta/summary.txt
     */
    public String getSummaryPath(Integer videoId) {
        return getMetaDir(videoId) + File.separator + "summary.txt";
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
     * 创建视频所有需要的目录
     */
    public void createVideoDirectories(Integer videoId) {
        ensureDirectoryExists(getSourceDir(videoId));
        ensureDirectoryExists(getTranscodedDir(videoId));
        ensureDirectoryExists(getThumbnailDir(videoId));
        ensureDirectoryExists(getAudioDir(videoId));
        ensureDirectoryExists(getSubtitleDir(videoId));
        ensureDirectoryExists(getMetaDir(videoId));
    }

    /**
     * 删除视频所有相关文件
     */
    public void deleteVideoDirectory(Integer videoId) {
        File videoDir = new File(getVideoDir(videoId));
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
     * 获取视频访问URL
     */
    public String getVideoUrl(Integer videoId, String resolution) {
        return "/uploads/" + videoId + "/videos/" + resolution + ".mp4";
    }

    /**
     * 获取封面访问URL
     */
    public String getCoverUrl(Integer videoId) {
        return "/uploads/" + videoId + "/thumbnails/cover.jpg";
    }

    /**
     * 获取字幕访问URL
     */
    public String getSubtitleUrl(Integer videoId, String lang) {
        return "/uploads/" + videoId + "/subtitles/" + lang + ".srt";
    }
}
