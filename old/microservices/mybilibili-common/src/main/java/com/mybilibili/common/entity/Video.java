package com.mybilibili.common.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("videos")
public class Video {
    private Integer id;
    private Integer manuscriptId;
    private Integer videoOrder;
    private String title;
    private String description;
    private String playUrlHd;
    private String playUrlSd;
    private String playUrlLd;
    private Integer status;
    private Integer reviewStatus;
    private Date uploadTime;
    private Date updatedAt;
    private Integer processProgress;
    private String processStage;
    private Integer hasSubtitle;
    private Integer hasSummary;
    private Integer processStatus;
    private String processError;
    private String sourceVideoUrl;
    private Integer durationSeconds;

    public static final int STATUS_PENDING_REVIEW = 0;
    public static final int STATUS_PROCESSING = 1;
    public static final int STATUS_PUBLISHED = 3;
    public static final int STATUS_REJECTED = 4;
    public static final int STATUS_PROCESS_FAILED = 5;
    public static final int STATUS_UNPUBLISHED = -1;

    public static final int REVIEW_STATUS_PENDING = 0;
    public static final int REVIEW_STATUS_APPROVED = 1;
    public static final int REVIEW_STATUS_REJECTED = 2;

    public static final int PROCESS_STATUS_PENDING = 0;
    public static final int PROCESS_STATUS_TRANSCODING = 1;
    public static final int PROCESS_STATUS_TRANSCODE_FAILED = 10;
    public static final int PROCESS_STATUS_TRANSCODE_SUCCESS = 11;
    public static final int PROCESS_STATUS_AUDIO_EXTRACTING = 2;
    public static final int PROCESS_STATUS_AUDIO_FAILED = 20;
    public static final int PROCESS_STATUS_AUDIO_SUCCESS = 21;
    public static final int PROCESS_STATUS_SUBTITLE_GENERATING = 3;
    public static final int PROCESS_STATUS_SUBTITLE_FAILED = 30;
    public static final int PROCESS_STATUS_SUBTITLE_SUCCESS = 31;
    public static final int PROCESS_STATUS_AI_SUMMARIZING = 4;
    public static final int PROCESS_STATUS_AI_FAILED = 40;
    public static final int PROCESS_STATUS_AI_SUCCESS = 41;
    public static final int PROCESS_STATUS_COMPLETED = 5;
}
