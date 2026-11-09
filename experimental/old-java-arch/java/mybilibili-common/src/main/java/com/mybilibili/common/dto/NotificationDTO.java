package com.mybilibili.common.dto;

import lombok.Data;

@Data
public class NotificationDTO {
    private Integer userId;
    private Integer type;
    private String title;
    private String content;
    private String relatedId;
}
