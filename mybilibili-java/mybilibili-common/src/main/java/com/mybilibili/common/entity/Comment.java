package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class Comment {
    private Integer id;
    private Integer manuscriptId;
    private Integer userId;
    private String content;
    private Integer likeCount;
    private Integer replyCount;
    private Integer status;
    private Date createdAt;
    private Date updatedAt;
}
