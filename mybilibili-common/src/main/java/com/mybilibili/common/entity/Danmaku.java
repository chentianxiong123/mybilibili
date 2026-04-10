package com.mybilibili.common.entity;

import lombok.Data;
import java.util.Date;

@Data
public class Danmaku {
    private Integer id;
    private Integer videoId;       // 弹幕仍然关联视频（分P视频）
    private Integer manuscriptId;  // 新增：关联稿件
    private Integer userId;
    private String content;
    private String time;
    private String color;
    private Integer mode;
    private Date createdAt;
}
