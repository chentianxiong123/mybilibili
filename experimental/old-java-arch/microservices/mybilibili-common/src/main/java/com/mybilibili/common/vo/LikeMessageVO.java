package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;

@Data
public class LikeMessageVO {
    private Integer id;
    private Integer userId;
    private String username;
    private String userAvatar;
    private String content;
    private Integer videoId;
    private String videoTitle;
    private String videoCover;
    private Integer manuscriptId;
    private Integer videoOrder;
    private Integer commentId;
    private String commentContent;
    private Boolean isRead;
    private Date createdAt;
}
