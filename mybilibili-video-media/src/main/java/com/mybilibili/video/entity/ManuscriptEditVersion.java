package com.mybilibili.video.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("manuscript_edit_versions")
public class ManuscriptEditVersion {
    public static final String STATUS_PENDING = "PENDING";
    public static final String STATUS_APPROVED = "APPROVED";
    public static final String STATUS_REJECTED = "REJECTED";

    private Long id;
    private Integer manuscriptId;
    private Integer userId;
    private String beforeSnapshot;
    private String afterSnapshot;
    private String changedFields;
    private String status;
    private Integer reviewerId;
    private String reviewReason;
    private Date reviewedAt;
    private Date createdAt;
    private Date updatedAt;
}
