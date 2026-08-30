package com.mybilibili.common.vo;

import lombok.Data;

/**
 * 创作者统计数据VO
 */
@Data
public class CreatorStatsVO {
    /**
     * 粉丝总数
     */
    private Integer totalFollowers;

    /**
     * 总播放量
     */
    private Integer totalViews;

    /**
     * 总评论数
     */
    private Integer totalComments;

    /**
     * 总点赞数
     */
    private Integer totalLikes;

    /**
     * 总分享数
     */
    private Integer totalShares;

    /**
     * 总收藏数
     */
    private Integer totalCollections;

    /**
     * 总投币数
     */
    private Integer totalCoins;

    /**
     * 稿件总数
     */
    private Integer totalManuscripts;
}
