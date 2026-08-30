package com.mybilibili.ai.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("ai_feedback")
public class AiFeedback {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Long sessionId;
    private Long userId;
    private Long skillId;
    private String question;
    private String answer;
    private Integer rating; // 1-5
    private String feedbackText;
    private String adminReply;
    private String status; // PENDING / PROCESSED
    private Date createdAt;
    private Date updatedAt;
}