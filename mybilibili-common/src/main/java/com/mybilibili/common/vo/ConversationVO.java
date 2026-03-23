package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;

@Data
public class ConversationVO {
    private Integer id;
    private Integer userId;
    private Integer targetUserId;
    private String targetUserName;
    private String targetUserAvatar;
    private Integer lastMessageId;
    private String lastMessageContent;
    private Date lastMessageTime;
    private Integer unreadCount;
    private Date createdAt;
    private Date updatedAt;
}
