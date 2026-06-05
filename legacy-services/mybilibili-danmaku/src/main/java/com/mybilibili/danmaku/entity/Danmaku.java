package com.mybilibili.danmaku.entity;

import lombok.Data;
import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.mapping.Document;

import java.io.Serializable;
import java.time.LocalDateTime;

@Data
@Document(collection = "danmakus")
public class Danmaku implements Serializable {
    @Id
    private String id;
    private Integer videoId;
    private Integer manuscriptId;
    private Integer userId;
    private String content;
    private Double time;
    private String color;
    private Integer mode;
    private Integer status = 0;
    private LocalDateTime createTime;
}
