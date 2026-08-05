package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;
import java.util.List;
import java.util.Map;

@Data
public class ManuscriptEditVersionVO {
    private Long id;
    private Integer manuscriptId;
    private Integer userId;
    private String status;
    private String reviewReason;
    private Integer reviewerId;
    private Date reviewedAt;
    private Date createdAt;
    private List<String> changedFields;
    private Map<String, Object> beforeSnapshot;
    private Map<String, Object> afterSnapshot;
}
