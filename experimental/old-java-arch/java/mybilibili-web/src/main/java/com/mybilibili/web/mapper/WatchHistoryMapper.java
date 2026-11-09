package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.WatchHistory;
import com.mybilibili.web.service.impl.RandomRecommendServiceImpl;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface WatchHistoryMapper {

    /**
     * 插入浏览历史记录
     */
    int insert(WatchHistory watchHistory);

    /**
     * 根据用户ID查询浏览历史列表（分页）
     */
    List<WatchHistory> selectByUserId(@Param("userId") Integer userId, @Param("offset") int offset, @Param("size") int size);

    /**
     * 根据用户ID删除所有浏览历史
     */
    int deleteByUserId(Integer userId);

    /**
     * 根据ID删除单条浏览历史
     */
    int deleteById(Integer id);

    /**
     * 更新观看进度
     */
    int updateProgress(@Param("userId") Integer userId, @Param("videoId") Integer videoId, @Param("progressSeconds") Integer progressSeconds);

    /**
     * 根据用户ID和视频ID查询浏览历史
     */
    WatchHistory selectByUserIdAndVideoId(@Param("userId") Integer userId, @Param("videoId") Integer videoId);

    /**
     * 更新浏览历史记录（用于更新观看时间）
     */
    int update(WatchHistory watchHistory);

    /**
     * 查询用户最近浏览历史（带稿件详情）
     * 用于推荐算法
     *
     * @param userId       用户ID
     * @param days         最近几天
     * @param limit        限制数量
     * @return 稿件行为信息列表
     */
    List<RandomRecommendServiceImpl.ManuscriptBehaviorInfo> selectRecentWatchHistoryWithDetails(
            @Param("userId") Integer userId,
            @Param("days") int days,
            @Param("limit") int limit);
}
