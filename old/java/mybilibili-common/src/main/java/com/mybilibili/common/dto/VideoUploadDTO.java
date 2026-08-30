package com.mybilibili.common.dto;

import lombok.Data;
import org.springframework.web.multipart.MultipartFile;

import java.util.List;

@Data
public class VideoUploadDTO {
    private String title;
    private String description;
    private MultipartFile cover;
    private MultipartFile video;
    private Integer categoryId;
    private List<String> tags;
}