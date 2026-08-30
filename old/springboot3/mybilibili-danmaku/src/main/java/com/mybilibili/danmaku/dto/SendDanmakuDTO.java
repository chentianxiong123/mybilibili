package com.mybilibili.danmaku.dto;

import lombok.Data;

@Data
public class SendDanmakuDTO {
    private Integer videoId;
    private String content;
    private String time;
    private String color;
    private Integer mode;
}
