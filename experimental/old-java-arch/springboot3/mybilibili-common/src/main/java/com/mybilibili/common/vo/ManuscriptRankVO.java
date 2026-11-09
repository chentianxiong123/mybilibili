package com.mybilibili.common.vo;

import lombok.Data;

@Data
public class ManuscriptRankVO {
    private Integer id;
    private String title;
    private String coverUrl;
    private Integer viewCount;
    private Integer likeCount;
    private Integer commentCount;
    private Integer danmakuCount;
    private Integer coinCount;
    private Integer collectCount;
    private Integer shareCount;
    private String uploadTime;
    private Double interactionRate;
}
