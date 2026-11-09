package com.mybilibili.common.dto;

import lombok.Data;

@Data
public class CreatorSettingsDTO {
    private Integer defaultCategoryId;
    private Boolean autoPublish;
    private Boolean commentNotify;
    private Boolean likeNotify;
    private Boolean followNotify;
}
