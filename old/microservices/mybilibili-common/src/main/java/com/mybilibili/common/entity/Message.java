package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class Message {
    private Long id;
    private Integer senderId;
    private Integer receiverId;
    private String content;
    private Integer messageType;
    private Integer targetId;
    private String mediaUrl;
    private Long conversationId;
    private Boolean isRead;
    private Integer commentId;
    private Date createdAt;
    private Date updatedAt;
}
