package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;

@Data
public class MessageVO {
    private Long id;
    private Integer senderId;
    private String senderName;
    private String senderAvatar;
    private Integer receiverId;
    private String content;
    private Integer messageType;
    private Integer targetId;
    private String mediaUrl;
    private Boolean isRead;
    private Date createdAt;
}
