package com.mybilibili.web.service;

import com.mybilibili.common.vo.CreatorStatsVO;
import com.mybilibili.common.vo.LatestCommentVO;
import com.mybilibili.common.vo.RankingVO;

import java.util.List;
import java.util.Map;

/**
 * 创作者统计服务接口
 */
public interface CreatorStatsService {

    /**
     * 获取创作者统计数据
     *
     * @param userId 用户ID
     * @return 统计数据
     */
    CreatorStatsVO getCreatorStats(Integer userId);

    /**
     * 获取数据概览
     *
     * @param userId 用户ID
     * @return 数据概览
     */
    Map<String, Object> getStatsOverview(Integer userId);

    /**
     * 获取趋势数据
     *
     * @param userId 用户ID
     * @param days   天数
     * @return 趋势数据
     */
    Map<String, Object> getStatsTrend(Integer userId, Integer days);

    /**
     * 获取稿件排行
     *
     * @param userId 用户ID
     * @param sortBy 排序方式
     * @param limit  限制数量
     * @return 排行列表
     */
    List<RankingVO> getManuscriptRanking(Integer userId, String sortBy, Integer limit);

    /**
     * 获取最新评论列表
     *
     * @param userId 用户ID
     * @param limit  限制数量
     * @return 最新评论列表
     */
    List<LatestCommentVO> getLatestComments(Integer userId, Integer limit);

    /**
     * 获取观看排行榜
     *
     * @param userId 用户ID
     * @param limit  限制数量
     * @return 观看排行榜
     */
    List<RankingVO> getViewRankings(Integer userId, Integer limit);

    /**
     * 获取互动排行榜
     *
     * @param userId 用户ID
     * @param limit  限制数量
     * @return 互动排行榜
     */
    List<RankingVO> getInteractionRankings(Integer userId, Integer limit);
}
