package com.mybilibili.web.service.impl;

import com.mybilibili.common.vo.CreatorStatsVO;
import com.mybilibili.common.vo.LatestCommentVO;
import com.mybilibili.common.vo.RankingVO;
import com.mybilibili.web.mapper.CreatorStatsMapper;
import com.mybilibili.web.service.CreatorStatsService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * 创作者统计服务实现
 */
@Slf4j
@Service
public class CreatorStatsServiceImpl implements CreatorStatsService {

    @Autowired
    private CreatorStatsMapper creatorStatsMapper;

    @Override
    public CreatorStatsVO getCreatorStats(Integer userId) {
        CreatorStatsVO stats = new CreatorStatsVO();

        // 获取粉丝总数
        Integer totalFollowers = creatorStatsMapper.countFollowers(userId);
        stats.setTotalFollowers(totalFollowers != null ? totalFollowers : 0);

        // 获取总播放量
        Integer totalViews = creatorStatsMapper.sumViewCount(userId);
        stats.setTotalViews(totalViews != null ? totalViews : 0);

        // 获取总评论数
        Integer totalComments = creatorStatsMapper.sumCommentCount(userId);
        stats.setTotalComments(totalComments != null ? totalComments : 0);

        // 获取总点赞数
        Integer totalLikes = creatorStatsMapper.sumLikeCount(userId);
        stats.setTotalLikes(totalLikes != null ? totalLikes : 0);

        // 获取总分享数
        Integer totalShares = creatorStatsMapper.sumShareCount(userId);
        stats.setTotalShares(totalShares != null ? totalShares : 0);

        // 获取总收藏数
        Integer totalCollections = creatorStatsMapper.sumCollectCount(userId);
        stats.setTotalCollections(totalCollections != null ? totalCollections : 0);

        // 获取总投币数
        Integer totalCoins = creatorStatsMapper.sumCoinCount(userId);
        stats.setTotalCoins(totalCoins != null ? totalCoins : 0);

        // 获取稿件总数
        Integer totalManuscripts = creatorStatsMapper.countManuscripts(userId);
        stats.setTotalManuscripts(totalManuscripts != null ? totalManuscripts : 0);

        return stats;
    }

    @Override
    public List<LatestCommentVO> getLatestComments(Integer userId, Integer limit) {
        if (limit == null || limit <= 0) {
            limit = 10;
        }
        List<LatestCommentVO> comments = creatorStatsMapper.selectLatestComments(userId, limit);
        return comments;
    }

    @Override
    public List<RankingVO> getViewRankings(Integer userId, Integer limit) {
        if (limit == null || limit <= 0) {
            limit = 10;
        }
        List<RankingVO> rankings = creatorStatsMapper.selectViewRankings(userId, limit);

        // 设置排名
        for (int i = 0; i < rankings.size(); i++) {
            rankings.get(i).setRank(i + 1);
        }

        return rankings;
    }

    @Override
    public List<RankingVO> getInteractionRankings(Integer userId, Integer limit) {
        if (limit == null || limit <= 0) {
            limit = 10;
        }
        List<RankingVO> rankings = creatorStatsMapper.selectInteractionRankings(userId, limit);

        // 设置排名
        for (int i = 0; i < rankings.size(); i++) {
            rankings.get(i).setRank(i + 1);
        }

        return rankings;
    }
}
