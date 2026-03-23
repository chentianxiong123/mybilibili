package com.mybilibili.admin.dto;

import io.swagger.v3.oas.annotations.media.Schema;

/**
 * SRT 保存到 MongoDB 请求 DTO
 */
@Schema(description = "SRT保存到MongoDB请求参数")
public class SrtSaveRequest {

    @Schema(description = "视频ID", example = "123", required = true)
    private Integer videoId;

    @Schema(description = "SRT文件内容", example = "1\n00:00:01,000 --> 00:00:04,000\n字幕内容", required = true)
    private String srtContent;

    @Schema(description = "语言代码", example = "zh-CN", defaultValue = "zh-CN")
    private String language;

    @Schema(description = "语言名称", example = "中文", defaultValue = "中文")
    private String languageName;

    @Schema(description = "来源标识", example = "admin", defaultValue = "admin")
    private String source;

    @Schema(description = "上传者ID，0表示系统生成", example = "0", defaultValue = "0")
    private Integer uploadedBy;

    public Integer getVideoId() {
        return videoId;
    }

    public void setVideoId(Integer videoId) {
        this.videoId = videoId;
    }

    public String getSrtContent() {
        return srtContent;
    }

    public void setSrtContent(String srtContent) {
        this.srtContent = srtContent;
    }

    public String getLanguage() {
        return language;
    }

    public void setLanguage(String language) {
        this.language = language;
    }

    public String getLanguageName() {
        return languageName;
    }

    public void setLanguageName(String languageName) {
        this.languageName = languageName;
    }

    public String getSource() {
        return source;
    }

    public void setSource(String source) {
        this.source = source;
    }

    public Integer getUploadedBy() {
        return uploadedBy;
    }

    public void setUploadedBy(Integer uploadedBy) {
        this.uploadedBy = uploadedBy;
    }
}
