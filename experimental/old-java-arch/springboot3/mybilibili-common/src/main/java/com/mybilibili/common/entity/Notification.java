package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class Notification {
    private Integer id;
    private Integer userId; // 接收通知的用户ID
    private Integer type; // 通知类型：1-互动通知，2-系统通知，3-私信通知，4-视频通知
    private String title;
    private String content;
    private String relatedId; // 相关资源ID，如视频ID、评论ID等
    private Integer isRead; // 0-未读，1-已读
    private Date createdAt;
    private Date readAt;
}
