package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;

@Data
public class NotificationVO {
    private Integer id;
    private Integer type;
    private String title;
    private String content;
    private String relatedId;
    private Integer isRead;
    private Date createdAt;
    private Date readAt;
    private String typeName; // 通知类型名称
}
