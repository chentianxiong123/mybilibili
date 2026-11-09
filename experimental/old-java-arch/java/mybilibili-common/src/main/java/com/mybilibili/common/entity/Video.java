package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class Video {
    private Integer id;
    private Integer manuscriptId;  // 所属稿件ID
    private Integer videoOrder;    // 在稿件中的排序（分P顺序）
    
    // 保留字段（兼容旧数据，实际使用manuscripts表的字段）
    private String title;
    private String description;
    private String coverUrl;
    private Integer userId;
    private Integer categoryId;
    
    // 视频特有字段
    private String playUrl;
    private String playUrlHd;
    private String playUrlSd;
    private String playUrlLd;
    
    // 统计字段（保留兼容）
    private Integer viewCount;
    private Integer likeCount;
    private Integer coinCount;
    private Integer collectCount;
    private Integer shareCount;
    private Integer commentCount;
    private Integer danmakuCount;
    
    private Integer durationSeconds;      // 时长秒数 (用于计算和显示)
    private Integer status;
    private Date uploadTime;
    private Date updatedAt;
    
    // 审核相关字段（保留兼容）
    private Integer reviewStatus;
    private String reviewReason;
    private Date reviewTime;
    private Integer reviewerId;
    
    // 处理进度相关字段
    private Integer processProgress;
    private String processStage;
    private Integer hasSubtitle;
    private Integer hasSummary;
    
    // 新增处理状态字段
    private Integer processStatus;      // 处理状态：0待处理 1转码中 2音频提取中 3字幕生成中 4AI总结中 5完成 6-9失败
    private String processError;        // 处理失败原因
    private String sourceVideoUrl;      // 源视频URL（审核预览用）
    
    // 非数据库字段 - 关联稿件信息（从manuscripts表查询）
    private Manuscript manuscript; // 所属稿件
    private String manuscriptTitle;
    private String manuscriptDescription;
    private String manuscriptCoverUrl;
    private Date manuscriptUploadTime;
    private Integer manuscriptViewCount;
    
    // 状态常量
    public static final int STATUS_PENDING_REVIEW = 0;      // 待审核
    public static final int STATUS_PROCESSING = 1;          // 审核通过-处理中
    public static final int STATUS_READY_TO_PUBLISH = 2;    // 处理完成-待上架
    public static final int STATUS_PUBLISHED = 3;           // 已上架
    public static final int STATUS_REJECTED = 4;            // 审核拒绝
    public static final int STATUS_PROCESS_FAILED = 5;      // 处理失败
    public static final int STATUS_UNPUBLISHED = -1;        // 已下架
    
    // 审核状态常量
    public static final int REVIEW_STATUS_PENDING = 0;      // 待审核
    public static final int REVIEW_STATUS_APPROVED = 1;     // 审核通过
    public static final int REVIEW_STATUS_REJECTED = 2;     // 审核拒绝
    
    // 处理状态常量（新规则）
    // 0x: 待处理/初始状态
    public static final int PROCESS_STATUS_PENDING = 0;             // 0-待处理
    
    // 1x: 视频转码
    public static final int PROCESS_STATUS_TRANSCODING = 1;         // 1-转码处理中
    public static final int PROCESS_STATUS_TRANSCODE_FAILED = 10;   // 10-转码失败
    public static final int PROCESS_STATUS_TRANSCODE_SUCCESS = 11;  // 11-转码成功
    
    // 2x: 音频提取
    public static final int PROCESS_STATUS_AUDIO_EXTRACTING = 2;    // 2-音频提取中
    public static final int PROCESS_STATUS_AUDIO_FAILED = 20;       // 20-音频提取失败
    public static final int PROCESS_STATUS_AUDIO_SUCCESS = 21;      // 21-音频提取成功
    
    // 3x: 字幕生成
    public static final int PROCESS_STATUS_SUBTITLE_GENERATING = 3; // 3-字幕生成中
    public static final int PROCESS_STATUS_SUBTITLE_FAILED = 30;    // 30-字幕生成失败
    public static final int PROCESS_STATUS_SUBTITLE_SUCCESS = 31;   // 31-字幕生成成功
    
    // 4x: AI总结
    public static final int PROCESS_STATUS_AI_SUMMARIZING = 4;      // 4-AI总结中
    public static final int PROCESS_STATUS_AI_FAILED = 40;          // 40-AI总结失败
    public static final int PROCESS_STATUS_AI_SUCCESS = 41;         // 41-AI总结成功
    
    // 5: 全部完成
    public static final int PROCESS_STATUS_COMPLETED = 5;           // 5-全部处理完成
}