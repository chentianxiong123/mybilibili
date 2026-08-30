package com.mybilibili.admin.dto;

import io.swagger.v3.oas.annotations.media.Schema;

/**
 * SRT 解析请求 DTO
 */
@Schema(description = "SRT解析请求参数")
public class SrtParseRequest {

    @Schema(description = "SRT文件内容", example = "1\n00:00:01,000 --> 00:00:04,000\n字幕内容", required = true)
    private String srtContent;

    public String getSrtContent() {
        return srtContent;
    }

    public void setSrtContent(String srtContent) {
        this.srtContent = srtContent;
    }
}
