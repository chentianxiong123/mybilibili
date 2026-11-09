package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class Message {
    private Integer id;
    private Integer senderId;
    private Integer receiverId;
    private String content;
    private Integer messageType;
    private Integer targetId;  // 目标ID（视频ID、评论ID等）
    private String mediaUrl;
    private Integer conversationId;
    private Boolean isRead;
    private Date createdAt;
    private Date updatedAt;
}
