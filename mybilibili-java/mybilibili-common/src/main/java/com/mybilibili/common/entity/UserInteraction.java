package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class UserInteraction {
    private Long id;
    private Integer userId;
    private String targetType;
    private Integer targetId;
    private String interactionType;
    private Date createdAt;
}
