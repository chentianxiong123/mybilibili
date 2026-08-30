package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;
import java.util.List;

@Data
public class Manuscript {
    private Integer id;
    private String title;
    private String description;
    private String coverUrl;
    private Integer userId;
    private Integer categoryId;
    
    // 统计字段
    private Integer viewCount;
    private Integer likeCount;
    private Integer coinCount;
    private Integer collectCount;
    private Integer shareCount;
    private Integer commentCount;
    private Integer danmakuCount;
    
    // 时长字段
    private String duration;              // 时长显示字符串 (如 "45:30")
    private Integer durationSeconds;      // 总时长秒数 (所有视频之和)
    
    // 稿件状态
    private Integer status;
    private Integer reviewStatus;
    private String reviewReason;
    private Date reviewTime;
    private Integer reviewerId;
    
    // 处理进度（已弃用，保留兼容）
    private Integer processProgress;
    private String processStage;
    
    // 注意：currentVideoId 和 processQueue 存储在Redis中
    // Redis Key: manuscript:process:{manuscriptId}:current
    // Redis Key: manuscript:process:{manuscriptId}:queue
    
    // 时间戳
    private Date uploadTime;
    private Date updatedAt;
    
    // 非数据库字段
    private List<Video> videos; // 稿件包含的视频列表
    private User user; // 上传者信息
    private Category category; // 分类信息
    private List<String> tags; // 标签列表（用于推荐算法）
    
    // 状态常量
    public static final int STATUS_PENDING_REVIEW = 0;      // 待审核
    public static final int STATUS_PROCESSING = 1;          // 处理中
    public static final int STATUS_READY_TO_PUBLISH = 2;    // 待上架
    public static final int STATUS_PUBLISHED = 3;           // 已上架
    public static final int STATUS_REJECTED = 4;            // 审核拒绝
    public static final int STATUS_PROCESS_FAILED = 5;      // 处理失败
    public static final int STATUS_UNPUBLISHED = -1;        // 已下架
    
    // 审核状态常量
    public static final int REVIEW_STATUS_PENDING = 0;      // 待审核
    public static final int REVIEW_STATUS_APPROVED = 1;     // 审核通过
    public static final int REVIEW_STATUS_REJECTED = 2;     // 审核拒绝
}
