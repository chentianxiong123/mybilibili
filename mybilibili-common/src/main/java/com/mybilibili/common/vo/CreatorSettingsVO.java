package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;

@Data
public class CreatorSettingsVO {
    private Integer id;
    private Integer userId;
    private Integer defaultCategoryId;
    private Boolean autoPublish;
    private Boolean commentNotify;
    private Boolean likeNotify;
    private Boolean followNotify;
    private Date createdAt;
    private Date updatedAt;
}
