package com.mybilibili.common.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.Data;
import org.springframework.web.multipart.MultipartFile;

import java.util.List;

@Data
public class VideoUploadDTO {
    @NotBlank(message = "标题不能为空")
    @Size(max = 80, message = "标题最多80个字符")
    private String title;

    @Size(max = 2000, message = "简介最多2000个字符")
    private String description;

    private MultipartFile cover;

    private MultipartFile video;

    @NotNull(message = "请选择���区")
    private Integer categoryId;

    @Size(max = 10, message = "标签最多10个")
    private List<String> tags;
}
