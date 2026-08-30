package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class UserDynamic {
    private Integer id;
    private Integer userId;
    private String content;
    private Integer dynamicType;
    private String imageUrl;
    private Integer refVideoId;
    private Integer refManuscriptId;
    private Integer likeCount;
    private Integer commentCount;
    private Integer shareCount;
    private Date createdAt;
    private Integer status;
}
