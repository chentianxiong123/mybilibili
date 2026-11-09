package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class DynamicComment {
    private Integer id;
    private Integer dynamicId;
    private Integer userId;
    private String content;
    private Integer parentId;
    private Integer replyUserId;
    private Integer likeCount;
    private Date createdAt;
    private Integer status;
}
