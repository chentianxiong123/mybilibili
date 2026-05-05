package com.mybilibili.video.service;

import com.mybilibili.common.vo.CreatorOverviewVO;
import com.mybilibili.common.vo.FansRankingVO;
import com.mybilibili.common.vo.FansTrendVO;
import com.mybilibili.common.vo.LatestCommentVO;
import com.mybilibili.common.vo.ManuscriptRankVO;
import com.mybilibili.common.vo.TrendDataVO;

import java.util.List;

public interface CreatorStatsService {
    
    CreatorOverviewVO getCreatorOverview(Integer userId);
    
    TrendDataVO getPlayTrend(Integer userId, Integer days);
    
    List<ManuscriptRankVO> getManuscriptRanking(Integer userId, String sortBy, Integer limit);
    
    List<LatestCommentVO> getLatestComments(Integer userId, Integer limit);
    
    List<FansRankingVO> getFansRanking(Integer userId, String type, Integer limit);

    FansTrendVO getFansTrend(Integer userId, Integer days);
}
