package com.mybilibili.common.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.Data;
import org.springframework.web.multipart.MultipartFile;

import java.util.List;

@Data
public class ManuscriptUpdateDTO {
    @NotBlank(message = "标题不能为空")
    @Size(max = 80, message = "标题最多80个字符")
    private String title;

    @Size(max = 2000, message = "简介最多2000个字符")
    private String description;

    @NotNull(message = "请选择分区")
    private Integer categoryId;

    @Size(max = 10, message = "标签最多10个")
    private List<String> tags;

    private List<VideoUpdateDTO> videos;

    private MultipartFile cover;

    @Data
    public static class VideoUpdateDTO {
        @NotNull(message = "分P ID不能为空")
        private Integer id;

        @Size(max = 80, message = "分P标题最多80个字符")
        private String title;

        private Integer videoOrder;

        private Integer durationSeconds;

        private MultipartFile file;
    }
}
