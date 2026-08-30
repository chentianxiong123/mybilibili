package com.mybilibili.ai.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("ai_skills")
public class AiSkill {
    @TableId(type = IdType.AUTO)
    private Long id;
    private String name;
    private String description;
    private String systemPrompt;
    private String fewShotExamples;
    private String type; // CUSTOMER_SERVICE / REVIEW / ADMIN
    private Boolean enabled;
    private Date createdAt;
    private Date updatedAt;
}