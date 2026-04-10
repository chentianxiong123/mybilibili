package com.mybilibili.common.vo;

import lombok.Data;

@Data
public class FansRankingVO {
    private Integer id;
    private String username;
    private String avatar;
    private Long totalDuration;
    private Integer interactionCount;
}
