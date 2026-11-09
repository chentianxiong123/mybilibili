package com.mybilibili.common.dto;

import lombok.Data;
import org.springframework.web.multipart.MultipartFile;

import java.util.List;

@Data
public class ManuscriptUploadDTO {
    private String title;
    private String description;
    private MultipartFile cover;
    private Integer categoryId;
    private List<String> tags;

    // 视频列表（支持多视频分P）
    private List<VideoItemDTO> videos;

    @Data
    public static class VideoItemDTO {
        private String title;           // 分P标题
        private String description;     // 分P描述
        private MultipartFile video;    // 视频文件
        private Integer videoOrder;     // 分P顺序
        private Integer durationSeconds; // 视频时长（秒）
    }
}
