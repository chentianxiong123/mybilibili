package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;

@Data
public class AtMessageVO {
    private Integer id;
    private Integer userId;
    private String username;
    private String userAvatar;
    private String content;
    private Integer videoId;
    private String videoTitle;
    private String videoCover;
    private Integer commentId;
    private Boolean isRead;
    private Date createdAt;
}
