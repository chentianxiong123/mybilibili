package com.mybilibili.common.entity;

import lombok.Data;

@Data
public class NotificationSetting {
    private Integer id;
    private Integer userId;
    private Integer likeNotification; // 点赞通知：1-开启，0-关闭
    private Integer commentNotification; // 评论通知：1-开启，0-关闭
    private Integer followNotification; // 关注通知：1-开启，0-关闭
    private Integer systemNotification; // 系统通知：1-开启，0-关闭
    private Integer videoNotification; // 视频通知：1-开启，0-关闭
}
