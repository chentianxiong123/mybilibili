package com.mybilibili.video.service.impl;

import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.vo.CreatorOverviewVO;
import com.mybilibili.common.vo.FansRankingVO;
import com.mybilibili.common.vo.FansTrendVO;
import com.mybilibili.common.vo.LatestCommentVO;
import com.mybilibili.common.vo.ManuscriptRankVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.TrendDataVO;
import com.mybilibili.video.mapper.CreatorStatsMapper;
import com.mybilibili.video.mapper.ManuscriptMapper;
import com.mybilibili.video.service.CreatorStatsService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.time.LocalDate;
import java.time.format.DateTimeFormatter;
import java.time.LocalDateTime;
import java.time.Duration;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeSet;

@Service
public class CreatorStatsServiceImpl implements CreatorStatsService {

    @Autowired
    private ManuscriptMapper manuscriptMapper;
    
    @Autowired
    private CreatorStatsMapper creatorStatsMapper;

    @Override
    public CreatorOverviewVO getCreatorOverview(Integer userId) {
        CreatorOverviewVO overview = new CreatorOverviewVO();
        
        Map<String, Object> totalStats = creatorStatsMapper.selectTotalStatsByUserId(userId);
        if (totalStats != null) {
            overview.setTotalViews(getIntValue(totalStats, "totalViews"));
            overview.setTotalLikes(getIntValue(totalStats, "totalLikes"));
            overview.setTotalCoins(getIntValue(totalStats, "totalCoins"));
            overview.setTotalCollections(getIntValue(totalStats, "totalCollections"));
            overview.setTotalShares(getIntValue(totalStats, "totalShares"));
            overview.setTotalComments(getIntValue(totalStats, "totalComments"));
            overview.setTotalDanmaku(getIntValue(totalStats, "totalDanmaku"));
            overview.setTotalManuscripts(getIntValue(totalStats, "totalManuscripts"));
        }
        
        Integer followerCount = creatorStatsMapper.selectFollowerCount(userId);
        overview.setTotalFollowers(followerCount != null ? followerCount : 0);
        
        LocalDate sevenDaysAgo = LocalDate.now().minusDays(7);
        Map<String, Object> increaseStats = creatorStatsMapper.selectIncreaseStats(userId, sevenDaysAgo.toString());
        if (increaseStats != null) {
            overview.setViewsIncrease(getIntValue(increaseStats, "viewsIncrease"));
            overview.setLikesIncrease(getIntValue(increaseStats, "likesIncrease"));
            overview.setCommentsIncrease(getIntValue(increaseStats, "commentsIncrease"));
            overview.setDanmakuIncrease(getIntValue(increaseStats, "danmakuIncrease"));
            overview.setSharesIncrease(getIntValue(increaseStats, "sharesIncrease"));
            overview.setCollectionsIncrease(getIntValue(increaseStats, "collectionsIncrease"));
            overview.setCoinsIncrease(getIntValue(increaseStats, "coinsIncrease"));
        }
        
        List<Map<String, Object>> fansTrend = creatorStatsMapper.selectFansTrendData(userId, 7);
        int followersIncrease = 0;
        for (Map<String, Object> row : fansTrend) {
            followersIncrease += getIntValue(row, "newFollowers");
        }
        overview.setFollowersIncrease(followersIncrease);
        
        overview.setUpdateTime(LocalDate.now().format(DateTimeFormatter.ofPattern("yyyy-MM-dd")));
        
        return overview;
    }

    @Override
    public TrendDataVO getPlayTrend(Integer userId, Integer days) {
        TrendDataVO trend = new TrendDataVO();
        List<String> dates = new ArrayList<>();
        List<Integer> views = new ArrayList<>();
        List<Integer> likes = new ArrayList<>();
        List<Integer> comments = new ArrayList<>();
        List<Integer> followers = new ArrayList<>();
        
        List<Map<String, Object>> manuscriptTrend = creatorStatsMapper.selectTrendData(userId, days);
        List<Map<String, Object>> fansTrend = creatorStatsMapper.selectFansTrendData(userId, days);
        
        Map<String, Map<String, Integer>> manuscriptMap = new HashMap<>();
        for (Map<String, Object> row : manuscriptTrend) {
            Object dateObj = row.get("date");
            String date = dateObj instanceof java.sql.Date ? ((java.sql.Date) dateObj).toString() : String.valueOf(dateObj);
            Map<String, Integer> dayData = new HashMap<>();
            dayData.put("views", getIntValue(row, "views"));
            dayData.put("likes", getIntValue(row, "likes"));
            dayData.put("comments", getIntValue(row, "comments"));
            manuscriptMap.put(date, dayData);
        }
        
        Map<String, Integer> fansMap = new HashMap<>();
        for (Map<String, Object> row : fansTrend) {
            Object dateObj = row.get("date");
            String date = dateObj instanceof java.sql.Date ? ((java.sql.Date) dateObj).toString() : String.valueOf(dateObj);
            fansMap.put(date, getIntValue(row, "newFollowers"));
        }
        
        Set<String> allDates = new TreeSet<>();
        allDates.addAll(manuscriptMap.keySet());
        allDates.addAll(fansMap.keySet());
        
        for (String date : allDates) {
            dates.add(date);
            
            Map<String, Integer> dayData = manuscriptMap.get(date);
            if (dayData != null) {
                views.add(dayData.get("views"));
                likes.add(dayData.get("likes"));
                comments.add(dayData.get("comments"));
            } else {
                views.add(0);
                likes.add(0);
                comments.add(0);
            }
            
            Integer dayFollowers = fansMap.get(date);
            followers.add(dayFollowers != null ? dayFollowers : 0);
        }
        
        trend.setDates(dates);
        trend.setViews(views);
        trend.setLikes(likes);
        trend.setComments(comments);
        trend.setFollowers(followers);
        
        return trend;
    }

    @Override
    public List<ManuscriptRankVO> getManuscriptRanking(Integer userId, String sortBy, Integer limit) {
        List<Manuscript> manuscripts;
        
        switch (sortBy) {
            case "views":
                manuscripts = creatorStatsMapper.selectTopByViews(userId, limit);
                break;
            case "likes":
                manuscripts = creatorStatsMapper.selectTopByLikes(userId, limit);
                break;
            case "comments":
                manuscripts = creatorStatsMapper.selectTopByComments(userId, limit);
                break;
            default:
                manuscripts = creatorStatsMapper.selectTopByViews(userId, limit);
        }
        
        List<ManuscriptRankVO> rankList = new ArrayList<>();
        for (Manuscript m : manuscripts) {
            ManuscriptRankVO vo = new ManuscriptRankVO();
            vo.setId(m.getId());
            vo.setTitle(m.getTitle());
            vo.setCoverUrl(m.getCoverUrl());
            vo.setViewCount(m.getViewCount());
            vo.setLikeCount(m.getLikeCount());
            vo.setCommentCount(m.getCommentCount());
            vo.setDanmakuCount(m.getDanmakuCount());
            vo.setCoinCount(m.getCoinCount());
            vo.setCollectCount(m.getCollectCount());
            vo.setShareCount(m.getShareCount());
            vo.setUploadTime(m.getUploadTime() != null ? m.getUploadTime().toString() : null);
            
            int totalInteractions = m.getLikeCount() + m.getCommentCount() + m.getDanmakuCount() 
                + m.getCoinCount() + m.getCollectCount() + m.getShareCount();
            double rate = m.getViewCount() > 0 ? (double) totalInteractions / m.getViewCount() * 100 : 0;
            vo.setInteractionRate(Math.round(rate * 100.0) / 100.0);
            
            rankList.add(vo);
        }
        
        return rankList;
    }
    
    @Override
    public List<LatestCommentVO> getLatestComments(Integer userId, Integer limit) {
        List<Map<String, Object>> comments = creatorStatsMapper.selectLatestComments(userId, limit);
        List<LatestCommentVO> result = new ArrayList<>();
        
        for (Map<String, Object> row : comments) {
            LatestCommentVO vo = new LatestCommentVO();
            vo.setId(getIntValue(row, "id"));
            vo.setUsername((String) row.get("username"));
            vo.setAvatar((String) row.get("avatar"));
            vo.setContent((String) row.get("content"));
            vo.setManuscriptTitle((String) row.get("manuscriptTitle"));
            
            Object timeObj = row.get("time");
            if (timeObj != null) {
                if (timeObj instanceof LocalDateTime) {
                    LocalDateTime commentTime = (LocalDateTime) timeObj;
                    vo.setTime(formatTimeAgo(commentTime));
                } else {
                    vo.setTime(timeObj.toString());
                }
            }
            
            result.add(vo);
        }
        
        return result;
    }
    
    @Override
    public List<FansRankingVO> getFansRanking(Integer userId, String type, Integer limit) {
        List<Map<String, Object>> ranking;
        
        if ("interaction".equals(type)) {
            ranking = creatorStatsMapper.selectInteractionRanking(userId, limit);
        } else {
            ranking = creatorStatsMapper.selectViewRanking(userId, limit);
        }
        
        List<FansRankingVO> result = new ArrayList<>();
        for (Map<String, Object> row : ranking) {
            FansRankingVO vo = new FansRankingVO();
            vo.setId(getIntValue(row, "id"));
            vo.setUsername((String) row.get("username"));
            vo.setAvatar((String) row.get("avatar"));
            
            if ("interaction".equals(type)) {
                vo.setInteractionCount(getIntValue(row, "interactionCount"));
            } else {
                vo.setInteractionCount(getIntValue(row, "commentCount"));
            }
            
            result.add(vo);
        }
        
        return result;
    }
    
    private String formatTimeAgo(LocalDateTime time) {
        if (time == null) return "";
        
        Duration duration = Duration.between(time, LocalDateTime.now());
        long minutes = duration.toMinutes();
        long hours = duration.toHours();
        long days = duration.toDays();
        
        if (minutes < 1) {
            return "刚刚";
        } else if (minutes < 60) {
            return minutes + "分钟前";
        } else if (hours < 24) {
            return hours + "小时前";
        } else if (days < 30) {
            return days + "天前";
        } else {
            return time.format(DateTimeFormatter.ofPattern("MM-dd HH:mm"));
        }
    }
    
    private Integer getIntValue(Map<String, Object> map, String key) {
        Object value = map.get(key);
        if (value == null) return 0;
        if (value instanceof Integer) return (Integer) value;
        if (value instanceof Long) return ((Long) value).intValue();
        if (value instanceof java.math.BigDecimal) return ((java.math.BigDecimal) value).intValue();
        return 0;
    }

    @Override
    public FansTrendVO getFansTrend(Integer userId, Integer days) {
        FansTrendVO vo = new FansTrendVO();
        List<String> dates = new ArrayList<>();
        List<Integer> newFollowers = new ArrayList<>();
        List<Integer> unfollows = new ArrayList<>();
        List<Integer> totalFollowers = new ArrayList<>();

        List<Map<String, Object>> trendData = creatorStatsMapper.selectFansTrendData(userId, days);

        Map<String, Map<String, Integer>> dataMap = new HashMap<>();
        for (Map<String, Object> row : trendData) {
            Object dateObj = row.get("date");
            String date = dateObj instanceof java.sql.Date ? ((java.sql.Date) dateObj).toString() : String.valueOf(dateObj);
            Map<String, Integer> dayData = new HashMap<>();
            dayData.put("newFollowers", getIntValue(row, "newFollowers"));
            dayData.put("unfollows", getIntValue(row, "unfollows"));
            dataMap.put(date, dayData);
        }

        Set<String> sortedDates = new TreeSet<>(dataMap.keySet());
        
        int runningTotal = 0;
        for (String date : sortedDates) {
            dates.add(date);
            Map<String, Integer> dayData = dataMap.get(date);
            if (dayData != null) {
                newFollowers.add(dayData.get("newFollowers"));
                unfollows.add(dayData.get("unfollows"));
                runningTotal += dayData.get("newFollowers") - dayData.get("unfollows");
            } else {
                newFollowers.add(0);
                unfollows.add(0);
            }
            totalFollowers.add(Math.max(0, runningTotal));
        }

        vo.setDates(dates);
        vo.setNewFollowers(newFollowers);
        vo.setUnfollows(unfollows);
        vo.setTotalFollowers(totalFollowers);

        Integer currentFollowers = creatorStatsMapper.selectFollowerCount(userId);
        vo.setCurrentFollowers(currentFollowers != null ? currentFollowers : 0);

        if (!newFollowers.isEmpty()) {
            vo.setNewFollowersToday(newFollowers.get(newFollowers.size() - 1));
            vo.setUnfollowsToday(unfollows.get(unfollows.size() - 1));
        }

        if (currentFollowers != null && currentFollowers > 0 && totalFollowers.size() >= 7) {
            int weekAgoFollowers = totalFollowers.get(totalFollowers.size() - 7);
            if (weekAgoFollowers > 0) {
                vo.setGrowthRate((double) (currentFollowers - weekAgoFollowers) / weekAgoFollowers * 100);
            }
        }

        return vo;
    }
}
