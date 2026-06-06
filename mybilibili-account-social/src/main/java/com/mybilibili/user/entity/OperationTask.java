package com.mybilibili.user.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@TableName("operation_tasks")
public class OperationTask {
    @TableId(type = IdType.AUTO)
    private Long id;
    private String taskKey;
    private String taskType;
    private String taskName;
    private String targetType;
    private String targetId;
    private String status;
    private Integer progress;
    private String stage;
    private String message;
    private String errorMessage;
    private Integer operatorId;
    private String operatorName;
    private LocalDateTime startedAt;
    private LocalDateTime finishedAt;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
}
