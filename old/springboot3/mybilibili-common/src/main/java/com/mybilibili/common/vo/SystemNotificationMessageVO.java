package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;

@Data
public class SystemNotificationMessageVO {
    private Integer id;
    private String title;
    private String content;
    private String type;
    private String actionText;
    private String actionUrl;
    private Boolean isRead;
    private Date createdAt;
}
