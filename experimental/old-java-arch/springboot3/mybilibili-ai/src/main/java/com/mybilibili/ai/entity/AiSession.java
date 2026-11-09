package com.mybilibili.ai.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("ai_sessions")
public class AiSession {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Long userId;
    private Long skillId;
    private String type;
    private Integer status;
    private Date createdAt;
    private Date updatedAt;
}