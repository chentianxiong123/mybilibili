package com.mybilibili.web.service;

import com.mybilibili.common.entity.WatchHistory;
import com.mybilibili.common.vo.WatchHistoryVO;

import java.util.List;

public interface WatchHistoryService {

    /**
     * 记录浏览历史
     * 当进度超过视频时长的10%时记录
     *
     * @param userId          用户ID
     * @param videoId         视频ID
     * @param progressSeconds 观看进度（秒）
     * @param videoDuration   视频总时长（秒）
     */
    void recordWatchHistory(Integer userId, Integer videoId, Integer progressSeconds, Integer videoDuration);

    /**
     * 获取浏览历史列表
     *
     * @param userId 用户ID
     * @param page   页码
     * @param size   每页大小
     * @return 浏览历史VO列表
     */
    List<WatchHistoryVO> getWatchHistoryList(Integer userId, Integer page, Integer size);

    /**
     * 清空浏览历史
     *
     * @param userId 用户ID
     */
    void clearWatchHistory(Integer userId);

    /**
     * 删除单条浏览历史
     *
     * @param id     记录ID
     * @param userId 用户ID
     */
    void deleteWatchHistory(Integer id, Integer userId);

    /**
     * 更新观看进度
     *
     * @param userId          用户ID
     * @param videoId         视频ID
     * @param progressSeconds 观看进度（秒）
     */
    void updateWatchProgress(Integer userId, Integer videoId, Integer progressSeconds);

    /**
     * 获取用户最近浏览的视频ID列表（从Redis）
     *
     * @param userId 用户ID
     * @param days   最近天数
     * @param limit  返回数量限制
     * @return 视频ID列表（按时间倒序）
     */
    List<Integer> getRecentWatchVideoIds(Integer userId, int days, int limit);
}
