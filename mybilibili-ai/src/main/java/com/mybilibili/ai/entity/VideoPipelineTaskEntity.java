package com.mybilibili.ai.entity;

import lombok.Data;

import java.time.LocalDateTime;

@Data
public class VideoPipelineTaskEntity {

    private Long id;
    private String taskKey;
    private Integer manuscriptId;
    private Integer videoId;
    private Integer uploaderId;
    private String videoTitle;
    private Integer currentStepIndex;
    private String status;
    private String workerId;
    private LocalDateTime lockedUntil;
    private String failedStep;
    private String errorMessage;
    private LocalDateTime createTime;
    private LocalDateTime startTime;
    private LocalDateTime endTime;
    private LocalDateTime updatedAt;
}
