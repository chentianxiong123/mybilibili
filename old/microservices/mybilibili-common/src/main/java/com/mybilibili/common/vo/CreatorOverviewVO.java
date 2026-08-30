package com.mybilibili.common.vo;

import lombok.Data;

@Data
public class CreatorOverviewVO {
    private Integer totalViews;
    private Integer totalLikes;
    private Integer totalCoins;
    private Integer totalCollections;
    private Integer totalShares;
    private Integer totalComments;
    private Integer totalDanmaku;
    private Integer totalFollowers;
    private Integer totalManuscripts;
    private Integer viewsIncrease;
    private Integer likesIncrease;
    private Integer followersIncrease;
    private Integer commentsIncrease;
    private Integer danmakuIncrease;
    private Integer sharesIncrease;
    private Integer collectionsIncrease;
    private Integer coinsIncrease;
    private Integer visitorCount;
    private Integer visitorIncrease;
    private String updateTime;
}
