package com.mybilibili.common.entity;

import lombok.Data;
import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.index.CompoundIndex;
import org.springframework.data.mongodb.core.index.Indexed;
import org.springframework.data.mongodb.core.mapping.Document;

import java.util.Date;

@Data
@Document(collection = "danmakus")
@CompoundIndex(name = "video_time_idx", def = "{'videoId': 1, 'time': 1}")
public class DanmakuDocument {

    @Id
    private String id;

    @Indexed
    private Integer videoId;

    private Integer manuscriptId;

    private Integer userId;

    private String content;

    private Double time;

    private String color;

    private Integer mode;

    private Date createTime;

    public DanmakuDocument() {
        this.createTime = new Date();
        this.color = "#ffffff";
        this.mode = 0;
    }
}
