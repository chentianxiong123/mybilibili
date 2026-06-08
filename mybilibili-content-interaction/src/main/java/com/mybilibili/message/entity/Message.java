package com.mybilibili.message.entity;

import com.baomidou.mybatisplus.annotation.*;
import lombok.Data;

import java.io.Serializable;
import java.time.LocalDateTime;

@Data
@TableName("messages")
public class Message implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Integer senderId;
    private Integer receiverId;
    private String content;
    private String messageType;
    private Integer targetId;
    private Integer commentId;
    private String mediaUrl;
    private Long conversationId;
    private Integer isRead;
    @TableField(fill = FieldFill.INSERT)
    private LocalDateTime createdAt;
    @TableField(fill = FieldFill.INSERT_UPDATE)
    private LocalDateTime updatedAt;
}