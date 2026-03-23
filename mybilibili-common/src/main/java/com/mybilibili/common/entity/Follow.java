package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class Follow {
    private Integer id;
    private Integer followerId; // 关注者ID
    private Integer followedId; // 被关注者ID
    private Date createdAt;
}