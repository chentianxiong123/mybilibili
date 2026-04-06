package com.mybilibili.web.service;

import com.mybilibili.common.vo.FanVO;
import com.mybilibili.common.vo.FansStatsVO;

import java.util.List;

/**
 * 创作者粉丝管理服务接口
 */
public interface CreatorFansService {

    /**
     * 获取粉丝列表（分页）
     *
     * @param userId   用户ID
     * @param page     页码
     * @param size     每页大小
     * @param mutual   是否只显示互关粉丝
     * @return 粉丝列表
     */
    List<FanVO> getFansList(Integer userId, Integer page, Integer size, Boolean mutual);

    /**
     * 获取粉丝统计数据
     *
     * @param userId 用户ID
     * @return 粉丝统计数据
     */
    FansStatsVO getFansStats(Integer userId);

    /**
     * 获取粉丝数量
     *
     * @param userId 用户ID
     * @param mutual 是否只统计互关粉丝
     * @return 粉丝数量
     */
    int getFansCount(Integer userId, Boolean mutual);
}
