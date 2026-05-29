package com.mybilibili.ai.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("ai_usage_logs")
public class AiUsageLog {

    @TableId(type = IdType.AUTO)
    private Long id;

    /** 功能类型: CHAT / REVIEW / SUMMARY / TEST */
    private String feature;

    private String model;

    private Integer inputTokens;

    private Integer outputTokens;

    private Integer totalTokens;

    /** 耗时(ms) */
    private Long durationMs;

    /** SUCCESS / FAILED */
    private String status;

    private String errorMessage;

    private Date createdAt;
}
