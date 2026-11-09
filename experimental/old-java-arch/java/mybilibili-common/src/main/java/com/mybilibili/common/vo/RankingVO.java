package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;

/**
 * 排行榜VO
 */
@Data
public class RankingVO {
    /**
     * 稿件ID
     */
    private Integer manuscriptId;

    /**
     * 稿件标题
     */
    private String title;

    /**
     * 稿件封面
     */
    private String coverUrl;

    /**
     * 播放量
     */
    private Integer viewCount;

    /**
     * 点赞数
     */
    private Integer likeCount;

    /**
     * 评论数
     */
    private Integer commentCount;

    /**
     * 投币数
     */
    private Integer coinCount;

    /**
     * 收藏数
     */
    private Integer collectCount;

    /**
     * 分享数
     */
    private Integer shareCount;

    /**
     * 互动率（点赞+评论+投币+收藏+分享）/播放量
     */
    private Double interactionRate;

    /**
     * 上传时间
     */
    private Date uploadTime;

    /**
     * 排名
     */
    private Integer rank;
}
