package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;

@Data
public class WatchHistoryVO {
    private Integer id;
    private Integer videoId;
    private Integer manuscriptId;
    private Integer progressSeconds;
    private Integer videoDuration;
    private Integer watchPercentage;
    private Date watchedAt;
    private VideoVO video;
}
