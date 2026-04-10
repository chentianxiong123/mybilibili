package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;

@Data
public class DynamicCommentVO {
    private Integer id;
    private Integer dynamicId;
    private Integer userId;
    private String content;
    private Integer parentId;
    private Integer replyUserId;
    private Integer likeCount;
    private Date createdAt;
    private Integer status;

    private String userName;
    private String userAvatar;
    private Integer followersCount;
    private Boolean isFollowing;
    private Boolean isMutual;
}