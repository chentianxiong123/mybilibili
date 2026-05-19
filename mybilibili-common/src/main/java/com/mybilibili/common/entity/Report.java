package com.mybilibili.common.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("reports")
public class Report {
    private Integer id;
    private Integer reporterId;
    private String targetType;
    private Integer targetId;
    private Integer manuscriptId;
    private String reason;
    private String description;
    private String status;
    private String adminRemark;
    private Date createdAt;
    private Date processedAt;
    private String aiReviewStatus;
    private String aiVerdict;
    private String aiRiskLevel;
    private Date aiReviewedAt;
}
