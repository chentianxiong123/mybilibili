package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class VideoTag {
    private Integer id;
    private Integer videoId;
    private Integer tagId;
    private Date createdAt;
}