package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class Conversation {
    private Integer id;
    private Integer userId;
    private Integer targetUserId;
    private Integer lastMessageId;
    private String lastMessageContent;
    private Date lastMessageTime;
    private Integer unreadCount;
    private Date createdAt;
    private Date updatedAt;
}
