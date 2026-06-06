package com.mybilibili.message.entity;

import com.baomidou.mybatisplus.annotation.*;
import lombok.Data;

import java.io.Serializable;
import java.time.LocalDateTime;

@Data
@TableName("notifications")
public class Notification implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Integer userId;
    private String notificationType;
    private String title;
    private String content;
    private Integer isRead;
    private Integer status;
    @TableField(fill = FieldFill.INSERT)
    private LocalDateTime createdAt;
}
