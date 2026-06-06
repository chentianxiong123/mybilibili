package com.mybilibili.analytics.service.impl;

import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.vo.*;
import com.mybilibili.analytics.feign.DanmakuClient;
import com.mybilibili.analytics.mapper.CreatorStatsMapper;
import com.mybilibili.analytics.service.CreatorStatsService;
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
import java.util.stream.Collectors;

@Service
public class CreatorStatsServiceImpl implements CreatorStatsService {

    @Autowired
    private CreatorStatsMapper creatorStatsMapper;

    @Autowired
    private DanmakuClient danmakuClient;

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
            overview.setTotalManuscripts(getIntValue(totalStats, "totalManuscripts"));
        }

        // Get real danmaku count from MongoDB via Feign
        List<Manuscript> userManuscripts = creatorStatsMapper.selectTopByViews(userId, 1000);
        if (!userManuscripts.isEmpty()) {
            List<Integer> manuscriptIds = userManuscripts.stream()
                    .map(Manuscript::getId).collect(Collectors.toList());
            try {
                Result<Map<Integer, Long>> result = danmakuClient.getDanmakuCountByManuscriptIds(manuscriptIds);
                if (result != null && result.getCode() == 200 && result.getData() != null) {
                    long totalDanmaku = result.getData().values().stream().mapToLong(Long::longValue).sum();
                    overview.setTotalDanmaku((int) totalDanmaku);
                }
            } catch (Exception e) {
                overview.setTotalDanmaku(0);
            }
        } else {
            overview.setTotalDanmaku(0);
        }

        Integer followerCount = creatorStatsMapper.selectFollowerCount(userId);
        overview.setTotalFollowers(followerCount != null ? followerCount : 0);

        // 7-day increases from real time data
        LocalDate sevenDaysAgo = LocalDate.now().minusDays(7);
        String startDate = sevenDaysAgo.toString();

        Integer viewsIncrease = creatorStatsMapper.selectViewIncreaseFromDailyMetrics(userId, startDate);
        overview.setViewsIncrease(viewsIncrease != null ? viewsIncrease : 0);

        Map<String, Object> interactionIncrease = creatorStatsMapper.selectInteractionIncrease(userId, startDate);
        if (interactionIncrease != null) {
            overview.setLikesIncrease(getIntValue(interactionIncrease, "likesIncrease"));
            overview.setCoinsIncrease(getIntValue(interactionIncrease, "coinsIncrease"));
            overview.setCollectionsIncrease(getIntValue(interactionIncrease, "collectionsIncrease"));
            overview.setSharesIncrease(getIntValue(interactionIncrease, "sharesIncrease"));
        }

        Map<String, Object> commentIncrease = creatorStatsMapper.selectCommentIncrease(userId, startDate);
        if (commentIncrease != null) {
            overview.setCommentsIncrease(getIntValue(commentIncrease, "commentsIncrease"));
        }

        // Danmaku 7-day increase via Feign
        if (!userManuscripts.isEmpty()) {
            List<Integer> manuscriptIds = userManuscripts.stream()
                    .map(Manuscript::getId).collect(Collectors.toList());
            try {
                Result<Map<String, Integer>> trendResult = danmakuClient.getDanmakuTrend(
                        manuscriptIds, startDate, LocalDate.now().toString());
                if (trendResult != null && trendResult.getCode() == 200 && trendResult.getData() != null) {
                    int danmakuIncrease = trendResult.getData().values().stream().mapToInt(Integer::intValue).sum();
                    overview.setDanmakuIncrease(danmakuIncrease);
                }
            } catch (Exception e) {
                overview.setDanmakuIncrease(0);
            }
        } else {
            overview.setDanmakuIncrease(0);
        }

        List<Map<String, Object>> fansTrend = creatorStatsMapper.selectFansTrendData(userId, sevenDaysAgo.toString());
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

        String startDate = LocalDate.now().minusDays(days - 1).toString();

        // View trends from daily metrics, not from manuscript upload-time snapshots.
        List<Map<String, Object>> viewTrend = creatorStatsMapper.selectViewTrendData(userId, startDate);
        // Interaction trends (likes, coins, collects, shares) from user_interactions
        List<Map<String, Object>> interactionTrend = creatorStatsMapper.selectInteractionTrendData(userId, startDate);
        // Comment trend from comments table
        List<Map<String, Object>> commentTrend = creatorStatsMapper.selectCommentTrendData(userId, startDate);
        // Fans trend (existing, unchanged)
        List<Map<String, Object>> fansTrend = creatorStatsMapper.selectFansTrendData(userId, startDate);

        // Danmaku trend from MongoDB
        List<Manuscript> userManuscripts = creatorStatsMapper.selectTopByViews(userId, 1000);
        Map<String, Integer> danmakuTrendMap = new HashMap<>();
        if (!userManuscripts.isEmpty()) {
            List<Integer> manuscriptIds = userManuscripts.stream()
                    .map(Manuscript::getId).collect(Collectors.toList());
            try {
                Result<Map<String, Integer>> result = danmakuClient.getDanmakuTrend(
                        manuscriptIds, startDate, LocalDate.now().toString());
                if (result != null && result.getCode() == 200 && result.getData() != null) {
                    danmakuTrendMap = result.getData();
                }
            } catch (Exception e) {
                // danmaku trend unavailable
            }
        }

        // Index by date
        Map<String, Integer> viewMap = new HashMap<>();
        for (Map<String, Object> row : viewTrend) {
            viewMap.put(row.get("date").toString(), getIntValue(row, "views"));
        }

        Map<String, Map<String, Integer>> interactionMap = new HashMap<>();
        for (Map<String, Object> row : interactionTrend) {
            String date = row.get("date").toString();
            Map<String, Integer> dayData = new HashMap<>();
            dayData.put("likes", getIntValue(row, "likes"));
            dayData.put("coins", getIntValue(row, "coins"));
            dayData.put("collects", getIntValue(row, "collects"));
            dayData.put("shares", getIntValue(row, "shares"));
            interactionMap.put(date, dayData);
        }

        Map<String, Integer> commentMap = new HashMap<>();
        for (Map<String, Object> row : commentTrend) {
            commentMap.put(row.get("date").toString(), getIntValue(row, "comments"));
        }

        Map<String, Integer> fansMap = new HashMap<>();
        for (Map<String, Object> row : fansTrend) {
            fansMap.put(row.get("date").toString(), getIntValue(row, "newFollowers"));
        }

        // Build full date range
        List<String> dates = new ArrayList<>();
        List<Integer> views = new ArrayList<>();
        List<Integer> likes = new ArrayList<>();
        List<Integer> comments = new ArrayList<>();
        List<Integer> followers = new ArrayList<>();
        List<Integer> danmaku = new ArrayList<>();
        List<Integer> coins = new ArrayList<>();
        List<Integer> collects = new ArrayList<>();
        List<Integer> shares = new ArrayList<>();

        LocalDate end = LocalDate.now();
        LocalDate start = LocalDate.now().minusDays(days - 1);
        for (LocalDate d = start; !d.isAfter(end); d = d.plusDays(1)) {
            String dateStr = d.toString();
            dates.add(dateStr);
            views.add(viewMap.getOrDefault(dateStr, 0));

            Map<String, Integer> dayInteraction = interactionMap.get(dateStr);
            if (dayInteraction != null) {
                likes.add(dayInteraction.get("likes"));
                coins.add(dayInteraction.get("coins"));
                collects.add(dayInteraction.get("collects"));
                shares.add(dayInteraction.get("shares"));
            } else {
                likes.add(0);
                coins.add(0);
                collects.add(0);
                shares.add(0);
            }

            comments.add(commentMap.getOrDefault(dateStr, 0));
            danmaku.add(danmakuTrendMap.getOrDefault(dateStr, 0));
            followers.add(fansMap.getOrDefault(dateStr, 0));
        }

        trend.setDates(dates);
        trend.setViews(views);
        trend.setLikes(likes);
        trend.setComments(comments);
        trend.setDanmaku(danmaku);
        trend.setFollowers(followers);
        trend.setCoins(coins);
        trend.setCollects(collects);
        trend.setShares(shares);

        return trend;
    }

    @Override
    public ManuscriptTrendVO getManuscriptTrend(Integer userId) {
        ManuscriptTrendVO vo = new ManuscriptTrendVO();

        List<Map<String, Object>> manuscripts = creatorStatsMapper.selectManuscriptListForTrend(userId);

        if (manuscripts.isEmpty()) {
            vo.setDates(new ArrayList<>());
            vo.setViews(new ArrayList<>());
            vo.setDanmaku(new ArrayList<>());
            vo.setTitles(new ArrayList<>());
            return vo;
        }

        List<String> dates = new ArrayList<>();
        List<Integer> views = new ArrayList<>();
        List<Integer> danmaku = new ArrayList<>();
        List<String> titles = new ArrayList<>();
        List<Integer> manuscriptIds = new ArrayList<>();

        // Sort by upload_time ASC and calculate cumulative views
        int cumulativeViews = 0;
        for (Map<String, Object> row : manuscripts) {
            Object timeObj = row.get("upload_time");
            if (timeObj != null) {
                if (timeObj instanceof java.sql.Date) {
                    dates.add(((java.sql.Date) timeObj).toLocalDate().toString());
                } else if (timeObj instanceof java.sql.Timestamp) {
                    dates.add(((java.sql.Timestamp) timeObj).toLocalDateTime().toLocalDate().toString());
                } else {
                    dates.add(timeObj.toString().substring(0, 10));
                }
            } else {
                dates.add("");
            }
            cumulativeViews += getIntValue(row, "views");
            views.add(cumulativeViews);
            titles.add((String) row.get("title"));
            manuscriptIds.add(getIntValue(row, "id"));
        }

        // Get danmaku counts from MongoDB
        if (!manuscriptIds.isEmpty()) {
            try {
                Result<Map<Integer, Long>> result = danmakuClient.getDanmakuCountByManuscriptIds(manuscriptIds);
                if (result != null && result.getCode() == 200 && result.getData() != null) {
                    for (Integer mid : manuscriptIds) {
                        Long count = result.getData().get(mid);
                        danmaku.add(count != null ? count.intValue() : 0);
                    }
                } else {
                    manuscriptIds.forEach(mid -> danmaku.add(0));
                }
            } catch (Exception e) {
                manuscriptIds.forEach(mid -> danmaku.add(0));
            }
        }

        vo.setDates(dates);
        vo.setViews(views);
        vo.setDanmaku(danmaku);
        vo.setTitles(titles);

        return vo;
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
            vo.setManuscriptId(getIntValue(row, "manuscriptId"));
            vo.setManuscriptTitle((String) row.get("manuscriptTitle"));

            Object timeObj = row.get("time");
            if (timeObj != null) {
                if (timeObj instanceof LocalDateTime) {
                    LocalDateTime commentTime = (LocalDateTime) timeObj;
                    vo.setTime(formatTimeAgo(commentTime));
                    vo.setCreateTime(commentTime.format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss")));
                } else {
                    vo.setTime(timeObj.toString());
                    vo.setCreateTime(timeObj.toString());
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

        String startDate = LocalDate.now().minusDays(days - 1).toString();
        List<Map<String, Object>> trendData = creatorStatsMapper.selectFansTrendData(userId, startDate);

        Map<String, Map<String, Integer>> dataMap = new HashMap<>();
        for (Map<String, Object> row : trendData) {
            Object dateObj = row.get("date");
            String date = dateObj instanceof java.sql.Date ? ((java.sql.Date) dateObj).toString() : String.valueOf(dateObj);
            Map<String, Integer> dayData = new HashMap<>();
            dayData.put("newFollowers", getIntValue(row, "newFollowers"));
            dayData.put("unfollows", getIntValue(row, "unfollows"));
            dataMap.put(date, dayData);
        }

        Integer currentFollowers = creatorStatsMapper.selectFollowerCount(userId);
        int runningFollowers = currentFollowers != null ? currentFollowers : 0;

        int totalNetInPeriod = 0;
        for (Map<String, Integer> dayData : dataMap.values()) {
            totalNetInPeriod += dayData.get("newFollowers") - dayData.get("unfollows");
        }
        int startFollowers = runningFollowers - totalNetInPeriod;

        LocalDate end = LocalDate.now();
        LocalDate start = LocalDate.now().minusDays(days - 1);
        for (LocalDate d = start; !d.isAfter(end); d = d.plusDays(1)) {
            String dateStr = d.toString();
            dates.add(dateStr);

            Map<String, Integer> dayData = dataMap.get(dateStr);
            if (dayData != null) {
                newFollowers.add(dayData.get("newFollowers"));
                unfollows.add(dayData.get("unfollows"));
                startFollowers += dayData.get("newFollowers") - dayData.get("unfollows");
            } else {
                newFollowers.add(0);
                unfollows.add(0);
            }
            totalFollowers.add(Math.max(0, startFollowers));
        }

        vo.setDates(dates);
        vo.setNewFollowers(newFollowers);
        vo.setUnfollows(unfollows);
        vo.setTotalFollowers(totalFollowers);
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
