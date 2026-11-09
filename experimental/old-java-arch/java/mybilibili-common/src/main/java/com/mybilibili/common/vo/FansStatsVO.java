package com.mybilibili.common.vo;

import lombok.Data;

/**
 * 粉丝统计数据VO
 */
@Data
public class FansStatsVO {
    /**
     * 总粉丝数
     */
    private Integer totalFans;

    /**
     * 互关粉丝数
     */
    private Integer mutualFans;

    /**
     * 新增粉丝数（近7天）
     */
    private Integer newFansWeek;

    /**
     * 新增粉丝数（近30天）
     */
    private Integer newFansMonth;
}
