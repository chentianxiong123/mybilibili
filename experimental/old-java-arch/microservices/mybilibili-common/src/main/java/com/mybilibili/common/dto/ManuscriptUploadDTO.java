package com.mybilibili.common.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.Data;
import org.springframework.web.multipart.MultipartFile;

import java.util.List;

@Data
public class ManuscriptUploadDTO {
    @NotBlank(message = "标题不能为空")
    @Size(max = 80, message = "标题最多80个字符")
    private String title;

    @Size(max = 2000, message = "简介最多2000个字符")
    private String description;

    private MultipartFile cover;

    @NotNull(message = "请选择分区")
    private Integer categoryId;

    @Size(max = 10, message = "标签最多10个")
    private List<String> tags;

    private List<VideoItemDTO> videos;

    @Data
    public static class VideoItemDTO {
        @Size(max = 80, message = "分P标题最多80个字符")
        private String title;

        @Size(max = 500, message = "分P描述最多500个字符")
        private String description;

        private MultipartFile video;

        private Integer videoOrder;

        private Integer durationSeconds;
    }
}
