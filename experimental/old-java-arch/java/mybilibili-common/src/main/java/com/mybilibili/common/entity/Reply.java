package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class Reply {
    private Integer id;
    private Integer commentId;
    private Integer userId;
    private Integer replyToUserId;
    private String content;
    private Integer likeCount;
    private String status;  // NORMAL-正常 REMOVED-已下架
    private Date createdAt;
    private Date updatedAt;
}