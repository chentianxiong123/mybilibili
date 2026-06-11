package com.mybilibili.common.entity;

import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;
import java.util.List;

@Data
@TableName("manuscripts")
public class Manuscript {
    private Integer id;
    private String title;
    private String description;
    private String coverUrl;
    private Integer userId;
    private Integer categoryId;

    private Integer viewCount;
    private Integer likeCount;
    private Integer coinCount;
    private Integer collectCount;
    private Integer shareCount;
    private Integer commentCount;
    private Integer danmakuCount;

    private String duration;
    private Integer durationSeconds;

    private Integer status;
    private Integer reviewStatus;
    private String reviewReason;
    private Date reviewTime;
    private Integer reviewerId;

    private Date uploadTime;
    private Date updatedAt;

    @TableField(exist = false)
    private List<Video> videos;
    @TableField(exist = false)
    private User user;
    @TableField(exist = false)
    private Category category;
    @TableField(exist = false)
    private List<String> tags;

    public static final int STATUS_PENDING_REVIEW = 0;
    public static final int STATUS_PROCESSING = 1;
    public static final int STATUS_PUBLISHED = 3;
    public static final int STATUS_REJECTED = 4;
    public static final int STATUS_PROCESS_FAILED = 5;
    public static final int STATUS_UNPUBLISHED = -1;

    public static final int REVIEW_STATUS_PENDING = 0;
    public static final int REVIEW_STATUS_APPROVED = 1;
    public static final int REVIEW_STATUS_REJECTED = 2;
}