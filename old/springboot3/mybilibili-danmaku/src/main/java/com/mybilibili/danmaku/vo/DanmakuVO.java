package com.mybilibili.danmaku.vo;

import lombok.Data;

import java.time.LocalDateTime;

@Data
public class DanmakuVO {
    private String id;
    private Integer videoId;
    private Integer userId;
    private String content;
    private Double time;
    private String color;
    private Integer mode;
    private LocalDateTime createTime;
}
