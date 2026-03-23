package com.mybilibili.web.service;

import com.mybilibili.common.vo.CreatorStatsVO;
import com.mybilibili.common.vo.LatestCommentVO;
import com.mybilibili.common.vo.RankingVO;

import java.util.List;

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
