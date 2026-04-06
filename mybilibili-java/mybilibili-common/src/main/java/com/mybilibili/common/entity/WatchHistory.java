package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class WatchHistory {
    private Integer id;
    private Integer userId;
    private Integer videoId;
    private Integer progressSeconds;
    private Date watchedAt;
}
