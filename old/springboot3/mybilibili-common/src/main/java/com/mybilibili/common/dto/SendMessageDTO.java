package com.mybilibili.common.dto;

import lombok.Data;

@Data
public class SendMessageDTO {
    private Integer receiverId;
    private String content;
    private Integer messageType;
    private Integer targetId;
    private String mediaUrl;
    private Integer commentId;
}
