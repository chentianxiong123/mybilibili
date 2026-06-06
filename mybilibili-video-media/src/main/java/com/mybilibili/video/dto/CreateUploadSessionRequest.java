package com.mybilibili.video.dto;

import lombok.Data;

import java.util.List;

@Data
public class CreateUploadSessionRequest {
    private String title;
    private String description;
    private Integer categoryId;
    private List<String> tags;
    private List<VideoPart> videos;

    @Data
    public static class VideoPart {
        private String title;
        private String fileName;
        private Long size;
        private Integer videoOrder;
        private Integer durationSeconds;
        private Integer totalChunks;
    }
}
